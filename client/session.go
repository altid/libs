package client

import (
	"context"
	"fmt"
	"net"

	"github.com/docker/go-p9p"
)

type session struct {
	p9     p9p.Session
	files  map[string]p9p.Fid
	root   p9p.Fid
	next   p9p.Fid
	iounit uint32
}

func attach(ctx context.Context, username, addr string) (*session, error) {
	s := &session{
		files:  make(map[string]p9p.Fid),
		next:   1,
		root:   1,
		iounit: 8192,
	}
	conn, err := net.Dial("tcp", addr+":564")
	if err != nil {
		return nil, fmt.Errorf("Unable to dial requested address: %v", err)
	}
	p9, err := p9p.NewSession(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("Unable to create session: %v", err)
	}
	s.p9 = p9
	if _, err := s.p9.Attach(ctx, s.next, p9p.NOFID, username, "/"); err != nil {
		return nil, fmt.Errorf("Error in Attach: %v", err)
	}
	io, _ := s.p9.Version()
	s.iounit = uint32(io)
	s.root = s.next
	s.next++
	if _, err := s.p9.Walk(ctx, s.root, s.next); err != nil {
		return nil, fmt.Errorf("Error doing initial walk: %v", err)
	}

	s.next++
	return s, nil
}

func (s *session) getNextFid(ctx context.Context, name string) (p9p.Fid, error) {
	var current p9p.Fid
	current = s.next
	s.next++
	if _, err := s.p9.Walk(ctx, s.root, current, name); err != nil {
		return 1, fmt.Errorf("Unable to walk to file %s: %v", name, err)
	}
	s.files[name] = current
	_, _, err := s.p9.Open(ctx, current, p9p.ORDWR)
	return current, err
}

func (s *session) readFile(ctx context.Context, name string, offset int64) ([]byte, error) {
	var err error
	fid, ok := s.files[name]
	if !ok {
		fid, err = s.getNextFid(ctx, name)
		if err != nil {
			return []byte(""), err
		}
	}

	p := make([]byte, s.iounit)
	if _, err := s.p9.Read(ctx, fid, p, offset); err != nil {
		return p, err
	}

	return p, nil
}

func (s *session) writeFile(ctx context.Context, name string, data []byte) error {
	var err error
	fid, ok := s.files[name]
	if !ok {
		fid, err = s.getNextFid(ctx, name)
		if err != nil {
			return err
		}
	}
	// Seek to end first!
	stat, err := s.p9.Stat(ctx, fid)
	if err != nil {
		return fmt.Errorf("Stat error in %s: %v", name, err)
	}
	_, err = s.p9.Write(ctx, fid, data, int64(stat.Length))
	return err
}
