package fs

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/altid/libs/markup"
)

type inputTestCtrl struct{}

func (i *inputTestCtrl) Handle(path string, c *markup.Lexer) error {
	d, err := c.String()
	if err != nil {
		return err
	}

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

	ew, _ := ioutil.TempFile("", "")
	i, err := NewMockInput(ctl, "foo", ew, false, reqs)
	if err != nil {
		t.Error(err)
	}

	i.Start()

	reqs <- "foo bar baz"
	reqs <- "baz bar foo"
	reqs <- "*boldly go*"
	reqs <- "-foo bar baz-"
	reqs <- "/baz bar foo/"

	i.Stop()
}
