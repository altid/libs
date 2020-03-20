package ninep

import (
	"github.com/altid/server/tail"
)

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

		// Send off our feed event as well
		s.feed <- struct{}{}
	}
}
