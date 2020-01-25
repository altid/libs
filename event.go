package main

import (
	"bytes"
	"fmt"
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

// REDO
func (t *tail) readlines() []*event {
	var events []*event

	lines := make([]byte, 2048)
	b := bytes.NewBuffer(lines)
	hs, _ := t.fd.Stat()

	t.size = hs.Size()
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

	for {
		line, err := b.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return events
		}

		if e := parseEvent(line, t.name); e != nil {
			events = append(events, e)
		}

		if err == io.EOF {
			return events
		}
	}
}

func parseEvent(line []byte, name string) *event {
	etype := noneEvent

	if bytes.Contains(line, isfeed) {
		etype = feedEvent
	} else if bytes.Contains(line, isnoti) {
		etype = notifyEvent
	}

	lines := bytes.Split(line, []byte("/"))

	l := len(lines)
	if l < 3 {
		fmt.Printf("Small read on %s with %s\n", lines, string(line))
		return nil
	}

	e := &event{
		service: name,
		etype:   etype,
		name:    string(lines[l-2]),
	}

	return e
}
