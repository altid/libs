package conf

import (
	"fmt"
	"os"
	"strings"

	"github.com/altid/libs/config/internal/entry"
	"github.com/altid/libs/config/internal/util"
)

type Conf struct {
	name    string
	entries []*entry.Entry
}

func FromConfig(debug func(string, ...interface{}), service string, confdir string) ([]*entry.Entry, error) {
	return entry.FromConfig(debug, service, confdir)
}

func (c *Conf) String() string {
	var entry strings.Builder

	fmt.Fprintf(&entry, "service=%s", c.name)

	for _, item := range c.entries {
		switch item.Value.(type) {
		case int, int64:
			fmt.Fprintf(&entry, " %s=%d", item.Key, item.Value)
		case bool:
			fmt.Fprintf(&entry, " %s=%t", item.Key, item.Value)
		default:
			fmt.Fprintf(&entry, " %s=%s", item.Key, item.Value)
		}
	}

	return entry.String()
}

func (c *Conf) WriteToFile() error {

	fp, err := os.OpenFile(util.GetConf(c.name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
