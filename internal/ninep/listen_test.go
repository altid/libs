package ninep

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/altid/server/client"
	"github.com/altid/server/command"
	"github.com/altid/server/internal/routes"
)

func TestListenCommands(t *testing.T) {
	cmds := make(chan *command.Command)
	defer close(cmds)

	fp, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}

	cm := &client.Manager{}
	s := &service{
		debug:   log.Printf,
		feed:    routes.NewFeed(),
		client:  cm,
		command: cmds,
	}

	go s.listenCommands(fp)
	c := cm.Client(0)

	cmds <- &command.Command{
		UUID:    c.UUID,
		CmdType: command.OpenCmd,
		From:    "none",
		Args:    []string{"test"},
	}

	time.Sleep(time.Millisecond * 100)
	if c.Current() != "test" {
		t.Error("unable to set buffer with open")
	}

	cmds <- &command.Command{
		UUID:    c.UUID,
		CmdType: command.OpenCmd,
		From:    "test",
		Args:    []string{"test2"},
	}

	cmds <- &command.Command{
		UUID:    c.UUID,
		CmdType: command.BufferCmd,
		From:    "test2",
		Args:    []string{"test"},
	}

	time.Sleep(time.Millisecond * 100)
	if c.Current() != "test" {
		t.Error("unable to change buffer")
	}

	if c.History()[0] != "test" && c.History()[1] != "test2" {
		t.Error("history incorrect")
	}

	cmds <- &command.Command{
		UUID:    c.UUID,
		CmdType: command.CloseCmd,
		From:    "test",
		Args:    []string{c.Current()},
	}

	time.Sleep(time.Millisecond * 100)
	if c.Current() != "test2" {
		t.Error("unable to change back to old buffer after close")
	}
}
