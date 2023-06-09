package main

import (
	"flag"
	"fmt"
	"log"

	"golang.org/x/net/context"

	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/raw"
)

var (
	flagHost = flag.String("host", "localhost", "The hostname to serve on")
	flagPort = flag.Int("port", 8080, "The port to listen on")
)

type rawHandler struct{}

func (rawHandler) Handle(ctx context.Context, args *raw.Args) (*raw.Res, error) {
	return &raw.Res{
		Arg2: args.Arg2,
		Arg3: args.Arg3,
	}, nil
}

func (rawHandler) OnError(ctx context.Context, err error) {
	log.Fatalf("OnError: %v", err)
}

func main() {
	flag.Parse()

	ch, err := tchannel.NewChannel("test_as_raw", nil)
	if err != nil {
		log.Fatalf("NewChannel failed: %v", err)
	}

	handler := raw.Wrap(rawHandler{})
	ch.Register(handler, "echo")
	ch.Register(handler, "streaming_echo")

	hostPort := fmt.Sprintf("%s:%v", *flagHost, *flagPort)
	if err := ch.ListenAndServe(hostPort); err != nil {
		log.Fatalf("ListenAndServe failed: %v", err)
	}

	fmt.Println("listening on", ch.PeerInfo().HostPort)
	select {}
}
