package garnet

import (
	"errors"
	"net"
	"time"

	"github.com/smallnest/log"
)

type Client struct {
	network    string
	addr       string
	connOption func(net.Conn) net.Conn
	conn       net.Conn
	codecs     []Codec
	handler    Handler

	DialTimeout time.Duration

	stopped int32
}

func (c *Client) SetConnOption(fn func(net.Conn) net.Conn) {
	c.connOption = fn
}

func (c *Client) AddCodec(codec Codec) {
	c.codecs = append(c.codecs, codec)
}

func (c *Client) SetHandler(handler Handler) {
	c.handler = handler
}

func (c *Client) Dial(network, addr string) (net.Conn, error) {
	if c.handler == nil {
		return nil, errors.New("handler has not set")
	}

	conn, err := net.DialTimeout(network, addr, c.DialTimeout)
	if err != nil {
		return nil, err
	}

	if c.connOption != nil {
		conn = c.connOption(conn)
	}

	go c.handleConn(conn)

	return conn, nil
}

func (c *Client) handleConn(conn net.Conn) {
	c.handler.Connected(conn)
	defer func() {
		conn.Close()
		c.handler.Disconnected(conn)
	}()

	var err error

	for {
		var v interface{} = conn
		for _, c := range c.codecs {
			v, err = c.Decode(v)
			if err != nil {
				log.Errorf("failed to decode from %s because of %v", conn.RemoteAddr(), err)
				return
			}
		}

		err = c.handler.Handle(conn, v)
		if err != nil {
			log.Errorf("failed to handle message from %s because of %v", conn.RemoteAddr(), err)
			return
		}
	}
}

func (c *Client) Close() {
	err := c.conn.Close()
	if err != nil {
		log.Error("failed to close %s", c.conn.RemoteAddr())
	}
}
