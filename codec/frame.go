package codec

import (
	"github.com/smallnest/goframe"
)

type FrameCodec struct {
	Conn goframe.FrameConn
}

func (c *FrameCodec) Encode(v interface{}) (interface{}, error) {

}

func (c *FrameCodec) Decode(v interface{}) (interface{}, error) {

}
