//Package mdns is a convenience wrapper over github.com/grandcat/zeroconf for listing Altid services over mdns
//
//		go get github.com/altid/server/mdns
//
package mdns

import (
	"fmt"

	"github.com/grandcat/zeroconf"
)

// An Entry indicates a service we want to broadcast over mDNS
type Entry struct {
	Addr    string
	Name    string
	Port    int
	service *zeroconf.Server
}

// Register adds the service listing to the intranet registry
func Register(srv *Entry) error {
	sname := fmt.Sprintf("_%s._tcp", srv.Name)

	entry, err := zeroconf.Register("altid", sname, srv.Addr, srv.Port, nil, nil)
	if err != nil {
		return err
	}

	srv.service = entry
	return nil
}

// Cleanup stops the broadcast
func (e *Entry) Cleanup() {
	e.service.Shutdown()
}
