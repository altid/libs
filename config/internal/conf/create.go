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
func Create(debug func(string, ...interface{}), service string, have []*entry.Entry, want []*request.Request, configFile string) (*Conf, error) {
	var entries []*entry.Entry

	// Range through and fill each entry with either the config data
	// or query the user for input
	for _, item := range want {
		// if an entry exists in the conf, don't create another
		if entry, ok := entry.Find(item, have); ok && entry.Key != "auth" {
			entries = append(entries, entry)
			continue
		}

		// This is ugly, we'll have to make this a separate function once we have time
		switch item.Defaults.(type) {
		case types.Auth:
			// Clean this up later if we can
			item.Pick = []string{"password", "factotum", "none"}
			if _, ok := entry.Find(item, have); !ok && item.Defaults.(types.Auth) == "password" {
				// We have no entry, check the defaults and send if there's a password
				i := &request.Request{
					Key:      "password",
					Prompt:   "Enter password:",
					Defaults: "password",
				}

				pass, err := fillEntry(debug, i)
				if err != nil {
					return nil, err
				}

				entries = append(entries, pass)
			} else {
				// we have an entry, see if it was set to password in the config
				en := entry.FixAuth(service, configFile)
				if en.Value.(string) == "password" {
					i := &request.Request{
						Key:      "password",
						Prompt:   "Enter password:",
						Defaults: "password",
					}

					if d, ok := entry.Find(i, have); ok {
						entries = append(entries, d)
					} else {
						pass, err := fillEntry(debug, i)
						if err != nil {
							return nil, err
						}

						entries = append(entries, pass)
					}
				}
			}
		}
		entry, err := fillEntry(debug, item)
		if err != nil {
			return nil, err
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

	debug("request key=\"%s\" default=\"%v\"", key, req.Defaults)

	switch {
	case req.Defaults == nil:
		return nil, errors.New("request defaults cannot be nil")
	case len(req.Prompt) < 1:
		entry.Value = req.Defaults
		fmt.Printf("using default %s=%v\n", key, req.Defaults)
		return entry, nil
	case len(req.Pick) > 0:
		fmt.Printf("%s (%s) [%v]: (press enter for default)\n", req.Prompt, strings.Join(req.Pick, "|"), req.Defaults)
	default:
		fmt.Printf("%s [%v]: (press enter for default)\n", req.Prompt, req.Defaults)
	}

	var value string
	var err error

	for i := 0; i < 3; i++ {
		value, err = readValue()
		if err != nil {
			return nil, err
		}

		// User pressed enter for default
		if value == "" || value == "\n" {
			entry.Value = req.Defaults
			debug("response key=\"%s\" value=\"%v\"", entry.Key, entry.Value)
			return entry, nil
		}

		if checkPicks(value, req.Pick) {
			break
		}

		if i < 2 {
			fmt.Println("unknown value selected, try again.")
			continue
		}

		return nil, errors.New("multiple unknown values entered, exiting")
	}

	switch req.Defaults.(type) {
	case bool:
		entry.Value, err = strconv.ParseBool(value)
		debug("response key=\"%s\" value=\"%t\"", entry.Key, entry.Value)
	case string:
		entry.Value = value
		debug("response key=\"%s\" value=\"%s\"", entry.Key, entry.Value)
	case types.Auth:
		debug("response key=\"%s\" value=\"%s\"", entry.Key, value)
		entry.Value = types.Auth(value)
	case types.Logdir:
		debug("response key=\"%s\" value=\"%s\"", entry.Key, value)
		entry.Value = types.Logdir(value)
	case types.ListenAddress:
		debug("response key=\"%s\" value=\"%s\"", entry.Key, value)
		entry.Value = types.ListenAddress(value)
	case float32:
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, err
		}

		entry.Value = v
		debug("response key=\"%s\" value=\"%f\"", v)
	case float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}

		entry.Value = v
		debug("response key=\"%s\" value=\"%f\"", v)
	default:
		v, e := tryInt(req.Defaults, value)
		if e != nil {
			return nil, e
		}

		entry.Value = v
		debug("response key=\"%s\" value=\"%d\"", entry.Key, entry.Value)
	}

	return entry, nil
}

func readValue() (string, error) {
	rd := bufio.NewReader(os.Stdin)

	value, err := rd.ReadString('\n')
	if err != nil {
		return "", err
	}

	value = value[:len(value)-1]
	return value, nil
}

func tryInt(req interface{}, value string) (v interface{}, err error) {
	switch req.(type) {
	case int:
		v, err = strconv.Atoi(value)
	case uint:
		v, err = strconv.ParseUint(value, 0, 0)
	case int8:
		v, err = strconv.ParseInt(value, 0, 8)
		if err != nil {
			return nil, err
		}

		v = int8(v.(int))
	case uint8:
		v, err = strconv.ParseUint(value, 0, 8)
		if err != nil {
			return nil, err
		}

		v = uint8(v.(uint))
	case int16:
		v, err = strconv.ParseInt(value, 0, 16)
		if err != nil {
			return nil, err
		}

		v = int16(v.(int))
	case uint16:
		v, err = strconv.ParseUint(value, 0, 16)
		if err != nil {
			return nil, err
		}

		v = uint16(v.(uint))
	case int32:
		v, err = strconv.ParseInt(value, 0, 32)
		if err != nil {
			return nil, err
		}

		v = int32(v.(int))
	case uint32:
		v, err = strconv.ParseUint(value, 0, 32)
		if err != nil {
			return nil, err
		}

		v = uint32(v.(uint))
	case int64:
		v, err = strconv.ParseInt(value, 0, 64)
		if err != nil {
			return nil, err
		}

		v = int64(v.(int))
	case uint64:
		v, err = strconv.ParseUint(value, 0, 64)
		if err != nil {
			return nil, err
		}

		v = uint64(v.(uint))
	}

	return
}

func checkPicks(value string, picks []string) bool {
	if len(picks) < 1 {
		return true
	}

	for _, pick := range picks {
		if value == pick {
			return true
		}
	}

	return false
}
