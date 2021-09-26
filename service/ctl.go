package service 

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	"github.com/altid/libs/service/input"
	"github.com/altid/libs/service/internal/command"
	"github.com/altid/libs/service/internal/control"
	"github.com/altid/libs/service/internal/store"
)

// Controller is our main type for controlling a session
type Controller interface {
	Manager
	input.Handler
}

// Listener instantiates a network server, listening for incoming clients
type Listener interface {
	Start() error
	// This almost certainly will change
	Quit()
}

// Manager wraps all interactions with the `ctl` file
type Manager interface {
	Run(*Control, *Command) error
	Quit()
}

type runner interface {
	Cleanup()
	SetCommands(cmd ...*command.Command) error
	BuildCommand(string) (*command.Command, error)
	Event(string) error
	Input(input.Handler, string) error
	CreateBuffer(string, string) error
	DeleteBuffer(string, string) error
	HasBuffer(string, string) bool
	Remove(string, string) error
	Notification(string, string, string) error
}

type writercloser interface {
	Errorwriter() (*store.WriteCloser, error)
	FileWriter(string, string) (*store.WriteCloser, error)
	ImageWriter(string, string) (*store.WriteCloser, error)
}

// Control type can be used to manage a running ctl file session
type Control struct {
	ctx    context.Context
	cancel context.CancelFunc
	req    chan string
	ctl    Manager
	input  input.Handler
	done   chan struct{}
	run    runner
	host   Listener	
	write  writercloser
	debug  func(ctlMsg, ...interface{})
	sync.Mutex
}

type ctlMsg int

const (
	ctlError ctlMsg = iota
	ctlEvent
	ctlCleanup
	ctlCreate
	ctlDelete
	ctlRemove
	ctlStart
	ctlNotify
	ctlDefault
)

// New sets up a ready-to-listen ctl file
// logdir is the directory to store copies of the contents of files created; specifically doctype. Logging any other type of data is left to implementation details, but is considered poor form for Altid's design.
// mtpt is the directory to create the file system in
// service is the subdirectory inside mtpt for the runtime fs
// This will return an error if a ctl file exists at the given directory, or if doctype is invalid.
func New(ctl interface{}, listener Listeneer, logdir, mtpt, service, doctype string, debug bool) (*Control, error) {
	if doctype != "document" && doctype != "feed" {
		return nil, fmt.Errorf("unknown doctype: %s", doctype)
	}

	manager, ok := ctl.(Manager)
	if !ok {
		return nil, errors.New("ctl missing Run/Quit method(s)")
	}

	var tab []string

	req := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	rtc := control.New(ctx, rundir, logdir, doctype, tab, req)

	c := &Control{
		ctx:    ctx,
		cancel: cancel,
		req:    req,
		done:   make(chan struct{}),
		run:    rtc,
		host:	listener,
		ctl:    manager,
		write:  rtc,
		debug:  func(ctlMsg, ...interface{}) {},
	}

	if debug {
		c.debug = ctlLogger
	}

	// Some services don't use input
	input, ok := ctl.(input.Handler)
	if ok {
		c.input = input
	}

	cmdlist := command.DefaultCommands
	cmdlist = append(cmdlist, &command.Command{
		Name:        service,
		Args:        []string{"<quit|restart|reload>"},
		Heading:     command.ServiceGroup,
		Description: "Control the lifecycle of a service",
	})

	rtc.SetCommands(cmdlist...)
	return c, nil
}

// Input starts an input file in the named buffer
// If Input is called before a buffer is created, an error will be returned
// If the Controller sent to Input does not implement a Handler this will panic
func (c *Control) Input(buffer string) error {
	return c.run.Input(c.input, buffer)
}

// Event appends the given string to the events file of Control's working directory.
// Strings cannot contain newlines, tabs, spaces, or control characters.
// Returns "$service: invalid event $eventmsg" or nil.
func (c *Control) Event(eventmsg string) error {
	c.debug(ctlEvent, eventmsg)
	return c.run.Event(eventmsg)
}

// Cleanup removes created symlinks and removes the main dir
// On plan9, it unbinds any file named 	"document" or "feed", prior to removing the directory itself.
func (c *Control) Cleanup() {
	c.debug(ctlCleanup)
	c.run.Cleanup()

}

// CreateBuffer creates a buffer of given name, as well as symlinking your file as follows:
// `os.Symlink(path.Join(logdir, name), path.Join(rundir, name, doctype))`
// This logged file will persist across reboots
// Calling CreateBuffer on a directory that already exists will return nil
func (c *Control) CreateBuffer(name, doctype string) error {
	c.debug(ctlCreate, name, doctype)
	return c.run.CreateBuffer(name, doctype)
}

// DeleteBuffer unlinks a document/buffer, and cleanly removes the directory
// Will return an error if it's unable to unlink on plan9, or if the remove fails.
func (c *Control) DeleteBuffer(name, doctype string) error {
	c.debug(ctlDelete, name)
	return c.run.DeleteBuffer(name, doctype)
}

// HasBuffer returns whether or not a buffer is present in the current control session
func (c *Control) HasBuffer(name, doctype string) bool {
	return c.run.HasBuffer(name, doctype)
}

// Remove removes a buffer from the runtime dir. If the buffer doesn't exist, this is a no-op
func (c *Control) Remove(buffer, filename string) error {
	c.debug(ctlRemove, buffer, filename)
	return c.run.Remove(buffer, filename)
}

// Listen starts a network listener for incoming clients
func (c *Control) Listen() error {
	go dispatch(c)

	c.debug(ctlStart, "listen")
	return c.run.Listen()
}

// SetCommands allows services to add additional commands
// Any client command encountered which matches will send
// The resulting command down to RunCommand
// Commands must include at least a name and a heading
// Running SetCommands after calling Start or Listen will have no effect
func (c *Control) SetCommands(cmd ...*Command) error {
	for _, comm := range cmd {
		if comm.Name == "" {
			return errors.New("command requires Name")
		}

		switch comm.Heading {
		case DefaultGroup, MediaGroup, ActionGroup:
			continue
		default:
			return errors.New("Unsupported or nil Heading set")
		}
	}

	if e := setCommands(c.run, cmd...); e != nil {
		return e
	}

	return nil
}

// Notification appends the content of msg to a buffers notification file
// Any errors encountered during file opening/creation will be returned
// The canonical form of notification can be found in the markup libs' Notification type,
// And the output of the Parse() method can be used directly here
// For example
//     ntfy, err := markup.NewNotifier(buff, from, msg)
//     if err != nil {
//         log.Fatal(err)
//     }
//     fs.Notification(ntfy.Parse())
func (c *Control) Notification(buff, from, msg string) error {
	c.debug(ctlNotify, buff, from, msg)
	return c.run.Notification(buff, from, msg)
}

// ErrorWriter returns a WriteCloser attached to a services' errors file
func (c *Control) ErrorWriter() (*store.WriteCloser, error) {
	return c.write.Errorwriter()
}

// StatusWriter returns a WriteCloser attached to a buffers status file, which will as well send the correct event to the events file
func (c *Control) StatusWriter(buffer string) (*store.WriteCloser, error) {
	return c.write.FileWriter(buffer, "status")
}

// SideWriter returns a WriteCloser attached to a buffers `aside` file, which will as well send the correct event to the events file
func (c *Control) SideWriter(buffer string) (*store.WriteCloser, error) {
	return c.write.FileWriter(buffer, "aside")
}

// NavWriter returns a WriteCloser attached to a buffers nav file, which will as well send the correct event to the events file
func (c *Control) NavWriter(buffer string) (*store.WriteCloser, error) {
	return c.write.FileWriter(buffer, "navi")
}

// TitleWriter returns a WriteCloser attached to a buffers title file, which will as well send the correct event to the events file
func (c *Control) TitleWriter(buffer string) (*store.WriteCloser, error) {
	return c.write.FileWriter(buffer, "title")
}

// ImageWriter returns a WriteCloser attached to a named file in the buffers' image directory
func (c *Control) ImageWriter(buffer, resource string) (*store.WriteCloser, error) {
	return c.write.ImageWriter(buffer, resource)
}

// MainWriter returns a WriteCloser attached to a buffers feed/document function to set the contents of a given buffers' document or feed file, which will as well send the correct event to the events file
func (c *Control) MainWriter(buffer, doctype string) (*store.WriteCloser, error) {
	return c.write.FileWriter(buffer, doctype)
}

// Context returns the underlying context of the service
// This will be closed after Quit() is called
func (c *Control) Context() context.Context {
	return c.ctx
}

func dispatch(c *Control) {
	// If close is requested on a file which is currently being opened, cancel open request
	// If open is requested on file which already exists, no-op
	ew, err := c.write.Errorwriter()
	if err != nil {
		log.Fatal(err)
	}

	defer ew.Close()

	for {
		select {
		case line := <-c.req:
			cmd, err := c.run.BuildCommand(line)
			if err != nil {
				fmt.Fprintf(ew, "%v\n", err)
				continue
			}

			real := translate(cmd)

			if real.Heading == ServiceGroup {
				serviceCommand(c, real, ew)
				continue
			}

			c.ctl.Run(c, real)
		case <-c.done:
			return
		case <-c.ctx.Done():
			return
		}
	}
}

func serviceCommand(c *Control, cmd *Command, ew *store.WriteCloser) {
	switch cmd.Args[0] {
	case "quit":
		defer c.cancel()
		// Close our local listeners, then
		close(c.done)
		c.ctl.Quit()
	// Eventually we may want to access these
	case "reload", "restart":
		c.ctl.Run(c, cmd)
	default:
		fmt.Fprintf(ew, "unsupported command: %s", cmd.Args[0])
	}
}

func translate(cmd *command.Command) *Command {
	return &Command{
		Name:        cmd.Name,
		Description: cmd.Description,
		From:        cmd.From,
		Args:        cmd.Args,
		Alias:       cmd.Alias,
		Heading:     ComGroup(cmd.Heading),
	}
}

func ctlLogger(msg ctlMsg, args ...interface{}) {
	l := log.New(os.Stdout, "ctl ", 0)

	switch msg {
	case ctlError:
		l.Printf("error: buffer=\"%s\" err=\"%v\"\n", args[0], args[1])
	case ctlEvent:
		l.Printf("event: msg=\"%s\"\n", args[0])
	case ctlCleanup:
		l.Println("cleanup: ending")
	case ctlCreate:
		l.Printf("create: buffer=\"%s\" doctype=%s\n", args[0], args[1])
	case ctlDelete:
		l.Printf("delete: buffer=\"%s\"\n", args[0])
	case ctlRemove:
		l.Printf("remove: buffer=\"%s\", filename=\"%s\"\n", args[0], args[1])
	case ctlStart:
		l.Printf("%s: starting\n", args[0])
	case ctlNotify:
		l.Printf("notify: buffer=\"%s\" from=\"%s\" msg=\"%s\"\n", args[0], args[1], args[2])
	case ctlDefault:
		cmd := args[0].(*Command)
		switch cmd.Heading {
		case ActionGroup:
			l.Printf("%s group=\"action\" arguments=\"%s\"\n", cmd.Name, cmd.Args)
		case MediaGroup:
			l.Printf("%s group=\"media\" arguments=\"%s\"\n", cmd.Name, cmd.Args)
		}
	}
}
