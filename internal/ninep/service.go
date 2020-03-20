package ninep

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/altid/libs/config"
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
	config   *config.Config
	tabs     *tabs.Manager
	feed     *routes.FeedHandler
	command  chan *command.Command
	events   chan *tail.Event
	basedir  string
	log      bool
	chatty   bool
	tls      bool
	factotum bool
	debug    func(string, ...interface{})
}

// Add the service to the client.Aux (yay for self-reference?)
func (s *service) run() error {
	//addr, port := s.config.DialString() returns found or defaults
	addr := ""
	port := 564

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
		c := s.client.Client(0)
		c.Aux = s

		c.SetBuffer(s.tabs.List()[0].Name)

		s.debug("client start id=\"%d\"", c.UUID)
		for sess.Next() {
			handleReq(c, sess.Request())
		}

		s.debug("client stop id=\"%d\"", c.UUID)
	})

	fp, err := os.OpenFile(path.Join(s.basedir, s.config.Name, "ctl"), os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	go s.listenEvents()
	go s.listenCommands(fp)

	switch s.tls {
	case true:
		cert, err := s.config.Search("cert")
		if err != nil {
			return err
		}

		key, err := s.config.Search("key")
		if err != nil {
			return err
		}

		if e := t.ListenAndServeTLS(cert, key); e != nil {
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
