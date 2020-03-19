package config

import (
	"github.com/mischief/ndb"
)

// ListAll returns every config
func ListAll() ([]*Config, error) {
	var configs []*Config

	conf, err := ndb.Open(getConfDir(""))
	if err != nil {
		return nil, err
	}

	for _, rec := range conf.Search("service", "") {
		c, err := New(nil, rec[0].Val, false)
		if err != nil {
			return nil, err
		}

		configs = append(configs, c)
	}

	return configs, nil
}
