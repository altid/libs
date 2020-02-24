// Simple wrapper package around docker-gop9p for use with gomobile
package client9p

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

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

type Qid struct {
	Qtype   int
	Version int
	Path    int
}

type Dir struct {
	Type int
	Dev  int
	Qid  *Qid
	Mode int

	AccessTime string
	ModTime    string

	Length int
	Name   string
	UID    string
	GID    string
	MUID   string
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

func (s *Session) Auth(afid int, uname, aname string) (*Qid, error) {
	q, err := s.session.Auth(s.ctx, p9p.Fid(afid), uname, aname)
	return toQid(q), err
}

func (s *Session) Attach(fid, afid int, uname, aname string) (*Qid, error) {
	q, err := s.session.Attach(s.ctx, p9p.Fid(fid), p9p.Fid(afid), uname, aname)
	return toQid(q), err
}

func (s *Session) Clunk(fid int) error {
	return s.session.Clunk(s.ctx, p9p.Fid(fid))
}

func (s *Session) Remove(fid int) error {
	return s.session.Remove(s.ctx, p9p.Fid(fid))
}

// We can't do varargs in gomobile yet, so we use a single name here
func (s *Session) Walk(fid, newfid int, names string) (*FileListing, error) {
	qids, err := s.session.Walk(s.ctx, p9p.Fid(fid), p9p.Fid(newfid), strings.Fields(names)...)
	if err != nil {
		return nil, err
	}

	f := &FileListing{
		qidpool: qids,
	}

	return f, nil
}

func (s *Session) Read(fid int, p []byte, offset int64) (int, error) {
	return s.session.Read(s.ctx, p9p.Fid(fid), p, offset)
}

func (s *Session) Write(fid int, p []byte, offset int64) (int, error) {
	return s.session.Write(s.ctx, p9p.Fid(fid), p, offset)
}

func (s *Session) Open(fid, mode int) (*File, error) {
	qid, id, err := s.session.Open(s.ctx, p9p.Fid(fid), p9p.Flag(mode))

	f := &File{
		qid: qid,
		id:  id,
	}

	return f, err
}

func (s *Session) Create(parent int, name string, perm, mode int) (*File, error) {
	qid, id, err := s.session.Create(s.ctx, p9p.Fid(parent), name, uint32(perm), p9p.Flag(mode))

	f := &File{
		qid: qid,
		id:  id,
	}

	return f, err
}

func (s *Session) Stat(fid int) (*Dir, error) {
	q, err := s.session.Stat(s.ctx, p9p.Fid(fid))
	return toDir(q), err
}

func (s *Session) WStat(fid int, dir *Dir) error {
	return s.session.WStat(s.ctx, p9p.Fid(fid), fromDir(dir))
}

func (s *Session) Version() string {
	_, v := s.session.Version()
	return v
}

func (s *Session) MSize() int {
	m, _ := s.session.Version()
	return m
}

func (f *FileListing) Next() *Qid {
	if f.current < len(f.qidpool) {
		item := f.qidpool[f.current]
		f.current++
		return toQid(item)
	}

	return nil
}

func toQid(q p9p.Qid) *Qid {
	return &Qid{
		Version: int(q.Version),
		Qtype:   int(q.Type),
		Path:    int(q.Path),
	}
}

func fromQid(d *Qid) p9p.Qid {
	return p9p.Qid{
		Version: uint32(d.Version),
		Type:    p9p.QType(d.Qtype),
		Path:    uint64(d.Path),
	}
}

func toDir(d p9p.Dir) *Dir {
	return &Dir{
		Type: int(d.Type),
		Dev:  int(d.Dev),
		Qid:  toQid(d.Qid),
		Mode: int(d.Mode),

		AccessTime: d.AccessTime.String(),
		ModTime:    d.ModTime.String(),

		Length: int(d.Length),
		Name:   d.Name,
		UID:    d.UID,
		GID:    d.GID,
		MUID:   d.MUID,
	}
}

func fromDir(d *Dir) p9p.Dir {
	t1, _ := time.Parse(time.ANSIC, d.AccessTime)
	t2, _ := time.Parse(time.ANSIC, d.ModTime)

	return p9p.Dir{
		Type: uint16(d.Type),
		Dev:  uint32(d.Dev),
		Qid:  fromQid(d.Qid),
		Mode: uint32(d.Mode),

		AccessTime: t1,
		ModTime:    t2,

		Length: uint64(d.Length),
		Name:   d.Name,
		UID:    d.UID,
		GID:    d.GID,
		MUID:   d.MUID,
	}
}
