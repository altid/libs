package fs

import (
	"context"
	"testing"
	"time"
)

type testctrl struct {
	cancel context.CancelFunc
}

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

func (c *testctrl) Refresh(ctl *Control) error {
	return nil
}

func (c *testctrl) Restart(ctl *Control) error {
	return nil
}

func (c *testctrl) Quit() {
	c.cancel()
}

func sendctl(reqs chan string) {
	// Commands are sent with the current buffer for context
	reqs <- "open current foo"
	reqs <- "open current bar"
	reqs <- "link current bar baz"
	reqs <- "gibberish"
	reqs <- "close current baz"
	reqs <- "test quit"
}

func TestCtl(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	reqs := make(chan string)
	ctl := &testctrl{cancel}

	c, err := MockCtlFile(ctx, ctl, reqs, "test", false)
	if err != nil {
		t.Error(err)
	}

	c.SetCommands()
	defer c.Cleanup()

	go sendctl(reqs)

	if e := c.Listen(); e != nil {
		t.Error(e)
	}
}

func TestWriters(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	reqs := make(chan string)
	ctl := &testctrl{cancel}

	c, err := MockCtlFile(ctx, ctl, reqs, "test", false)
	if err != nil {
		t.Error(err)
	}

	defer c.Cleanup()

	go func() {
		// `reqs <- "open foo"` is a race condition, but on a real client there will always
		// be an Open called before MainWriter (generally you call MainWriter in your client's Open method);
		// So we explicitly call c.CreateBuffer to avoid in the mock client tests
		c.CreateBuffer("foo", "feed")
		mw, err := c.MainWriter("foo", "feed")
		if err != nil {
			t.Error(err)
		}

		mw.Write([]byte("test"))
		mw.Close()
		reqs <- "test quit"
	}()

	if e := c.Listen(); e != nil {
		t.Error(e)
	}
}

func TestCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	reqs := make(chan string)
	ctl := &testctrl{cancel}

	c, err := MockCtlFile(ctx, ctl, reqs, "test", false)
	if err != nil {
		t.Error(err)
	}

	defer c.Cleanup()

	time.AfterFunc(time.Second*2, func() {
		cancel()
	})

	if e := c.Listen(); e != nil {
		t.Error(e)
	}
}
