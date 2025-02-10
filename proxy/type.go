package proxy

import (
	"context"
	"net"
)

type DialOutFunc func(ctx context.Context) (net.Conn, error)
