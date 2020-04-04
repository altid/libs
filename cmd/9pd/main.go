package main

import (
	"context"
	"flag"
	"log"
	"os"

	_ "net/http/pprof"

	"github.com/altid/server/internal/ninep"
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
		cert = ""
		key = ""
	}

	// Send all our flags up to the libs
	// if the build fails there isn't any chance to recover
	// best approach will be just having the user try again
	set := ninep.NewSettings(*debug, *chatty, cert, key, *dir, *factotum, *usetls)
	if e := set.BuildServices(ctx); e != nil {
		log.Fatal(e)
	}

	// This will error if there are no services running
	// in the future we may want to facilitate service discovery
	// during run time
	srv, err := ninep.NewServer(ctx, set)
	if err != nil {
		log.Fatal(err)
	}

	if e := srv.Run(); e != nil {
		log.Fatal(e)
	}

}
