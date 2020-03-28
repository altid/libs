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
	ReloadCmd
	RestartCmd
	BufferCmd
	CloseCmd
	LinkCmd
	QuitCmd
)

// There are more possible commands, but these are the only ones
// the server itself intercepts in any meaningful way
var items = map[string]CmdType{
	"open":    OpenCmd,
	"close":   CloseCmd,
	"buffer":  BufferCmd,
	"restart": RestartCmd,
	"link":    LinkCmd,
	"quit":    QuitCmd,
	"reload":  ReloadCmd,
	"other":   OtherCmd,
}

// Command is a ctl message
type Command struct {
	// The sending client ID
	UUID    uint32
	CmdType CmdType
	// The buffer that the command came from
	From string
	// The arguments to the command
	Args string
	// The raw command bytes
	Data []byte
}

// New - helper func with varargs
// If there are 2 args, the first is assumed to be the "From" path, the second "Args"
// If only one, it is set as "Args"
//
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
}

// WriteOut will write the command correctly to the Writer
// If a command is poorly formatted, it will return an error
func (c *Command) WriteOut(w io.Writer) error {
	// check the args count before writing
	switch c.CmdType {
	case OpenCmd:
		fmt.Fprintf(w, "open \"%s\" \"%s\"\n", c.From, c.Args)
	case RestartCmd:
		return errors.New("restart command not writable. Did you mean service restart?")
	case ReloadCmd:
		return errors.New("reload command not writable. Did you mean service reload?")
	case BufferCmd:
		fmt.Fprintf(w, "buffer \"%s\" \"%s\"\n", c.From, c.Args)
	case CloseCmd:
		fmt.Fprintf(w, "close \"%s\"\n", c.Args)
	case LinkCmd:
		fmt.Fprintf(w, "link \"%s\" \"%s\"\n", c.From, c.Args)
	case QuitCmd:
		fmt.Fprintln(w, "quit")
	default:
		fmt.Fprintf(w, "%s\n", c.Data)
	}

	return nil
}
