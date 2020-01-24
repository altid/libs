// build !plan9

package main

import (
	"context"
	"io"
	"log"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
)

var tails map[string]*tail

func tailEvents(ctx context.Context, services map[string]*service) (chan *event, error) {
	events := make(chan *event)
	tails = make(map[string]*tail)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				if _, ok := tails[ev.Name]; !ok {
					log.Printf("Unknown event: %s", ev.Name)
					continue
				}
				for _, event := range tails[ev.Name].readlines() {
					events <- event
				}
			case err := <-watcher.Errors:
				ctx.Done()
				log.Print(err)
			}
		}
	}()
	for _, svc := range services {
		dir := path.Join(*inpath, svc.name, "event")
		f, err := os.Open(dir)
		if err != nil {
			log.Printf("%s: Entry found, but no service running\n", svc.name)
			continue
		}
		stat, err := f.Stat()
		if err != nil {
			return nil, err
		}
		f.Seek(0, io.SeekEnd)
		tail := &tail{
			fd:   f,
			name: svc.name,
			size: stat.Size(),
		}
		tails[dir] = tail
		watcher.Add(dir)
	}
	return events, nil
}
