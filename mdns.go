package main

// List Avahi/Bonjour/Zeroconf mDNS entry for each service.
import (
	"fmt"

	"github.com/grandcat/zeroconf"
)

var entries[] *zeroconf.Server

func registerMDNS(cfg *config) error {
	for _, service := range cfg.listServices() {
		sname := fmt.Sprintf("_%s._tcp", service)
		entry, err := zeroconf.Register("altid", sname, "local.", 564, nil, nil)
		if err != nil {
			return err
		}
		entries = append(entries, entry)
	}
	return nil
}

func cleanupMDNS(){
	for _, service := range entries {
		service.Shutdown()
	}
}