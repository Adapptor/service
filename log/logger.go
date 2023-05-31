package log

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"time"
)

// Singleton LoggerSet
var loggerSet = NewLoggerSet(Info)

// A public reference to the singleton LoggerSet if required for injection
var L = loggerSet

// A simple logging interface
type Logger interface {
	// Log logs a message with the specified log level and error (if applicable), along with the provided context.
	//
	// level: The log level for the message.
	// message: The message to log.
	// err: An optional error associated with the message (if any).
	// ctx: An optional context to log with the message.
	Log(level LogLevel, message string, err error, ctx context.Context)

	// Logf logs a formatted message with the specified log level and error (if applicable), along with the provided context.
	Logf(level LogLevel, err error, ctx context.Context, format string, args ...interface{})

	// Logln logs a message with the specified log level and error (if applicable), along with the provided context.
	Logln(level LogLevel, err error, ctx context.Context, args ...interface{})

	// SetMinimumLevel sets the minimum log level; logs below this level will be ignored
	SetMinimumLevel(level LogLevel)

	// GetMinimumLevel returns the most recently set minimum log level
	GetMinimumLevel() LogLevel

	// SetUserPropertiesToLog sets the user properties to log
	SetUserPropertiesToLog(userPropertiesToLog *[]UserProperty)

	// GetUserPropertiesToLog returns the most recently set user properties to log
	GetUserPropertiesToLog() *[]UserProperty

	// If applicable to the logger type, closes the logger with a timeout to flush all buffered log entries before returning
	Close(timeout time.Duration) error
}

func AddLogger(logger Logger) {
	loggerSet.AddLogger(logger)
}

// Convenience function to set the minimum log level for all
// current log sinks.
//
// Note: Log sinks added after this call will not be affected
func SetMinimumLevel(logLevel LogLevel) {
	loggerSet.SetMinimumLevel(logLevel)
}

// Convenience function to get the minimum log level for all
// current log sinks.
//
// Note: Log sinks added after the most recent call of SetMinimumLevel
// can have different minimum log levels.
func GetMinimumLevel() LogLevel {
	return loggerSet.GetMinimumLevel()
}

// Convenience function to set the user properties to log for all
// current log sinks.
//
// Note: Log sinks added after this call will not be affected
func SetUserPropertiesToLog(userPropertiesToLog *[]UserProperty) {
	loggerSet.userPropertiesToLog = userPropertiesToLog

	for _, logger := range loggerSet.loggers {
		logger.SetUserPropertiesToLog(userPropertiesToLog)
	}
}

// Convenience function to get the user properties to log
// for all current log sinks
//
// Note: Log sinks added after the most recent call of SetUserPropertiesToLog
// can have different values.
func GetUserPropertiesToLog() *[]UserProperty { return loggerSet.userPropertiesToLog }

func Log(level LogLevel, message string, err error, ctx context.Context) {
	loggerSet.Log(level, message, err, ctx)
}

func Logf(level LogLevel, err error, ctx context.Context, format string, args ...interface{}) {
	loggerSet.Logf(level, err, ctx, format, args...)
}

func Logln(level LogLevel, err error, ctx context.Context, args ...interface{}) {
	loggerSet.Logln(level, err, ctx, args...)
}

func Close(timeout time.Duration) error {
	return loggerSet.Close(timeout)
}

// Log the contents of a reader
func LogReader(level LogLevel, reader io.Reader, prefix string) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	loggerSet.Log(level, buf.String(), nil, nil)
}

// Add a file logger with the given file name
//
// Deprecated: use NewFileLogger instead
func SetupLog(logfile string, minLogLevel LogLevel) Logger {
	fileLogger := NewFileLogger(logfile, minLogLevel, 500, 3, 28)
	loggerSet.AddLogger(fileLogger)
	return loggerSet
}

// Write the given JSON object to the standard log
//
// Deprecated: use Log instead
func LogDump(logger *log.Logger, obj interface{}) {
	js, _ := json.Marshal(obj)
	log.Printf("%v\n", string(js))
}
