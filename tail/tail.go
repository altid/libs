// build !plan9

package tail

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
)

func watchEvents(ctx context.Context, dir, service string) (chan *Event, error) {
	events := make(chan *Event)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fp := path.Join(dir, service, "event")

	f, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("%s: entry found, but no service running", service)
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	_, err = f.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	go func() {
		t := &tail{
			fd:   f,
			name: service,
			size: stat.Size(),
		}

		for {
			select {
			case <-watcher.Events:
				eventlist, err := t.readlines()
				if err != nil {
					ctx.Done()
					log.Print(err)
					break
				}
				for _, event := range eventlist {
					events <- event
				}
			case err := <-watcher.Errors:
				ctx.Done()
				log.Print(err)
				break
			}
		}
	}()

	if err = watcher.Add(fp); err != nil {
		return nil, err
	}

	return events, nil
}
