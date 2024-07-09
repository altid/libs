package control

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/controller"
)

type Control struct {
	ctl io.ReadWriteCloser
	cb callback.Callback
	ctx context.Context
}

func (c *Control) Listen(ctl func(*commander.Command)) error {
	// TODO: We should select on both our context, and the scanner
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

func (c *Control) WithCallbacks(cb callback.Callback) {
	c.cb = cb
}

func (c *Control) WithContext(ctx context.Context) {
	c.ctx = ctx
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

func cmd(c *Control, cmd string) error {
	n, err := fmt.Fprint(c.ctl, cmd)
	if n < len(cmd) {
		return fmt.Errorf("short write on ctl: %s", cmd)
	}
	return err
}
