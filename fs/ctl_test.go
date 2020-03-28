package fs

import (
	"context"
	"testing"
)

type testctrl struct {
	cancel context.CancelFunc
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

func (c *testctrl) Quit() {
	c.cancel()
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
