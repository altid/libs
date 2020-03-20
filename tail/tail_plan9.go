// build plan9
package tail

import (
	"context"
	"log"
	"os"
	"path"
	"time"
)

type runner struct {
	t      *tail
	events chan *Event
}

func (r *runner) item() error {
	hs, err := r.t.fd.Stat()
	if err != nil {
		return err
	}

	if hs.Size() == r.t.size {
		defer time.Sleep(250 * time.Millisecond)
		return nil
	}

	eventlist, err := r.t.readlines()
	if err != nil {
		return err
	}

	for _, event := range eventlist {
		r.events <- event
	}

	r.t.size = hs.Size()
	return nil
}

func watchEvents(ctx context.Context, dir, service string) (chan *Event, error) {
	events := make(chan *Event)

	dir = path.Join(dir, service, "event")

	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	go func(f *os.File, stat os.FileInfo, events chan *Event) {
		// Make sure we tail from the end of the file
		tail := &tail{
			fd:   f,
			name: service,
			size: stat.Size(),
		}

		defer f.Close()

		r := &runner{
			t:      tail,
			events: events,
		}

		for {
			select {
			case <-ctx.Done():
				close(events)
				break
			default:
				if e := r.item(); e != nil {
					log.Println(e)
					return
				}
			}
		}
	}(f, stat, events)

	return events, nil
}
