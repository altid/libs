package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/altid/server/signal"
)

var factotum = flag.Bool("f", false, "Enable client authentication via a plan9 factotum")
var dir = flag.String("m", "/tmp/altid", "Path to Altid services")
var port = flag.Int("p", 564, "Port to listen on")
var usetls = flag.Bool("t", false, "Use TLS")
var debug = flag.Bool("d", false, "Debug")
var chatty = flag.Bool("D", false, "Chatty")

func main() {
	flag.Parse()

	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(0)
	}

	sigs := signal.Setup()
	ctx, cancel := context.WithCancel(context.Background())

	srv, err := server.New(ctx, *debug, *chatty)
	if err != nil {
		log.Fatal(err)
	}

	srv.Start(*dir, *port, *usetls, *factotum)

	for {
		select {
		case e := srv.Err():
			log.Fatal(e)
		case sig := <-sigs:
			signal.Handle(cancel, sig.String(), *debug)
		case <-ctx.Done():
			break
		}
	}
}
