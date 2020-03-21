package fs

import (
	"testing"
	"time"
)

func TestCommands(t *testing.T) {
	reqs := make(chan string)
	ctl := &testctrl{}

	c, err := MockCtlFile(ctl, reqs, "test", false)
	if err != nil {
		t.Error(err)
	}

	var cmdlist []*Command
	cmdlist = append(cmdlist, testMakeCmd("foo", []string{"<1>"}, ActionGroup, []string{}))
	cmdlist = append(cmdlist, testMakeCmd("bar", []string{"<1>", "<2>"}, MediaGroup, []string{}))
	cmdlist = append(cmdlist, testMakeCmd("baz", []string{"<2>", "<1>"}, ActionGroup, []string{}))
	cmdlist = append(cmdlist, testMakeCmd("banana", []string{}, MediaGroup, []string{}))
	cmdlist = append(cmdlist, testMakeCmd("nocomm", []string{}, ActionGroup, []string{"yacomm", "maybecomm"}))

	if e := c.SetCommands(cmdlist...); e != nil {
		t.Error(e)
	}

	time.AfterFunc(time.Second*1, func() {
		reqs <- "foo current test"
		reqs <- "bar current test this"
		reqs <- "baz current test this"
		reqs <- "nocomm current jump up jump up and get down"
		reqs <- "foo current too many args should log error"
		reqs <- "banana current"
		reqs <- "test quit"
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
