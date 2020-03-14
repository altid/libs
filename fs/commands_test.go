package fs

import (
	"testing"
	"time"
)

func TestCommands(t *testing.T) {
	reqs := make(chan string)
	ctl := &testctrl{}

	c, err := MockCtlFile(ctl, reqs, true)
	if err != nil {
		t.Error(err)
	}

	var cmdlist []*Command
	cmdlist = append(cmdlist, testMakeCmd("foo", []string{"1"}, ActionGroup))
	cmdlist = append(cmdlist, testMakeCmd("bar", []string{"1", "2"}, MediaGroup))
	cmdlist = append(cmdlist, testMakeCmd("baz", []string{"2", "1"}, ActionGroup))

	if e := c.SetCommands(cmdlist...); e != nil {
		t.Error(e)
	}

	time.AfterFunc(time.Second*10, func() {
		reqs <- "quit"
	})

	defer c.Cleanup()

	if e := c.Listen(); e != nil {
		t.Error(e)
	}
}

func testMakeCmd(name string, args []string, com ComGroup) *Command {
	return &Command{
		Name:    name,
		Args:    args,
		Heading: com,
	}
}
