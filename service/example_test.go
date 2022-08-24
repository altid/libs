package service_test

import (
	"log"

	"github.com/altid/libs/service"
	"github.com/altid/libs/service/listener"
	"github.com/altid/libs/store"
)

type Manager struct {}

func (m *Manager) Run(*service.Command, *service.Control) error {
	// Callback - do something with a given command
	// [...]
	return nil
}

func (m *Manager) Quit() {
	// Clean up any state here
	// [...]
}

func ExampleControl_Listen() {
	// Register everything we use to the service controller
	listen, err := listener.NewListen9p("127.0.0.1", "", "")
	if err != nil {
		log.Fatal(err)
	}

	s := store.NewRamStore()
	manage := Manager{}

	ctl, err := service.New(manage, s, listen, "", false)
	if err != nil {
		log.Fatal(err)
	}

	// Before we can call ctl.Listen, we need to register a store for our listener
	// Not doing this will return an error
	if e := listen.Register(s, nil); e != nil {
		log.Fatal(e)
	}

	// Finally call our embedded listener
	ctl.Listen()
}
