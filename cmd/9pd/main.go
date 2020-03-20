package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/altid/9pd/internal/ninep"
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

	ctx := context.Background()

	settings := ninep.NewSettings(*debug, *chatty, *dir, *port, *factotum, *usetls)
	if e := settings.BuildServices(ctx); e != nil {
		log.Fatal(e)
	}

	srv, err := ninep.NewServer(ctx, settings)
	if err != nil {
		log.Fatal(err)
	}

	if e := srv.Run(); e != nil {
		log.Fatal(e)
	}
}
