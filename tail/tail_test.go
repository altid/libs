package tail

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

func TestTail(t *testing.T) {
	ctx := context.Background()

	dir, err := ioutil.TempDir(os.TempDir(), "test")
	if err != nil {
		t.Error(err)
		return
	}

	defer os.RemoveAll(dir) // clean up

	fp, err := os.Create(path.Join(dir, "event"))
	if err != nil {
		t.Error(err)
		return
	}

	ev, err := WatchEvents(ctx, path.Dir(dir), path.Base(dir))
	if err != nil {
		t.Error(err)
		return
	}

	go mockEvents(fp, dir)

	first := <-ev
	if first.Name != "test1" || first.EventType != FeedEvent {
		t.Error(errors.New("test1 failed"))
	}

	second := <-ev
	if second.Name != "test2" || second.EventType != FeedEvent {
		t.Error(errors.New("test2 failed"))
	}

	third := <-ev
	if third.Name != "test3" || third.EventType != DocEvent {
		t.Error(errors.New("test3 failed"))
	}

	fourth := <-ev
	if fourth.Name != "test4" || fourth.EventType != NotifyEvent {
		t.Error(errors.New("test4 failed"))
	}

	fifth := <-ev
	if fifth.Name != "test5" || fifth.EventType != NoneEvent {
		t.Error(errors.New("test5 failed"))
	}

	ctx.Done()
}

func mockEvents(fp *os.File, dir string) {
	first := path.Join(dir, "test1", "feed")
	fp.WriteString(first + "\n")

	time.Sleep(125 * time.Millisecond)
	second := path.Join(dir, "test2", "feed")
	fp.WriteString(second + "\n")

	time.Sleep(125 * time.Millisecond)
	third := path.Join(dir, "test3", "document")
	fp.WriteString(third + "\n")

	time.Sleep(125 * time.Millisecond)
	fourth := path.Join(dir, "test4", "notification")
	fp.WriteString(fourth + "\n")

	time.Sleep(125 * time.Millisecond)
	fifth := path.Join(dir, "test5", "banana")
	fp.WriteString(fifth + "\n")
}
