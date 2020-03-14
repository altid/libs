package fs

import (
	"testing"
)

type testctrl struct{}

func (c *testctrl) Open(ctl *Control, buf string) error {
	return ctl.CreateBuffer(buf, "test")
}

func (c *testctrl) Close(ctl *Control, buf string) error {
	return ctl.DeleteBuffer(buf, "test")
}

func (c *testctrl) Link(ctl *Control, to, from string) error {
	if e := ctl.DeleteBuffer(from, "test"); e != nil {
		return e
	}

	return ctl.CreateBuffer(to, "test")
}

func (c *testctrl) Default(ctl *Control, cmd *Command) error {
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
	ctl := &testctrl{}

	c, err := MockCtlFile(ctl, reqs, true)
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
	ctl := &testctrl{}

	c, err := MockCtlFile(ctl, reqs, true)
	if err != nil {
		t.Error(err)
	}

	defer c.Cleanup()

	go func() {
		// `reqs <- "open foo"` is a race condition, but on a real client there will always
		// be an Open called before MainWriter (generally you call MainWriter in your client's Open method);
		// So we explicitely call c.CreateBuffer to avoid in the mock client tests
		c.CreateBuffer("foo", "feed")
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
