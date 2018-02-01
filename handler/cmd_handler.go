package handler

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/smallnest/garnet/codec"
)

type CmdHandler struct {
	ConnectedFn    func(net.Conn)
	DisconnectedFn func(net.Conn)
	ErrorCaughtFn  func(err error)
	handlers       map[string]func(frameCodec codec.FrameCodec, v interface{}) error
}

func NewCmdHandler() *CmdHandler {
	return &CmdHandler{
		handlers: make(map[string]func(frameCodec codec.FrameCodec, v interface{}) error),
	}
}

func (h *CmdHandler) Register(cmd string, handler func(frameCodec codec.FrameCodec, v interface{}) error) {
	if h.handlers == nil {
		h.handlers = make(map[string]func(frameCodec codec.FrameCodec, v interface{}) error)
	}

	h.handlers[cmd] = handler
}

func (h *CmdHandler) Connected(conn net.Conn) {
	if h.ConnectedFn != nil {
		h.ConnectedFn(conn)
	}
}
func (h *CmdHandler) Disconnected(conn net.Conn) {
	if h.DisconnectedFn != nil {
		h.DisconnectedFn(conn)
	}
}
func (h *CmdHandler) Handle(frameCodec codec.FrameCodec, v interface{}) error {
	if h.handlers == nil {
		return errors.New("handlers have not been configured")
	}
	data, ok := v.([]byte)
	if !ok {
		return errors.New("CmdHandler only handle []byte data")
	}

	log.Printf("@@@@:%s", data)

	// parse cmd (LV format)
	if len(data) < 2 {
		return errors.New("data is too short for parsing length")
	}

	l := binary.BigEndian.Uint16(data[:2])
	if len(data) < int(l)+2 {
		return errors.New("data is too short for parsing cmd")
	}
	cmd := string(data[2 : l+2])
	hand := h.handlers[cmd]
	if hand == nil {
		return fmt.Errorf("handler has not been configured for %s", cmd)
	}

	return hand(frameCodec, data[l+2:])
}
func (h *CmdHandler) ErrorCaught(err error) {
	if h.ErrorCaughtFn != nil {
		h.ErrorCaughtFn(err)
	}
}

func WrapCmdData(cmd string, data []byte) []byte {
	cmdBytes := []byte(cmd)
	l := len(cmdBytes)
	v := make([]byte, l+2+len(data))

	binary.BigEndian.PutUint16(v[:2], uint16(l))
	copy(v[2:l+2], cmdBytes)
	copy(v[l+2:], data)

	return v
}
