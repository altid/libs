package fs

import (
	"github.com/altid/libs/fs/internal/command"
)

// ComGroup is a logical grouping of commands
// To add a ComGroup, please do so in a PR
type ComGroup int

const (
	DefaultGroup ComGroup = iota
	ActionGroup
	MediaGroup
)

// Command represents an available command to a service
type Command struct {
	Name        string
	Description string
	Heading     ComGroup
	Args        []string
	Alias       []string
	From        string
}

func ParseCommands(b []byte) ([]*Command, error) {

}

// Conversion functions for our internal command type
func setCommands(r runner, cmds ...*Command) error {
	// Parse into command structure and set
	var cmdlist []*command.Command

	for _, cmd := range cmds {
		c := &command.Command{
			Name:        cmd.Name,
			Description: cmd.Description,
			Heading:     command.ComGroup(cmd.Heading),
			Args:        cmd.Args,
			Alias:       cmd.Alias,
			From:        cmd.From,
		}

		cmdlist = append(cmdlist, c)
	}

	return r.SetCommands(cmdlist...)
}

func cmd2Command(cmd *command.Command) *Command {
	return &Command{
		Name:        cmd.Name,
		Description: cmd.Description,
		Heading:     ComGroup(cmd.Heading),
		Args:        cmd.Args,
		Alias:       cmd.Alias,
		From:        cmd.From,
	}
}
