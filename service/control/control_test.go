package control

import (
	"io/ioutil"
	"testing"

	"github.com/altid/libs/service/command"
	"github.com/altid/libs/service/listener"
	"github.com/altid/libs/store"
)

type testctrl struct {
}

func (c *testctrl) Run(ctrl *Control, cmd *command.Command) error {
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

// Assure that our writers work with the store as expected
func TestWriters(t *testing.T) {
	ctl := &testctrl{}

	l, err := listener.NewListen9p("127.0.0.1:12345", "", "")
	if err != nil {
		t.Error(err)
	}

	s := store.NewRamStore()
	l.Register(s, nil)

	c, err := New(ctl, s, l, "", false)
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

	if _, e := mw.Write([]byte("test/status\n")); e != nil {
		t.Error(e)
	}
	if _, e := mw.Write([]byte("There is no spoon")); e != nil {
		t.Error(e)
	}

	tw, err := c.TitleWriter("test")
	if err != nil {
		t.Error(err)
	}

	if _, e := tw.Write([]byte("chicken")); e != nil {
		t.Error(e)
	}

	li := s.List()
	t.Logf("%s\n", li)
	if len(li) != 4 {
		t.Errorf("Expected 4 files, found %d\n", len(li))
	}

	f, err := s.Open("test")
	if err != nil {
		t.Error(err)
	}

	f.Write([]byte(" nuggets"))

	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Error(err)
	}


	t.Logf("Result: %s\n", b)
	if e := mw.Close(); e != nil {
		t.Error(e)
	}
}
