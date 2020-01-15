package main

// Abstracted back a bit from raw ndb so that we can change the interface if we ever need

import (
	"errors"
	"log"

	"github.com/mischief/ndb"
)

type config struct {
	*ndb.Ndb
}

func newConfig(filename string) (*config, error) {
	conf, err := ndb.Open(filename)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		log.Fatal(errors.New("Error parsing Altid config"))
	}
	c := &config{conf}
	return c, nil
}

func (c *config) listServices() []string {
	var results []string
	for _, rs := range c.Search("service", "") {
		if rs[0].Attr != "service" {
			log.Fatal("Incorrectly formatted Altid config")
		}
		results = append(results, rs[0].Val)
	}
	return results
}

func (c *config) refresh() error {
	changed, err := c.Changed()
	if err != nil {
		return err
	}
	if changed {
		return c.Reopen()
	}
	return nil
}

func (c *config) getAddress(name string) string {
	rs := c.Search("service", name)
	if rs == nil {
		log.Fatal(errors.New("No services found"))
	}
	if address := rs.Search("listen_address"); address != "" {
		return address
	}
	return ""
}
