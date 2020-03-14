package fs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

// UserShareDir returns the default root directory to use for user-specific application data. Users should create their own application-specific subdirectory within this one and use that.
// On Unix systems, it returns $XDG_DATA_HOME as specified by https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html if non-empty, else $HOME/.local/share. On Darwin, it returns $HOME/Library. On Windows, it returns %LocalAppData%. On Plan 9, it returns $home/lib.
func UserShareDir() (string, error) {
	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = os.Getenv("LocalAppData")
		if dir == "" {
			return "", errors.New("%LocalAppData% is not defined")
		}
	case "darwin":
		dir = os.Getenv("HOME")
		if dir == "" {
			return "", errors.New("$HOME is not defined")
		}
		dir += "/Library"
	case "plan9":
		dir = os.Getenv("home")
		if dir == "" {
			return "", errors.New("$home is not defined")
		}
		dir += "/lib"
	default: // Unix
		dir = os.Getenv("XDG_DATA_HOME")
		if dir == "" {
			dir = os.Getenv("HOME")
			if dir == "" {
				return "", errors.New("neither $XDG_DATA_HOME nor $HOME is defined")
			}
			dir += "/.local/share"
		}
	}
	return dir, nil
}

// UserConfDir returns the default root directory to use for user-specific configuration data. Users should create their own application-specific subdirectory within this one and use that.
// On Unix systems, it returns $XDG_CONFIG_HOME as specified by https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html if non-empty, else $HOME/.config. On Darwin, it returns $HOME/Library/Preferences. On Windows, it returns %LocalAppData%. On Plan 9, it returns $home/lib.
func UserConfDir() (string, error) {
	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = os.Getenv("LocalAppData")
		if dir == "" {
			return "", errors.New("%LocalAppData% is not defined")
		}
	case "darwin":
		dir = os.Getenv("HOME")
		if dir == "" {
			return "", errors.New("$HOME is not defined")
		}
		dir += "/Library/Preferences"
	case "plan9":
		dir = os.Getenv("home")
		if dir == "" {
			return "", errors.New("$home is not defined")
		}
		dir += "/lib"
	default: // Unix
		dir = os.Getenv("XDG_CONFIG_HOME")
		if dir == "" {
			dir = os.Getenv("HOME")
			if dir == "" {
				return "", errors.New("neither $XDG_CONFIG_HOME nor $HOME is defined")
			}
			dir += "/.config"
		}
	}
	return dir, nil
}

func symlink(logname, feedname string) error {
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

func unlink(feedname string) error {
	if runtime.GOOS == "plan9" {
		command := exec.Command("/bin/unmount", feedname)
		return command.Run()
	}

	return os.Remove(feedname)
}

func validateString(path string) error {
	if _, e := os.Stat(path); e != nil {
		return e
	}

	return nil
}

func sigwatch(c *Control) {
	d := c.watch
	d.SigHandle(c)
}

func dispatch(c *Control) {
	// TODO: wrap with waitgroups
	// If close is requested on a file which is currently being opened, cancel open request
	// If open is requested on file which already exists, no-op
	ew, err := c.write.errorwriter()
	if err != nil {
		log.Fatal(err)
	}

	defer ew.Close()

	for {
		select {
		case line := <-c.req:
			run(c, ew, line)
		case <-c.done:
			return
		}
	}
}

func run(c *Control, ew *WriteCloser, line string) {
	c.Lock()
	defer c.Unlock()

	token := strings.Fields(line)
	if len(token) < 1 {
		return
	}

	switch token[0] {
	case "open":
		if len(token) < 2 {
			return
		}

		if e := c.ctl.Open(c, token[1]); e != nil {
			c.debug(ctlError, token[1], e)
			fmt.Fprintf(ew, "open: %v\n", e)
		}
	case "close":
		if len(token) < 2 {
			return
		}

		// We need to get to these still somehow
		if e := c.ctl.Close(c, token[1]); e != nil {
			c.debug(ctlError, token[1], e)
			fmt.Fprintf(ew, "close: %v\n", e)
		}

	case "link":
		if len(token) < 2 {
			return
		}

		if e := c.ctl.Link(c, token[1], token[2]); e != nil {
			c.debug(ctlError, token[1], e)
			fmt.Fprintf(ew, "link: %v\n", e)
		}

	default:
		cmd, err := c.run.buildCommand(line)
		if err != nil {
			c.debug(ctlError, token[0], errors.New("unsupported command"))
			fmt.Fprintf(ew, "unsupported command")
			return
		}

		if e := c.ctl.Default(c, cmd); e != nil {
			c.debug(ctlError, token[0], e)
			fmt.Fprintf(ew, "%s: %v\n", token[0], e)
		}
	}
}
