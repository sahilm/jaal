package jaal

import (
	"net"
	"time"
)

type Event struct {
	UnixTime       int64
	Timestamp      string
	CorrelationID  string
	RemoteAddr     string
	RemoteHostName string
	LocalAddr      string
	LocalHostName  string
	Type           string
	Summary        string
	Data           interface{}
}

func (e *Event) WithMetadata(t time.Time) *Event {
	e.RemoteHostName = lookupAddr(e.RemoteAddr)
	e.LocalHostName = lookupAddr(e.LocalAddr)
	e.UnixTime = t.Unix()
	e.Timestamp = t.UTC().Format(time.RFC3339)
	return e
}

func lookupAddr(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "unknown"
	}
	hostNames, err := net.LookupAddr(host)
	if err != nil {
		return "unknown"
	}
	return hostNames[0]
}
