package main

import (
	"anytls/proxy"
	"anytls/proxy/padding"
	"anytls/proxy/session"
	"context"
	"crypto/tls"
	"encoding/binary"
	"net"

	"github.com/sagernet/sing/common/buf"
	M "github.com/sagernet/sing/common/metadata"
)

type myClient struct {
	serverAddr    string
	tlsConfig     *tls.Config
	sessionClient *session.Client
}

func NewMyClient(ctx context.Context, serverAddr string, tlsConfig *tls.Config) *myClient {
	s := &myClient{
		serverAddr: serverAddr,
		tlsConfig:  tlsConfig,
	}
	s.sessionClient = session.NewClient(ctx)
	return s
}

func (c *myClient) CreateProxy(ctx context.Context, destination M.Socksaddr) (net.Conn, error) {
	conn, err := c.sessionClient.CreateStream(ctx, c.CreateOutboundTLSConnection)
	if err != nil {
		return nil, err
	}
	err = M.SocksaddrSerializer.WriteAddrPort(conn, destination)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func (c *myClient) CreateOutboundTLSConnection(ctx context.Context) (net.Conn, error) {
	conn, err := proxy.SystemDialer.DialContext(ctx, "tcp", c.serverAddr)
	if err != nil {
		return nil, err
	}

	b := buf.NewPacket()
	b.Write(passwordSha256)
	var paddingLen int
	if pad := padding.DefaultPaddingFactory.Load().GenerateRecordPayloadSizes(0); len(pad) > 0 {
		paddingLen = pad[0]
	}
	binary.BigEndian.PutUint16(b.Extend(2), uint16(paddingLen))
	if paddingLen > 0 {
		b.WriteZeroN(paddingLen)
	}

	conn = tls.Client(conn, c.tlsConfig)

	_, err = b.WriteTo(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}
