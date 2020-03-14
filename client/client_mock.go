package client

import (
	"io"
	"log"
	"os"
	"time"

	fuzz "github.com/google/gofuzz"
)

type mock struct {
	afid  io.ReadWriteCloser
	addr  string
	debug func(int, ...interface{})
}

func (c *mock) cleanup() {}

func (c *mock) connect(debug int) error {
	if debug > 0 {
		c.debug = mockLogging
	}

	c.debug(CmdConnect, c.addr)
	return nil
}

// Test the afid here
func (c *mock) attach() error {
	// Read on RPC for
	c.debug(CmdAttach, true)
	return nil
}

func (c *mock) auth() error {
	// TODO(halfwit): We want to flag in factotum use and hand it an afid
	c.debug(CmdAuth, true)
	return nil
}

// We want to eventually create and track tabs internally to the library
func (c *mock) ctl(cmd int, args ...string) (int, error) {
	data, err := runClientCtl(cmd, args...)
	if err != nil {
		return 0, err
	}

	c.debug(cmd, data)
	return 0, nil
}

func (c *mock) tabs() ([]byte, error) {
	//c.debug(CmdTabs, c.tablist)
	return nil, nil
}

func (c *mock) title() ([]byte, error) {
	//c.debug(CmdTitle, c.current)
	//return c.current, nil
	return nil, nil
}

func (c *mock) status() ([]byte, error) {
	c.debug(CmdStatus)
	return []byte("status"), nil
}

func (c *mock) document() ([]byte, error) {
	b := make([]byte, 4096)
	fuzz := fuzz.New()

	fuzz.Fuzz(&b)
	c.debug(CmdDocument, b)
	return b, nil
}

func (c *mock) aside() ([]byte, error) {
	c.debug(CmdAside)
	return []byte("aside"), nil
}

func (c *mock) input(data []byte) (int, error) {
	c.debug(CmdInput, data)
	return len(data), nil
}

func (c *mock) notifications() ([]byte, error) {
	c.debug(CmdNotify)
	return []byte("notifications"), nil
}

func (c *mock) feed() (io.ReadCloser, error) {
	data := make(chan []byte)
	done := make(chan struct{})

	go func() {
		var b []byte
		fuzz := fuzz.New()
		defer close(data)

		for {
			select {
			case <-done:
				return
			default:
				fuzz.Fuzz(&b)
				c.debug(CmdFeed, b)
				data <- b
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()

	f := &feed{
		data: data,
		done: done,
	}

	return f, nil
}

func mockLogging(cmd int, args ...interface{}) {
	l := log.New(os.Stdout, "client ", 0)

	switch cmd {
	case CmdConnect:
		l.Printf("connect addr=\"%s\"\n", args[0])
	case CmdAttach:
		l.Printf("attach success=%t\n", args[0])
	case CmdAuth:
		l.Printf("auth success=%t\n", args[0])
	case CmdBuffer, CmdOpen, CmdClose, CmdLink:
		l.Printf("cmd %s", args[0])
		l.Println()
	case CmdTabs:
		l.Printf("tabs list=\"%s\"\n", args[0])
	case CmdTitle:
		l.Printf("title data=\"%s\"\n", args[0])
	case CmdStatus:
		l.Println("status data=nil")
	case CmdAside:
		l.Println("aside data=nil")
	case CmdInput:
		l.Printf("input data=\"%s\"\n", args[0])
	case CmdNotify:
		l.Println("notification data=nil")
	case CmdFeed:
		l.Printf("feed data=\"%s\"\n", args[0])
	case CmdDocument:
		l.Printf("document data=\"%s\"\n", args[0])
	}
}
