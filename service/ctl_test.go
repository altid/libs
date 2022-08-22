package service

import (
	"os"
	"testing"

	"github.com/altid/libs/service/listener"
	"github.com/altid/libs/store"
)

type testctrl struct {
}

func (c *testctrl) Run(ctrl *Control, cmd *Command) error {
	switch cmd.Name {
	case "open":
	case "close":
	case "buffer":
	case "link":
	case "reload":
	case "restart":
	default:
	}

	return nil
}

func (c *testctrl) Quit() {}

func TestWriters(t *testing.T) {
	ctl := &testctrl{}

	l, err := listener.NewListen9p("127.0.0.1", "", "")
	if err != nil {
		t.Error(err)
	}

	l.Register(store.NewRamStore(), nil)

	td, _ := os.MkdirTemp("", "")
	c, err := New(ctl, l, td, true)
	if err != nil {
		t.Error(err)
	}

	defer c.Cleanup()

	go func() {
		c.CreateBuffer("foo")

		mw, err := c.MainWriter("foo")
		if err != nil {
			t.Error(err)
		}

		mw.Write([]byte("test"))
		mw.Write([]byte("input:foo:There is no spoon"))
		mw.Close()
	}()

	if e := c.Listen(); e != nil {
		t.Error(e)
	}
}
