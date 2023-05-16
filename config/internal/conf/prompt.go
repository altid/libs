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

type Prompter interface {
	Query(*request.Request) (*entry.Entry, error)
}

// Prompt just queries for things
type Prompt struct {
	debug func(string, ...any)
}

func NewPrompt(debug func(string, ...any)) *Prompt {
	return &Prompt{
		debug: debug,
	}
}

func (p *Prompt) Query(req *request.Request) (*entry.Entry, error) {
	key := strings.ToLower(req.Key)
	entry := &entry.Entry{
		Key: key,
	}

	p.debug("request key=\"%s\" default=\"%v\"", key, req.Defaults)

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
			p.debug("response key=\"%s\" value=\"%v\"", entry.Key, entry.Value)
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
		if err != nil {
			p.debug("error: %v\n", err)
		}
		p.debug("response key=\"%s\" value=\"%t\"", entry.Key, entry.Value)
	case string:
		entry.Value = value
		p.debug("response key=\"%s\" value=\"%s\"", entry.Key, entry.Value)
	case types.Auth:
		p.debug("response key=\"%s\" value=\"%s\"", entry.Key, value)
		entry.Value = types.Auth(value)
	case types.Logdir:
		p.debug("response key=\"%s\" value=\"%s\"", entry.Key, value)
		entry.Value = types.Logdir(value)
	case types.ListenAddress:
		p.debug("response key=\"%s\" value=\"%s\"", entry.Key, value)
		entry.Value = types.ListenAddress(value)
	case float32:
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, err
		}

		entry.Value = v
		p.debug("response key=\"%s\" value=\"%f\"", v)
	case float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}

		entry.Value = v
		p.debug("response key=\"%s\" value=\"%f\"", v)
	default:
		v, e := tryInt(req.Defaults, value)
		if e != nil {
			return nil, e
		}

		entry.Value = v
		p.debug("response key=\"%s\" value=\"%d\"", entry.Key, entry.Value)
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

func tryInt(req any, value string) (v any, err error) {
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
