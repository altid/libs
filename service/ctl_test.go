package service

import (
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

	l, err := listener.NewListen9p("127.0.0.1:12345", "", "")
	if err != nil {
		t.Error(err)
	}

	l.Register(store.NewRamStore(), nil)

	c, err := New(ctl, l, "", true)
	if err != nil {
		t.Error(err)
	}

	defer c.Cleanup()

	if e := c.CreateBuffer("test"); e != nil {
		t.Error(e)
	}

	mw, err := c.MainWriter("test")
	if err != nil {
		t.Error(err)
	}

	if _, e := mw.Write([]byte("test/status")); e != nil {
		t.Error(e)
	}
	if _, e := mw.Write([]byte("There is no spoon")); e != nil {
		t.Error(e)
	}
	
	if e := mw.Close(); e != nil {
		t.Error(e)
	}

}
