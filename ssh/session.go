package ssh

import (
	"io"

	"fmt"

	"strings"

	"github.com/sahilm/jaal/jaal"
	"github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type session struct {
	ctx        *sshContext
	newChannel gossh.NewChannel
	logger     *logrus.Logger
	eventchan  chan<- *jaal.Event
}

func (s *session) handle() error {
	l := s.logger.WithFields(logrus.Fields{"correlationID": s.ctx.SessionID()})
	ch, reqs, err := s.newChannel.Accept()
	if err != nil {
		return err
	}
	go func() {
		for r := range reqs {
			switch r.Type {
			case "exec":
				s.exec(r, ch)
			case "tcpip-forward":
				s.tcpipForward(r)
			case "env":
				s.env(r)
			case "shell":
				s.shell(ch)
			}
			if r.WantReply {
				err := r.Reply(true, nil)
				if err != nil && err != io.EOF {
					l.WithFields(logrus.Fields{
						"event": "ssh reply",
					}).Error("failed to send reply", err)
				}
			}
		}
	}()
	return nil
}

func (s *session) env(r *gossh.Request) {
	data := struct{ Key, Value string }{}
	err := gossh.Unmarshal(r.Payload, &data)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"event": "ssh env",
		}).Error("failed to unmarshal", err)
	} else {
		event := getEvent(s.ctx)
		event.Summary = fmt.Sprintf("env %v=%v", data.Key, data.Value)
		event.Data = data
		s.eventchan <- event
	}
}

func (s *session) tcpipForward(r *gossh.Request) {
	data := struct {
		BindAddress string
		BindPort    uint32
	}{}
	err := gossh.Unmarshal(r.Payload, &data)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"event": "ssh tcpip-forward",
		}).Error("failed to unmarshal", err)
	} else {
		event := getEvent(s.ctx)
		event.Summary = fmt.Sprintf("port forward %v:%v", data.BindAddress, data.BindPort)
		event.Data = data
		s.eventchan <- event
	}
}

const uname string = "Linux host 4.4.0-1022 #31-Ubuntu SMP Tue Jun 27 11:27:55 UTC 2017 x86_64 x86_64 x86_64 GNU/Linux\n"

func (s *session) exec(r *gossh.Request, ch gossh.Channel) {
	data := struct{ Command string }{}
	err := gossh.Unmarshal(r.Payload, &data)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"event": "ssh exec",
		}).Error("failed to unmarshal", err)
	} else {
		command := data.Command
		if strings.HasPrefix(command, "uname") {
			ch.Write([]byte(uname))
		}
		event := getEvent(s.ctx)
		event.Summary = fmt.Sprintf("command %v", data.Command)
		event.Data = data
		s.eventchan <- event
	}
	exit(0, ch)
}

func (s *session) shell(closer io.ReadWriteCloser) {
	term := terminal.NewTerminal(closer, "$ ")
	l := s.logger.WithFields(logrus.Fields{"event": "ssh shell"})
loop:
	for {
		shellcmd, err := term.ReadLine()
		switch err {
		case io.EOF:
			closer.Close()
			break loop
		case nil:
			event := getEvent(s.ctx)
			event.Summary = fmt.Sprintf("shell %v", shellcmd)
			event.Data = shellcmd
			s.eventchan <- event
		default:
			l.Error("failed to read line", err)
			closer.Close()
			break loop
		}
	}
}

func exit(code int, ch gossh.Channel) error {
	status := struct{ Status uint32 }{uint32(code)}
	_, err := ch.SendRequest("exit-status", false, gossh.Marshal(&status))
	if err != nil {
		return err
	}
	return ch.Close()
}
