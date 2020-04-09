package command

import (
	"errors"
	"fmt"
	"io"
)

// CmdType is a type of default command
type CmdType int

// Currently supported cmdtypes
const (
	OtherCmd CmdType = iota
	OpenCmd
<<<<<<< HEAD:command/command.go
	ReloadCmd
	RestartCmd
	BufferCmd
=======
>>>>>>> dev:internal/command/command.go
	CloseCmd
	BufferCmd
	LinkCmd
	QuitCmd
	ReloadCmd
	RestartCmd
)

// There are more possible commands, but these are the only ones
// the server itself intercepts in any meaningful way
var items = map[string]CmdType{
<<<<<<< HEAD:command/command.go
	"open":    OpenCmd,
	"close":   CloseCmd,
	"buffer":  BufferCmd,
	"restart": RestartCmd,
	"link":    LinkCmd,
	"quit":    QuitCmd,
	"reload":  ReloadCmd,
	"other":   OtherCmd,
=======
	"other":   OtherCmd,
	"open":    OpenCmd,
	"close":   CloseCmd,
	"buffer":  BufferCmd,
	"link":    LinkCmd,
	"quit":    QuitCmd,
	"reload":  ReloadCmd,
	"restart": RestartCmd,
>>>>>>> dev:internal/command/command.go
}

// Command is a ctl message
type Command struct {
	// The sending client ID
	UUID    uint32
	CmdType CmdType
<<<<<<< HEAD:command/command.go
	// The buffer that the command came from
	From string
	// The arguments to the command
	Args string
	// The raw command bytes
	Data []byte
=======
	// Arguments for known commands, such as buffer/open/close/link
	Args []string
	// The active buffer of the client who sent the command
	From string
>>>>>>> dev:internal/command/command.go
}

// New - helper func with varargs
// If there are 2 args, the first is assumed to be the "From" path, the second "Args"
// If only one, it is set as "Args"
//
<<<<<<< HEAD:command/command.go
//		cmdChannel <- New(clientuuid, BufferCmd, nil, "oldbuffer", "newbuffer")
//		cmdChannel <- New(clientuuid, CloseCmd, nil, "oldbuffer")
//		cmdChannel <- New(clientuuid, OtherCmd, []byte("emote I swapped buffers"))
//		cmdChannel <- New(clientuuid, LinkCmd, nil, "newbuffer", "oldbuffer")
//
func New(uuid uint32, cmdType CmdType, data []byte, args ...string) *Command {
	c := &Command{
		UUID:    uuid,
		CmdType: cmdType,
		Data:    data,
	}

	switch len(args) {
	case 1:
		c.Args = args[0]
	case 2:
		c.From = args[0]
		c.Args = args[1]
	}

	return c
=======
//		cmdChannel <- New(clientuuid, BufferCmd, "somebuffer", "newbuffer")
//		cmdChannel <- New(clientuuid, CloseCmd, "somebuffer", "oldbuffer")
//		cmdChannel <- New(clientuuid, LinkCmd, "oldbuffer", "newbuffer")
//
func New(uuid uint32, cmdType CmdType, from string, args ...string) *Command {
	return &Command{
		UUID:    uuid,
		CmdType: cmdType,
		From:    from,
		Args:    args,
	}
}

// FromBytes returns a command from byte input, or an error if it was unable to parse
func FromBytes(uuid uint32, from string, b []byte) (*Command, error) {
	return FromString(uuid, from, string(b))
}

// FromString retruns a command from string input, or an error if it was unable to parse
//
//		cmdChannel <- FromString(clientuuid, "buffer newbuffer")
//		cmdChannel <- FromString(clientuuid, "link new old")
//
func FromString(uuid uint32, from, s string) (*Command, error) {
	c := &Command{
		From: from,
		UUID: uuid,
	}

	t := strings.Fields(s)

	ct, ok := items[t[0]]
	if !ok {
		c.Args = t[1:]
		c.CmdType = OtherCmd
		return c, nil
	}

	c.CmdType = ct

	switch c.CmdType {
	case OpenCmd, BufferCmd, CloseCmd, LinkCmd:
		c.Args = t[1:]
	}

	return c, nil
>>>>>>> dev:internal/command/command.go
}

// WriteOut will write the command correctly to the Writer
// If a command is poorly formatted, it will return an error
func (c *Command) WriteOut(w io.Writer) error {
	// check the args count before writing
	switch c.CmdType {
	case OpenCmd:
<<<<<<< HEAD:command/command.go
		fmt.Fprintf(w, "open \"%s\" \"%s\"\n", c.From, c.Args)
=======
		if len(c.Args) != 1 {
			return errors.New("incorrect arguments to open")
		}

		fmt.Fprintf(w, "open %s %s\n", c.From, c.Args[0])
>>>>>>> dev:internal/command/command.go
	case RestartCmd:
		return errors.New("restart command not writable. Did you mean service restart?")
	case ReloadCmd:
		return errors.New("reload command not writable. Did you mean service reload?")
	case BufferCmd:
<<<<<<< HEAD:command/command.go
		fmt.Fprintf(w, "buffer \"%s\" \"%s\"\n", c.From, c.Args)
	case CloseCmd:
		fmt.Fprintf(w, "close \"%s\"\n", c.Args)
	case LinkCmd:
		fmt.Fprintf(w, "link \"%s\" \"%s\"\n", c.From, c.Args)
=======
		if len(c.Args) != 1 {
			return errors.New("incorrect arguments to buffer")
		}

		fmt.Fprintf(w, "buffer %s %s\n", c.From, c.Args[0])
	case CloseCmd:
		if len(c.Args) != 1 {
			return errors.New("incorrect arguments to close")
		}

		fmt.Fprintf(w, "close %s %s\n", c.From, c.Args[0])
	case LinkCmd:
		if len(c.Args) != 1 {
			return errors.New("missing to/from argument pair")
		}

		fmt.Fprintf(w, "link %s %s\n", c.From, c.Args[0])
>>>>>>> dev:internal/command/command.go
	case QuitCmd:
		fmt.Fprintln(w, "quit")
	default:
		fmt.Fprintf(w, "%s %s %s\n", c.Args[0], c.From, c.Args[1])
	}

	return nil
}
