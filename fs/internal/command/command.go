package command

import (
	"errors"
	"io"
	"log"
	"text/template"
)

// ComGroup is a logical grouping of commands
// To add a ComGroup, please do so in a PR
type ComGroup int

const (
	DefaultGroup ComGroup = iota
	ActionGroup
	MediaGroup
	ServiceGroup
)

//TODO(halfiwt) i18n
var DefaultCommands = []*Command{
	{
		Name:        "open",
		Args:        []string{"<buffer>"},
		Heading:     DefaultGroup,
		Description: "Open and change buffers to a given service",
	},
	{
		Name:        "close",
		Args:        []string{"<buffer>"},
		Heading:     DefaultGroup,
		Description: "Close a buffer and return to the last opened previously",
	},
	{
		Name:        "buffer",
		Args:        []string{"<buffer>"},
		Heading:     DefaultGroup,
		Description: "Change to the named buffer",
	},
	{
		Name:        "link",
		Args:        []string{"<current>", "<buffer>"},
		Heading:     DefaultGroup,
		Description: "Overwrite the current buffer with the named",
	},
	{
		Name:        "quit",
		Args:        []string{},
		Heading:     DefaultGroup,
		Description: "Exits the client",
	},
}

const commandTemplate = `{{range .}}	{{.Name}}{{if .Alias}}{{range .Alias}}|{{.}}{{end}}{{end}}{{if .Args}}	{{range .Args}}{{.}} {{end}}{{end}}{{if .Description}}	# {{.Description}}{{end}}
{{end}}`

// Command represents an available command to a service
type Command struct {
	Name        string
	Description string
	Heading     ComGroup
	Args        []string
	Alias       []string
	From        string
}

type CmdList []*Command

func (a CmdList) Len() int           { return len(a) }
func (a CmdList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CmdList) Less(i, j int) bool { return a[i].Heading < a[j].Heading }

func PrintCtlFile(cmdlist []*Command, to io.WriteCloser) error {
	var last int

	curr := cmdlist[0].Heading
	tp := template.Must(template.New("entry").Parse(commandTemplate))

	for n, comm := range cmdlist {
		// 0, 0 and comm.Heading != curr; we want to set a heading
		if comm.Heading != curr {
			cmdHeading(to, curr)
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
		cmdHeading(to, cmdlist[last].Heading)
		if e := tp.Execute(to, cmdlist[last:]); e != nil {
			return e
		}
	}

	return nil
}

// So for this, we want to parse out a proper cmd - each arg can have spaces if it's wrapped in \" \"
func ParseCmd(cmd string, cmdlist []*Command) (*Command, error) {
	name, from, args, err := parseCmd(cmd)
	if err != nil {
		return nil, err
	}

	for _, comm := range cmdlist {

		if comm.Name == name {
			return newFrom(comm, from, args)
		}

		for _, alias := range comm.Alias {
			if alias == name {
				return newFrom(comm, from, args)
			}
		}
	}

	return nil, errors.New("command not supported")
}

func cmdHeading(to io.WriteCloser, heading ComGroup) {
	switch heading {
	case ActionGroup:
		to.Write([]byte("emotes:\n"))
	case DefaultGroup:
		to.Write([]byte("general:\n"))
	case MediaGroup:
		to.Write([]byte("media:\n"))
	case ServiceGroup:
		to.Write([]byte("service:\n"))
	default:
		log.Fatal("Group not implemented")
	}
}

func newFrom(comm *Command, from string, args []string) (*Command, error) {
	if comm.Heading == ServiceGroup {
		c := &Command{
			Name:        comm.Name,
			Description: comm.Description,
			Heading:     ServiceGroup,
			Args:        args,
		}

		return c, nil
	}

	c := &Command{
		Name:        comm.Name,
		Description: comm.Description,
		Heading:     comm.Heading,
		Args:        args,
		Alias:       comm.Alias,
		From:        from,
	}
	return c, nil
}
