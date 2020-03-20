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
	ReloadCmd
	BufferCmd
	CloseCmd
	LinkCmd
	QuitCmd
)

// There are more possible commands, but these are the only ones
// the server itself intercepts in any meaningful way
var items = map[string]CmdType{
	"open":   OpenCmd,
	"close":  CloseCmd,
	"buffer": BufferCmd,
	"link":   LinkCmd,
	"quit":   QuitCmd,
	"reload": ReloadCmd,
	"other":  OtherCmd,
}

// Command is a ctl message
type Command struct {
	// The sending client ID
	UUID    uint32
	CmdType CmdType
	// Arguments for known commands, such as buffer/open/close/link
	Args []string
	// The raw command bytes
	Data []byte
}

// New - helper func with varargs
//
//		cmdChannel <- New(clientuuid, BufferCmd, nil, "newbuffer")
//		cmdChannel <- New(clientuuid, CloseCmd, nil, "oldbuffer")
//		cmdChannel <- New(clientuuid, OtherCmd, []byte("emote I swapped buffers"))
//		cmdChannel <- New(clientuuid, LinkCmd, nil, "newbuffer", "oldbuffer")
//
func New(uuid uint32, cmdType CmdType, data []byte, args ...string) *Command {
	return &Command{
		UUID:    uuid,
		CmdType: cmdType,
		Data:    data,
		Args:    args,
	}
}

// FromBytes returns a command from byte input, or an error if it was unable to parse
func FromBytes(uuid uint32, b []byte) (*Command, error) {
	return FromString(uuid, string(b))
}

// FromString retruns a command from string input, or an error if it was unable to parse
//
//		cmdChannel <- FromString(clientuuid, "buffer newbuffer")
//		cmdChannel <- FromString(clientuuid, "link new old")
//
func FromString(uuid uint32, s string) (*Command, error) {
	c := &Command{
		UUID: uuid,
		Data: []byte(s),
	}

	t := strings.Fields(s)

	var ok bool
	c.CmdType, ok = items[t[0]]
	if !ok {
		c.CmdType = OtherCmd
		return c, nil
	}

	switch c.CmdType {
	case OpenCmd:
		return add(c, t, 1)
	case CloseCmd:
		return add(c, t, 1)
	case BufferCmd:
		return add(c, t, 1)
	case LinkCmd:
		return add(c, t, 2)
	case QuitCmd:
		return c, nil
	case ReloadCmd:
		return c, nil
	default:
		panic("should not happen")
	}
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

		fmt.Fprintf(w, "open %s\n", c.Args[0])
	case ReloadCmd:
		return errors.New("reload command not writable. Did you mean service reload?")
	case BufferCmd:
		if len(c.Args) != 1 {
			return errors.New("incorrect arguments to buffer")
		}

		fmt.Fprintf(w, "buffer %s\n", c.Args[0])
	case CloseCmd:
		if len(c.Args) != 1 {
			return errors.New("incorrect arguments to close")
		}

		fmt.Fprintf(w, "close %s\n", c.Args[0])
	case LinkCmd:
		if len(c.Args) != 2 {
			return errors.New("missing to/from argument pair")
		}

		fmt.Fprintf(w, "link %s %s\n", c.Args[0], c.Args[1])
	case QuitCmd:
		fmt.Fprint(w, "quit")
	default:
		fmt.Fprintf(w, "%s\n", c.Data)
	}

	return nil
}

func add(c *Command, t []string, count int) (*Command, error) {
	// don't count the command token, just the args
	if len(t)-1 != count {
		return nil, fmt.Errorf("missing argument(s) to command %s", t[0])
	}

	switch count {
	// Open, close, buffer take 1 argument
	// t[0] holds the command name
	case 1:
		c.Args = t[1:2]
	// Link takes 2 arguments
	case 2:
		c.Args = t[1:3]
	}

	return c, nil
}
