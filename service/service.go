package service

import (
	"bufio"
	"io"
	"log"
)

// We may have to do config rework, but html and markup should be fine.
// We can take the config parsing out to the library
// Just export a function to get that dir --> fd, then pass an fd with handlers for ctl and input messages
// From there, we could convenience wrap our append/create/write/etc

type service struct {
	fd io.ReadWriteCloser
	handler Handler
}

type Handler interface {
	Input([]byte) // add Markup, etc
	Ctl([]byte)   // commender comes in here
}

func Start(name string, handler Handler) (io.WriteCloser, error) {
	fd, err := connectService(name) // from _plan9.go, for example
	if err != nil {
		return nil, err
	}

	s := &service{
		fd: fd,
		handler: handler,
	}

	go s.handleIncoming()
	return s.fd, nil
}

func (s service) handleIncoming() {
	scanner := bufio.NewScanner(s.fd)
	for scanner.Scan() {
		log.Print(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
}