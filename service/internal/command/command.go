package command

import (
	"errors"
	"fmt"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/internal/parse"
	"html/template"
	"io"
	"strings"
)

type Command struct {
	SendCommand     func(*commander.Command) error
	CtrlDataCommand func() []byte
}

const commandTemplate = `{{range .}}	{{.Name}}{{if .Alias}}{{range .Alias}}|{{.}}{{end}}{{end}}{{if .Args}}	{{range .Args}}{{.}} {{end}}{{end}}{{if .Description}}	# {{.Description}}{{end}}
{{end}}`

// FindCommands within a byte array
// It returns an error if it encounters malformed input
func (c *Command) FindCommands(b []byte) ([]*commander.Command, error) {
	var cmdlist []*commander.Command
	cl, err := parse.ParseCtlFile(b)
	if err != nil {
		return nil, err
	}
	for _, comm := range cl {
		if comm.Heading < 0 {
			return nil, fmt.Errorf("unable to find a heading for %s", comm.Name)
		}
		c := &commander.Command{
			Name:        comm.Name,
			Description: comm.Description,
			Heading:     commander.ComGroup(comm.Heading),
			Args:        comm.Args,
			Alias:       comm.Alias,
			From:        comm.From,
		}
		cmdlist = append(cmdlist, c)
	}
	return cmdlist, nil
}

func (c *Command) Exec(cmd *commander.Command) error { return c.SendCommand(cmd) }
func (c *Command) CtrlData() func() []byte           { return c.CtrlDataCommand }
func (c *Command) FromBytes(input []byte) (*commander.Command, error) {
	return c.FromString(string(input))
}

// FromString returns a partially filled command
// It will have a Heading type of DefaultGroup
func (c *Command) FromString(input string) (*commander.Command, error) {
	name, from, args, err := parse.ParseCmd(input)
	if err != nil {
		return nil, err
	}
	if from != "" && len(args) < 1 {
		args = strings.Fields(from)
		from = ""
	}
	comm := &commander.Command{
		Name:    name,
		From:    from,
		Args:    args,
		Heading: commander.DefaultGroup,
	}
	return comm, nil
}

func (c *Command) WriteCommands(cmdlist []*commander.Command, to io.Writer) error {
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
func (c *Command) FindCommand(cmd string, cmdlist []*commander.Command) (*commander.Command, error) {
	name, from, args, err := parse.ParseCmd(cmd)
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

func cmdHeading(to io.Writer, heading commander.ComGroup) {
	switch heading {
	case commander.ActionGroup:
		to.Write([]byte("emotes:\n"))
	case commander.DefaultGroup:
		to.Write([]byte("general:\n"))
	case commander.MediaGroup:
		to.Write([]byte("media:\n"))
	case commander.ServiceGroup:
		to.Write([]byte("service:\n"))
	}
}

func newFrom(comm *commander.Command, from string, args []string) (*commander.Command, error) {
	if comm.Heading == commander.ServiceGroup {
		c := &commander.Command{
			Name:        comm.Name,
			Description: comm.Description,
			Heading:     commander.ServiceGroup,
			Args:        args,
		}
		return c, nil
	}
	c := &commander.Command{
		Name:        comm.Name,
		Description: comm.Description,
		Heading:     comm.Heading,
		Args:        args,
		Alias:       comm.Alias,
		From:        from,
	}
	return c, nil
}
