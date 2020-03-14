package fs

import (
	"testing"
	"time"
)

func TestCommands(t *testing.T) {
	reqs := make(chan string)
	ctl := &testctrl{}

	c, err := MockCtlFile(ctl, reqs, false)
	if err != nil {
		t.Error(err)
	}

	var cmdlist []*Command
	cmdlist = append(cmdlist, testMakeCmd("foo", []string{"<1>"}, ActionGroup, []string{}))
	cmdlist = append(cmdlist, testMakeCmd("bar", []string{"<1>", "<2>"}, MediaGroup, []string{}))
	cmdlist = append(cmdlist, testMakeCmd("baz", []string{"<2>", "<1>"}, ActionGroup, []string{}))
	cmdlist = append(cmdlist, testMakeCmd("banana", []string{}, MediaGroup, []string{}))
	cmdlist = append(cmdlist, testMakeCmd("nocomm", []string{}, ActionGroup, []string{"yacomm"}))

	if e := c.SetCommands(cmdlist...); e != nil {
		t.Error(e)
	}

	time.AfterFunc(time.Second*5, func() {
		reqs <- "foo test"
		reqs <- "bar test this"
		reqs <- "baz test this"
		reqs <- "nocomm"
		reqs <- "foo too many args should log error"
		reqs <- "banana"
		reqs <- "quit"
	})

	defer c.Cleanup()

	if e := c.Listen(); e != nil {
		t.Error(e)
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
