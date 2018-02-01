package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/smallnest/garnet"
	"github.com/smallnest/garnet/codec"
	"github.com/smallnest/garnet/handler"
)

var (
	host = flag.String("host", "127.0.0.1:8972", "server address")
)

var (
	cmdHandler = handler.NewCmdHandler()
)

func main() {
	flag.Parse()

	cmdHandler.Register("reply", func(frameCodec codec.FrameCodec, v interface{}) error {
		log.Printf("received reply: %s\n", v)
		return nil
	})

	client := &garnet.Client{}

	client.SetFrameFunc(codec.NewLineBasedFrameFunc())
	client.SetHandler(cmdHandler)

	err := client.Dial("tcp", *host)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		err := client.Write(handler.WrapCmdData("say", []byte("hello "+strconv.Itoa(i))))
		if err != nil {
			log.Fatal(err)
		}
	}

	select {}
}
