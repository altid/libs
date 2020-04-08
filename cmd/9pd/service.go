package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/altid/server/client"
	"github.com/altid/server/files"

	"github.com/go9p/styx"
)

type service struct {
	listen string
	port   string
	chatty bool
	tls    bool
	log    bool
	client *client.Manager
	cert   string
	key    string
}

func (s *service) Run(ctx context.Context, r *files.Files) error {
	t := &styx.Server{
		Addr: fmt.Sprintf("%s:%s", s.listen, s.port),
	}

	//if s.factotum {
	//	t.Auth, t.OpenAuth = factotum.Start(auth.OpenRPC, "p9any")
	//}

	if s.chatty {
		t.TraceLog = log.New(os.Stdout, "styx: ", 0)
	}

	t.Handler = styx.HandlerFunc(func(sess *styx.Session) {
		c := s.client.Client(0)
		for sess.Next() {
			handleReq(c, r, sess.Request(), s.listen)
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

