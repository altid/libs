package service

import (
	"os"
	"testing"

	"github.com/altid/libs/service/listener"
	"github.com/altid/libs/store"
)

func TestCommands(t *testing.T) {
	ctl := &testctrl{}
	l, err := listener.NewListen9p("127.0.0.1", "", "")
	if err != nil {
		t.Error(err)
	}

	s := store.NewRamStore()
	l.Register(s, nil)
	p, _ := os.MkdirTemp("", "")

	c, err := New(ctl, s, l, p, false)
	if err != nil {
		t.Error("failed to create Control", err)
	}

	var cmdlist2 []*Command
	cmdlist2 = append(cmdlist2, testMakeCmd("foo", []string{"<1>"}, ActionGroup, []string{}))
	cmdlist2 = append(cmdlist2, testMakeCmd("bar", []string{"<1>", "<2>"}, MediaGroup, []string{}))
	cmdlist2 = append(cmdlist2, testMakeCmd("baz", []string{"<2>", "<1>"}, ActionGroup, []string{}))
	cmdlist2 = append(cmdlist2, testMakeCmd("banana", []string{}, MediaGroup, []string{}))
	cmdlist2 = append(cmdlist2, testMakeCmd("nocomm", []string{}, ActionGroup, []string{"yacomm", "maybecomm"}))

	if e := c.SetCommands(cmdlist2...); e != nil {
		t.Error("error in setting commands", e)
	}
}

func testMakeCmd(name string, args []string, com ComGroup, alias []string) *Command {
	return &Command{
		Name:    name,
		Args:    args,
		Heading: com,
		Alias:   alias,
	}
}
