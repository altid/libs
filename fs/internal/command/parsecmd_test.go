package command

import (
	"testing"
)

var cmdlist = []*Command{
	{
		Name:        "link",
		Args:        []string{"<to>", "<from>"},
		Description: "Tests",
		Alias:       []string{"foo", "bar"},
		Heading:     DefaultGroup,
	},
	{
		Name: "quit",
	},
}

func TestParseCmd(t *testing.T) {

	cmd, err := ParseCmd(`link "my buffer" "new buffer"`, cmdlist)
	if err != nil {
		t.Error(err)
		return
	}

	if cmd.From != "my buffer" {
		t.Error("unable to parse compound")
	}

	cmd, err = ParseCmd(`foo bar`, cmdlist)
	if err != nil {
		t.Error(err)
	}

	if cmd.Args[0] != "bar" {
		t.Error("unable to parse simple")
	}

	cmd, err = ParseCmd(`quit`, cmdlist)
	if err != nil {
		t.Error(err)
	}

	if cmd.Name != "quit" && cmd.From != "" {
		t.Error("Unable to parse cmd with no args")
	}
}
