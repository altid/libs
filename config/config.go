package main

import (
	"log"

	"github.com/mischief/ndb"
)

type Config struct {
	*ndb.Ndb
}

// Switch to altid/libs/config
// We want a config.All, to get back all config entries in an array
// Then we can walk them and see all the names trivially
// get the error messages back, etc

func NewConfig(filename string) (*config, error) {
	conf, err := ndb.Open(filename)
	if err != nil {
		return nil, err
	}

	if conf == nil {
		log.Fatal("error parsing Altid config")
	}

	c := &config{conf}

	return c, nil
}

func (c *Config) ListServices() []string {
	var results []string

	for _, rs := range c.Search("service", "") {
		if rs[0].Attr != "service" {
			log.Fatal("incorrectly formatted Altid config")
		}

		results = append(results, rs[0].Val)
	}

	return results
}

func (c *Config) GetAddress(name string) string {
	rs := c.Search("service", name)
	if rs == nil {
		log.Fatal("no service entry found")
	}

	if address := rs.Search("listen_address"); address != "" {
		return address
	}

	return ""
}
