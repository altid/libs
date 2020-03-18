package mdns

import (
	"fmt"

	"github.com/grandcat/zeroconf"
)

type Entry struct {
	Addr    string
	Port    string
	Name    string
	service *zeroconf.Server
}

// Register adds the service listing to the intranet registry
func Register(srv *Entry) error {
	for _, s := range srv {
		sname := fmt.Sprintf("_%s._tcp", s.Name)

		entry, err := zeroconf.Register("altid", sname, s.Addr, s.Port, nil, nil)
		if err != nil {
			return err
		}

		s.service = entry
	}

	return nil
}

// Cleanup stops the broadcast
func (e *Entry) Cleanup() {
	for _, service := range mdnsEntries {
		e.service.Shutdown()
	}
}
