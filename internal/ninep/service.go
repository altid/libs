package ninep

import (
	"fmt"
	"log"
	"os"

	"github.com/altid/libs/config"
	"github.com/altid/server/client"
	"github.com/altid/server/command"
	"github.com/altid/server/files"
	"github.com/altid/server/tabs"
	"github.com/go9p/styx"
)

type service struct {
	client   *client.Manager
	files    *files.Handler
	config   *config.Config
	tabs     *tabs.Manager
	events   chan *tail.Event
	commands chan *command.Command
	feed     chan struct{}
	basedir  string
	log      bool
	chatty   bool
	tls      bool
	factotum bool
	debug    func(string, ...interface{})
}

// Add the service to the client.Aux (yay for self-reference?)
func (s *service) run() error {
	addr := s.config.Addr()
	port := s.config.Port()

	t := &styx.Server{
		Addr: addr + fmt.Sprintf(":%d", port),
	}

	//if s.factotum {
	//	t.Auth, t.OpenAuth = factotum.Start(auth.OpenRPC, "p9any")
	//}

	if s.chatty {
		t.TraceLog = log.New(os.Stderr, "", 0)
	}

	if s.log {
		s.debug = serviceDebugLog
		t.ErrorLog = log.New(os.Stderr, "", 0)
	}

	t.Handler = styx.HandlerFunc(func(sess *styx.Session) {
		c := s.client.Add(0)
		c.Aux = s

		s.debug("client start id=\"%d\"", s.client.UUID)
		for sess.Next() {
			handleReq(c, sess.Request())
		}

		s.debug("client stop id=\"%d\"", s.client.UUID))
	})

	go s.handleCommands()
	go s.listenEvents()

	switch s.tls {
	case true:
		if e := t.ListenAndServeTLS("none", "none"); e != nil {
			return e
		}
	case false:
		if e := t.ListenAndServe(); e != nil {
			return e
		}
	}

	return nil
}

func serviceDebugLog(format string, v ...interface{}) {
	l := log.New(os.Stdout, "9pd service :", 0)
	l.Printf(format, v...)
}
