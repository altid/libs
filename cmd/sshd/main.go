package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/altid/server"
)

var factotum = flag.Bool("f", false, "Enable client authentication via factotum")
var debug = flag.Bool("d", false, "Debug")

var keys = flag.String("k", "~/.ssh/authorized_keys", "Path to authorized_keys file")
var port = flag.String("p", "22", "Port to listen on ")
var addr = flag.String("l", "", "Address to listen on")
var rsa = flag.String("r", "~/.ssh/id_rsa", "Path to id_rsa file")
var dir = flag.String("m", "/tmp/altid", "Path to Altid services")

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(0)
	}

	ctx := context.Background()
	svc := &service{
		logger: func(string, ...interface{}) {},
	}

	if e := svc.setup(); e != nil {
		log.Fatal(e)
	}

	srv, err := server.NewServer(ctx, svc, *dir)
	if err != nil {
		log.Fatal(err)
	}

	if *debug {
		svc.logger = log.Printf
		srv.Logger = log.Printf
	}

	if e := srv.Listen(); e != nil {
		log.Fatal(e)
	}
}
