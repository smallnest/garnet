package garnet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/smallnest/garnet/codec"
	"github.com/smallnest/garnet/handler"
	"github.com/smallnest/rpcx/log"
)

type Client struct {
	network    string
	addr       string
	connOption func(net.Conn) net.Conn
	conn       net.Conn
	frameConn  codec.FrameCodec
	codecs     []codec.Codec
	frameFunc  codec.FrameFunc
	handler    handler.Handler

	DialTimeout time.Duration

	stopped int32
}

func (c *Client) SetConnOption(fn func(net.Conn) net.Conn) {
	c.connOption = fn
}

func (c *Client) SetFrameFunc(frameFunc codec.FrameFunc) {
	c.frameFunc = frameFunc
}

func (c *Client) AddCodec(codec codec.Codec) {
	c.codecs = append(c.codecs, codec)
}

func (c *Client) SetHandler(handler handler.Handler) {
	c.handler = handler
}

func (c *Client) Dial(network, addr string) error {
	if c.handler == nil {
		return errors.New("handler has not set")
	}

	conn, err := net.DialTimeout(network, addr, c.DialTimeout)
	if err != nil {
		return err
	}

	if c.connOption != nil {
		conn = c.connOption(conn)
	}

	if c.frameFunc == nil {
		return fmt.Errorf("frameCodec must not be nil")
	}

	c.conn = conn
	c.frameConn = c.frameFunc(conn)

	go c.handleConn(c.frameConn)

	return nil
}

func (c *Client) handleConn(frameConn codec.FrameCodec) {
	conn := frameConn.Conn()
	c.handler.Connected(conn)
	defer func() {
		conn.Close()
		c.handler.Disconnected(conn)
	}()

	for {
		v, err := frameConn.ReadFrame()
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			log.Errorf("failed to read a frame because of %v", err)
			c.handler.ErrorCaught(err)
			return
		}

		for _, cc := range c.codecs {
			v, err = cc.Decode(v)
			if err != nil {
				log.Errorf("failed to decode from %s because of %v", conn.RemoteAddr(), err)
				c.handler.ErrorCaught(err)
				return
			}
		}

		err = c.handler.Handle(frameConn, v)
		if err != nil {
			log.Errorf("failed to handle message from %s because of %v", conn.RemoteAddr(), err)
			c.handler.ErrorCaught(err)
			return
		}
	}
}

func (c *Client) Write(data interface{}) error {
	if c.frameFunc == nil && c.conn == nil {
		return errors.New("connection not found")
	}

	var err error
	l := len(c.codecs)
	var v interface{} = data
	for i := l - 1; i >= 0; i-- {
		cc := c.codecs[i]
		v, err = cc.Encode(v)
		if err != nil {
			return err
		}
	}

	var result []byte
	var ok bool
	if result, ok = v.([]byte); !ok {
		return fmt.Errorf("the final type of encoded data must be []byte but got %T", v)
	}

	if c.frameConn != nil {
		err = c.frameConn.WriteFrame(result)
		return err
	}

	_, err = c.conn.Write(result)
	return err
}

func (c *Client) Close() {
	err := c.conn.Close()
	if err != nil {
		log.Error("failed to close %s", c.conn.RemoteAddr())
	}
	c.handler.Disconnected(c.conn)

	atomic.StoreInt32(&c.stopped, 1)
}
