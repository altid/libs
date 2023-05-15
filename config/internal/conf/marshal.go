package conf

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/altid/libs/config/internal/entry"
	"github.com/altid/libs/config/internal/request"
	"github.com/altid/libs/config/types"
)

func Marshal(debug func(string, ...any), requests any, have []*entry.Entry, p Prompter) error {
	// Loop through entries and attempt to fill the struct
	// Make sure we turn any ints that are supposed to be strings,
	// back into strings!
	want, err := request.Build(requests)
	if err != nil {
		return err
	}

	for _, item := range want {
		en, err := query(item, p, have)
		if err != nil {
			return err
		}
		if item.Key == "auth" {
			// We handle password, factotum, and none types currently
			for _, pick := range item.Pick {
				switch pick {
				case "password", "factotum", "none":
				default:
					return errors.New("unsupported auth mechanism selected")
				}
			}
			// Try to request the password= entry
			if string(en.Value.(types.Auth)) == "password" {
				i := & request.Request{
					Key:	  "password",
					Prompt:	  "Enter password:",
					Defaults: "password",
				}
				pw, err := query(i, p, have)
				if err != nil {
					return err
				}
				en.Value = types.Auth(pw.Value.(string))
			}
			// Try to request the token= entry
			if string(en.Value.(types.Auth)) == "token" {
				i := &request.Request{
					Key:		"token",
					Prompt:		"Enter token:",
					Defaults:	"",
				}
				token, err := query(i, p, have)
				if err != nil {
					return err
				}
				en.Value = types.Auth(token.Value.(string))
			}
		}
		if e := push(requests, en); e != nil {
			return e
		}
	}
	return nil
}

func query(item *request.Request, p Prompter, have []*entry.Entry) (*entry.Entry, error) {
	if(p != nil && item.Prompt != "no_prompt") {
		return p.Query(item)
	}
	if en, ok := entry.Find(item, have); ok {
		return en, pick(item, en)
	}
	if item.Defaults != nil {
		entry := &entry.Entry{
			Key:	item.Key,
			Value:	item.Defaults,
		}
		return entry, nil
	}
	return nil, fmt.Errorf("no config entry for %s", item.Key)
}

// Add the actual value to the struct, being careful that
// the strconv.Atoi from before is guarded against
func push(requests interface{}, entry *entry.Entry) error {
	s := reflect.ValueOf(requests)
	t := reflect.Indirect(s).Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.FieldByIndex([]int{i})
		tag := f.Tag.Get("altid")
		if tag == "" {
			continue
		}
		// Use the tag, Luke
		req, err := request.ParseTag(tag)
		if err != nil {
			continue
		}
		if req.Key != entry.Key {
			continue
		}
		d := reflect.Indirect(s).Field(i)
		switch f.Type.Name() {
		case "string":
			d.SetString(entry.Value.(string))
		case "Auth":
			d.Set(reflect.ValueOf(entry.Value.(types.Auth)))
		case "Logdir":
			d.Set(reflect.ValueOf(entry.Value.(types.Logdir)))
		case "ListenAddress":
			d.Set(reflect.ValueOf(entry.Value.(types.ListenAddress)))
		default:
			d.Set(reflect.ValueOf(entry.Value))
		}
		return nil
	}
	return fmt.Errorf("could not add %s entry to struct", entry.Key)
}

// We want to make sure request.Pick matches the value of entry.Value
func pick(req *request.Request, entry *entry.Entry) error {
	v := entry.String()

	if len(req.Pick) < 1 {
		return nil
	}

	for _, pick := range req.Pick {
		if pick == v {
			return nil
		}
	}

	return fmt.Errorf("invalid choice for %s", entry.Key)
}
