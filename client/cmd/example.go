package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/altid/libs/client"
)

var debug = flag.Bool("d", false, "enable debug output")
var addr = flag.String("s", "127.0.0.1", "address to connect to")
var errBadArgs = errors.New("Incorrect arguments to command")

func main() {
	flag.Parse()

	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(0)
	}

	cl := client.NewClient(*addr)

	withDebug := 0
	if *debug {
		withDebug = 1
	}

	if e := cl.Connect(withDebug); e != nil {
		log.Fatal(e)
	}

	// calling auth on a server without authentication will always succeed
	if e := cl.Auth(); e != nil {
		log.Fatal(e)
	}

	if e := cl.Attach(); e != nil {
		log.Fatal(e)
	}

	getFeed := func() {
		f, err := cl.Feed()
		if err != nil {
			log.Print(err)
			return
		}

		defer f.Close()

		for {
			// Ensure your buffer is MSIZE
			b := make([]byte, client.MSIZE)

			_, err := f.Read(b)
			if err != nil && err != io.EOF {
				log.Print(err)
				return
			}

			fmt.Printf("%s", b)
		}
	}

	go getFeed()

	rd := bufio.NewReader(os.Stdin)

	for {
		line, err := rd.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}

		args := strings.Fields(line)

		switch args[0] {
		case "/quit":
			os.Exit(0)
		case "/buffer":
			if len(args) != 2 {
				log.Print(errBadArgs)
				continue
			}
			if _, err := cl.Buffer(args[1]); err != nil {
				log.Println(err)
				continue
			}

			time.Sleep(time.Millisecond * 200)
			go getFeed()
		case "/tabs":
			t, err := cl.Tabs()
			if err != nil {
				log.Println(err)
				continue
			}

			fmt.Printf("%s", t)
		case "/open":
			if len(args) != 2 {
				log.Print(errBadArgs)
				continue
			}
			if _, err := cl.Open(args[1]); err != nil {
				log.Println(err)
			}
		case "/close":
			if len(args) != 2 {
				log.Print(errBadArgs)
				continue
			}
			if _, err := cl.Close(args[1]); err != nil {
				log.Println(err)
				continue
			}

			time.Sleep(time.Millisecond * 200)
			go getFeed()
		case "/link":
			if len(args) != 3 {
				log.Println(errBadArgs)
				continue
			}
			if _, err := cl.Link(args[1], args[2]); err != nil {
				log.Println(err)
				continue
			}

			time.Sleep(time.Millisecond * 200)
			go getFeed()
		case "/title":
			t, err := cl.Title()
			if err != nil {
				log.Println(err)
				continue
			}

			fmt.Printf("%s", t)
		case "/status":
			t, err := cl.Status()
			if err != nil {
				log.Println(err)
				continue
			}

			fmt.Printf("%s", t)
		case "/aside":
			t, err := cl.Aside()
			if err != nil {
				log.Println(err)
				continue
			}

			fmt.Printf("%s", t)
		case "/notify":
			t, err := cl.Notifications()
			if err != nil {
				log.Println(err)
				continue
			}

			fmt.Printf("%s", t)
		default:
			if line[0] == '/' {
				//cl.Ctl([]byte(line[1:]))
				continue
			}
			cl.Input([]byte(line))
		}
	}
}
