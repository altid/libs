package mock

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/altid/libs/fs/input"
	"github.com/altid/libs/fs/internal/command"
	"github.com/altid/libs/fs/internal/util"
	"github.com/altid/libs/fs/internal/writer"
)

type tab struct {
	name    string
	doctype string
	data    []byte
	input   chan string
	cancel  context.CancelFunc
	ctx     context.Context
}

type Control struct {
	ctx     context.Context
	cmdlist []*command.Command
	done    chan struct{}
	reqs    chan string
	cmds    chan string
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

func (t *tab) Read(p []byte) (n int, err error) {
	select {
	case b := <-t.input:
		return copy(p, []byte(b)), nil
	case <-t.ctx.Done():
		return 0, io.EOF
	}
}

func NewControl(ctx context.Context, errs chan error, reqs, cmds chan string, done chan struct{}) *Control {
	return &Control{
		ctx:  ctx,
		err:  errs,
		reqs: reqs,
		cmds: cmds,
		done: done,
	}
}

func (c *Control) Cleanup() {}

func (c *Control) Event(ev string) error { return nil }

func (c *Control) Input(handler input.Handler, buffer string) error {
	for _, t := range c.tabs {
		if t.name == buffer {
			util.RunInput(t.ctx, buffer, handler, t, ioutil.Discard)
			return nil
		}
	}

	return errors.New("Input called on buffer before creation")
}

func (c *Control) SetCommands(cmd ...*command.Command) error {
	for _, comm := range cmd {
		c.cmdlist = append(c.cmdlist, comm)
	}

	sort.Sort(command.CmdList(c.cmdlist))
	return nil
}

func (c *Control) BuildCommand(cmd string) (*command.Command, error) {
	return command.ParseCmd(cmd, c.cmdlist)
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

	if e := command.PrintCtlFile(c.cmdlist, os.Stdout); e != nil {
		return e
	}

	for {
		select {
		case cmd := <-c.reqs:
			t := strings.Split(cmd, ":")
			if t[0] != "input" {
				c.cmds <- cmd
				continue
			}

			if len(t) < 2 {
				continue
			}

			for _, item := range c.tabs {
				if item.name == t[1] {
					item.input <- t[2]
				}
			}
		case err := <-c.err:
			return err
		case <-c.done:
			return nil
		case <-c.ctx.Done():
			return nil
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
			defer close(c.tabs[n].input)
			c.tabs[n].cancel()
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

	ctx, cancel := context.WithCancel(c.ctx)

	t := &tab{
		input:   make(chan string),
		name:    tabname,
		doctype: doctype,
		ctx:     ctx,
		cancel:  cancel,
	}

	c.tabs = append(c.tabs, t)
	return nil
}

func (c *Control) ErrorWriter() (*writer.WriteCloser, error) {
	w := writer.New(c.Event, os.Stdout, "errors")

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
