package build

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/altid/libs/config/internal/entry"
	"github.com/altid/libs/config/internal/request"
	"github.com/altid/libs/config/types"
)

func Marshal(debug func(string, ...interface{}), requests interface{}, have []*entry.Entry) error {
	// Loop through entries and attempt to fill the struct
	// Make sure we turn any ints that are supposed to be strings,
	// back into strings!
	want, err := request.Build(requests)
	if err != nil {
		return err
	}

	for _, item := range want {
		// We need special handling for Auth
		// as the en.Value will be populated with a usable value
		// either from the password field or factotum
		if item.Key == "auth" {
			for _, pick := range item.Pick {
				switch pick {
				case "password", "factotum", "none":
					break
				default:
					return errors.New("unsupported auth mechanism selected")
				}
			}

			if en, ok := entry.Find(item, have); ok {
				if e := push(requests, en); e != nil {
					return e
				}

				continue
			}
		}

		// We have a good item
		if en, ok := entry.Find(item, have); ok {
			if e := pick(item, en); e != nil {
				return e
			}

			if e := push(requests, en); e != nil {
				return e
			}

			continue
		}

		if item.Defaults != nil {
			entry := &entry.Entry{
				Key:   item.Key,
				Value: item.Defaults,
			}

			if e := push(requests, entry); e != nil {
				return e
			}

			continue
		}

		return fmt.Errorf("missing config entry for %s", item.Key)
	}
	return nil
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
