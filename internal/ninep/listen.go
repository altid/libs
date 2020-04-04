package ninep

import (
	"io"

	"github.com/altid/server/command"
	"github.com/altid/server/tail"
)

func (s *service) listenCommands(fp io.WriteCloser) {
	defer fp.Close()

	for cmd := range s.command {
		c := s.client.Client(cmd.UUID)

		switch cmd.CmdType {
		case command.OtherCmd:
			go cmd.WriteOut(fp)
		case command.OpenCmd:
			c.SetBuffer(cmd.Args[0])
			go cmd.WriteOut(fp)
			s.update(cmd.UUID)
		case command.BufferCmd:
			c.SetBuffer(cmd.Args[0])
			s.update(cmd.UUID)
		case command.CloseCmd:
			if cmd.Args[0] == c.Current() {
				if e := c.Previous(); e != nil {
					c.SetBuffer("none")
				}
			}

			s.update(cmd.UUID)
			go cmd.WriteOut(fp)
		case command.LinkCmd:
			c.SetBuffer(cmd.Args[1])
			go cmd.WriteOut(fp)
			s.update(cmd.UUID)
		case command.ReloadCmd:
			// TODO (halfwit): We want to recreate everything but save our client connections
			// possibly we'll be loading more services, etc
		case command.QuitCmd:
			go cmd.WriteOut(fp)
		}
	}
}

// We need to send feed commands at very least
func (s *service) listenEvents() {
	for e := range s.events {
		t := s.tabs.Tab(e.Name)
		s.debug("event name=\"%s\" service=\"%s\"", e.Name, e.Service)

		switch e.EventType {
		case tail.NotifyEvent:
			t.Alert = true
		case tail.DocEvent:
		case tail.FeedEvent:
		default:
		}

		if !t.Active {
			t.Unread++
		}

		go s.sendFeeds(e)
	}
}

// sendFeeds loops through all connected clients
// If they're listening to the current buffer try to send an event
// Shortcut to a bool comparison if inactive
func (s *service) sendFeeds(e *tail.Event) {
	for _, c := range s.client.List() {
		if c.Active && c.Current() == e.Name {
			s.debug("feed name=\"%s\" id=\"%d\"", e.Name, c.UUID)
			s.feed.Send(c.UUID)
		}
	}
}

func (s *service) update(uuid uint32) {
	s.debug("feed update id=\"%d\"", uuid)
	s.feed.Done(uuid)
}
