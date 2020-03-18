package tail

import (
	"bytes"
	"context"
	"io"
	"os"
)

// We may want to revisit more of these events, since we care about all events

// EventType indicites which type of file the event is for
type EventType int

// Available EventTypes - this will likely increase
const (
	FeedEvent EventType = iota
	NotifyEvent
	DocEvent
	NoneEvent
)

var isfeed = []byte("feed")
var isdocu = []byte("document")
var isnoti = []byte("notification")

// Event represents a single event that occured on a service
// Such as new data to feed or a notification
type Event struct {
	Service string
	Etype   EventType
	Name    string
}

// WatchEvents watches a given directory, returning a channel of all events written to the events file
func WatchEvents(ctx context.Context, dir, service string) (chan *Event, error) {
	return watchEvents(ctx, dir, service)
}

type tail struct {
	fd   *os.File
	name string
	size int64
}

func (t *tail) readlines() ([]*Event, error) {
	var events []*Event
	hs, _ := t.fd.Stat()

	if hs.Size() < t.size {
		t.size = 0
	}

	lines := make([]byte, 2048)
	n, err := t.fd.ReadAt(lines, t.size)
	if err != nil && err != io.EOF {
		return nil, err
	}

	t.size += int64(n)

	b := bytes.NewBuffer(lines)

	for {
		line, err := b.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return events, nil
		}

		if e := parseEvent(line, t.name); e != nil {
			events = append(events, e)
		}

		if err == io.EOF {
			return events, nil
		}
	}

	return events, nil
}

func parseEvent(line []byte, name string) *Event {
	etype := NoneEvent

	switch {
	case bytes.Contains(line, isfeed):
		etype = FeedEvent
	case bytes.Contains(line, isdocu):
		etype = DocEvent
	case bytes.Contains(line, isnoti):
		etype = NotifyEvent
	}

	lines := bytes.Split(line, []byte("/"))

	l := len(lines)
	if l < 3 {
		return nil
	}

	e := &Event{
		Service: name,
		Etype:   etype,
		Name:    string(lines[l-2]),
	}

	return e
}
