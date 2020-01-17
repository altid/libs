package main

import (
	"fmt"

	"github.com/grandcat/zeroconf"
)

var mdnsEntries []*zeroconf.Server

func registerMDNS(srv map[string]*service) error {
	for _, s := range srv {
		sname := fmt.Sprintf("_%s._tcp", s.name)
		entry, err := zeroconf.Register("altid", sname, s.addr, *listenPort, nil, nil)
		if err != nil {
			return err
		}
		mdnsEntries = append(mdnsEntries, entry)
	}
	return nil
}

func cleanupMDNS() {
	for _, service := range mdnsEntries {
		service.Shutdown()
	}
}
