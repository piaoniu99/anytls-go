package main

import (
	"anytls/proxy"
	"context"
	"net"

	"github.com/sagernet/sing/common/bufio"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/common/uot"
	"github.com/sirupsen/logrus"
)

func proxyOutboundTCP(ctx context.Context, conn net.Conn, destination M.Socksaddr) {
	c, err := proxy.SystemDialer.DialContext(ctx, "tcp", destination.String())
	if err != nil {
		logrus.Debugln("proxyOutboundTCP DialContext:", err)
		return
	}
	bufio.CopyConn(ctx, conn, c)
}

func proxyOutboundUoT(ctx context.Context, conn net.Conn, destination M.Socksaddr) {
	request, err := uot.ReadRequest(conn)
	if err != nil {
		logrus.Debugln("proxyOutboundUoT ReadRequest:", err)
		return
	}

	c, err := net.ListenPacket("udp", "")
	if err != nil {
		logrus.Debugln("proxyOutboundUoT ListenPacket:", err)
		return
	}

	bufio.CopyPacketConn(ctx, uot.NewConn(conn, *request), bufio.NewPacketConn(c))
}
