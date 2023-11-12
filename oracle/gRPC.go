package oracle

import (
	"context"
	"net"
	"strings"
)

func dialerFunc(ctx context.Context, addr string) (net.Conn, error) {
	return Connect(addr)
}

func Connect(protoAddr string) (net.Conn, error) {
	proto, address := ProtocolAndAddress(protoAddr)
	conn, err := net.Dial(proto, address)
	return conn, err
}

func ProtocolAndAddress(listenAddr string) (string, string) {
	protocol, address := "tcp", listenAddr

	parts := strings.SplitN(address, "://", 2)
	if len(parts) == 2 {
		protocol, address = parts[0], parts[1]
	}

	return protocol, address
}
