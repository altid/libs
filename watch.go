// build !plan9

package main

import (
	"io"
	"log"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
)

var tails map[string]*tail

func listen(cfg *config) (chan *event, chan struct{}) {
	events := make(chan *event)
	done := make(chan struct{})
	tails = make(map[string]*tail)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				if _, ok := tails[ev.Name]; !ok {
					log.Printf("Unknown event: %s", ev.Name)
					continue
				}

				events <- tails[ev.Name].readlines()
			case err := <-watcher.Errors:
				close(done)
				log.Print(err)
			}
		}
	}()
	for _, conf := range cfg.listServices() {
		dir := path.Join(*inpath, conf, "event")
		f, err := os.Open(dir)
		if err != nil {
			log.Printf("%s: Entry found, but no service running\n", conf)
			continue
		}
		stat, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		f.Seek(0, io.SeekEnd)
		tail := &tail{
			fd:   f,
			name: conf,
			size: stat.Size(),
		}
		tails[dir] = tail
		watcher.Add(dir)
	}
	return events, done
}
