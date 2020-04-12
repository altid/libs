package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/altid/libs/markup"
	"github.com/altid/server"
	"github.com/altid/server/client"
	"github.com/altid/server/command"
	"github.com/altid/server/files"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type service struct {
	svc    *server.Service
	listen net.Listener
	config *ssh.ServerConfig
	logger func(string, ...interface{})
}

func (s *service) setup() error {
	// Public key authentication is done by comparing
	// the public key of a received connection
	// with the entries in the authorized_keys file.
	kfp := *keys

	if kfp[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}

		kfp = path.Join(home, kfp[1:])
	}

	akbytes, err := ioutil.ReadFile(kfp)
	if err != nil {
		return fmt.Errorf("failed to load authorized_keys, err: %v", err)
	}

	akmap := map[string]bool{}
	for len(akbytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(akbytes)
		if err != nil {
			return err
		}

		akmap[string(pubKey.Marshal())] = true
		akbytes = rest
	}

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		// TODO: Use factotum when enabled
		/*
		   // Remove to disable password auth.
		   PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
		       // Should use constant-time compare (or better, salt+hash) in
		       // a production setting.
		       if c.User() == "testuser" && string(pass) == "tiger" {
		           return nil, nil
		       }
		       return nil, fmt.Errorf("password rejected for %q", c.User())
		   },*/

		// Remove to disable public key auth.
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if akmap[string(pubKey.Marshal())] {
				return &ssh.Permissions{
					// Record the public key used for authentication.
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}

	rsafp := *rsa
	if rsafp[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}

		rsafp = path.Join(home, rsafp[1:])
	}

	privateBytes, err := ioutil.ReadFile(rsafp)
	if err != nil {
		return fmt.Errorf("failed to load private key: %v", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", *addr+":"+*port)
	if err != nil {
		return fmt.Errorf("failed to listen for connection: %v", err)
	}

	s.listen = listener
	s.config = config

	return nil
}

func (s *service) Run(ctx context.Context, svc *server.Service) error {
	// Wait for incoming connections
	// As they arrive, assing and defer the client connection
	s.svc = svc

	for {
		nc, err := s.listen.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept incoming connection: %v", err)
		}

		conn, chans, reqs, err := ssh.NewServerConn(nc, s.config)
		if err != nil {
			log.Fatal("failed to handshake: ", err)
		}

		s.logger("logged in with key %s", conn.Permissions.Extensions["pubkey-fp"])
		c := svc.Client.Client(0)
		c.SetBuffer(svc.Default())

		go ssh.DiscardRequests(reqs)
		go s.handleChannels(chans, c)
	}
}

func (s *service) handleChannels(chans <-chan ssh.NewChannel, c *client.Client) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go s.handleChannel(newChannel, c)
	}
}

func (s *service) handleChannel(newChannel ssh.NewChannel, c *client.Client) {
	defer s.svc.Client.Remove(c.UUID)

	// We don't support the tcp-ip/forwarded or x11
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	connection, requests, err := newChannel.Accept()
	if err != nil {
		return
	}

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				// We only accept the default shell
				// (i.e. no command in the Payload)
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req":
				// Make sure we resize pty
				req.Reply(true, nil)
			case "window-change":
				// pty resized
			}
		}
	}()

	term := terminal.NewTerminal(connection, fmt.Sprintf("%s > ", c.Current()))
	go readFeed(c.Current(), c.UUID, term, s.svc.Files)

	defer connection.Close()

	for {
		line, err := term.ReadLine()
		if err != nil || len(line) < 1 {
			continue
		}

		tokens := strings.Fields(line)

		switch tokens[0] {
		case "/quit":
			return
		case "/tabs":
			tabs := readFile(c.Current(), "/tabs", c.UUID, s.svc.Files)
			term.Write(tabs)
		case "/title":
			title := readFile(c.Current(), "/title", c.UUID, s.svc.Files)
			term.Write(title)
		case "/buffer":
			target := strings.Join(tokens[1:], " ")
			s.svc.Commands <- &command.Command{
				UUID:    c.UUID,
				CmdType: command.BufferCmd,
				Args:    []string{target},
				From:    c.Current(),
			}

			term.SetPrompt(fmt.Sprintf("%s > ", target))
			time.AfterFunc(time.Millisecond*50, func() {
				readFeed(target, c.UUID, term, s.svc.Files)
			})
		case "/open":
			target := strings.Join(tokens[1:], " ")
			s.svc.Commands <- &command.Command{
				UUID:    c.UUID,
				CmdType: command.OpenCmd,
				Args:    []string{target},
				From:    c.Current(),
			}

			term.SetPrompt(fmt.Sprintf("%s > ", target))
			time.AfterFunc(time.Millisecond*50, func() {
				readFeed(target, c.UUID, term, s.svc.Files)
			})
		case "/close": // Case closed!
			s.svc.Commands <- &command.Command{
				UUID:    c.UUID,
				CmdType: command.CloseCmd,
				From:    c.Current(),
			}

			time.AfterFunc(time.Millisecond*50, func() {
				term.SetPrompt(fmt.Sprintf("%s > ", c.Current()))
				readFeed(c.Current(), c.UUID, term, s.svc.Files)
			})
		default:
			if tokens[0][0] == '/' {
				s.svc.Commands <- &command.Command{
					UUID:    c.UUID,
					CmdType: command.OtherCmd,
					Args:    tokens,
					From:    c.Current(),
				}
				continue
			}
			input(c.Current(), line, c.UUID, s.svc.Files)
		}
	}
}

func (s *service) Address() (string, string) {
	return *addr, *port
}

func readFeed(buffer string, uuid uint32, term *terminal.Terminal, files *files.Files) {
	var offset int64

	fp, err := files.Normal(buffer, "/feed", uuid)
	if err != nil {
		return
	}

	for {
		b := make([]byte, 2048)

		switch v := fp.(type) {
		case io.ReaderAt:
			n, err := v.ReadAt(b, offset)
			offset += int64(n)

			if err != nil {
				return
			}

			l, err := markup.NewLexer(b[:n]).Bytes()
			if err != nil {
				return
			}
			term.Write(l)
		}
	}
}

func readFile(buffer string, name string, uuid uint32, files *files.Files) []byte {
	var err error
	var bits bytes.Buffer
	var offset int64

	// Ladder down if we don't find a feed file or doc
	fp, err := files.Normal(buffer, name, uuid)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	for {
		var n int
		// Tune this
		b := make([]byte, 2048)

		switch v := fp.(type) {
		case io.ReaderAt:
			n, err = v.ReadAt(b, offset)
			offset += int64(n)
		case io.Reader:
			n, err = v.Read(b)
		case io.ReadWriter:
			n, err = v.Read(b)
		case io.ReadWriteSeeker:
			v.Seek(offset, os.SEEK_SET)
			n, err = v.Read(b)
			offset += int64(n)
		case io.ReadWriteCloser:
			n, err = v.Read(b)
		case io.ReadCloser:
			n, err = v.Read(b)
		case io.ReadSeeker:
			v.Seek(offset, os.SEEK_SET)
			n, err = v.Read(b)
			offset += int64(n)
		default:
			return nil
		}

		switch err {
		case io.EOF:
			return bits.Bytes()
		case nil:
			bits.Write(b)
		default:
			return nil
		}

		if n < 1 {
			return bits.Bytes()
		}
	}
}

func input(buffer, line string, uuid uint32, files *files.Files) error {
	fp, err := files.Normal(buffer, "/input", uuid)
	if err != nil {
		return err
	}

	switch v := fp.(type) {
	case io.WriterAt:
		_, err := v.WriteAt([]byte(line), 0)
		return err
	}

	return nil
}
