package garnet

import (
	"net"

	"github.com/smallnest/garnet/codec"
)

type Handler interface {
	Connected(net.Conn)
	Disconnected(net.Conn)
	Handle(frameCodec codec.FrameCodec, v interface{}) error
	ErrorCaught(err error)
}
