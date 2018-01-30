package codec

import (
	"net"

	"github.com/smallnest/goframe"
)

func NewDelimiterBasedFrameFunc(delimiter byte) func(net.Conn) FrameCodec {
	return func(conn net.Conn) FrameCodec {
		return &WrappedFrameConn{goframe.NewDelimiterBasedFrameConn(delimiter, conn)}
	}
}

func NewFixedLengthFrameFunc(frameLength int) func(net.Conn) FrameCodec {
	return func(conn net.Conn) FrameCodec {
		return &WrappedFrameConn{goframe.NewFixedLengthFrameConn(frameLength, conn)}
	}
}

func NewLengthFieldBasedFrameFunc(encoderConfig goframe.EncoderConfig, decoderConfig goframe.DecoderConfig) func(net.Conn) FrameCodec {
	return func(conn net.Conn) FrameCodec {
		return &WrappedFrameConn{goframe.NewLengthFieldBasedFrameConn(encoderConfig, decoderConfig, conn)}
	}
}

func NewLineBasedFrameFunc() func(net.Conn) FrameCodec {
	return func(conn net.Conn) FrameCodec {
		return &WrappedFrameConn{goframe.NewLineBasedFrameConn(conn)}
	}
}

type WrappedFrameConn struct {
	FrameConn goframe.FrameConn
}

func (c *WrappedFrameConn) ReadFrame() (interface{}, error) {
	return c.FrameConn.ReadFrame()
}

func (c *WrappedFrameConn) WriteFrame(v interface{}) error {
	return c.FrameConn.WriteFrame(v.([]byte))
}

func (c *WrappedFrameConn) Conn() net.Conn {
	return c.FrameConn.Conn()
}

func (c *WrappedFrameConn) Close() error {
	return c.FrameConn.Close()
}
