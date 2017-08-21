package interfaces

// Logger defines a logging abstraction.
type Logger interface {
	WithField(key string, value interface{}) Logger
	WithError(err error) Logger
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}
