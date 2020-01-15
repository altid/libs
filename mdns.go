package main

import (
	"fmt"

	"github.com/grandcat/zeroconf"
)

var entries []*zeroconf.Server

// Make our listings, altid and _servname._tcp listening on 564 for whichever IP 
func registerMDNS(srv []*service) error {
	for _, s := range srv {
		ip := fmt.Sprintf(".%s", s.addr)
		if ip == ".dhcp" {
			ip = ".local"
		}
		sname := fmt.Sprintf("_%s._tcp", s.name)
		entry, err := zeroconf.Register("altid", sname, ip, 564, nil, nil)
		if err != nil {
			return err
		}
		entries = append(entries, entry)
	}
	return nil
}

func cleanupMDNS() {
	for _, service := range entries {
		service.Shutdown()
	}
}
