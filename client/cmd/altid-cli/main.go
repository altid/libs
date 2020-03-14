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

const usage = `
Commands are entered simply by typing a leading slash
All other input is sent to the input channel of the current buffer.
Commands are:
/quit				# exit
/buffer <target>	# swap to named buffer, if it exists
/open <target>		# open and swap to named buffer
/close <target>		# close named buffer
/link <to> <from>	# close current buffer and replace with named buffer
/title		# print the title of the current buffer
/aside		# print the aside data for the current buffer
/status		# print the status of the current buffer
/tabs		# display a list of all connected buffers
/notify		# display any pending notifications and clear them
`

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

	// Ideally we would call auth here, when it's properly supported

	if e := cl.Attach(); e != nil {
		log.Fatal(e)
	}

	getDocument := func() {
		f, err := cl.Document()
		if err != nil {
			log.Println("Unable to find a feed or document for this buffer")
			return
		}

		fmt.Printf("%s\n", f)
	}

	getFeed := func() {
		f, err := cl.Feed()
		if err != nil {
			getDocument()
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
		case "/help":
			fmt.Print(usage)
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

			time.Sleep(time.Millisecond * 200)
			go getFeed()
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
		// TODO(halfwit): We want to track the current buffer
		// and only send the `from` field internally
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
