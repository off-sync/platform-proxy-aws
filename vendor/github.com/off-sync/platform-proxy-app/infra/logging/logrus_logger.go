package logging

import (
	"github.com/sirupsen/logrus"
	"github.com/off-sync/platform-proxy-app/interfaces"
)

type logrusLogger struct {
	l *logrus.Logger
}

// NewLogrusLogger creates a new interfaces.Logger using the provided logrus
// logger.
func NewLogrusLogger(logger *logrus.Logger) interfaces.Logger {
	return &logrusLogger{l: logger}
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.l.Debug(args...)
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.l.Info(args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.l.Warn(args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.l.Error(args...)
}

func (l *logrusLogger) Fatal(args ...interface{}) {
	l.l.Fatal(args...)
}

func (l *logrusLogger) WithField(field string, value interface{}) interfaces.Logger {
	return &logrusEntry{e: l.l.WithField(field, value)}
}

func (l *logrusLogger) WithError(err error) interfaces.Logger {
	return &logrusEntry{e: l.l.WithError(err)}
}

type logrusEntry struct {
	e *logrus.Entry
}

func (e *logrusEntry) Debug(args ...interface{}) {
	e.e.Debug(args...)
}

func (e *logrusEntry) Info(args ...interface{}) {
	e.e.Info(args...)
}

func (e *logrusEntry) Warn(args ...interface{}) {
	e.e.Warn(args...)
}

func (e *logrusEntry) Error(args ...interface{}) {
	e.e.Error(args...)
}

func (e *logrusEntry) Fatal(args ...interface{}) {
	e.e.Fatal(args...)
}

func (e *logrusEntry) WithField(field string, value interface{}) interfaces.Logger {
	return &logrusEntry{e: e.e.WithField(field, value)}
}

func (e *logrusEntry) WithError(err error) interfaces.Logger {
	return &logrusEntry{e: e.e.WithError(err)}
}
