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

	events, err := listenEvents(ctx, config)
	if err != nil {
		log.Fatal(err)
	}
	//input, err := listenInput(config, ctx)
	//control, err := listenControl(config, ctx)
	//client, err := listenClients(ctx)

	err = registerMDNS(config)
	if err != nil {
		// Do we want to try n times to register here?
		log.Print(err)
	}

	for {
		select {
		case event := <-events:
			handleEvent(event)
		//case in := <-input:
		//handleInput(in)
		//case ctl := <-control:
		//handleCtl(ctl)
		//case cli := <- client:
		//handleClient(cli)
		case sig := <-signals:
			handleSig(ctx, sig.String())
		case <-ctx.Done():
			cleanupMDNS()
			break
		}
	}
}
