package service

import (
	"errors"
	"os"
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
