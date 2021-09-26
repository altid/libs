package service 

import (
	"testing"
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
	reqs := make(chan string)
	ctl := &testctrl{}

	c, err := Mock(ctl, reqs, "test", true)
	if err != nil {
		t.Error(err)
	}

	defer c.Cleanup()

	go func() {
		// `reqs <- "open foo"` is a race condition, but on a real client there will always
		// be an Open called before MainWriter (generally you call MainWriter in your client's Open method);
		// So we explicitly call c.CreateBuffer to avoid in the mock client tests
		c.CreateBuffer("foo")

		if e := c.Input("foo"); e != nil {
			t.Error(e)
		}

		mw, err := c.MainWriter("foo")
		if err != nil {
			t.Error(err)
		}

		mw.Write([]byte("test"))
		mw.Write([]byte("input:foo:There is no spood"))
		mw.Close()
		reqs <- "test quit"
	}()

	if e := c.Listen(); e != nil {
		t.Error(e)
	}
}
