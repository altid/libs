package conf

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/altid/libs/config/internal/entry"
	"github.com/altid/libs/config/internal/request"
	"github.com/altid/libs/config/types"
)

// Monster function, clean up later
func Create(debug func(string, ...interface{}), service string, have []*entry.Entry, want []*request.Request) (*Conf, error) {
	var entries []*entry.Entry

	// Range through and fill each entry with either the config data
	// or query the user for input
	for _, item := range want {
		// if an entry exists in the conf, don't create another
		if entry, ok := entry.Find(item, have); ok {
			entries = append(entries, entry)
			continue
		}

		entry, err := fillEntry(debug, item)
		if err != nil {
			return nil, err
		}

		// We want to catch auth, since it require additional fields
		switch item.Defaults.(type) {
		case types.Auth:
			if entry.Value.(types.Auth) == "password" {
				i := &request.Request{
					Key:      "password",
					Prompt:   "Enter password:",
					Defaults: "none",
				}

				pass, err := fillEntry(debug, i)
				if err != nil {
					return nil, err
				}

				entries = append(entries, pass)
			}
		}

		entries = append(entries, entry)
	}

	c := &Conf{
		name:    service,
		entries: entries,
	}

	return c, nil
}

func fillEntry(debug func(string, ...interface{}), req *request.Request) (*entry.Entry, error) {
	key := strings.ToLower(req.Key)
	entry := &entry.Entry{
		Key: key,
	}

	debug("request key=\"%s\" default=\"%s\"", key, req.Defaults)

	switch {
	case req.Defaults == nil:
		return nil, errors.New("request defaults cannot be nil")
	case len(req.Prompt) < 1:
		entry.Value = req.Defaults
		fmt.Printf("using default %s=%v\n", key, req.Defaults)
		return entry, nil
	default:
		fmt.Printf("%s [%v]: (press enter for default)\n", req.Prompt, req.Defaults)
	}

	rd := bufio.NewReader(os.Stdin)

	value, err := rd.ReadString('\n')
	if err != nil {
		return nil, err
	}

	value = value[:len(value)-1]

	// User pressed enter for default
	if value == "" || value == "\n" {
		entry.Value = req.Defaults
		debug("response key=\"%s\" value=\"%v\"", entry.Key, entry.Value)
		return entry, nil
	}

	switch req.Defaults.(type) {
	case int:
		entry.Value, err = strconv.Atoi(value)
		debug("response key=\"%s\" value=\"%d\"", entry.Key, entry.Value)
	case bool:
		entry.Value, err = strconv.ParseBool(value)
		debug("response key=\"%s\" value=\"%t\"", entry.Key, entry.Value)
	case string:
		entry.Value = value
		debug("response key=\"%s\" value=\"%s\"", entry.Key, entry.Value)
	case types.Auth:
		entry.Value = types.Auth(value)
		debug("response key=\"%s\" value=\"%s\"", entry.Key, value)
	case types.Logdir:
		entry.Value = types.Logdir(value)
		debug("response key=\"%s\" value=\"%s\"", entry.Key, value)
	case types.ListenAddress:
		entry.Value = types.ListenAddress(value)
		debug("response key=\"%s\" value=\"%s\"", entry.Key, value)

	}

	return entry, nil
}
