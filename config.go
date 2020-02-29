package main

// Abstracted back a bit from raw ndb so that we can change the interface if we ever need

import (
	"io"
	"log"

	"github.com/altid/libs/auth"
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
		log.Fatal("error parsing Altid config")
	}

	c := &config{conf}

	return c, nil
}

func (c *config) mockFactotum() (io.ReadWriteCloser, error) {
	mock, err := auth.NewRPC("9pd")
	if err != nil {
		return nil, err
	}

	rs := c.Search("auth", "oauth2")
	if rs != nil {
		oa2e := &auth.OAuth2Entry{
			Provider: rs.Search("provider"),
			Key:      rs.Search("key"),
			Secret:   rs.Search("secret"),
		}

		if e := mock.AddEntry(oa2e); e != nil {
			return nil, e
		}
	}

	rs = c.Search("auth", "key")
	if rs != nil {
		keye := &auth.KeyEntry{
			User:    rs.Search("user"),
			Dom:     rs.Search("dom"),
			KeyType: rs.Search("type"),
		}

		if e := mock.AddEntry(keye); e != nil {
			return nil, e
		}
	}

	return mock, nil
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
