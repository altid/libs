package server

import (
	"context"
	"log"
	"testing"
	"time"
)

type service struct {
}

// Walk through and test read/writes to the files
func (s *service) Run(ctx context.Context, svc *Service) error {
	return nil
}

func (s *service) Address() (string, string) {
	return "localhost", "564"
}

func TestServer(t *testing.T) {
	service := &service{}
	ctx, cancel := context.WithCancel(context.Background())

	s, err := NewServer(ctx, service, "resources")
	if err != nil {
		panic(err)
	}

	go time.AfterFunc(time.Second*3, func() { cancel() })
	s.Logger = log.Printf

	if e := s.Listen(); e != nil {
		t.Error(e)
	}
}
