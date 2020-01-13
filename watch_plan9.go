package main

import (
	"context"
	"io"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type tailer struct {
	tails []*tail
	out   chan *event
	ctx   context.Context
	sync.Mutex
}

func listenEvents(ctx context.Context, cfg *config) (chan *event, err) {
	events := make(chan *event)
	t := &tailer{
		out: events,
		ctx: ctx,
	}
	services := cfg.listServices()
	for _, service := range services {
		err := t.add(service)
		if err != nil {
			log.Printf("%s: Entry found, but no service running\n", service)
		}
	}
	go t.tail()
	return events, nil
}

func (t *tailer) add(name string) error {
	dir := path.Join(*inpath, name, "event")
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	f.Seek(0, io.SeekEnd)
	// Make sure we tail from the end of the file
	tail := &tail{
		fd:   f,
		name: name,
		size: stat.Size(),
	}
	t.Lock()
	t.tails = append(t.tails, tail)
	t.Unlock()
	return nil
}

func (t *tailer) tail() {
	for {
		select {
		case <-t.ctx.Done():
			t.cleanup()
			break
		case <-time.After(125 * time.Millisecond):
			checkAll(t)
		}
	}
}

func (t *tailer) cleanup() {
	for _, tail := range t.tails {
		tail.fd.Close()
	}
}

func checkAll(t *tailer) {
	for n, tail := range t.tails {
		hs, err := tail.fd.Stat()
		if err != nil {
			// Assume we have a dead events file, pop from slice
			// TODO(halfwit) Gather data and try to recover
			defer tail.fd.Close()
			t.tails = append(t.tails[:n], t.tails[n+1:]...)
			return
		}
		if hs.Size() == tail.size {
			continue
		}
		t.out <- tail.readlines()
		tail.size = hs.Size()
	}
}
