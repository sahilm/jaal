package jaal

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Server interface {
	ListenAndServe() (chan *Event, chan error)
}

func ListenAndLog(eventLogger *EventLogger, logger *logrus.Logger, servers ...Server) {
	eventchans, errchans := runAll(servers)
	eventchan := mergeEvents(eventchans)
	errchan := mergeErrs(errchans)

	go func() {
		for err := range errchan {
			switch err := err.(type) {
			case FatalError:
				logger.Fatal(err)
			default:
				logger.Error(err)
			}
		}
	}()

	go func() {
		for e := range eventchan {
			eventLogger.Log(time.Now(), e)
		}
	}()

	select {} //block forever
}

func runAll(servers []Server) ([]chan *Event, []chan error) {
	var eventchans []chan *Event
	var errchans []chan error
	for _, s := range servers {
		eventc, errc := s.ListenAndServe()
		eventchans = append(eventchans, eventc)
		errchans = append(errchans, errc)
	}
	return eventchans, errchans
}

func mergeEvents(cs []chan *Event) chan *Event {
	out := make(chan *Event)

	output := func(c <-chan *Event) {
		for n := range c {
			out <- n
		}
	}
	for _, c := range cs {
		go output(c)
	}
	return out
}

func mergeErrs(cs []chan error) chan error {
	out := make(chan error)

	output := func(c <-chan error) {
		for n := range c {
			out <- n
		}
	}

	for _, c := range cs {
		go output(c)
	}
	return out
}

type FatalError struct {
	Err error
}

func (f FatalError) Error() string {
	return f.Err.Error()
}
