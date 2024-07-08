package conf

import (
	"fmt"
	"reflect"

	"github.com/altid/libs/config/internal/entry"
	"github.com/altid/libs/config/internal/request"
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
		var en *entry.Entry
		en, err = query(item, p, have)
		if err != nil {
			return err
		}
		if e := push(requests, en); e != nil {
			return e
		}
	}
	return nil
}

func query(item *request.Request, p Prompter, have []*entry.Entry) (*entry.Entry, error) {
	if p != nil && item.Prompt != "no_prompt" {
		return p.Query(item)
	}
	if en, ok := entry.Find(item, have); ok {
		return en, pick(item, en)
	}
	if item.Defaults != nil {
		entry := &entry.Entry{
			Key:   item.Key,
			Value: item.Defaults,
		}
		return entry, nil
	}
	return nil, fmt.Errorf("no config entry for %s", item.Key)
}

// Add the actual value to the struct, being careful that
// the strconv.Atoi from before is guarded against
func push(requests any, entry *entry.Entry) error {
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
