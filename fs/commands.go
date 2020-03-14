package fs

import (
	"fmt"
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

	return nil, nil
}

func printCmdList(cmdlist []*Command) {
	for _, cmd := range cmdlist {
		fmt.Printf("%s %d\n", cmd.Name, cmd.Heading)
	}
}
