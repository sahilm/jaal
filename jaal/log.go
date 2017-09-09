package jaal

import (
	"io"

	"encoding/json"
	"fmt"

	"time"

	"github.com/sirupsen/logrus"
)

type UTCFormatter struct {
	logrus.Formatter
}

func (u UTCFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}

type EventLogger struct {
	l         *logrus.Logger
	notifiers []EventNotifier
}

type EventNotifier interface {
	notify(event *Event)
}

func NewEventLogger(out io.Writer, notifiers ...EventNotifier) *EventLogger {
	l := logrus.New()
	l.Out = out
	l.Formatter = &eventLogFormatter{}
	return &EventLogger{l, notifiers}
}

func (el EventLogger) Log(now time.Time, event *Event) {
	el.l.WithField("data", event.WithMetadata(now)).Info("")
	for _, notifier := range el.notifiers {
		notifier.notify(event)
	}
}

type eventLogFormatter struct{}

func (f eventLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var serialized []byte
	var err error
	serialized, err = json.Marshal(entry.Data["data"])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}
