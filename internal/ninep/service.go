package ninep

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/altid/server/client"
	"github.com/altid/server/command"
	"github.com/altid/server/files"
	"github.com/altid/server/internal/routes"
	"github.com/altid/server/tabs"
	"github.com/altid/server/tail"
	"github.com/go9p/styx"
)

type service struct {
	client   *client.Manager
	files    *files.Files
	tabs     *tabs.Manager
	feed     *routes.FeedHandler
	command  chan *command.Command
	events   chan *tail.Event
	cert     string
	key      string
	basedir  string
	listen   string
	name     string
	port     string
	log      bool
	chatty   bool
	tls      bool
	factotum bool
	debug    func(string, ...interface{})
}

// Add the service to the client.Aux (yay for self-reference?)
func (s *service) run() error {
	t := &styx.Server{
		Addr: fmt.Sprintf("%s:%s", s.listen, s.port),
	}

	//if s.factotum {
	//	t.Auth, t.OpenAuth = factotum.Start(auth.OpenRPC, "p9any")
	//}

	if s.chatty {
		t.TraceLog = log.New(os.Stdout, "styx: ", 0)
	}

	if s.log {
		s.debug = serviceDebugLog
		t.ErrorLog = log.New(os.Stdout, "styx: ", 0)
	}

	t.Handler = styx.HandlerFunc(func(sess *styx.Session) {
		c := s.client.Client(0)
		c.Aux = s

		defer s.client.Remove(c.UUID)
		s.debug("client set default=\"%s\"", s.tabs.List()[0].Name)
		c.SetBuffer(s.tabs.List()[0].Name)

		s.debug("client start id=\"%d\"", c.UUID)
		for sess.Next() {
			handleReq(c, sess.Request())
		}

		s.debug("client stop id=\"%d\"", c.UUID)
	})

	fp, err := os.OpenFile(path.Join(s.basedir, s.name, "ctl"), os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	go s.listenEvents()
	go s.listenCommands(fp)

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

func serviceDebugLog(format string, v ...interface{}) {
	l := log.New(os.Stdout, "9pd ", 0)
	l.Printf(format, v...)
}
