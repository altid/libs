package fs

import (
	"errors"
	"strings"
)

// ComGroup is a logical grouping of commands
type ComGroup int

const (
	// DefaultGroup includes normal commands
	// `Open`, `Close`, `Buffer`, `Link`, `Quit`
	DefaultGroup ComGroup = iota
	// ActionGroup is for chat emotes
	ActionGroup
	// MediaGroup is for media control
	// Such as `play`, `pause` and can be used in a client
	// to make media control clusters
	MediaGroup
)

// Command represents an avaliable command to a service
type Command struct {
	Name        string
	Description string
	Heading     ComGroup
	Args        []string
	Alias       []string
}

type cmdList []*Command

func (a cmdList) Len() int           { return len(a) }
func (a cmdList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a cmdList) Less(i, j int) bool { return a[i].Heading < a[j].Heading }

func buildCommand(cmd string, cmdlist []*Command) (*Command, error) {
	var name string
	var args []string

	items := strings.Fields(cmd)
	name = items[0]

	if len(items) > 1 {
		args = items[1:]
	}

	for _, comm := range cmdlist {
		if comm.Name == name {
			return newCommand(comm, args), nil
		}

		for _, alias := range comm.Alias {
			if alias == name {
				return newCommand(comm, args), nil
			}
		}
	}

	return nil, errors.New("command not supported")
}

func newCommand(comm *Command, args []string) *Command {
	return &Command{
		Name:        comm.Name,
		Description: comm.Description,
		Heading:     comm.Heading,
		Args:        args,
		Alias:       comm.Alias,
	}
}
