package fs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"text/template"
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

const commandTemplate = `{{range .}}	{{.Name}}{{if .Alias}}{{range .Alias}}|{{.}}{{end}}{{end}}{{if .Args}}	{{range .Args}}{{.}} {{end}}{{end}}{{if .Description}}	# {{.Description}}{{end}}
{{end}}`

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
			return newCommand(comm, args)
		}

		for _, alias := range comm.Alias {
			if alias == name {
				return newCommand(comm, args)
			}
		}
	}

	return nil, errors.New("command not supported")
}

func newCommand(comm *Command, args []string) (*Command, error) {
	if len(comm.Args) != len(args) && len(comm.Args) > 0 {
		return nil, fmt.Errorf("expected %d arguments: received %d", len(comm.Args), len(args))
	}

	c := &Command{
		Name:        comm.Name,
		Description: comm.Description,
		Heading:     comm.Heading,
		Args:        args,
		Alias:       comm.Alias,
	}
	return c, nil
}

func printCtlFile(cmdlist []*Command, to io.WriteCloser) error {
	var last int

	curr := cmdlist[0].Heading
	tp := template.Must(template.New("entry").Parse(commandTemplate))

	for n, comm := range cmdlist {
		// 0, 0 and comm.Heading != curr; we want to set a heading
		if comm.Heading != curr {
			switch curr {
			case ActionGroup:
				to.Write([]byte("emotes:\n"))
			case DefaultGroup:
				to.Write([]byte("general:\n"))
			case MediaGroup:
				to.Write([]byte("media:\n"))
			default:
				log.Fatal("Group not implemented")
			}

			for j, subcomm := range cmdlist[last:] {
				if subcomm.Heading != comm.Heading {
					if n+j > last {
						if e := tp.Execute(to, cmdlist[last:n+j]); e != nil {
							return e
						}
						last = n + j
					}
					break
				}
			}

			curr = comm.Heading
		}
	}

	// We have one Grouping remaining, print
	if last < len(cmdlist) {
		switch cmdlist[last].Heading {
		case ActionGroup:
			to.Write([]byte("emotes:\n"))
		case DefaultGroup:
			to.Write([]byte("general:\n"))
		case MediaGroup:
			to.Write([]byte("media:\n"))
		default:
			log.Fatal("Group not implemented")
		}
		if e := tp.Execute(to, cmdlist[last:]); e != nil {
			return e
		}
	}

	return nil
}
