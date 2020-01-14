package main

import (
	"io"
	"log"
	"os"
)

// event will be a batch of all received events since the last
type event struct {
	lines []byte
	name  string
}

type tail struct {
	fd   *os.File
	name string
	size int64
}

func (t *tail) readlines() *event {
	var lines = make([]byte, 2048)
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
	return &event{
		lines: lines,
		name:  t.name,
	}
}
