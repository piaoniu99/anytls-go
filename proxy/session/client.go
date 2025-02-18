package session

import (
	"anytls/proxy"
	"anytls/util"
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chen3feng/stl4go"
	"github.com/sagernet/sing/common"
)

type Client struct {
	die       context.Context
	dieCancel context.CancelFunc

	dialOut proxy.DialOutFunc

	sessionCounter  atomic.Uint64
	idleSession     *stl4go.SkipList[uint64, *Session]
	idleSessionLock sync.Mutex
}

func NewClient(ctx context.Context, dialOut proxy.DialOutFunc) *Client {
	c := &Client{
		dialOut: dialOut,
	}
	c.die, c.dieCancel = context.WithCancel(ctx)
	c.idleSession = stl4go.NewSkipList[uint64, *Session]()
	util.StartRoutine(c.die, time.Second*30, c.idleCleanup)
	return c
}

func (c *Client) CreateStream(ctx context.Context) (net.Conn, error) {
	select {
	case <-c.die.Done():
		return nil, io.ErrClosedPipe
	default:
	}

	var session *Session
	var stream *Stream
	var err error

	for i := 0; i < 3; i++ {
		session, err = c.findSession(ctx)
		if session == nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
		stream, err = session.OpenStream()
		if err != nil {
			common.Close(session, stream)
			continue
		}
		break
	}
	if session == nil || stream == nil {
		return nil, fmt.Errorf("too many closed session: %w", err)
	}

	stream.dieHook = func() {
		if session.IsClosed() {
			if session.dieHook != nil {
				session.dieHook()
			}
		} else {
			c.idleSessionLock.Lock()
			session.idleSince = time.Now()
			c.idleSession.Insert(math.MaxUint64-session.seq, session)
			c.idleSessionLock.Unlock()
		}
	}

	return stream, nil
}

func (c *Client) findSession(ctx context.Context) (*Session, error) {
	var idle *Session

	c.idleSessionLock.Lock()
	if !c.idleSession.IsEmpty() {
		it := c.idleSession.Iterate()
		idle = it.Value()
		c.idleSession.Remove(it.Key())
	}
	c.idleSessionLock.Unlock()

	if idle == nil {
		s, err := c.createSession(ctx)
		return s, err
	}
	return idle, nil
}

func (c *Client) createSession(ctx context.Context) (*Session, error) {
	underlying, err := c.dialOut(ctx)
	if err != nil {
		return nil, err
	}

	session := NewClientSession(underlying)
	session.seq = c.sessionCounter.Add(1)
	session.dieHook = func() {
		//logrus.Debugln("session died", session)
		c.idleSessionLock.Lock()
		c.idleSession.Remove(math.MaxUint64 - session.seq)
		c.idleSessionLock.Unlock()
	}
	session.Run()
	return session, nil
}

func (c *Client) Close() error {
	c.dieCancel()
	go c.idleCleanupExpTime(time.Now())
	return nil
}

func (c *Client) idleCleanup() {
	c.idleCleanupExpTime(time.Now().Add(-time.Second * 30))
}

func (c *Client) idleCleanupExpTime(expTime time.Time) {
	var sessionToRemove = make([]*Session, 0)

	c.idleSessionLock.Lock()
	it := c.idleSession.Iterate()
	for it.IsNotEnd() {
		session := it.Value()
		if session.idleSince.Before(expTime) {
			sessionToRemove = append(sessionToRemove, session)
			c.idleSession.Remove(it.Key())
		}
		it.MoveToNext()
	}
	c.idleSessionLock.Unlock()

	for _, session := range sessionToRemove {
		session.Close()
	}
}
