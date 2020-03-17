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
	"strings"

	"github.com/altid/libs/auth"
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
	debug  func(value string, args ...interface{})
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
	if debug {
		configLogger("mock starting")
	}

	conf, err := createConfigFile(c, service)
	if err != nil {
		return nil, err
	}

	if debug {
		conf.debug = configLogger
	} else {
		conf.debug = func(string, ...interface{}) {}
	}

	conf.debug("success")

	return conf, nil
}

// New returns a valid config for a given service. If one is not found, the Configurator
// will be called to interactively create one
func New(c Configurator, service string, debug bool) (*Config, error) {
	// Since this is a library to create services, we can expect there to be flags passed in
	if debug {
		configLogger("starting")
	}

	flag.Parse()
	if *createConfig {
		if debug {
			configLogger("start configurator")
		}

		conf, err := createConfigFile(c, service)
		if err != nil {
			return nil, err
		}

		if e := writeToFile(conf); e != nil {
			return nil, e
		}

		return conf, err
	}

	conf, err := ndb.Open(getConfDir(service))
	if err != nil {
		return nil, fmt.Errorf(ErrNoConfigure, os.Args[0])
	}

	if entry := conf.Search("service", service).Search("service"); entry == "" {

		return nil, fmt.Errorf(ErrNoConfigure, os.Args[0])
	}

	cf, err := parseConfig(service)
	if err != nil {
		return nil, err
	}

	if debug {
		cf.debug = configLogger
	} else {
		cf.debug = func(string, ...interface{}) {}
	}

	cf.debug("success")
	return cf, nil
}

// Password queries the database for a password
// the format in the config is auth=pass=mypassword or auth=factotum
// when auth=factotum is found, it will attempt to query the factotum for a response
func (c *Config) Password() (string, error) {
	// The factotum bits should really move out of here
	c.debug("request key=\"password\"")

	pass, err := c.Search("auth")
	if err != nil {
		c.debug("response key=\"password\" error=\"%v\"", err)
		return "", err
	}

	if len(pass) > 5 && pass[:5] == "pass=" {
		pass = pass[5:]
	}

	if pass == "factotum" {
		c.debug("request key=\"factotum\"")
		userPwd, err := auth.Getuserpasswd(
			"proto=pass service=%s",
			c.Name,
		)
		if err != nil {
			c.debug("response key=\"password\" error=\"%v\"", err)
			return "", err
		}

		pass = userPwd.Password
		c.debug("response key=\"factotum\" value=\"success\"")
	}

	c.debug("response key=\"password\" value=\"success\"")
	return pass, nil
}

// SSLCert returns a tls.Certificate based on successfully finding
// a cert and key file listen in the configuration
func (c *Config) SSLCert() (tls.Certificate, error) {
	c.debug("request type=ssl")
	cert, err := c.Search("cert")
	if err != nil {
		c.debug("response type=ssl error=\"%v\"", err)
		return tls.Certificate{}, err
	}

	key, err := c.Search("key")
	if err != nil {
		c.debug("response type=ssl error=\"%v\"", err)
		return tls.Certificate{}, err
	}

	c.debug("response type=ssl value=\"success\"")
	return tls.LoadX509KeyPair(cert, key)
}

// Search queries for an entry in the config matching key
// Returning the value if it exists, or a "no such key" error
func (c *Config) Search(key string) (string, error) {
	c.debug("request key=\"%s\"", key)

	if key == "service" {
		c.debug("response key=\"%s\" value=\"%s\"", key, c.Name)
		return c.Name, nil
	}

	for _, k := range c.Values {
		if k.Key == key {
			c.debug("response key=\"%s\" value=\"%s\"", key, k.Value)
			return k.Value, nil
		}
	}

	c.debug("response key=\"%s\" error=\"%s\"", ErrNoSuchKey)
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
	c.debug("request key=\"log\"")

	dir, err := c.Search("log")
	if err != nil {
		c.debug("response key=\"log\" value=\"none\"")
		return "none"
	}

	c.debug("response key=\"log\" value=\"%s\"", dir)

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

func configLogger(format string, v ...interface{}) {
	l := log.New(os.Stdout, "config: ", 0)
	l.Printf(format+"\n", v...)
}
