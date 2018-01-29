package garnet

import "net"

type Handler interface {
	Connected(net.Conn)
	Disconnected(net.Conn)
	Handle(conn net.Conn, v interface{}) error
	ErrorCaught(err error)
}
