package services

import (
	"io"

	"github.com/altid/server/command"
	"github.com/altid/server/internal/tail"
)

// tabs.Manager unref or so would really help with this
func (s *Service) listenCommands(fp io.WriteCloser) {
	defer fp.Close()

	for cmd := range s.Command {
		switch cmd.CmdType {
		case command.OtherCmd:
			go cmd.WriteOut(fp)
		case command.OpenCmd, command.LinkCmd:
			openCmd(s, cmd, fp)
		case command.BufferCmd:
			bufferCmd(s, cmd)
		case command.CloseCmd:
			closeCmd(s, cmd, fp)
		case command.ReloadCmd, command.RestartCmd:
		// TODO (halfwit): We want to recreate everything but save our client connections
		// possibly we'll be loading more services, etc
		case command.QuitCmd:
			cmd.WriteOut(fp)
		}
	}
}

// We need to send feed commands at very least
func (s *Service) listenEvents() {
	for e := range s.Events {
		if e.Name == "." {
			continue
		}

		t := s.Tabs.Tab(e.Name)
		s.Debug("event name=\"%s\" service=\"%s\"", e.Name, e.Service)

		switch e.EventType {
		case tail.NotifyEvent:
			t.Alert()
		case tail.DocEvent, tail.FeedEvent:
			// This may be useful in the future
		default:
		}

		t.Activity()

		go s.sendFeeds(e)
	}
}

// sendFeeds loops through all connected clients
// If they're listening to the current buffer try to send an event
// Shortcut to a bool comparison if inactive
func (s *Service) sendFeeds(e *tail.Event) {
	for _, c := range s.Client.List() {
		if c.Active && c.Current() == e.Name {
			s.Debug("feed name=\"%s\" id=\"%d\"", e.Name, c.UUID)
			s.Feed.Send(c.UUID)
		}
	}
}

func (s *Service) update(uuid uint32) {
	s.Debug("feed update id=\"%d\"", uuid)
	s.Feed.Done(uuid)
}

func openCmd(s *Service, cmd *command.Command, fp io.Writer) {
	go cmd.WriteOut(fp)

	c := s.Client.Client(cmd.UUID)
	s.Tabs.Done(c.Current())

	c.SetBuffer(cmd.Args[0])
	s.Tabs.Active(cmd.Args[0])

	s.update(cmd.UUID)
}

func bufferCmd(s *Service, cmd *command.Command) {
	c := s.Client.Client(cmd.UUID)
	s.Debug("buffer: %s, %s", c.Current(), cmd.Args[0])
	s.Tabs.Done(c.Current())

	c.SetBuffer(cmd.Args[0])
	s.Tabs.Active(cmd.Args[0])

	s.update(cmd.UUID)
}

func closeCmd(s *Service, cmd *command.Command, fp io.Writer) {
	go cmd.WriteOut(fp)

	c := s.Client.Client(cmd.UUID)
	s.Tabs.Remove(c.Current())

	c.Previous()
	s.Tabs.Active(c.Current())

	s.update(cmd.UUID)
}
