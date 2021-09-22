package service 

import (
	"testing"
)

func TestCommands(t *testing.T) {
	reqs := make(chan string)
	ctl := &testctrl{}

	c, err := Mock(ctl, reqs, "cmdtest", false)
	if err != nil {
		t.Error(err)
	}

	var cmdlist2 []*Command
	cmdlist2 = append(cmdlist2, testMakeCmd("foo", []string{"<1>"}, ActionGroup, []string{}))
	cmdlist2 = append(cmdlist2, testMakeCmd("bar", []string{"<1>", "<2>"}, MediaGroup, []string{}))
	cmdlist2 = append(cmdlist2, testMakeCmd("baz", []string{"<2>", "<1>"}, ActionGroup, []string{}))
	cmdlist2 = append(cmdlist2, testMakeCmd("banana", []string{}, MediaGroup, []string{}))
	cmdlist2 = append(cmdlist2, testMakeCmd("nocomm", []string{}, ActionGroup, []string{"yacomm", "maybecomm"}))

	if e := c.SetCommands(cmdlist2...); e != nil {
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
