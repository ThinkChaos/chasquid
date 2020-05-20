// Package maillog implements a log specifically for email.
package maillog

import (
	"fmt"
	"log/syslog"
	"net"
	"sync"
	"time"

	"blitiri.com.ar/go/chasquid/internal/trace"
	"blitiri.com.ar/go/log"
)

// Global event logs.
var (
	authLog = trace.NewEventLog("Authentication", "Incoming SMTP")
)

// Logger contains a backend used to log data to, such as a file or syslog.
// It implements various user-friendly methods for logging mail information to
// it.
type Logger struct {
	inner *log.Logger
	once  sync.Once
}

// New creates a new Logger which will write messages to the given writer.
func NewFile(path string) (*Logger, error) {
	inner, err := log.NewFile(path)
	if err != nil {
		return nil, err
	}

	// Only include time in output
	inner.SetOptions(log.ShowTime)

	return &Logger{inner: inner}, nil
}

// NewSyslog creates a new Logger which will write messages to syslog.
func NewSyslog() (*Logger, error) {
	inner, err := log.NewSyslog(syslog.LOG_INFO|syslog.LOG_MAIL, "chasquid")
	if err != nil {
		return nil, err
	}

	inner.SetOptions(log.ShowMessageOnly)

	return &Logger{inner: inner}, nil
}

func (l *Logger) printf(format string, args ...interface{}) {
	err := l.inner.Log(log.Info, 2, format, args...)
	if err != nil {
		l.once.Do(func() {
			log.Errorf("failed to write to maillog: %v", err)
			log.Errorf("(will not report this again)")
		})
	}
}

// Listening logs that the daemon is listening on the given address.
func (l *Logger) Listening(a string) {
	l.printf("daemon listening on %s\n", a)
}

// Auth logs an authentication request.
func (l *Logger) Auth(netAddr net.Addr, user string, successful bool) {
	res := "succeeded"
	if !successful {
		res = "failed"
	}
	msg := fmt.Sprintf("%s auth %s for %s\n", netAddr, res, user)
	l.printf(msg)
	authLog.Debugf(msg)
}

// Rejected logs that we've rejected an email.
func (l *Logger) Rejected(netAddr net.Addr, from string, to []string, err string) {
	if from != "" {
		from = fmt.Sprintf(" from=%s", from)
	}
	toStr := ""
	if len(to) > 0 {
		toStr = fmt.Sprintf(" to=%v", to)
	}
	l.printf("%s rejected%s%s - %v\n", netAddr, from, toStr, err)
}

// Queued logs that we have queued an email.
func (l *Logger) Queued(netAddr net.Addr, from string, to []string, id string) {
	l.printf("%s from=%s queued ip=%s to=%v\n", id, from, netAddr, to)
}

// SendAttempt logs that we have attempted to send an email.
func (l *Logger) SendAttempt(id, from, to string, err error, permanent bool) {
	if err == nil {
		l.printf("%s from=%s to=%s sent\n", id, from, to)
	} else {
		t := "(temporary)"
		if permanent {
			t = "(permanent)"
		}
		l.printf("%s from=%s to=%s failed %s: %v\n", id, from, to, t, err)
	}
}

// QueueLoop logs that we have completed a queue loop.
func (l *Logger) QueueLoop(id, from string, nextDelay time.Duration) {
	if nextDelay > 0 {
		l.printf("%s from=%s completed loop, next in %v\n", id, from, nextDelay)
	} else {
		l.printf("%s from=%s all done\n", id, from)
	}
}

// Default logger, used in the following top-level functions.
var Default *Logger = nil

// Listening logs that the daemon is listening on the given address.
func Listening(a string) {
	if Default == nil { return }
	Default.Listening(a)
}

// Auth logs an authentication request.
func Auth(netAddr net.Addr, user string, successful bool) {
	if Default == nil { return }
	Default.Auth(netAddr, user, successful)
}

// Rejected logs that we've rejected an email.
func Rejected(netAddr net.Addr, from string, to []string, err string) {
	if Default == nil { return }
	Default.Rejected(netAddr, from, to, err)
}

// Queued logs that we have queued an email.
func Queued(netAddr net.Addr, from string, to []string, id string) {
	if Default == nil { return }
	Default.Queued(netAddr, from, to, id)
}

// SendAttempt logs that we have attempted to send an email.
func SendAttempt(id, from, to string, err error, permanent bool) {
	if Default == nil { return }
	Default.SendAttempt(id, from, to, err, permanent)
}

// QueueLoop logs that we have completed a queue loop.
func QueueLoop(id, from string, nextDelay time.Duration) {
	if Default == nil { return }
	Default.QueueLoop(id, from, nextDelay)
}
