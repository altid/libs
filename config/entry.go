package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mischief/ndb"
)

type entry struct {
	key   string
	value interface{}
}

func haveEntry(debug func(string, ...interface{}), req *request, entries []*entry) (*entry, bool) {
	// TODO(halfwit) Also ensure that the types match
	for _, entry := range entries {
		if entry.key == req.key {
			return entry, true
		}
	}

	return nil, false
}

func fillEntry(debug func(string, ...interface{}), rw io.ReadWriter, req *request) (*entry, error) {
	key := strings.ToLower(req.key[:1]) + req.key[1:]
	entry := &entry{
		key: key,
	}

	debug("request key=\"%s\" default=\"%s\"", key, req.defaults)

	switch {
	case req.defaults == nil:
		return nil, errors.New("request defaults cannot be nil")
	case req.prompt == "":
		entry.value = req.defaults
		fmt.Fprintf(rw, "using default %s=%v\n", key, req.defaults)
		return entry, nil
	default:
		fmt.Fprintf(rw, "%s [%v]: (press enter for default)\n", req.prompt, req.defaults)
	}

	rd := bufio.NewReader(rw)

	value, err := rd.ReadString('\n')
	if err != nil {
		return nil, err
	}

	// User pressed enter for default
	if value == "" || value == "\n" {
		entry.value = req.defaults
		debug("response key=\"%s\" value=\"%v\"", entry.key, entry.value)
		return entry, nil
	}

	switch req.defaults.(type) {
	case int:
		entry.value, err = strconv.Atoi(value)
		debug("response key=\"%s\" value=\"%d\"", entry.key, entry.value)
	case bool:
		entry.value, err = strconv.ParseBool(value)
		debug("response key=\"%s\" value=\"%t\"", entry.key, entry.value)
	case string:
		entry.value = value
	case Auth, Logdir, ListenAddress:
		entry.value = value
		debug("response key=\"%s\" value=\"%s\"", entry.key, entry.value)
	}

	return entry, nil
}

func fromConfig(debug func(string, ...interface{}), service string, confdir string) ([]*entry, error) {
	dir := getConf(service)

	if confdir != "" {
		dir = confdir
	}
	
	conf, err := ndb.Open(dir)
	if err != nil {
		return nil, err
	}

	recs := conf.Search("service", service)

	switch len(recs) {
	case 0:
		return nil, errors.New(ErrNoEntries)
	case 1:
		return fromNdb(debug, recs, service)
	default:
		return nil, errors.New(ErrMultiEntries)
	}
}

// This will error if auth=password has no complementary password=field
// And if the tls.FromX509Pair
func fromNdb(debug func(string, ...interface{}), recs ndb.RecordSet, service string) ([]*entry, error) {
	var values []*entry

	for _, tup := range recs[0] {
		switch tup.Attr {
		case "auth":
			pass, err := findAuth(debug, service, recs)
			if err != nil {
				return nil, err
			}

			tup.Val = pass
		case "logdir":
			tup.Val = findLogdir(debug, recs)
		case "listen_dir":
			tup.Val = findListen(debug, recs)
		}

		v := &entry{
			key:   tup.Attr,
			value: tup.Val,
		}

		// Bool can't fail when it's true or false
		if tup.Val == "true" || tup.Val == "false" {
			v.value, _ = strconv.ParseBool(tup.Val)
		}

		// Since we don't have access to the req type, make sure we
		// cast back to string in the case where numerical strings exist
		// so this is a benign, albeit expensive step to ensure we have
		// an int type when we really want it
		if num, err := strconv.ParseInt(tup.Val, 0, 0); err == nil {
			v.value = int(num)
		}

		values = append(values, v)
	}

	return values, nil
}
