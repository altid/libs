package main

// Abstracted back a bit from raw ndb so that we can change the interface if we ever need

import (
	"log"

	"github.com/mischief/ndb"
)

type config struct {
	*ndb.Ndb
}

func newConfig(filename string) (*config, error) {
	var c *config

	conf, err := ndb.Open(filename)
	if err != nil {
		return nil, err
	}

	if conf == nil {
		log.Fatal("error parsing Altid config")
	}

	c.Ndb = conf

	return c, nil
}

func (c *config) listServices() []string {
	var results []string

	for _, rs := range c.Search("service", "") {
		if rs[0].Attr != "service" {
			log.Fatal("incorrectly formatted Altid config")
		}

		results = append(results, rs[0].Val)
	}

	return results
}

func (c *config) getAddress(name string) string {
	rs := c.Search("service", name)
	if rs == nil {
		log.Fatal("no service entry found")
	}

	if address := rs.Search("listen_address"); address != "" {
		return address
	}

	return ""
}
