package main

import (
	"context"
	"flag"
	"log"
	"os"

	_ "net/http/pprof"

	"github.com/altid/server"
)

var factotum = flag.Bool("f", false, "Enable client authentication via a plan9 factotum")
var dir = flag.String("m", "/tmp/altid", "Path to Altid services")
var port = flag.String("p", "564", "Port to listen on")
var usetls = flag.Bool("t", false, "Use TLS")
var debug = flag.Bool("d", false, "Debug")
var chatty = flag.Bool("D", false, "Chatty")
var addr = flag.String("l", "", "Address to listen on")
var cert string
var key string

func main() {
	flag.Parse()

	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(0)
	}

	ctx := context.Background()

	if *usetls {
		// TODO(halfwit) config.ServerTLS()
	}

	svc := &service{
		addr:   *addr,
		port:   *port,
		listen: *dir,
		chatty: *chatty,
		tls:    *usetls,
	}

	// This will error if there are no services running
	// in the future we may want to facilitate service discovery
	// during run time
	srv, err := server.NewServer(ctx, svc, *dir)
	if err != nil {
		log.Fatal(err)
	}

	srv.Logger = log.Printf

	if e := srv.Listen(); e != nil {
		log.Fatal(e)
	}
}
