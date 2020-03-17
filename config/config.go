// Package config manage and create configs for Altid services
package config

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/fs"
	"github.com/mischief/ndb"
)

// Errors
const (
	ErrBadConfigurator = "configurator nil or invalid, cannot continue"
	ErrNoConfigure     = "unable to find or create config for this service. To create one, please run %s -conf"
	ErrNoSuchKey       = "no such key"
	ErrNoEntries       = "unable to find config entry for this service."
	ErrMultiEntries    = "config contains duplicate entries for this service"
)

var createConfig = flag.Bool("conf", false, "Create configuration file for service")

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
type Configurator func(io.ReadWriter) (*Config, error)

// Mock returns a mock config for a given service for testing
// It always calls Configurator to create a Config, and will not write to file
func Mock(c Configurator, service string, debug bool) (*Config, error) {
	return nil, nil
}

// New returns a valid config for a given service. If one is not found, the Configurator
// will be called to interactively create one
func New(c Configurator, service string, debug bool) (*Config, error) {
	// Since this is a library to create services, we can expect
	// there to be flags passed in
	flag.Parse()
	if *createConfig {
		return createConfigFile(c, service)
	}

	conf, err := ndb.Open(getConfDir(service))
	if err != nil {
		return nil, fmt.Errorf(ErrNoConfigure, os.Args[0])
	}

	if entry := conf.Search("service", service).Search("service"); entry != "" {
		return parseConfig(service)
	}

	return nil, fmt.Errorf(ErrNoConfigure, os.Args[0])
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

// SSLCert returns a tls.Certificate based on successfully finding
// a cert and key file listen in the configuration
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

// MustSearch returns a value or an empty string, if not found
func (c *Config) MustSearch(key string) string {
	val, err := c.Search(key)
	if err != nil {
		return ""
	}

	return val
}

// Log returns the configured Log, or "none"
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

func (c *Config) writeToFile() error {
	fp, err := os.OpenFile(getConfDir(c.Name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer fp.Close()
	// NOTE(halfwit) We always want an extra newline to separate entries
	if _, e := fmt.Fprintf(fp, "%s\n\n", c.String()); e != nil {
		return e
	}

	return nil
}

// GetLogDir returns a canonical directory for a user log, searching first altid/config
// If no entry is found or the file is missing, it will return "none"
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

func createConfigFile(c Configurator, service string) (*Config, error) {
	if c == nil {
		return nil, errors.New(ErrBadConfigurator)
	}

	fp, err := os.OpenFile(getConfDir(service), os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	fp.Close()
	return buildConfigEntry(c, service)
}

func buildConfigEntry(c Configurator, service string) (*Config, error) {
	rw := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}

	conf, err := c(rw)
	if err != nil {
		return nil, err
	}

	conf.Name = service

	if e := conf.writeToFile(); e != nil {
		return nil, e
	}

	return conf, nil
}
