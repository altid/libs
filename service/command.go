package service 

import (
	"fmt"
	"strings"

	"github.com/altid/libs/service/internal/command"
)

// ComGroup is a logical grouping of commands
// To add a ComGroup, please do so in a PR
type ComGroup int

// Currently supported ComGroups
const (
	DefaultGroup ComGroup = iota
	ActionGroup
	MediaGroup
	ServiceGroup
)

// Command represents an available command to a service
// The From field should generally be populated, except in the case of a ServiceGroup command
type Command struct {
	Name        string
	Description string
	Heading     ComGroup
	Args        []string
	Alias       []string
	From        string
}

// FindCommands within a byte array
// It returns an error if it encounters malformed input
func FindCommands(b []byte) ([]*Command, error) {
	var cmdlist []*Command

	cl, err := command.ParseCtlFile(b)
	if err != nil {
		return nil, err
	}

	for _, cmd := range cl {
                if cmd.Heading < 0 {
			return nil, fmt.Errorf("Unable to find a heading for %s", cmd.Name)
		}
		c := &Command{
			Name:        cmd.Name,
			Description: cmd.Description,
			Heading:     ComGroup(cmd.Heading),
			Args:        cmd.Args,
			Alias:       cmd.Alias,
			From:        cmd.From,
		}

		cmdlist = append(cmdlist, c)
	}

	return cmdlist, nil
}

// FromString returns a partially filled command
// It will have a Heading type of DefaultGroup
func FromString(cmd string) (*Command, error) {
	name, from, args, err := command.FindParts(cmd)
	if err != nil {
		return nil, err
	}

	if from != "" && len(args) < 1 {
		args = strings.Fields(from)
		from = ""
	}

	c := &Command{
		Name:    name,
		From:    from,
		Args:    args,
		Heading: DefaultGroup,
	}

	return c, nil
}

func (c *Command) String() string {
	args := strings.Join(c.Args, " ")
	if c.From != "" {
		return fmt.Sprintf("%s \"%s\" \"%s\"\n", c.Name, c.From, args)
	}
	return fmt.Sprintf("%s \"%s\"\n", c.Name, args)
}

// Bytes - Return a byte representation of a command
func (c *Command) Bytes() []byte {
	return []byte(c.String())
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
