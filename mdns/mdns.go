// Package mdns is a convenience wrapper over github.com/grandcat/zeroconf for listing Altid services over mdns
//
//	go get github.com/altid/server/mdns
package mdns

import (
	"fmt"
	"net/url"
	"strconv"

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

func ParseURL(addr, svc string) (*Entry, error) {
	u, err := url.Parse(addr)
	if err != nil {
		u, err = url.Parse(fmt.Sprintf("https://%s", addr))
		if err != nil {
			return nil, err
		}
	}
	e := &Entry{
		Addr: u.Hostname(),
		Name: svc,
		Port: 564,
	}
	fmt.Println(e.Addr)
	if u.Port() != "" {
		e.Port, err = strconv.Atoi(u.Port())
	}
	return e, err
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
