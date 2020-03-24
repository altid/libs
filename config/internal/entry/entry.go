package entry

import (
	"errors"
	"strconv"

	"github.com/altid/libs/config/internal/util"
	"github.com/mischief/ndb"
)

// Errors
const (
	ErrBadConfigurator = "configurator nil or invalid, cannot continue"
	ErrNoConfigure     = "unable to find or create config for this service. To create one, please run %s -conf"
	ErrNoSuchKey       = "no such key"
	ErrNoEntries       = "unable to find config entry for this service."
	ErrMultiEntries    = "config contains duplicate entries for this service. Please edit your altid/config file"
)

type Entry struct {
	Key   string
	Value interface{}
}

func FromConfig(debug func(string, ...interface{}), service string, confdir string) ([]*Entry, error) {
	dir := util.GetConf(service)

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
func fromNdb(debug func(string, ...interface{}), recs ndb.RecordSet, service string) ([]*Entry, error) {
	var values []*Entry

	for _, tup := range recs[0] {
		v := &Entry{
			Key:   tup.Attr,
			Value: tup.Val,
		}

		switch tup.Attr {
		case "auth":
			pass, err := findAuth(debug, service, recs)
			if err != nil {
				return nil, err
			}
			v.Value = pass
		case "logdir":
			v.Value = findLogdir(debug, recs)
		case "listen_dir":
			v.Value = findListen(debug, recs)
		}

		// Bool can't fail when it's true or false
		if tup.Val == "true" || tup.Val == "false" {
			v.Value, _ = strconv.ParseBool(tup.Val)
		}

		// Since we don't have access to the req type, make sure we
		// cast back to string in the case where numerical strings exist
		// so this is a benign, albeit expensive step to ensure we have
		// an int type when we really want it
		if num, err := strconv.ParseInt(tup.Val, 10, 0); err == nil {
			v.Value = int(num)
		}

		values = append(values, v)
	}

	return values, nil
}
