package util

import (
	"os"
	"os/exec"
	"path"
	"runtime"
)

func Symlink(logname, feedname string) error {
	if _, err := os.Stat(logname); os.IsNotExist(err) {
		os.MkdirAll(path.Dir(logname), 0755)
		fp, err := os.Create(logname)
		if err != nil {
			return err
		}

		fp.Close()
	}

	if runtime.GOOS == "plan9" {
		command := exec.Command("/bin/bind", logname, feedname)
		return command.Run()
	}

	return os.Symlink(logname, feedname)
}

func Unlink(feedname string) error {
	if runtime.GOOS == "plan9" {
		command := exec.Command("/bin/unmount", feedname)
		return command.Run()
	}

	return os.Remove(feedname)
}
