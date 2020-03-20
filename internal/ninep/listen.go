package ninep

import "github.com/altid/server/tail"

func (s *service) listenCommands() {
	for _, cmd := range s.commands {
		switch cmd.CmdType {
		/* mostly send cmds down for now
		OtherCmd
		OpenCmd
		ReloadCmd
		BufferCmd
		CloseCmd
		LinkCmd
		QuitCmd
		*/
		}
	}
}

// We need to send feed commands at very least
func (s *service) listenEvents() {
	for e := range s.events {

		t := s.tabs.Tab(e.Name)

		switch e.EventType {
		case tail.NotifyEvent:
			t.Alert = true
		case tail.DocEvent:
		case tail.FeedEvent:
		default:
			// None event here
		}

		if !t.Active {
			t.Unread++
		}
	}
}
