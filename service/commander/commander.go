package commander

import (
	"fmt"
	"io"
	"strings"

	"github.com/altid/libs/service/callback"
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

type Commander interface {
	FindCommands(b []byte) ([]*Command, error)
	FromString(string) (*Command, error)
	FromBytes([]byte) (*Command, error)
	FindCommand(string, []*Command) (*Command, error)
	WriteCommands([]*Command, io.Writer) error
	RunCommand() func(*Command) error
	CtrlData() func() []byte
}

// Allow sorting of our lists
type CmdList []*Command

func (a CmdList) Len() int           { return len(a) }
func (a CmdList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CmdList) Less(i, j int) bool { return a[i].Heading < a[j].Heading }

// Command represents an available command to a service
// The From field should generally be populated, except in the case of a ServiceGroup command
type Command struct {
	Name        string
	Description string
	Heading     ComGroup
	Args        []string
	Alias       []string
	From        string
	Sender      callback.Sender
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
