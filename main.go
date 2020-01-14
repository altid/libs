package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path"

	"github.com/altid/fslib"
)

var enableFactotum = flag.Bool("f", false, "Enable client authentication via a plan9 factotum")
var inpath = flag.String("m", "/tmp/altid", "Path to Altid services")

func main() {
	confdir, err := fslib.UserConfDir()
	if err != nil {
		log.Fatal(err)
	}

	config, err := newConfig(path.Join(confdir, "altid", "config"))
	if err != nil {
		log.Fatal(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals)
	ctx := context.Background()

	srv, err := newServer(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	err = registerMDNS(config)
	if err != nil {
		// Do we want to try n times to register here?
		log.Print(err)
	}

	go srv.listenAndServe()

	for {
		select {
		case input := <-srv.inputs:
			handleInput(input)
		case control := <-srv.controls:
			handleControl(ctx, control)
		case client := <-srv.clients:
			handleClient(client)
		case sig := <-signals:
			handleSig(ctx, sig.String())
		case <-ctx.Done():
			cleanupMDNS()
			break
		}
	}
}
