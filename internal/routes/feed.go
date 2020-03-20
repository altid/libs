package routes

import (
	"errors"
	"io"
	"os"
	"path"
	"sync"

	"github.com/altid/server/client"
	"github.com/altid/server/command"
	"github.com/altid/server/files"
)

type FeedHandler struct {
	clients  *client.Manager
	commands chan *command.Command
	feeds    map[uint32]chan struct{}
	sync.Mutex
}

func NewFeed(clients *client.Manager, commands chan *command.Command, fp *os.File) *FeedHandler {
	fh := &FeedHandler{
		clients:  clients,
		commands: commands,
		feeds:    make(map[uint32]chan struct{}),
	}

	go fh.listenCommands(fp)
	return fh
}

func (fh *FeedHandler) Normal(msg *files.Message) (interface{}, error) {
	done := make(chan struct{})
	event := make(chan struct{})
	f := &feed{
		event: event,
		path:  path.Join(msg.Service, msg.Buffer, "feed"),
		buff:  path.Join(msg.Service, msg.Buffer),
		done:  done,
	}

	// Feed to match this specific client
	fh.feeds[msg.UUID] = event

	return f, nil

}

func (*FeedHandler) Stat(msg *files.Message) (os.FileInfo, error) {
	return os.Lstat(path.Join(msg.Service, msg.Buffer, "feed"))
}

// Commands only effect feed file
func (fh *FeedHandler) listenCommands(fp *os.File) {
	defer fp.Close()
	for cmd := range fh.commands {
		switch cmd.CmdType {
		case command.OtherCmd:
			cmd.WriteOut(fp)
		case command.OpenCmd:
			c := fh.clients.Client(client.UUID(cmd.UUID))
			if c != nil {
				fh.switchBuffer(c, cmd)
				cmd.WriteOut(fp)
			}
		case command.BufferCmd:
			c := fh.clients.Client(client.UUID(cmd.UUID))
			if c != nil {
				fh.switchBuffer(c, cmd)
			}
		case command.CloseCmd:
			c := fh.clients.Client(client.UUID(cmd.UUID))
			if c != nil {
				// Pop back to the last buffer
				history := c.History()
				if len(history) < 1 {
					c.SetBuffer("none")
				} else {
					c.SetBuffer(history[len(history)-1])
				}
				cmd.WriteOut(fp)
				close(fh.feeds[cmd.UUID])
				fh.feeds[cmd.UUID] = make(chan struct{})
			}
			/* mostly send cmds down for now
			ReloadCmd
			CloseCmd
			LinkCmd
			QuitCmd
			*/
		}
	}
}

func (fh *FeedHandler) switchBuffer(c *client.Client, cmd *command.Command) {
	c.SetBuffer(cmd.Args[0])
	close(fh.feeds[cmd.UUID])
	fh.feeds[cmd.UUID] = make(chan struct{})
}

// feed files are special in that they're blocking
type feed struct {
	event   chan struct{}
	tailing bool
	path    string
	buff    string
	done    chan struct{}
}

func (f *feed) ReadAt(p []byte, off int64) (n int, err error) {
	fp, err := os.Open(f.path)
	if err != nil {
		return 0, err
	}

	defer fp.Close()

	if !f.tailing {
		n, err = fp.ReadAt(p, off)
		if err != nil && err != io.EOF {
			return
		}

		if err == io.EOF {
			f.tailing = true
		}
		return n, nil
	}

	for range f.event {
		n, err = fp.ReadAt(p, off)
		if err == io.EOF {
			return n, nil
		}

		return
	}

	return 0, io.EOF
}

func (f *feed) WriteAt(p []byte, off int64) (int, error) {
	return 0, errors.New("writing to feed files is currently unsupported")
}

func (f *feed) Close() error { return nil }
