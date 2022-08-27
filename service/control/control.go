package control

// TODO: Refactor into just a public interface
import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/altid/libs/service/command"
	"github.com/altid/libs/service/input"
	"github.com/altid/libs/service/internal/cmd"
	"github.com/altid/libs/service/internal/ctrl"
	"github.com/altid/libs/service/listener"
	"github.com/altid/libs/store"
)

type Manager interface {
	Run(*Control, *command.Command) error
	Quit()
}

type Sender interface {
	Send(string)
}

// Control type can be used to manage a running ctl file session
type Control struct {
	ctx			context.Context
	cancel		context.CancelFunc
	ctl			Manager
	input		input.Handler
	listener	listener.Listener
	run			*ctrl.Control
	done		chan struct{}
	req			chan string
	debug  func(ctlMsg, ...interface{})
}

type ctlMsg int

const (
	ctlError ctlMsg = iota
	ctlCleanup
	ctlCreate
	ctlDelete
	ctlRemove
	ctlStart
	ctlNotify
	ctlDefault
)

// New sets up a ready-to-listen ctl file
// logdir is the directory to store the contents written to the main element of a buffer. Logging any other type of data is left to implementation details, but is considered poor form for Altid's design.
// This will return an error if a ctl file exists at the given directory
func New(ctl interface{}, store store.Filer, listener listener.Listener, logdir string, debug bool) (*Control, error) {

	manager, ok := ctl.(Manager)
	if !ok {
		return nil, errors.New("ctl missing Run/Quit method(s)")
	}

	req := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	rtc := ctrl.New(ctx, store, logdir, "", "", nil)

	c := &Control{
		ctx:        ctx,
		cancel:     cancel,
		req:        req,
		done:       make(chan struct{}),
		run:        rtc,
		listener:	listener,
		ctl:        manager,
		debug:      func(ctlMsg, ...interface{}) {},
	}

	if debug {
		c.debug = ctlLogger
	}

	// Some services don't use input
	input, ok := ctl.(input.Handler)
	if ok {
		c.input = input
	}

	cmdlist := cmd.DefaultCommands
	cmdlist = append(cmdlist, &cmd.Command{
		Name:        "main",
		Args:        []string{"<quit|restart|reload>"},
		Heading:     cmd.ServiceGroup,
		Description: "Control the lifecycle of a service",
	})

	rtc.SetCommands(cmdlist...)
	return c, nil
}

// Cleanup flushes anything pending to logs, and disconnects any remaining clients
func (c *Control) Cleanup() {
	c.debug(ctlCleanup)
	c.run.Cleanup()

}

// CreateBuffer creates a buffer of given name, as well as symlinking your file as follows:
// `os.Symlink(path.Join(logdir, name), path.Join(rundir, name))`
// This logged file will persist across reboots
// Calling CreateBuffer on a directory that already exists will return nil
func (c *Control) CreateBuffer(name string) error {
	c.debug(ctlCreate, name)
	return c.run.CreateBuffer(name)
}

// DeleteBuffer unlinks a document/buffer, and cleanly removes the directory
// Will return an error if it's unable to unlink on plan9, or if the remove fails.
func (c *Control) DeleteBuffer(name string) error {
	c.debug(ctlDelete, name)
	return c.run.DeleteBuffer(name)
}

// HasBuffer returns whether or not a buffer is present in the current control session
func (c *Control) HasBuffer(name string) bool {
	return c.run.HasBuffer(name) 
}

// Remove removes a buffer from the runtime dir. If the buffer doesn't exist, this is a no-op
func (c *Control) Remove(buffer, filename string) error {
	c.debug(ctlRemove, buffer, filename)
	return c.run.Remove(buffer, filename)
}

func (c *Control) SendCommand(input string) error {
	cmd, err := c.run.BuildCommand(input)
	if err != nil {
		return err
	}

	real := translate(cmd)
	if real.Heading == command.ServiceGroup {
		return serviceCommand(c, real)
	}

	return c.ctl.Run(c, real)
}

// Listen starts a network listener for incoming clients
func (c *Control) Listen() error {
	c.debug(ctlStart, "listen")
	return c.listener.Listen()
}

// SetCommands allows services to add additional commands
// Any client command encountered which matches will send
// The resulting command down to RunCommand
// Commands must include at least a name and a heading
// Running SetCommands after calling Start or Listen will have no effect
func (c *Control) SetCommands(cmd ...*command.Command) error {
	for _, comm := range cmd {
		if comm.Name == "" {
			return errors.New("command requires Name")
		}

		switch comm.Heading {
		case command.DefaultGroup, command.MediaGroup, command.ActionGroup:
			continue
		default:
			return errors.New("unsupported or nil Heading set")
		}
	}

	// TODO: Investigate how to do this
	//if e := setCommands(c.run, cmd...); e != nil {
	//	return e
	//}

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
func (c *Control) ErrorWriter() (*ctrl.WriteCloser, error) {
	return c.run.Errorwriter()
}

// StatusWriter returns a WriteCloser attached to a buffers status file
func (c *Control) StatusWriter(buffer string) (*ctrl.WriteCloser, error) {
	return c.run.FileWriter(buffer, "status")
}

// SideWriter returns a WriteCloser attached to a buffers `aside` file
func (c *Control) SideWriter(buffer string) (*ctrl.WriteCloser, error) {
	return c.run.FileWriter(buffer, "aside")
}

// NavWriter returns a WriteCloser attached to a buffers nav file
func (c *Control) NavWriter(buffer string) (*ctrl.WriteCloser, error) {
	return c.run.FileWriter(buffer, "navi")
}

// TitleWriter returns a WriteCloser attached to a buffers title file
func (c *Control) TitleWriter(buffer string) (*ctrl.WriteCloser, error) {
	return c.run.FileWriter(buffer, "title")
}

// ImageWriter returns a WriteCloser attached to a named file in the buffers' image directory
func (c *Control) ImageWriter(buffer, resource string) (*ctrl.WriteCloser, error) {
	return c.run.ImageWriter(buffer, resource)
}

// MainWriter returns a WriteCloser attached to a buffer's main output
func (c *Control) MainWriter(buffer string) (*ctrl.WriteCloser, error) {
	return c.run.FileWriter(buffer, "main")
}

// Context returns the underlying context of the service
// This will be closed after Quit() is called
func (c *Control) Context() context.Context {
	return c.ctx
}

func serviceCommand(c *Control, cmd *command.Command) error {
	switch cmd.Args[0] {
	case "quit":
		defer c.cancel()
		// Close our local listeners, then
		close(c.done)
		c.ctl.Quit()
		return nil
	// Eventually we may want to access these
	case "reload", "restart":
		c.ctl.Run(c, cmd)
		return nil
	default:
		return fmt.Errorf("unsupported command: %s", cmd.Args[0])
	}
}

func translate(cmd *command.Command) *command.Command {
	return &command.Command{
		Name:        cmd.Name,
		Description: cmd.Description,
		From:        cmd.From,
		Args:        cmd.Args,
		Alias:       cmd.Alias,
		Heading:     command.ComGroup(cmd.Heading),
	}
}

func ctlLogger(msg ctlMsg, args ...interface{}) {
	l := log.New(os.Stdout, "ctl ", 0)

	switch msg {
	case ctlError:
		l.Printf("error: buffer=\"%s\" err=\"%v\"\n", args[0], args[1])
	case ctlCleanup:
		l.Println("cleanup: ending")
	case ctlCreate:
		l.Printf("create: buffer=\"%s\"", args[0])
	case ctlRemove:
		l.Printf("remove: buffer=\"%s\", filename=\"%s\"\n", args[0], args[1])
	case ctlStart:
		l.Printf("%s: starting\n", args[0])
	case ctlNotify:
		l.Printf("notify: buffer=\"%s\" from=\"%s\" msg=\"%s\"\n", args[0], args[1], args[2])
	case ctlDefault:
		cmd := args[0].(*command.Command)
		switch cmd.Heading {
		case command.ActionGroup:
			l.Printf("%s group=\"action\" arguments=\"%s\"\n", cmd.Name, cmd.Args)
		case command.MediaGroup:
			l.Printf("%s group=\"media\" arguments=\"%s\"\n", cmd.Name, cmd.Args)
		}
	}
}
