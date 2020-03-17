package config

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/altid/libs/fs"
	"github.com/mischief/ndb"
)

func getConfDir(service string) string {
	confdir, err := fs.UserConfDir()
	if err != nil {
		log.Fatal(err)
	}

	return path.Join(confdir, "altid", "config")
}

func parseConfig(service string) (*Config, error) {
	conf, err := ndb.Open(getConfDir(service))
	if err != nil {
		return nil, err
	}

	recs := conf.Search("service", service)

	switch len(recs) {
	case 0:
		return nil, errors.New(ErrNoEntries)
	case 1:
		return buildConfigFromNdb(recs[0], service)
	default:
		return nil, errors.New(ErrMultiEntries)
	}
}

func buildConfigFromNdb(recs ndb.Record, service string) (*Config, error) {
	var values []*Entry
	c := &Config{
		Name:   service,
		Values: values,
	}

	for _, tup := range recs {
		v := &Entry{
			Key:   tup.Attr,
			Value: tup.Val,
		}
		c.Values = append(c.Values, v)
	}

	return c, nil
}

func createConfigFile(c Configurator, service string) (*Config, error) {
	if c == nil {
		return nil, errors.New(ErrBadConfigurator)
	}

	fp, err := os.OpenFile(getConfDir(service), os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	fp.Close()
	return buildConfigFromConfigurator(c, service)
}

func buildConfigFromConfigurator(c Configurator, service string) (*Config, error) {
	rw := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}

	conf, err := c(rw)
	if err != nil {
		return nil, err
	}

	conf.Name = service

	return conf, nil
}

func writeToFile(c *Config) error {
	fp, err := os.OpenFile(getConfDir(c.Name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	c.debug("write file=\"%s\"", fp.Name())

	defer fp.Close()
	// NOTE(halfwit) We always want an extra newline to separate entries
	if _, e := fmt.Fprintf(fp, "%s\n\n", c.String()); e != nil {
		return e
	}

	return nil
}
