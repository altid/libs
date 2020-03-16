package defaults

import (
	"bufio"
	"context"
	"fmt"
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
	cmdlist []*command.Command
	rundir  string
	logdir  string
	doctype string
	tabs    []string
	req     chan string
	done    chan struct{}
}

type controlrunner struct {
	ctx     context.Context
	scanner *bufio.Scanner
	done    chan struct{}
	req     chan string
	cancel  context.CancelFunc
}

func NewControl(r, l, d string, t []string, req chan string, done chan struct{}) *Control {
	return &Control{
		rundir:  r,
		logdir:  l,
		doctype: d,
		tabs:    t,
		req:     req,
		done:    done,
	}
}

func (c *Control) Event(eventmsg string) error {
	file := path.Join(c.rundir, "event")
	if _, err := os.Stat(path.Dir(file)); os.IsNotExist(err) {
		os.MkdirAll(path.Dir(file), 0755)
	}

	if e := validateString(file); e != nil {
		sname := path.Base(c.rundir)
		return fmt.Errorf("%s: invalid event %s", sname, eventmsg)
	}

	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer f.Close()
	f.WriteString(eventmsg + "\n")

	return nil
}

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
	ctx, cancel := context.WithCancel(context.Background())

	cr, err := newControlRunner(ctx, cancel, c)
	if err != nil {
		return err
	}

	cr.listen()

	return nil
}

func (c *Control) Start() (context.Context, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cr, err := newControlRunner(ctx, cancel, c)
	if err != nil {
		return nil, err
	}

	go cr.listen()

	return ctx, nil
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
	for n := range c.tabs {
		if c.tabs[n] == tabname {
			c.tabs = append(c.tabs[:n], c.tabs[n+1:]...)
			return tabs(c)
		}
	}
	return fmt.Errorf("entry not found: %s", tabname)
}

func (c *Control) pushTab(tabname string) error {
	for n := range c.tabs {
		if c.tabs[n] == tabname {
			return fmt.Errorf("entry already exists: %s", tabname)
		}
	}
	c.tabs = append(c.tabs, tabname)

	return tabs(c)
}

func tabs(c *Control) error {
	// Create truncates and opens file in a single step, utilize this.
	file := path.Join(c.rundir, "tabs")

	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()
	f.WriteString(strings.Join(c.tabs, "\n") + "\n")
	c.Event(file)

	return nil
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

func newControlRunner(ctx context.Context, cancel context.CancelFunc, c *Control) (*controlrunner, error) {
	if e := os.MkdirAll(c.rundir, 0755); e != nil {
		return nil, e
	}

	cfile := path.Join(c.rundir, "ctl")
	c.Event(cfile)

	ctl, err := os.Create(cfile)
	if err != nil {
		return nil, err
	}

	if e := command.PrintCtlFile(c.cmdlist, ctl); e != nil {
		return nil, e
	}

	ctl.Close()

	r, err := reader.New(cfile)
	if err != nil {
		return nil, err
	}

	cr := &controlrunner{
		scanner: bufio.NewScanner(r),
		req:     c.req,
		ctx:     ctx,
		cancel:  cancel,
		done:    c.done,
	}

	return cr, nil
}

func (c *controlrunner) listen() {
	defer close(c.req)

	for c.scanner.Scan() {
		select {
		case <-c.ctx.Done():
			break
		default:
			line := c.scanner.Text()
			if line == "quit" {
				c.cancel()
				close(c.done)
				break
			}

			c.req <- line
		}
	}
}

func validateString(path string) error {
	if _, e := os.Stat(path); e != nil {
		return e
	}

	return nil
}
