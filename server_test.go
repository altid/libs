package server

import (
	"context"
	"crypto/tls"
	"log"
	"testing"
	"time"

	"github.com/altid/server/files"
	"github.com/altid/server/settings"
)

type service struct {
}

// Walk through and test read/writes to the files
func (s *service) Run(ctx context.Context, h *files.Files) error {
	return nil
}

func TestServer(t *testing.T) {
	service := &service{}

	settings := settings.NewSettings("resources", false, tls.Certificate{})
	ctx, cancel := context.WithCancel(context.Background())

	s := NewServer(ctx, service)
	if e := s.Config(settings); e != nil {
		t.Error(e)
	}

	go time.AfterFunc(time.Second*3, func() { cancel() })
	s.Logger = log.Printf
	if e := s.Listen(); e != nil {
		t.Error(e)
	}
}
