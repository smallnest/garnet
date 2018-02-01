package main

import (
	"flag"
	"log"

	"github.com/smallnest/garnet"
	"github.com/smallnest/garnet/codec"
	"github.com/smallnest/garnet/handler"
)

var (
	addr = flag.String("addr", ":8972", "listened address")
)

var (
	cmdHandler = handler.NewCmdHandler()
	server     *garnet.Server
)

func main() {
	flag.Parse()

	cmdHandler.Register("say", func(frameCodec codec.FrameCodec, v interface{}) error {
		log.Printf("received: %s", v)
		server.Write(frameCodec, handler.WrapCmdData("reply", v.([]byte)))

		return nil
	})

	server = garnet.NewServer()
	server.SetFrameFunc(codec.NewLineBasedFrameFunc())
	server.SetHandler(cmdHandler)

	server.Serve("tcp", *addr)
}
