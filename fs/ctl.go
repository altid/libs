package fs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/altid/libs/fs/internal/command"
	"github.com/altid/libs/fs/internal/defaults"
	"github.com/altid/libs/fs/internal/mock"
	"github.com/altid/libs/fs/internal/writer"
)

// Controller is our main type for controlling a session
// Open is called when a control message starting with 'open' or 'join' is written to the ctl file
// Close is called when a control message starting with 'close or 'part' is written to the ctl file
// Link is called when a control message starting with 'link' is written to the ctl file
// Default is called when any other control message is written to the ctl file.
// If a client attempts to write an invalid control message, it will return a generic error
// When Open is called, a file will be created with a path of `mountpoint/msg/document (or feed)`, containing initially a file named what you've set doctype to.. Calls to open are expected to populate that file, as well as create any supplementary files needed, such as title, aside, status, input, etc
// When Link is called, the content of the current buffer is expected to change, and the name of the current tab will be removed, replaced with msg
// The main document or feed file is also symlinked into the given log directory, under service/msgs, so for example, an expensive parse would only have to be completed once for a given request, even across separate runs; or a chat log could have history from previous sessions accessible.
// The message provided to all three functions is all of the message, less 'open', 'join', 'close', or 'part'.
type Controller interface {
	Open(c *Control, msg string) error
	Close(c *Control, msg string) error
	Link(c *Control, from, msg string) error
	Default(c *Control, cmd *Command) error
	Restart(c *Control) error
	Refresh(c *Control) error
	Quit()
}

type runner interface {
	Cleanup()
	SetCommands(cmd ...*command.Command) error
	BuildCommand(string) (*command.Command, error)
	Event(string) error
	CreateBuffer(string, string) error
	DeleteBuffer(string, string) error
	HasBuffer(string, string) bool
	Listen() error
	Remove(string, string) error
	Notification(string, string, string) error
}

type writercloser interface {
	Errorwriter() (*writer.WriteCloser, error)
	FileWriter(string, string) (*writer.WriteCloser, error)
	ImageWriter(string, string) (*writer.WriteCloser, error)
}

// Control type can be used to manage a running ctl file session
type Control struct {
	ctx   context.Context
	req   chan string
	ctl   Controller
	done  chan struct{}
	run   runner
	write writercloser
	debug func(ctlMsg, ...interface{})
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

// MockCtlFile returns a type that can be used for testing services
// it will track in-memory and behave like a file-backed control
// It will wait for messages on reqs which act as ctl messages
// By default it writes to Stdout + Stderr with each WriteCloser
// If debug is true, all logs will be written to stdout
func MockCtlFile(ctx context.Context, ctl Controller, reqs chan string, service string, debug bool) (*Control, error) {

	done := make(chan struct{})
	cmds := make(chan string)
	errs := make(chan error)
	t := mock.NewControl(ctx, errs, reqs, cmds, done)

	c := &Control{
		ctx:   ctx,
		ctl:   ctl,
		done:  make(chan struct{}),
		req:   cmds,
		run:   t,
		write: t,
		debug: func(ctlMsg, ...interface{}) {},
	}

	if debug {
		c.debug = ctlLogger
	}

	cmdlist := command.DefaultCommands
	cmdlist = append(cmdlist, &command.Command{
		Name:        service,
		Args:        []string{"<quit|restart|reload>"},
		Heading:     command.ServiceGroup,
		Description: "Control the lifecycle of a service",
	})

	t.SetCommands(cmdlist...)

	return c, nil
}

// CreateCtlFile sets up a ready-to-listen ctl file
// logdir is the directory to store copies of the contents of files created; specifically doctype. Logging any other type of data is left to implementation details, but is considered poor form for Altid's design.
// mtpt is the directory to create the file system in
// service is the subdirectory inside mtpt for the runtime fs
// This will return an error if a ctl file exists at the given directory, or if doctype is invalid.
func CreateCtlFile(ctx context.Context, ctl Controller, logdir, mtpt, service, doctype string, debug bool) (*Control, error) {
	if doctype != "document" && doctype != "feed" {
		return nil, fmt.Errorf("unknown doctype: %s", doctype)
	}

	rundir := path.Join(mtpt, service)

	if _, e := os.Stat(path.Join(rundir, "ctl")); os.IsNotExist(e) {
		var tab []string
		req := make(chan string)
		rtc := defaults.NewControl(ctx, rundir, logdir, doctype, tab, req)

		c := &Control{
			ctx:   ctx,
			req:   req,
			done:  make(chan struct{}),
			run:   rtc,
			ctl:   ctl,
			write: rtc,
			debug: func(ctlMsg, ...interface{}) {},
		}

		if debug {
			c.debug = ctlLogger
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

	return nil, fmt.Errorf("Control file already exist at %s", rundir)
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

// Listen creates a file named "ctl" inside RunDirectory, after making sure the directory exists
// Any text written to the ctl file will be parsed, line by line.
// Messages handled internally are as follows: open (or join), close (or part), and quit, which causes Listen() to return.
// This will return an error if we're unable to create the ctlfile itself, and will log any error relating to control messages.
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
func (c *Control) ErrorWriter() (*writer.WriteCloser, error) {
	return c.write.Errorwriter()
}

// StatusWriter returns a WriteCloser attached to a buffers status file, which will as well send the correct event to the events file
func (c *Control) StatusWriter(buffer string) (*writer.WriteCloser, error) {
	return c.write.FileWriter(buffer, "status")
}

// SideWriter returns a WriteCloser attached to a buffers `aside` file, which will as well send the correct event to the events file
func (c *Control) SideWriter(buffer string) (*writer.WriteCloser, error) {
	return c.write.FileWriter(buffer, "aside")
}

// NavWriter returns a WriteCloser attached to a buffers nav file, which will as well send the correct event to the events file
func (c *Control) NavWriter(buffer string) (*writer.WriteCloser, error) {
	return c.write.FileWriter(buffer, "navi")
}

// TitleWriter returns a WriteCloser attached to a buffers title file, which will as well send the correct event to the events file
func (c *Control) TitleWriter(buffer string) (*writer.WriteCloser, error) {
	return c.write.FileWriter(buffer, "title")
}

// ImageWriter returns a WriteCloser attached to a named file in the buffers' image directory
func (c *Control) ImageWriter(buffer, resource string) (*writer.WriteCloser, error) {
	return c.write.ImageWriter(buffer, resource)

}

// MainWriter returns a WriteCloser attached to a buffers feed/document function to set the contents of a given buffers' document or feed file, which will as well send the correct event to the events file
func (c *Control) MainWriter(buffer, doctype string) (*writer.WriteCloser, error) {
	return c.write.FileWriter(buffer, doctype)
}

func dispatch(c *Control) {
	// TODO: wrap with waitgroups
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
			run(c, ew, line)
		case <-c.done:
			return
		case <-c.ctx.Done():
			return
		}
	}
}

func run(c *Control, ew *writer.WriteCloser, line string) {
	c.Lock()
	defer c.Unlock()

	token := strings.Fields(line)
	if len(token) < 1 {
		return
	}

	switch token[0] {
	case "open":
		if len(token) < 2 {
			return
		}

		if e := c.ctl.Open(c, token[1]); e != nil {
			c.debug(ctlError, token[1], e)
			fmt.Fprintf(ew, "open: %v\n", e)
		}
	case "close":
		if len(token) < 2 {
			return
		}

		// We need to get to these still somehow
		if e := c.ctl.Close(c, token[1]); e != nil {
			c.debug(ctlError, token[1], e)
			fmt.Fprintf(ew, "close: %v\n", e)
		}

	case "link":
		if len(token) < 2 {
			return
		}

		if e := c.ctl.Link(c, token[1], token[2]); e != nil {
			c.debug(ctlError, token[1], e)
			fmt.Fprintf(ew, "link: %v\n", e)
		}

	default:
		cmd, err := c.run.BuildCommand(line)
		if err != nil {
			c.debug(ctlError, token[0], errors.New("unsupported command"))
			fmt.Fprintf(ew, "unsupported command: %s", token[0])
			return
		}

		if cmd.Heading == command.ServiceGroup {
			serviceCommand(c, cmd, ew)
			return
		}

		defaultCommand(c, cmd, ew, token[0])
	}
}

func serviceCommand(c *Control, cmd *command.Command, ew *writer.WriteCloser) {
	switch cmd.Args[0] {
	case "quit":
		// Close our local listeners, then
		close(c.done)
		c.ctl.Quit()
	case "restart":
		c.ctl.Restart(c)
	case "reload":
		c.ctl.Refresh(c)
	default:
		fmt.Fprintf(ew, "unsupported command: %s", cmd.Args[0])
	}
}

func defaultCommand(c *Control, cmd *command.Command, ew *writer.WriteCloser, name string) {
	cmd2 := &Command{
		Name:        cmd.Name,
		Description: cmd.Description,
		Args:        cmd.Args,
		Alias:       cmd.Alias,
		From:        cmd.From,
	}

	cmd2.Heading = ComGroup(cmd.Heading)
	c.debug(ctlDefault, cmd2)
	if e := c.ctl.Default(c, cmd2); e != nil {
		c.debug(ctlError, name, e)
		fmt.Fprintf(ew, "%s: %v\n", name, e)
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
