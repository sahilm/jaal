package jaal

import (
	"fmt"
	"net/url"

	slack "github.com/huguesalary/slack-go"
	"github.com/sirupsen/logrus"
)

type SlackNotifier struct {
	u      *url.URL
	logger *logrus.Logger
}

func NewSlackNotifier(rawurl string, logger *logrus.Logger) (*SlackNotifier, error) {
	u, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return nil, FatalError{Err: fmt.Errorf("invalid slack url: %v", rawurl)}
	}
	return &SlackNotifier{u, logger}, nil
}

func (s *SlackNotifier) notify(event *Event) {
	client := slack.NewClient(s.u.String())
	msg := &slack.Message{}
	summary := fmt.Sprintf("[%v] %v", event.Type, event.Summary)
	msg.Text = summary

	attach := msg.NewAttachment()
	attach.Fallback = summary
	attach.Pretext = fmt.Sprintf("New %v event", event.Type)
	attach.Color = "warning"
	attach.Text = summary

	go func() {
		err := client.SendMessage(msg)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"event": "slack notify",
			}).Error(err)
		}
	}()
}
