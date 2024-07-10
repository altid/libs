package control

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/controller"
	"github.com/altid/libs/service/internal/command"
)

type Control struct {
	ctl io.ReadWriteCloser
	cb callback.Callback
	ctx context.Context
	cmds commander.Commander
	cmdlist []*commander.Command
}


func (c *Control) Listen() error {
	c.cmds = &command.Command{
		SendCommand: c.sendCommand,
		CtrlDataCommand: c.ctrlData,
	}

	// TODO: We should select on our context, scanner, and any err back from start
	go c.cb.Start(c)
	scanner := bufio.NewScanner(c.ctl)
	for scanner.Scan() {
		// TODO: Check on our format on the fs for how these come in exactly
		// it may be that we need multiple scan lines at once to handle this correctly
		fmt.Printf("New command: %s\n", scanner.Bytes())
		//c.cb.Handle for input
	}

	if err:= scanner.Err(); err != nil {
		return err
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

	return c.cmds.Exec(cmd)
}

func (c *Control) ctrlData() (b []byte) {
	cw := bytes.NewBuffer(b)
	c.cmds.WriteCommands(c.cmdlist, cw)

	return cw.Bytes()
}

func cmd(c *Control, cmd string) error {
	n, err := fmt.Fprint(c.ctl, cmd)
	if n < len(cmd) {
		return fmt.Errorf("short write on ctl: %s", cmd)
	}
	return err
}
