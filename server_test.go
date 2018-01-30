package garnet

import (
	"net"
	"testing"
	"time"

	"github.com/orcaman/concurrent-map"
	"github.com/smallnest/garnet/codec"
)

func TestCommnunication(t *testing.T) {
	serverHandler := &testHandler{}
	server := &Server{
		frameFunc: codec.NewLineBasedFrameFunc(),
		handler:   serverHandler,
		clients:   cmap.New(),
	}
	serverHandler.server = server
	server.AddCodec(&codec.StringCodec{})
	go server.Serve("tcp", ":8972")

	clientHandler := &testHandler{}
	client := &Client{
		frameFunc: codec.NewLineBasedFrameFunc(),
		handler:   clientHandler,
	}
	client.AddCodec(&codec.StringCodec{})

	err := client.Dial("tcp", "127.0.0.1:8972")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1; i++ {
		err := client.Write("hello")
		if err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Second)
	server.Close()
	client.Close()

	t.Log(serverHandler.connected)
	t.Log(serverHandler.disconnected)
	t.Log(serverHandler.errcaught)
	t.Log(serverHandler.messages[0])

	t.Log(clientHandler.connected)
	t.Log(clientHandler.disconnected)
	t.Log(clientHandler.errcaught)
	t.Log(clientHandler.messages[0])
}

type testHandler struct {
	server       *Server
	messages     []string
	connected    int32
	disconnected int32
	errcaught    int32
}

func (h *testHandler) Connected(conn net.Conn) {
	h.connected++
}
func (h *testHandler) Disconnected(conn net.Conn) {
	h.disconnected++
}
func (h *testHandler) Handle(frameCodec codec.FrameCodec, v interface{}) error {
	h.messages = append(h.messages, v.(string))
	if h.server != nil {
		h.server.Write(frameCodec, v)
	}
	return nil
}
func (h *testHandler) ErrorCaught(err error) {
	h.errcaught++
}
