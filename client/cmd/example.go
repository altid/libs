package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/altid/libs/client"
)

func main() {
	cl := client.NewClient(os.Args[1])

	if e := cl.Connect(0); e != nil {
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
			if _, err := cl.Ctl(client.CmdBuffer, args[1]); err != nil {
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
			if _, err := cl.Ctl(client.CmdOpen, args[1]); err != nil {
				log.Println(err)
			}
		case "/close":
			if _, err := cl.Ctl(client.CmdClose, args[1]); err != nil {
				log.Println(err)
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
