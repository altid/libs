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
var usetls = flag.Bool("t", false, "Use TLS")
var certfile = flag.String("c", "/etc/ssl/certs/altid.pem", "Path to certificate file")
var keyfile = flag.String("k", "/etc/ssl/private/altid.pem", "Path to key file")
var dir = flag.String("d", "/tmp/altid", "Directory to watch")
var username = flag.String("u", "", "Run as another user")

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(0)
	}
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

	if len(srv.services) < 1 {
		log.Fatal("Found no running services, exiting")
	}

	err = registerMDNS(srv.services)
	if err != nil {
		log.Fatal(err)
	}

	go srv.listenEvents()
	go srv.start()

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
