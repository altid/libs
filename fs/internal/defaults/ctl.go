package defaults

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/altid/libs/fs/internal/command"
	"github.com/altid/libs/fs/internal/reader"
	"github.com/altid/libs/fs/internal/util"
	"github.com/altid/libs/fs/internal/writer"
)

type Control struct {
	event   *os.File
	tabs    *os.File
	cmdlist []*command.Command
	scanner *bufio.Scanner
	done    chan struct{}
	rundir  string
	logdir  string
	doctype string
	tablist []string
	req     chan string
	ctx     context.Context
}

func NewControl(ctx context.Context, r, l, d string, t []string, req chan string) *Control {
	if _, err := os.Stat(r); os.IsNotExist(err) {
		os.MkdirAll(r, 0755)
	}

	ef, _ := os.OpenFile(path.Join(r, "event"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	tf, _ := os.OpenFile(path.Join(r, "tabs"), os.O_CREATE|os.O_WRONLY, 0644)

	return &Control{
		ctx:     ctx,
		done:    make(chan struct{}),
		rundir:  r,
		event:   ef,
		logdir:  l,
		doctype: d,
		tabs:    tf,
		tablist: t,
		req:     req,
	}
}

func (c *Control) Event(eventmsg string) error {
	_, err := c.event.WriteString(eventmsg + "\n")
	return err
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

func (c *Control) Cleanup() {
	if runtime.GOOS == "plan9" {
		glob := path.Join(c.rundir, "*", c.doctype)

		files, err := filepath.Glob(glob)
		if err != nil {
			log.Print(err)
		}

		for _, f := range files {
			command := exec.Command("/bin/unmount", f)
			log.Print(command.Run())
		}
	}

	c.event.Close()
	os.RemoveAll(c.rundir)
}

func (c *Control) CreateBuffer(name, doctype string) error {
	if name == "" {
		return fmt.Errorf("no buffer name given")
	}

	fp := path.Join(c.rundir, name)
	d := path.Join(fp, doctype)

	if _, e := os.Stat(fp); e != nil && !os.IsNotExist(e) {
		return e
	}

	if e := os.MkdirAll(fp, 0755); e != nil {
		return e
	}

	if e := ioutil.WriteFile(d, []byte("Welcome!\n"), 0644); e != nil {
		return e
	}

	if e := c.pushTab(name); e != nil {
		return e
	}

	// If there is no log, we're done otherwise create symlink
	if c.logdir == "none" {
		return nil
	}

	logfile := path.Join(c.logdir, name)

	return util.Symlink(logfile, d)
}

func (c *Control) DeleteBuffer(name, doctype string) error {
	if c.logdir != "none" {
		d := path.Join(c.rundir, name, doctype)
		if e := util.Unlink(d); e != nil {
			return e
		}
	}

	defer os.RemoveAll(path.Join(c.rundir, name))

	return c.popTab(name)
}

func (c *Control) HasBuffer(name, doctype string) bool {
	d := path.Join(c.rundir, name, doctype)

	if _, e := os.Stat(d); os.IsNotExist(e) {
		return false
	}

	return true
}

func (c *Control) Listen() error {
	if e := c.setup(); e != nil {
		return e
	}

	c.Event(path.Join(c.rundir, "ctl"))
	if e := c.listen(); e != nil {
		return e
	}

	return nil
}

func (c *Control) Remove(buffer, filename string) error {
	doc := path.Join(c.rundir, buffer, filename)
	// Don't try to delete that which isn't there
	if _, e := os.Stat(doc); os.IsNotExist(e) {
		return nil
	}

	c.Event(doc)
	return os.Remove(doc)
}

func (c *Control) Notification(buff, from, msg string) error {
	nfile := path.Join(c.rundir, buff, "notification")
	if _, e := os.Stat(path.Dir(nfile)); os.IsNotExist(e) {
		os.MkdirAll(path.Dir(nfile), 0755)
	}

	f, err := os.OpenFile(nfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	c.Event(nfile)
	fmt.Fprintf(f, "%s\n%s\n", from, msg)

	return nil
}

func (c *Control) popTab(tabname string) error {
	for n := range c.tablist {
		if c.tablist[n] == tabname {
			c.tablist = append(c.tablist[:n], c.tablist[n+1:]...)
			return tabs(c)
		}
	}
	return fmt.Errorf("entry not found: %s", tabname)
}

func (c *Control) pushTab(tabname string) error {
	for n := range c.tablist {
		if c.tablist[n] == tabname {
			return fmt.Errorf("entry already exists: %s", tabname)
		}
	}
	c.tablist = append(c.tablist, tabname)

	return tabs(c)
}

func tabs(c *Control) error {
	tabdata := strings.Join(c.tablist, "\n") + "\n"

	if _, e := c.tabs.Seek(0, io.SeekStart); e != nil {
		return e
	}

	if _, e := c.tabs.WriteString(tabdata); e != nil {
		return e
	}

	if e := c.Event(path.Join(c.rundir, "tabs")); e != nil {
		return e
	}

	return c.tabs.Truncate(int64(len(tabdata)))
}

func (c *Control) Errorwriter() (*writer.WriteCloser, error) {
	ep := path.Join(c.rundir, "errors")

	fp, err := os.OpenFile(ep, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {

		return nil, err
	}

	w := writer.New(c.Event, fp, "errors")

	return w, nil
}

func (c *Control) FileWriter(buffer, doctype string) (*writer.WriteCloser, error) {
	doc := path.Join(c.rundir, buffer, doctype)
	if doctype == "feed" {
		fp, err := os.OpenFile(doc, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {

			return nil, err
		}

		w := writer.New(c.Event, fp, doc)

		return w, nil
	}

	// Abuse truncation semantics of Create so we clear any old data
	fp, err := os.Create(doc)
	if err != nil {
		return nil, err
	}

	w := writer.New(c.Event, fp, doc)

	return w, nil
}

func (c *Control) ImageWriter(buffer, resource string) (*writer.WriteCloser, error) {
	os.MkdirAll(path.Dir(path.Join(c.rundir, buffer, "images", resource)), 0755)
	return c.FileWriter(buffer, path.Join("images", resource))
}

func (c *Control) setup() error {
	if e := os.MkdirAll(c.rundir, 0755); e != nil {
		return e
	}

	cfile := path.Join(c.rundir, "ctl")
	c.Event(cfile)

	ctl, err := os.Create(cfile)
	if err != nil {
		return err
	}

	if e := command.PrintCtlFile(c.cmdlist, ctl); e != nil {
		return e
	}

	ctl.Close()

	r, err := reader.Cmd(c.rundir)
	if err != nil {
		return err
	}

	c.scanner = bufio.NewScanner(r)

	return nil
}

func (c *Control) listen() error {
	defer close(c.req)

	for c.scanner.Scan() {
		select {
		case <-c.ctx.Done():
			return nil
		case <-c.done:
			return nil
		default:
			c.req <- c.scanner.Text()
		}
	}

	return nil
}
