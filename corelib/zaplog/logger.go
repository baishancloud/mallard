package zaplog

// Logger provide same apis to log message
type Logger interface {
	Debug(msg string, values ...interface{})
	Info(msg string, values ...interface{})
	Warn(msg string, values ...interface{})
	Error(msg string, values ...interface{})
	Fatal(msg string, values ...interface{})
	Sync()
	NewPrefix(string) Logger
	SetDebug(bool)
}

// NullLogger implements Logger but null operation
type NullLogger struct{}

// Debug implements Logger.Debug
func (nl *NullLogger) Debug(msg string, values ...interface{}) {}

// Info implements Logger.Info
func (nl *NullLogger) Info(msg string, values ...interface{}) {}

// Warn implements Logger.Warn
func (nl *NullLogger) Warn(msg string, values ...interface{}) {}

// Error implements Logger.Error
func (nl *NullLogger) Error(msg string, values ...interface{}) {}

// Fatal implements Logger.Fatal
func (nl *NullLogger) Fatal(msg string, values ...interface{}) {}

// Sync implements Logger.Sync
func (nl *NullLogger) Sync() {}

// NewPrefix implements Logger.NewPrefix
func (nl *NullLogger) NewPrefix(prefix string) Logger { return nl }

// SetDebug implements Logger.SetDebug
func (nl *NullLogger) SetDebug(debug bool) {}

// Null creates no usage logger
func Null() Logger {
	return &NullLogger{}
}
