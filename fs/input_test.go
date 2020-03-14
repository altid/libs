package fs

import (
	"fmt"
	"testing"

	"github.com/altid/libs/markup"
)

type inputTestCtrl struct{}

func (i *inputTestCtrl) Handle(path string, c *markup.Lexer) error {
	d := c.String()

	switch d {
	case "foo bar baz":
	case "baz bar foo":
	case "boldly go":
	default:
		return fmt.Errorf("Incorrect string result %s", d)
	}

	return nil
}

func TestInput(t *testing.T) {
	reqs := make(chan string)
	ctl := &inputTestCtrl{}

	i, err := NewMockInput(ctl, "foo", true, reqs)
	if err != nil {
		t.Error(err)
	}

	i.Start()

	reqs <- "foo bar baz"
	reqs <- "baz bar foo"
	reqs <- "*boldly go*"
	reqs <- "-foo bar baz-"
	reqs <- "/baz bar foo/"

	if e := i.Errs(); len(e) > 0 {
		for _, err := range e {
			t.Error(err)
		}
	}
	i.Stop()
}
