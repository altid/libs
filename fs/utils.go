package fslib

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"runtime"
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

func event(c *Control, eventmsg string) error {
	err := validateString(eventmsg)
	if err != nil {
		return err
	}
	file := path.Join(c.rundir, "event")
	if _, err := os.Stat(path.Dir(file)); os.IsNotExist(err) {
		os.MkdirAll(path.Dir(file), 0755)
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	f.WriteString(eventmsg + "\n")
	return nil
}

func symlink(logname, feedname string) error {
	if _, err := os.Stat(logname); os.IsNotExist(err) {
		os.MkdirAll(path.Dir(logname), 0755)
		fp, err := os.Create(logname)
		defer fp.Close()
		if err != nil {
			return err
		}
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

func validateString(test string) error {
	if valid.MatchString(test) {
		return errors.New("error - invalid string")
	}
	return nil
}
