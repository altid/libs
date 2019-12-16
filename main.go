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
	events, done := listen(config)
	for {
		select {
		case event := <-events:
			if event == nil {
				continue
			}
			// events will have the service and any (possibly > 1) line of files with changes.
			fmt.Printf("%s - %s", event.name, event.lines)
		case <-done:
			break
		}
	}
}
