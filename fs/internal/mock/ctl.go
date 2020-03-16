package mock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/altid/libs/fs/internal/command"
	"github.com/altid/libs/fs/internal/writer"
)

type tab struct {
	name    string
	doctype string
	data    []byte
}

type Control struct {
	cmdlist []*command.Command
	reqs    chan string
	cmds    chan string
	done    chan struct{}
	err     chan error
	tabs    []*tab
}

func (t *tab) Write(p []byte) (n int, err error) {
	n = copy(p, t.data)
	return
}

func (t *tab) Close() error {
	return nil
}

func NewControl(errs chan error, reqs, cmds chan string, done chan struct{}) *Control {
	return &Control{
		err:  errs,
		reqs: reqs,
		cmds: cmds,
		done: done,
	}
}

func (c *Control) Cleanup() {}

func (c *Control) Event(ev string) error { return nil }

func (c *Control) SetCommands(cmd ...*command.Command) error {
	for _, comm := range cmd {
		c.cmdlist = append(c.cmdlist, comm)
	}

	sort.Sort(command.CmdList(c.cmdlist))
	return nil
}

func (c *Control) BuildCommand(cmd string) (*command.Command, error) {
	return command.BuildFrom(cmd, c.cmdlist)
}

func (c *Control) CreateBuffer(name, doctype string) error {
	return c.pushTab(name, doctype)
}

func (c *Control) DeleteBuffer(name, doctype string) error {
	return c.popTab(name)
}

func (c *Control) HasBuffer(name, doctype string) bool {
	for _, i := range c.tabs {
		if i.name == name {
			return true
		}
	}

	return false
}

func (c *Control) Remove(name, doctype string) error {
	return c.popTab(name)
}

func (c *Control) Listen() error {
	defer close(c.err)
	defer close(c.done)

	if e := command.PrintCtlFile(c.cmdlist, os.Stdout); e != nil {
		return e
	}

	for {
		select {
		case cmd := <-c.reqs:
			if cmd == "quit" {
				return nil
			}
			c.cmds <- cmd
		case err := <-c.err:
			return err
		}
	}
}

func (c *Control) Start() (context.Context, error) {
	return nil, errors.New("please use listen for testing")
}

func (c *Control) Notification(string, string, string) error {
	return nil
}

func (c *Control) popTab(tabname string) error {
	for n := range c.tabs {
		if c.tabs[n].name == tabname {
			c.tabs = append(c.tabs[:n], c.tabs[n+1:]...)
			return nil
		}
	}

	return fmt.Errorf("entry not found: %s", tabname)
}

func (c *Control) pushTab(tabname, doctype string) error {
	for n := range c.tabs {
		if c.tabs[n].name == tabname {
			return fmt.Errorf("entry already exists: %s", tabname)
		}
	}

	t := &tab{
		name:    tabname,
		doctype: doctype,
	}

	c.tabs = append(c.tabs, t)
	return nil
}

func (c *Control) Errorwriter() (*writer.WriteCloser, error) {
	w := writer.New(c.Event, &tab{}, "errors")

	return w, nil
}
func (c *Control) FileWriter(buffer, doctype string) (*writer.WriteCloser, error) {
	tab, err := c.findTab(buffer)
	if err != nil {
		c.err <- err
		return nil, err
	}

	w := writer.New(c.Event, tab, buffer)

	return w, nil
}
func (c *Control) ImageWriter(buffer, resource string) (*writer.WriteCloser, error) {
	tab, err := c.findTab(buffer)
	if err != nil {
		c.err <- err
		return nil, err
	}

	w := writer.New(c.Event, tab, buffer)

	return w, nil
}

func (c *Control) findTab(buffer string) (*tab, error) {
	for _, tab := range c.tabs {
		if tab.name == buffer {
			return tab, nil
		}
	}

	return nil, errors.New("No such tab")
}
