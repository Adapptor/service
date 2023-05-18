package log

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"
)

// Singleton LogSet
var loggerSet = NewLoggerSet(Info)

// A public reference to the singleton LogSet if required for injection
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

	// SetMinimumLevel sets the minimum log level; logs below this level will be ignored
	SetMinimumLevel(level LogLevel)

	// If applicable to the logger type, closes the logger with a timeout to flush all buffered log entries before returning
	Close(timeout time.Duration) error
}

// A set of log sinks (e.g., stdout, file, Sentry, etc.)
// Logs are sent to all sinks
type LoggerSet struct {
	loggers []Logger
}

// Create a new log set, with the standard logger
func NewLoggerSet(minimumLevel LogLevel) *LoggerSet {
	logSet := &LoggerSet{}
	logSet.AddLogger(NewStandardLogger(minimumLevel))

	return logSet
}

func (l *LoggerSet) AddLogger(logger Logger) {
	_addLogger(l, logger)
}

func AddLogger(logger Logger) {
	_addLogger(loggerSet, logger)
}

func _addLogger(loggerSet *LoggerSet, logger Logger) {
	if logger != nil {
		loggerSet.loggers = append(loggerSet.loggers, logger)
	}
}

// Convenience function to set the minimum log level for all
// current log sinks.
// Log sinks added after this call will not be affected
func (l *LoggerSet) SetMinimumLevel(logLevel LogLevel) {
	_setMinimumLevel(l, logLevel)
}

// Convenience function to set the minimum log level for all
// current log sinks.
// Log sinks added after this call will not be affected
func SetMinimumLevel(logLevel LogLevel) {
	_setMinimumLevel(loggerSet, logLevel)
}

func _setMinimumLevel(loggerSet *LoggerSet, logLevel LogLevel) {
	for _, logger := range loggerSet.loggers {
		logger.SetMinimumLevel(logLevel)
	}
}

func (l *LoggerSet) Log(level LogLevel, message string, err error, ctx context.Context) {
	_log(l, level, message, err, ctx)
}

func Log(level LogLevel, message string, err error, ctx context.Context) {
	_log(loggerSet, level, message, err, ctx)
}

func _log(loggerSet *LoggerSet, level LogLevel, message string, err error, ctx context.Context) {
	for _, logger := range loggerSet.loggers {
		logger.Log(level, message, err, ctx)
	}
}

func (l *LoggerSet) Close(timeout time.Duration) error {
	return _close(l, timeout)
}

func Close(timeout time.Duration) error {
	return _close(loggerSet, timeout)
}

func _close(loggerSet *LoggerSet, timeout time.Duration) error {
	errors := make(chan error, len(loggerSet.loggers))
	var wg sync.WaitGroup
	wg.Add(len(loggerSet.loggers))

	for _, logger := range loggerSet.loggers {
		go func(logger Logger) {
			defer wg.Done()
			errors <- logger.Close(timeout)
		}(logger)
	}

	go func() {
		wg.Wait()
		close(errors)
	}()

	for err := range errors {
		// return the first error
		if err != nil {
			return err
		}
	}

	return nil
}

// Add a file logger with the given file name
func SetupLog(logfile string, minLogLevel LogLevel) Logger {
	fileLogger := NewFileLogger(logfile, minLogLevel, 500, 3, 28)
	loggerSet.AddLogger(fileLogger)
	return loggerSet
}

// Write the given JSON object to the standard log
func LogDump(logger *log.Logger, obj interface{}) {
	js, _ := json.Marshal(obj)
	log.Printf("%v\n", string(js))
}

// Log the contents of a reader
func LogReader(level LogLevel, reader io.Reader, prefix string) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	loggerSet.Log(level, buf.String(), nil, nil)
}
