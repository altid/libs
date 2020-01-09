package main

import (
	"flag"
	"fmt"
	"log"
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
	//TODO(halfwit) switch to context here for all threads
	events, done := listenEvents(config)
	//events, err := listenEvents(config, ctx)
	//input, err := listenInput(config, ctx)
	//control, err := listenControl(config, ctx)
	//client, err := listenClients(ctx)
	for {
		// Use the select here to keep all clients in scope so messages can go vhere they need to
		select {
		case event := <-events:
			if event == nil {
				continue
			}
			// events will have the service and any (possibly > 1) line of files with changes.
			fmt.Printf("%s - %s", event.name, event.lines)
		//case in := <-input:
		//case ctl := <-control:
		//case client := <- client:
		case <-done:
			break
		}
	}
}
