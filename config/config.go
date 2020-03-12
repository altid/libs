// manage and create configs for Altid services
package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/fs"
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
	Values []*Entry
}

// Entry is a single tuple in a services configuration
type Entry struct {
	Key   string
	Value string
}

// Configurator is called when no entry is found for a given service
// It should query the user for each required value, returning a
// complete and usable Config
type Configurator func() (*Config, error)

// New returns a valid config for a given service. If one is not found, the Configurators Configure method
// will be called to interactively create one
func New(c Configurator, service string) (*Config, error) {
	conf, err := ndb.Open(getConfDir(service))
	if err != nil {
		if err != os.ErrNotExist {
			return nil, err
		}

		cf, err := c()
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

	cf, err := c()
	if err != nil {
		return nil, err
	}

	return cf.writeToFile()
}

// Password queries the database for a password, and uses the factotum
func (c *Config) Password() (string, error) {
	pass, err := c.Search("auth")
	if err != nil {
		return "", err
	}

	if len(pass) > 5 && pass[:5] == "pass=" {
		pass = pass[5:]
	}

	if pass == "factotum" {
		userPwd, err := auth.Getuserpasswd(
			"proto=pass service=%s",
			c.Name,
		)
		if err != nil {
			return "", err
		}

		pass = userPwd.Password
	}

	return pass, nil
}

func (c *Config) SSLCert() (tls.Certificate, error) {
	cert, err := c.Search("cert")
	if err != nil {
		return tls.Certificate{}, err
	}

	key, err := c.Search("key")
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.LoadX509KeyPair(cert, key)
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

func (c *Config) MustSearch(key string) string {
	val, err := c.Search(key)
	if err != nil {
		return ""
	}

	return val
}

func (c *Config) Log() string {
	dir, err := c.Search("log")
	if err != nil {
		return "none"
	}

	return dir
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
		return "none"
	}

	logdir := conf.Search("service", service).Search("log")
	if logdir != "" {
		return path.Join(logdir, service)
	}

	return "none"
}

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
		return buildconf(recs[0], service)
	default:
		return nil, errors.New(ErrMultiEntries)
	}
}

func buildconf(recs ndb.Record, service string) (*Config, error) {
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
