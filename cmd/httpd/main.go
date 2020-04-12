package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/altid/server"
)

var factotum = flag.Bool("f", false, "Enable client authentication via factotum")
var usetls = flag.Bool("t", false, "Enable TLS")
var debug = flag.Bool("d", false, "Debug")

var port = flag.String("p", "22", "Port to listen on ")
var addr = flag.String("l", "", "Address to listen on")
var dir = flag.String("m", "/tmp/altid", "Path to Altid services")

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(0)
	}

	ctx := context.Background()
	svc := &service{}

	srv, err := server.NewServer(ctx, svc, *dir)
	if err != nil {
		log.Fatal(err)
	}

	if *debug {
		srv.Logger = log.Printf
	}

	if e := srv.Listen(); e != nil {
		log.Fatal(e)
	}
}
