package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"os"

	_ "net/http/pprof"

	"github.com/altid/server"
	"github.com/altid/server/settings"
)

var factotum = flag.Bool("f", false, "Enable client authentication via a plan9 factotum")
var dir = flag.String("m", "/tmp/altid", "Path to Altid services")
var usetls = flag.Bool("t", false, "Use TLS")
var debug = flag.Bool("d", false, "Debug")
var chatty = flag.Bool("D", false, "Chatty")
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

	// Send all our flags up to the libs
	// if the build fails there isn't any chance to recover
	// best approach will be just having the user try again
	set := settings.NewSettings(*dir, *factotum, tls.Certificate{})

	// This will error if there are no services running
	// in the future we may want to facilitate service discovery
	// during run time
	srv := server.NewServer(ctx, &service{
		listen: *dir,
		chatty: *chatty,
		tls:    *usetls,
	})
	
	srv.Logger = log.Printf

	if e := srv.Config(set); e != nil {
		log.Fatal(e)
	}

	if e := srv.Listen(); e != nil {
		log.Fatal(e)
	}
}
