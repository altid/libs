package fs

import (
	"testing"

	"github.com/altid/libs/fs"
)

type ctrl struct{}

func (c *ctrl) Open(ctl *fs.Control, buf string) error {
	return ctl.CreateBuffer(buf, "test")
}

func (c *ctrl) Close(ctl *fs.Control, buf string) error {
	return ctl.DeleteBuffer(buf, "test")
}

func (c *ctrl) Link(ctl *fs.Control, to, from string) error {
	if e := ctl.DeleteBuffer(from, "test"); e != nil {
		return e
	}

	return ctl.CreateBuffer(to, "test")
}

func (c *ctrl) Default(ctl *fs.Control, cmd, from, msg string) error {
	return nil
}

func sendctl(reqs chan string) {
	reqs <- "open foo"
	reqs <- "open bar"
	reqs <- "link bar baz"
	reqs <- "gibberish"
	reqs <- "close baz"
	reqs <- "quit"
}

func TestCtl(t *testing.T) {
	reqs := make(chan string)
	ctl := &ctrl{}

	c, err := fs.MockCtlFile(ctl, reqs, true)
	if err != nil {
		t.Error(err)
	}

	defer c.Cleanup()

	go sendctl(reqs)

	if e := c.Listen(); e != nil {
		t.Error(e)
	}
}

func TestWriters(t *testing.T) {
	reqs := make(chan string)
	ctl := &ctrl{}

	c, err := fs.MockCtlFile(ctl, reqs, true)
	if err != nil {
		t.Error(err)
	}

	defer c.Cleanup()
	
	go func() {
		reqs <- "open foo"
		
		mw, err := c.MainWriter("foo", "feed")
		if err != nil {
			t.Error(err)
		}

		mw.Write([]byte("test"))
		mw.Close()
		reqs <- "quit"
	}()

	if e := c.Listen(); e != nil {
		t.Error(e)
	}
}
