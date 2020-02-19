package fs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/mischief/ndb"
)

const (
	ErrNoConfigure  = "unable to find or create config for this service"
	ErrNoSuchKey    = "no such key"
	ErrNoEntries    = "unable to find config entries for this service"
	ErrMultiEntries = "config contains duplicate entries for this service"
)

// Config defines a services' configuration in a given config file
type Config struct {
	Name   string
	Values []*ConfigEntry
}

// ConfigEntry is a single tuple in a services configuration
type ConfigEntry struct {
	Key   string
	Value string
}

// Configurator when satisfied will call Configure when an entry is not found, creating an entry and loading it
type Configurator interface {
	Configure() (*Config, error)
}

// NewConfig returns a valid config for a given service. If one is not found, the Configurators Configure method
// will be called to interactively create one
// Any errors in Configure() should be caught by the client and won't be returned
func NewConfig(c Configurator, service string) (*Config, error) {
	conf, err := ndb.Open(getConfDir(service))
	if err != nil {
		if err != os.ErrNotExist {
			return nil, err
		}

		cf, err := c.Configure()
		if err != nil {
			return nil, err
		}

		return cf.writeToFile()
	}

	if entry := conf.Search("service", service).Search("service"); entry != "" {
		return parseConfig(service)
	}

	if c == nil {
		return nil, errors.New(ErrNoConfigure)
	}

	cf, err := c.Configure()
	if err != nil {
		return nil, err
	}

	return cf.writeToFile()
}

// Search queries for an entry in the config matching key
// Returning the value if it exists, or a "no such key" error
func (c *Config) Search(key string) (string, error) {
	if key == "service" {
		return c.Name, nil
	}

	for _, k := range c.Values {
		if k.Key == key {
			return k.Value, nil
		}
	}

	return "", errors.New(ErrNoSuchKey)
}

// String returns our entry tuples in the form of key=value
func (c *Config) String() string {
	var entry strings.Builder

	fmt.Fprintf(&entry, "service=%s", c.Name)

	for _, item := range c.Values {
		fmt.Fprintf(&entry, " %s=%s", item.Key, item.Value)
	}

	return entry.String()
}

func (c *Config) writeToFile() (*Config, error) {
	fp, err := os.OpenFile(getConfDir(c.Name), os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	defer fp.Close()
	// NOTE(halfwit) We always want an extra newline to separate entries
	fmt.Fprintf(fp, "%s\n\n", c.String())

	return c, nil
}

// GetLogDir returns a canonical directory for a user log, searching first altid/config
// If no entry is found or the file is missing, it will return a default path depending on the current operating system. Refer to UserShareDir documentation for what that is for your system
func GetLogDir(service string) string {
	conf, err := ndb.Open(getConfDir(service))
	if err != nil {
		return logDir(service)
	}

	logdir := conf.Search("service", service).Search("log")
	if logdir != "" {
		return path.Join(logdir, service)
	}

	return logDir(service)
}

// If we can't get a valid dir it should be fatal
func logDir(service string) string {
	userdir, err := UserShareDir()
	if err != nil {
		log.Fatal(err)
	}

	return path.Join(userdir, "altid", service)
}

func getConfDir(service string) string {
	confdir, err := UserConfDir()
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
		return buildconf(recs[0], service)
	default:
		return nil, errors.New(ErrMultiEntries)
	}
}

func buildconf(recs ndb.Record, service string) (*Config, error) {
	c := &Config{
		Name:   service,
		Values: make([]*ConfigEntry, len(recs)),
	}

	for _, tup := range recs {
		v := &ConfigEntry{
			Key:   tup.Attr,
			Value: tup.Val,
		}
		c.Values = append(c.Values, v)
	}

	return c, nil
}
