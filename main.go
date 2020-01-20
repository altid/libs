package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path"

	"github.com/altid/fslib"
)

//var enableFactotum = flag.Bool("f", false, "Enable client authentication via a plan9 factotum")
var inpath = flag.String("m", "/tmp/altid", "Path to Altid services")
var listenPort = flag.Int("p", 564, "Port to listen on")
var usetls = flag.Bool("t", false, "Use TLS")
var certfile = flag.String("c", "/etc/ssl/certs/altid.pem", "Path to certificate file")
var keyfile = flag.String("k", "/etc/ssl/private/altid.pem", "Path to key file")
var username = flag.String("u", "", "Run as another user")

var defaultUid string
var defaultGid string

func init() {
	us, _ := user.Current()
	defaultUid = us.Uid
	defaultGid = us.Gid
}

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(0)
	}
	if *username != "" {
		us, err := user.Lookup(*username)
		if err != nil {
			log.Fatal(err)
		}
		defaultUid = us.Uid
		defaultGid = us.Gid
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

	ctx, cancel := context.WithCancel(context.Background())
	srv, err := newServer(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	if len(srv.services) < 1 {
		log.Fatal("Found no running services, exiting")
	}

	/*
		err = registerMDNS(srv.services)
		if err != nil {
			log.Fatal(err)
		}
	*/
	go srv.listenEvents()
	go srv.start()

	for {
		select {
		case sig := <-signals:
			handleSig(cancel, sig.String())
		case <-ctx.Done():
			cleanupMDNS()
			os.Exit(0)
		}
	}
}
