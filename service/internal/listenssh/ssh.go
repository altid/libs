package listenssh

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/store"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/muesli/termenv"

	tea "github.com/charmbracelet/bubbletea"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
)

// Strong chance we'll need to declare some UI in this
// But it would be good to abstract it outside of the scope of this file
// Ideally, we would pass it in at compile time to our parent listener
type ErrSSH string

const (
	ErrSSHNoOpen = ErrSSH("store does not implement Opener")
	ErrSSHNoList = ErrSSH("store does not implement Lister")
)

func (e ErrSSH) Error() string {
	return string(e)
}

var l *log.Logger

type sessionMsg int

const (
	sessionStart sessionMsg = iota
	sessionClient
	sessionBuffer
	sessionOpen
	sessionInfo
	sessionErr
)

type Session struct {
	address  string
	id_rsa   string
	authkeys string
	cmd      commander.Commander
	cb       callback.Callback
	list     store.Lister
	open     store.Opener
	delete   store.Deleter
	progs    []*tea.Program
	act      func(string)
	debug    func(sessionMsg, ...any)
	*ssh.Server
}

func NewSession(address, id_rsa, authkeys string, debug bool) (*Session, error) {
	s := &Session{
		address:  address,
		id_rsa:   id_rsa,
		authkeys: authkeys,
		debug:    func(sessionMsg, ...any) {},
	}
	if debug {
		s.debug = sessionLogger
		l = log.New(os.Stdout, "listenssh ", 0)
	}
	srv, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", address, 22)),
		wish.WithHostKeyPath(id_rsa),
		wish.WithMiddleware(
			bm.MiddlewareWithProgramHandler(s.ProgramHandler, termenv.ANSI256),
			lm.Middleware(),
		),
	)
	s.Server = srv
	return s, err
}

func (s *Session) Auth(ap *auth.Protocol) error {
	return nil
}

func (s *Session) Address() string              { return s.address }
func (s *Session) SetActivity(act func(string)) { s.act = act }
func (s *Session) Listen() error {
	var err error
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			s.debug(sessionErr, err)
			return
		}
	}()
	<-done
	// Make sure this works correctly.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	return s.Shutdown(ctx)
}

func (s *Session) ProgramHandler(sh ssh.Session) *tea.Program {
	if _, _, active := sh.Pty(); !active {
		wish.Fatalln(sh, "terminal is not active")
	}
	model := initialModel()
	model.send = s.send
	model.id = sh.User()
	p := tea.NewProgram(model, tea.WithOutput(sh), tea.WithInput(sh))
	s.progs = append(s.progs, p)
	return p
}

func (s *Session) Register(filer store.Filer, cmd commander.Commander, cb callback.Callback) error {
	list, ok := filer.(store.Lister)
	if !ok {
		return ErrSSHNoList
	}
	open, ok := filer.(store.Opener)
	if !ok {
		return ErrSSHNoOpen
	}
	if delete, ok := filer.(store.Deleter); ok {
		s.delete = delete
	}
	s.open = open
	s.list = list
	s.cmd = cmd
	s.cb = cb
	return nil
}

func (s *Session) send(msg tea.Msg) {
	for _, p := range s.progs {
		go p.Send(msg)
	}
}

func sessionLogger(msg sessionMsg, args ...any) {
	switch msg {
	case sessionErr:
		l.Printf("error: %e", args[0])
	case sessionInfo:
		l.Printf("info: %v", args)
	case sessionStart:
		l.Println("starting session")
	case sessionOpen:
		l.Printf("open: %s", args[0])
	case sessionBuffer:
		if cmd, ok := args[0].(*commander.Command); ok {
			l.Printf("buffer: name=\"%s\" args=\"%s\" from=\"%s\"", cmd.Name, cmd.Args[0], cmd.From)
		}
	case sessionClient:
		//if client, ok := args[0].(*client); ok {
		//	l.Printf("client: user=\"%s\" buffer=\"%s\" id=\"%s\"\n", client.name, client.current, client.uuid.String())
		//}
	}
}

type (
	errMsg  error
	chatMsg struct {
		id   string
		text string
	}
)

type model struct {
	send        func(tea.Msg)
	viewport    viewport.Model
	messages    []string
	id          string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

func initialModel() model {
	ta := textarea.New()
	// We don't need this, but it doesn't matter
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = " > "
	ta.CharLimit = 280
	ta.SetWidth(30)
	ta.SetHeight(1)
	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	vp := viewport.New(30, 5)
	// TODO: Real data instead
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)
	ta.KeyMap.InsertNewline.SetEnabled(false)
	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)
	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.send(chatMsg{
				id:   m.id,
				text: m.textarea.Value(),
			})
			m.textarea.Reset()
		}
	case chatMsg:
		m.messages = append(m.messages, m.senderStyle.Render(msg.id)+": "+msg.text)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}
