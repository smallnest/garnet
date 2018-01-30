package codec

import (
	"errors"
	"net"
)

type FrameFunc func(net.Conn) FrameCodec

type FrameCodec interface {
	ReadFrame() (interface{}, error)
	WriteFrame(p interface{}) error
	Conn() net.Conn
	Close() error
}

type Codec interface {
	Encode(v interface{}) (interface{}, error)
	Decode(v interface{}) (interface{}, error)
}

type StringCodec struct {
}

func (StringCodec) Encode(v interface{}) (interface{}, error) {
	if s, ok := v.(string); ok {
		return []byte(s), nil
	}

	return nil, errors.New("input is not a string")
}

func (StringCodec) Decode(v interface{}) (interface{}, error) {
	if s, ok := v.([]byte); ok {
		return string(s), nil
	}

	return nil, errors.New("input is not a []byte]")
}
