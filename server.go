package garnet

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/orcaman/concurrent-map"
	"github.com/smallnest/log"
)

type Server struct {
	network    string
	addr       string
	connOption func(net.Conn) net.Conn
	codecs     []Codec
	handler    Handler
	clients    cmap.ConcurrentMap

	stopped int32
}

func (s *Server) SetConnOption(fn func(net.Conn) net.Conn) {
	s.connOption = fn
}

func (s *Server) AddCodec(codec Codec) {
	s.codecs = append(s.codecs, codec)
}

func (s *Server) SetHandler(handler Handler) {
	s.handler = handler
}

func (s *Server) Serve(network, addr string) error {
	if s.handler == nil {
		return errors.New("handler has not set")
	}

	s.addr = addr
	ln, err := makeListener(network, addr)
	if err != nil {
		return err
	}

	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Warnf("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			if atomic.LoadInt32(&s.stopped) == 1 { // it's stopping, then discard any error.
				return nil
			}
			log.Errorf("motan server accept error: %v", err)
			return err
		}
		tempDelay = 0

		if atomic.LoadInt32(&s.stopped) != 0 {
			conn.Close()
			return nil
		}

		if tc, ok := conn.(*net.TCPConn); ok {
			tc.SetKeepAlive(true)
			tc.SetKeepAlivePeriod(3 * time.Minute)
		}

		s.clients.Set(conn.RemoteAddr().String(), conn)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	s.handler.Connected(conn)
	defer func() {
		conn.Close()
		s.handler.Disconnected(conn)
	}()

	if s.connOption != nil {
		conn = s.connOption(conn)
	}

	var err error

	for {
		var v interface{} = conn
		for _, c := range s.codecs {
			v, err = c.Decode(v)
			if err != nil {
				log.Errorf("failed to decode from %s because of %v", conn.RemoteAddr(), err)
				return
			}
		}

		err = s.handler.Handle(conn, v)
		if err != nil {
			log.Errorf("failed to handle message from %s because of %v", conn.RemoteAddr(), err)
			return
		}
	}
}

func (s *Server) Close() {
	s.clients.IterCb(func(key string, v interface{}) {
		err := v.(net.Conn).Close()
		if err != nil {
			log.Error("failed to close %s", v.(net.Conn).RemoteAddr())
		}
	})
}

func makeListener(network, addr string) (net.Listener, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return net.Listen(network, addr)
	case "udp", "udp4", "udp6":
		return net.Listen(network, addr)
	case "unix":
		return net.Listen(network, addr)
	case "unixpacket":
		return net.Listen(network, addr)
	}

	return nil, fmt.Errorf("unsupported network: %s", network)
}
