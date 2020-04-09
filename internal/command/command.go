package command

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// CmdType is a type of default command
type CmdType int

// Currently supported cmdtypes
const (
	OtherCmd CmdType = iota
	OpenCmd
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
	"other":   OtherCmd,
	"open":    OpenCmd,
	"close":   CloseCmd,
	"buffer":  BufferCmd,
	"link":    LinkCmd,
	"quit":    QuitCmd,
	"reload":  ReloadCmd,
	"restart": RestartCmd,
}

// Command is a ctl message
type Command struct {
	// The sending client ID
	UUID    uint32
	CmdType CmdType
	// Arguments for known commands, such as buffer/open/close/link
	Args []string
	// The active buffer of the client who sent the command
	From string
}

// New - helper func with varargs
// If there are 2 args, the first is assumed to be the "From" path, the second "Args"
// If only one, it is set as "Args"
//
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
}

// WriteOut will write the command correctly to the Writer
// If a command is poorly formatted, it will return an error
func (c *Command) WriteOut(w io.Writer) error {
	// check the args count before writing
	switch c.CmdType {
	case OpenCmd:
		if len(c.Args) != 1 {
			return errors.New("incorrect arguments to open")
		}

		fmt.Fprintf(w, "open %s %s\n", c.From, c.Args[0])
	case RestartCmd:
		return errors.New("restart command not writable. Did you mean service restart?")
	case ReloadCmd:
		return errors.New("reload command not writable. Did you mean service reload?")
	case BufferCmd:
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
	case QuitCmd:
		fmt.Fprintln(w, "quit")
	default:
		fmt.Fprintf(w, "%s %s %s\n", c.Args[0], c.From, c.Args[1])
	}

	return nil
}
