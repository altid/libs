package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/altid/server"

	"github.com/go9p/styx"
)

type service struct {
	listen string
	chatty bool
	tls    bool
	log    bool
	cert   string
	key    string
}

func (s *service) Address() (string, string) {
	return *addr, *port
}

func (s *service) Run(ctx context.Context, svc *server.Service) error {
	t := &styx.Server{
		Addr: fmt.Sprintf("%s:%s", *addr, *port),
	}

	//if s.factotum {
	//	t.Auth, t.OpenAuth = factotum.Start(auth.OpenRPC, "p9any")
	//}

	if s.chatty {
		t.TraceLog = log.New(os.Stdout, "styx: ", 0)
	}

	t.Handler = styx.HandlerFunc(func(sess *styx.Session) {
		c := svc.Client.Client(0)
		c.SetBuffer(svc.Default())

		for sess.Next() {
			handleReq(c, svc.Files, sess.Request(), s.listen)
		}
	})

	switch s.tls {
	case true:
		if e := t.ListenAndServeTLS(s.cert, s.key); e != nil {
			return e
		}
	case false:
		if e := t.ListenAndServe(); e != nil {
			return e
		}
	}

	return nil
}
