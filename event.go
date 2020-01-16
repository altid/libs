package main

import (
	"bytes"
	"io"
	"log"
	"os"
)

type eventType int8

const (
	feedEvent eventType = iota
	notifyEvent
	noneEvent
)

var isfeed = []byte("feed")
var isnoti = []byte("notification")

// event will be a batch of all received events since the last
type event struct {
	service string
	etype   eventType
	name    string
}

type tail struct {
	fd   *os.File
	name string
	size int64
}

func (t *tail) readlines() []*event {
	var lines = make([]byte, 2048)
	var events []*event
	hs, _ := t.fd.Stat()
	// Assume truncation
	if hs.Size() < t.size {
		t.size = 0
	}
	_, err := t.fd.ReadAt(lines, t.size)
	if err != nil && err != io.EOF {
		// NOTE(halfwit) We set t.size to 0 on truncation so the logs are correct
		// Later it is set to the correct file size
		log.Printf("Error reading from file %s at offset %d: %v", t.name, t.size, err)
		return nil
	}
	t.size = hs.Size()
	b := bytes.NewBuffer(lines)
	// TODO(halfwit): Switch to raw byte manipulation here.
	for {
		etype := noneEvent
		line, err := b.ReadBytes('\n')
		if err != nil {
			return events
		}
		if bytes.Contains(line, isfeed) {
			etype = feedEvent
		} else if bytes.Contains(line, isnoti) {
			etype = notifyEvent
		}
		lines := bytes.Split(line, []byte("/"))
		l := len(lines)
		if l < 3 {
			return events
		}
		e := &event{
			service: t.name,
			etype:   etype,
			name:    string(lines[l-2]),
		}
		events = append(events, e)
	}
}
