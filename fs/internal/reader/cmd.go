package reader

import (
	"io"
	"os"
	"path"
	"time"
)

type CmdReader struct {
	f    *os.File
	size int64
}

// Cmd returns a new reader, ready to read from
// it will truncate the file after every read back to the initial size
func Cmd(rundir string) (*CmdReader, error) {
	os.MkdirAll(rundir, 0755)
	fp := path.Join(rundir, "ctl")

	f, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	if _, err := f.Seek(0, os.SEEK_END); err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	c := &CmdReader{
		f:    f,
		size: s.Size(),
	}

	return c, err
}

// Clear out any new data and seek back to end for next read
func (r *CmdReader) Read(p []byte) (n int, err error) {
	for {
		n, err := r.f.Read(p)
		if n > 0 {
			return n, nil
		} else if err != io.EOF {
			return n, err
		}

		r.f.Truncate(r.size)
		r.f.Seek(0, os.SEEK_END)

		time.Sleep(500 * time.Millisecond)
	}
}
