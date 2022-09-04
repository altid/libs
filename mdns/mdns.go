//Package mdns is a convenience wrapper over github.com/grandcat/zeroconf for listing Altid services over mdns
//
//		go get github.com/altid/server/mdns
//
package mdns

import (
	"github.com/grandcat/zeroconf"
)

// An Entry indicates a service we want to broadcast over mDNS
type Entry struct {
	Addr    string
	Name    string
	Txt     []string
	Port    int
	service *zeroconf.Server
}

// Register adds the service listing to the intranet registry
func Register(srv *Entry) error {
	entry, err := zeroconf.Register(srv.Name, "_altid._tcp", "local.", srv.Port, srv.Txt, nil)
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
