package threads

import (
	"errors"
	"os"
	"syscall"
)

func Start(fn func() error, fg bool) error {
	if fg {
		return fn()
	}

	if os.Getenv("RFORK") == "" {
		r, _, err := syscall.Syscall(syscall.SYS_RFORK, uintptr(syscall.RFPROC|syscall.RFNOWAIT|syscall.RFNOTEG|syscall.RFNAMEG), 0, 0)
		if err.Error() != "" {
			return errors.New(err.Error())
		}

		if r == 0 {
			os.Setenv("RFORK", "1")
			return syscall.Exec(os.Args[0], os.Args, append(os.Environ(), "RFORK=1"))
		} else {
			return nil
		}	
	} else {
		return fn()
	}
}
