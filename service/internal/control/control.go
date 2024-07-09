package control

import (
	"fmt"
	"io"

	"github.com/altid/libs/service/commander"
)

type Control struct {
	ctl io.ReadWriteCloser
}

func (c *Control) Listen(input func([]byte), ctl func(*commander.Command)) {
	//read on ctl, call the appropriate function
}

func (c *Control) CreateBuffer(name string) error {
	return cmd(c, "create " + name)
}

func (c *Control) DeleteBuffer(name string) error {
	return cmd(c, "delete " + name)
}

// TODO: Was this ever used?
func (c *Control) Remove(string, string) error {
	return nil
}

func (c *Control) Notification(string, string, string) error {
	//return cmd(c, fmt.Sprintf("notify %s\n%s", buffer, data))
	return nil
}

func (c *Control) WriteError(data string) error {
	return cmd(c, fmt.Sprintf("error %s\n", data))
}

func (c *Control) WriteStatus(buffer string, data string) error {
	return cmd(c, fmt.Sprintf("status %s\n%s", buffer, data))
}

func (c *Control) WriteAside(buffer string, data string) error {
	return cmd(c, fmt.Sprintf("aside %s\n%s", buffer, data))
}

func (c *Control) WriteNav(buffer string, data string) error {
	return cmd(c, fmt.Sprintf("navi %s\n%s", buffer, data))
}

func (c *Control) WriteTitle(buffer string, data string) error {
	return cmd(c, fmt.Sprintf("title %s\n%s", buffer, data))
}

// TODO: Implement images
func (c *Control) WriteImage(string, string, io.ReadCloser) error {
	return nil
}

func (c *Control) WriteMain(buffer string, data io.ReadCloser) error {
	return cmd(c, fmt.Sprintf("main %s\n%s", buffer, data))
}

func (c *Control) WriteFeed(buffer string, data io.ReadCloser) error {
	return cmd(c, fmt.Sprintf("feed %s\n%s", buffer, data))
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
