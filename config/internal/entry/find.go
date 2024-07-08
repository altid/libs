package entry

import (
	"github.com/altid/libs/config/internal/request"
)

func Find(req *request.Request, entries []*Entry) (*Entry, bool) {
	for _, entry := range entries {
		if entry.Key == req.Key {
			return entry, true
		}
	}
	return nil, false
}
