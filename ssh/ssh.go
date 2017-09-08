package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"time"

	"fmt"

	"io/ioutil"

	"github.com/sahilm/jaal/jaal"
	"github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type Server struct {
	Addr        string
	HostKeyFile string
	HostSigner  gossh.Signer
	IdleTimeout time.Duration
	MaxTimeout  time.Duration
	Version     string
	Logger      *logrus.Logger
}

func (s Server) ListenAndServe() (chan *jaal.Event, chan error) {
	eventchan := make(chan *jaal.Event)
	errchan := make(chan error, 1)
	addr := s.Addr
	if addr == "" {
		addr = ":22"
	}
	s.getLogger().Info(fmt.Sprintf("starting ssh server on: %v", addr))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		errchan <- jaal.FatalError{Err: err}
		return eventchan, errchan
	}
	go s.serve(listener, eventchan, errchan)
	return eventchan, errchan
}

func (s *Server) serve(l net.Listener, eventchan chan<- *jaal.Event, errchan chan error) {
	defer l.Close()
	if err := s.ensureHostSigner(); err != nil {
		errchan <- err
		return
	}
	var tempDelay time.Duration
	for {
		conn, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			errchan <- err
			return
		}
		go s.handleConn(conn, eventchan)
	}
}

func (s *Server) handleConn(newConn net.Conn, eventchan chan<- *jaal.Event) {
	ctx, cancel := newContext()
	conn := &serverConn{
		Conn:          newConn,
		idleTimeout:   s.IdleTimeout,
		closeCanceler: cancel,
	}
	if s.MaxTimeout > 0 {
		conn.maxDeadline = time.Now().Add(s.MaxTimeout)
	}
	defer func() {
		eventchan <- logoutEvent(ctx)
		conn.Close()
	}()
	sshConn, chans, reqs, err := gossh.NewServerConn(conn, s.config(ctx))
	if err != nil {
		return
	}
	eventchan <- loginEvent(ctx)
	ctx.applyConnMetadata(sshConn, s.getLogger())
	go gossh.DiscardRequests(reqs)
	for ch := range chans {
		session := &session{
			newChannel: ch,
			logger:     s.Logger,
			ctx:        ctx,
			eventchan:  eventchan,
		}
		go session.handle()
	}
}

func logoutEvent(ctx *sshContext) *jaal.Event {
	event := getEvent(ctx)
	event.Summary = "logout"
	event.Data = struct{}{}
	return event
}

func loginEvent(ctx *sshContext) *jaal.Event {
	event := getEvent(ctx)
	event.Summary = fmt.Sprintf("login username: %v, password: %v", ctx.User(), ctx.Password())
	event.Data = struct{ ClientVersion string }{ctx.ClientVersion()}
	return event
}

func (s *Server) config(ctx *sshContext) *gossh.ServerConfig {
	c := &gossh.ServerConfig{

		// Allow everyone to login. This is a honeypot 😀
		PasswordCallback: func(cm gossh.ConnMetadata, pass []byte) (*gossh.Permissions, error) {
			ctx.applyConnMetadata(cm, s.getLogger())
			ctx.applyPassword(string(pass))
			return ctx.Permissions(), nil
		},
	}
	if s.Version != "" {
		c.ServerVersion = "SSH-2.0-" + s.Version
	}
	c.AddHostKey(s.HostSigner)
	return c
}

func (s *Server) ensureHostSigner() error {
	if s.HostKeyFile == "" {
		signer, err := generateSigner()
		if err != nil {
			return err
		}
		s.HostSigner = signer
	} else {
		keyBytes, err := ioutil.ReadFile(s.HostKeyFile)
		if err != nil {
			return err
		}

		signer, err := gossh.ParsePrivateKey(keyBytes)
		if err != nil {
			return err
		}
		s.HostSigner = signer
	}
	return nil
}

func generateSigner() (gossh.Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	signer, err := gossh.NewSignerFromKey(key)
	if err != nil {
		return nil, err
	}
	return signer, nil
}

func (s *Server) getLogger() *logrus.Logger {
	if s.Logger == nil {
		s.Logger = logrus.New()
	}
	return s.Logger
}

func getEvent(ctx *sshContext) *jaal.Event {
	return &jaal.Event{
		RemoteAddr:    ctx.RemoteAddr().String(),
		Type:          "ssh",
		CorrelationID: ctx.SessionID(),
		LocalAddr:     ctx.LocalAddr().String(),
	}
}
