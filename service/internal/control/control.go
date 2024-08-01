package control

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/altid/libs/markup"
	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/controller"
	"github.com/altid/libs/service/internal/command"
)

type Control struct {
	l sync.Mutex
	done chan bool
	errs chan error
	cmds chan *commander.Command
	ctl *os.File
	cb callback.Callback
	ctx context.Context
	commander commander.Commander
	cmdlist []*commander.Command
}

func (c *Control) Listen() error {
	defer c.ctl.Close()

	c.commander = &command.Command{
		SendCommand: c.sendCommand,
		CtrlDataCommand: c.ctrlData,
	}

	go c.ReadCommands()
	go func(c *Control) {
		c.cb.Start(c)
		c.done <- true
	}(c)

	go func(c *Control) {
		for cmd := range c.cmds {
			if cmd.Name == "input" {
				l := markup.NewLexer(cmd.ArgBytes())
				c.cb.Handle(cmd.From, l)
			} else {
				fmt.Printf("Incoming command: %s, %v\n", cmd.Name, cmd.Args)
			}
		}
	}(c)

	select {
	case e := <-c.errs:
		return e
	case <- c.done:
		return nil
	}
	return nil
}

func (c *Control) SetCallbacks(cb callback.Callback) {
	c.cb = cb
}

func (c *Control) SetCommands(cmds []*commander.Command) {
	c.cmdlist = append(c.cmdlist, cmds...)
	sort.Sort(commander.CmdList(c.cmdlist))
}

func (c *Control) CreateBuffer(name string) error {
	return cmd(c, "create " + name)
}

func (c *Control) DeleteBuffer(name string) error {
	return cmd(c, "delete " + name)
}

// TODO: Research usage
func (c *Control) Remove(string, string) error {
	return nil
}

// TODO: Research usage
func (c *Control) Notification(string, string, string) error {
	return nil
}

func (c *Control) ErrorWriter() (controller.WriteCloser, error) {
	return newPrefix(c, errorFmt)
}

func (c *Control) StatusWriter(buffer string) (controller.WriteCloser, error) {
	return newPrefix(c, statusFmt, buffer)
}

func (c *Control) SideWriter(buffer string) (controller.WriteCloser, error) {
	return newPrefix(c, sideFmt, buffer)
}

func (c *Control) NavWriter(buffer string) (controller.WriteCloser, error) {
	return newPrefix(c, navFmt, buffer)
}

func (c *Control) TitleWriter(buffer string) (controller.WriteCloser, error) {
	return newPrefix(c, titleFmt, buffer)
}

func (c *Control) ImageWriter (buffer string, name string) (controller.WriteCloser, error) {
	return newPrefix(c, imageFmt, buffer, name)
}

func (c *Control) MainWriter(buffer string) (controller.WriteCloser, error) {
	return newPrefix(c, mainFmt, buffer)
}

func (c *Control) FeedWriter(buffer string) (controller.WriteCloser, error) {
	return newPrefix(c, feedFmt, buffer)
}

// TODO: We don't really need this anymore
func (c *Control) HasBuffer(string) bool {
	return false
}

func (c *Control) sendCommand(cmd *commander.Command) error {
	switch cmd.Name {
	case "shutdown":
		c.ctx.Done()
		return nil
	case "reload":
	case "restart":
		return nil
	}

	return c.commander.Exec(cmd)
}

func (c *Control) ctrlData() (b []byte) {
	c.l.Lock()
	defer c.l.Unlock()
	cw := bytes.NewBuffer(b)
	c.commander.WriteCommands(c.cmdlist, cw)

	return cw.Bytes()
}

func cmd(c *Control, cmd string) error {
	c.l.Lock()
	defer c.l.Unlock()
	_, err := fmt.Fprintf(c.ctl, "%s\x00", cmd)
	return err
}
