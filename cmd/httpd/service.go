package main

import (
	"context"

	"github.com/altid/server"
)

type service struct {
}

func (s *service) Run(ctx context.Context, svc *server.Service) error {
	return nil
}

func (s *service) Address() (string, string) {
	return *addr, *port
}
