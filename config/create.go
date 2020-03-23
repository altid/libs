package config

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
)

type create struct {
	name    string
	entries []*entry
}

func (c *create) String() string {
	var entry strings.Builder

	fmt.Fprintf(&entry, "service=%s", c.name)

	for _, item := range c.entries {
		fmt.Fprintf(&entry, " %s=%s", item.key, item.value)
	}

	return entry.String()
}

func (c *create) writeToFile() error {

	fp, err := os.OpenFile(getConf(c.name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

// Monster function, clean up later
func createConfFile(debug func(string, ...interface{}), service string, have []*entry, want []*request) (*create, error) {
	var entries []*entry

	// Open up stdin/stdout in case we need access to them to create entries
	rw := bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout))

	// Range through and fill each entry with either the config data
	// or query the user for input
	for _, item := range want {
		// if an entry exists in the conf, don't create another
		if entry, ok := haveEntry(debug, item, have); ok {
			entries = append(entries, entry)
			continue
		}

		entry, err := fillEntry(debug, rw, item)
		if err != nil {
			return nil, err
		}

		// We want to catch auth and tls, since they require additional fields
		switch item.defaults.(type) {
		case Auth:
			if entry.value == "password" {
				i := &request{
					key:      "password",
					prompt:   "Enter password:",
					defaults: "none",
				}

				if entry, ok := haveEntry(debug, i, have); ok {
					entries = append(entries, entry)
					continue
				}

				pass, err := fillEntry(debug, rw, i)
				if err != nil {
					return nil, err
				}

				entries = append(entries, pass)
			}
		// We just create the ndb entries here, not the actual certificate
		case tls.Certificate:
			k := &request{
				key:    "keyfile",
				prompt: "Enter the path to your TLS Key file",
			}

			if entry, ok := haveEntry(debug, k, have); ok {
				entries = append(entries, entry)
			} else {
				key, err := fillEntry(debug, rw, k)
				if err != nil {
					return nil, err
				}

				entries = append(entries, key)
			}

			c := &request{
				key:    "certfile",
				prompt: "Enter the path to your TLS Certificate file",
			}

			if entry, ok := haveEntry(debug, c, have); ok {
				entries = append(entries, entry)
			} else {
				entry, err = fillEntry(debug, rw, c)
				if err != nil {
					return nil, err
				}
				entries = append(entries, entry)
				continue
			}
		}

		entries = append(entries, entry)
	}

	c := &create{
		name:    service,
		entries: entries,
	}

	return c, nil
}
