package fs

import (
	"os"
	"os/signal"
	"syscall"
)

type watcher struct{}

func (w *watcher) SigHandle(c *Control) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGKILL, syscall.SIGINT)
	for sig := range sigs {
		switch sig {
		case syscall.SIGKILL, syscall.SIGINT:
			c.run.cleanup()
			//case syscall.SIGUSR
		}
	}
}
