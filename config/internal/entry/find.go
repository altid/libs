package entry

import (
	"errors"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/config/internal/request"
	"github.com/altid/libs/config/types"
	"github.com/mischief/ndb"
)

func Find(req *request.Request, entries []*Entry) (*Entry, bool) {
	for _, entry := range entries {
		if entry.Key == req.Key {
			return entry, true
		}
	}

	return nil, false
}

func findAuth(debug func(string, ...interface{}), service string, c ndb.RecordSet) (types.Auth, error) {
	switch c.Search("auth") {
	case "password":
		pass := c.Search("password")
		if pass == "" {
			return "", errors.New("unable to find password")
		}

		return types.Auth(pass), nil
	case "factotum":
		debug("request key=\"factotum\"")
		userPwd, err := auth.Getuserpasswd("proto=pass service=%s", service)
		if err != nil {
			debug("response key=\"password\" error=\"%v\"", err)
			return "", err
		}

		debug("response key=\"factotum\" value=\"success\"")
		return types.Auth(userPwd.Password), nil
	// If we're here, either we have a "none" value, or auth isn't listed
	default:
		return "", nil
	}
}

func findLogdir(debug func(string, ...interface{}), c ndb.RecordSet) types.Logdir {
	debug("request key=\"logdir\"")

	dir := c.Search("logdir")
	if dir == "" {
		debug("response key=\"logdir\" value=\"none\"")
		return "none"
	}

	debug("response key=\"logdir\" value=\"%s\"", dir)
	return types.Logdir(dir)
}

func findListen(debug func(string, ...interface{}), c ndb.RecordSet) types.ListenAddress {
	debug("request key=\"listen_address\"")
	dir := c.Search("listen_address")
	if dir == "" {
		debug("response key=\"listen_address\" value=\"none\"")
		return "none"
	}

	debug("response key=\"listen_address\" value=\"%s\"", dir)
	return types.ListenAddress(dir)
}
