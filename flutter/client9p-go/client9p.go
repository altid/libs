// Simple wrapper package around docker-gop9p for use with gomobile
package client9p

import (
	"context"
	"fmt"
	"net"

	"github.com/docker/go-p9p"
)

type Session struct {
	session p9p.Session
	ctx     context.Context
	conn    net.Conn
}

type File struct {
	qid p9p.Qid
	id  uint32
}

type FileListing struct {
	qidpool []p9p.Qid
	current int
}

func NewSession(addr string) (*Session, error) {
	conn, err := net.Dial("tcp", addr+":564")
	if err != nil {
		return nil, fmt.Errorf("Unable to dial requested address: %v", err)
	}

	ctx := context.Background()
	s, err := p9p.NewSession(ctx, conn)
	if err != nil {
		return nil, err
	}
	session := &Session{
		ctx:     ctx,
		conn:    conn,
		session: s,
	}

	return session, nil
}

func (s *Session) Hangup() error {
	s.ctx.Done()
	return s.conn.Close()
}

func (s *Session) Auth(afid p9p.Fid, uname, aname string) (p9p.Qid, error) {
	return s.session.Auth(s.ctx, afid, uname, aname)
}

func (s *Session) Attach(fid, afid p9p.Fid, uname, aname string) (p9p.Qid, error) {
	return s.session.Attach(s.ctx, fid, afid, uname, aname)
}

func (s *Session) Clunk(fid p9p.Fid) error {
	return s.session.Clunk(s.ctx, fid)
}

func (s *Session) Remove(fid p9p.Fid) error {
	return s.session.Remove(s.ctx, fid)
}

func (s *Session) Walk(fid, newfid p9p.Fid, names ...string) (*FileListing, error) {
	qids, err := s.session.Walk(s.ctx, fid, newfid, names...)
	if err != nil {
		return nil, err
	}

	f := &FileListing{
		qidpool: qids,
	}

	return f, nil
}

func (s *Session) Read(fid p9p.Fid, p []byte, offset int64) (int, error) {
	return s.session.Read(s.ctx, fid, p, offset)
}

func (s *Session) Write(fid p9p.Fid, p []byte, offset int64) (int, error) {
	return s.session.Write(s.ctx, fid, p, offset)
}

func (s *Session) Open(fid p9p.Fid, mode p9p.Flag) (*File, error) {
	qid, id, err := s.session.Open(s.ctx, fid, mode)

	f := &File{
		qid: qid,
		id:  id,
	}

	return f, err
}

func (s *Session) Create(parent p9p.Fid, name string, perm uint32, mode p9p.Flag) (*File, error) {
	qid, id, err := s.session.Create(s.ctx, parent, name, perm, mode)

	f := &File{
		qid: qid,
		id:  id,
	}

	return f, err
}

func (s *Session) Stat(fid p9p.Fid) (p9p.Dir, error) {
	return s.session.Stat(s.ctx, fid)
}

func (s *Session) WStat(fid p9p.Fid, dir p9p.Dir) error {
	return s.session.WStat(s.ctx, fid, dir)
}

func (s *Session) Version() (int, string) {
	return s.session.Version()
}

func (f *FileListing) Next() p9p.Qid {
	if f.current < len(f.qidpool) {
		item := f.qidpool[f.current]
		f.current++
		return item
	}

	return p9p.Qid{}
}
