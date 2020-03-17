package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
)

var debugLogger func(string, ...interface{})

type request struct {
	key      string
	prompt   string
	defaults string
}

// Repl accepts a struct with String/Int/Bool types
// The fields should contain the default values, which will be presented to the client during the repl
//
// For example:
//
//		mytype := struct {
//			Address string `IP address to dial`
//			Nickname string `Username for service`
//			Password string `Password for service`
//		}{"127.0.0.1", "guest", "hunter2"}
//
// will start a repl with the client over the io.ReadWriter, using the struct tags as the prompt
//
//		$ IP address to dial [127.0.0.1]: (enter to use default)
//		$ Username for service[guest]: (enter to use default)
//		$ Password for service[hunter2]: (enter to use default)
//
// On successful completion, a Config will be returned
// If debug is true Repl will output debug information
func Repl(rw io.ReadWriter, req interface{}, debug bool) (*Config, error) {
	var values []*Entry

	c := new(Config)

	if debug {
		debugLogger = replLogger
	} else {
		debugLogger = func(string, ...interface{}) {}
	}

	debugLogger("starting")
	reqs, err := toArray(req)
	if err != nil {
		return nil, err
	}

	if len(reqs) < 1 {
		return nil, errors.New("unable to find any configuration entries")
	}

	for _, req := range reqs {
		entry, err := runRepl(rw, req)
		if err != nil {
			fmt.Fprintf(rw, "%v\n", err)
			continue
		}

		values = append(values, entry)
	}

	c.Values = values
	debugLogger("success")

	return c, nil
}

func runRepl(rw io.ReadWriter, req request) (*Entry, error) {
	key := strings.ToLower(req.key[:1]) + req.key[1:]
	entry := &Entry{
		Key: key,
	}

	debugLogger("request key=\"%s\" default=\"%s\"", key, req.defaults)

	switch {
	case req.prompt == "":
		entry.Value = req.defaults
		fmt.Fprintf(rw, "found no struct tag for key \"%s\", using default \"%s\"\n", key, req.defaults)
		return entry, nil
	case req.defaults == "":
		fmt.Fprintf(rw, "%s\n", req.prompt)
	default:
		fmt.Fprintf(rw, "%s [%s]: (press enter for default)\n", req.prompt, req.defaults)
	}

	rd := bufio.NewReader(rw)

	value, err := rd.ReadString('\n')
	if err != nil {
		return nil, err
	}

	entry.Value = value

	// User pressed enter for default
	if value == "" || value == "\n" {
		entry.Value = req.defaults
	}

	debugLogger("response key=\"%s\" value=\"%s\"", entry.Key, entry.Value)

	return entry, nil
}

func toArray(req interface{}) ([]request, error) {
	var reqs []request
	t := reflect.TypeOf(req)
	s := reflect.ValueOf(req)
	for i := 0; i < t.NumField(); i++ {
		f := t.FieldByIndex([]int{i})
		req := request{
			key:    f.Name,
			prompt: string(f.Tag),
		}
		switch f.Type.Name() {
		case "string":
			req.defaults = s.Field(i).Interface().(string)
		case "int":
			i := s.Field(i).Interface().(int)
			req.defaults = fmt.Sprintf("%d", i)
		case "bool":
			b := s.Field(i).Interface().(bool)
			req.defaults = fmt.Sprintf("%t", b)
		default:
			return nil, fmt.Errorf("unknown type for config: %s", f.Type.Name())
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

func replLogger(format string, v ...interface{}) {
	l := log.New(os.Stdout, "repl: ", 0)
	l.Printf(format+"\n", v...)
}
