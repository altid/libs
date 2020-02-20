// gomobile-compatible library for creating clients
package client

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
)

var debug = flag.Int("d", 0, "debug level")
var addr = flag.String("a", "127.0.0.1", "IP to dial")

func run() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(0)
	}
	ctx := context.Background()
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	s, err := attach(ctx, u.Username)
	if err != nil {
		log.Fatal(err)
	}

	var off int64

	for {
		b, err := s.readFile(ctx, "feed", off)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		off += int64(len(b) - 1)
		fmt.Printf("%s", b)
	}
}
