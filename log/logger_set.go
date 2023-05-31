package log

import (
	"context"
	"sync"
	"time"
)

// A set of log sinks (e.g., stdout, file, Sentry, etc.)
// Logs are sent to all sinks
type LoggerSet struct {
	loggers             []Logger
	minimumLevel        LogLevel
	userPropertiesToLog *[]UserProperty
}

// Create a new log set, with the standard logger
func NewLoggerSet(minimumLevel LogLevel) *LoggerSet {
	logSet := &LoggerSet{}
	logSet.AddLogger(NewStandardLogger(minimumLevel))

	return logSet
}

func (l *LoggerSet) AddLogger(logger Logger) {
	if logger != nil {
		l.loggers = append(l.loggers, logger)
	}
}

// Convenience function to set the minimum log level for all
// current log sinks.
//
// Note: Log sinks added after this call will not be affected
func (l *LoggerSet) SetMinimumLevel(logLevel LogLevel) {
	for _, logger := range l.loggers {
		logger.SetMinimumLevel(logLevel)
	}
	l.minimumLevel = logLevel
}

// Convenience function to get the minimum log level
// for all current log sinks
//
// Note: Log sinks added after the most recent call of SetMinimumLevel
// can have different minimum log levels.
func (l *LoggerSet) GetMinimumLevel() LogLevel {
	return l.minimumLevel
}

// Convenience function to set the user properties to log for all
// current log sinks.
//
// Note: Log sinks added after this call will not be affected
func (l *LoggerSet) SetUserPropertiesToLog(userPropertiesToLog *[]UserProperty) {
	l.userPropertiesToLog = userPropertiesToLog

	for _, logger := range l.loggers {
		logger.SetUserPropertiesToLog(userPropertiesToLog)
	}
}

// Convenience function to get the user properties to log
// for all current log sinks
//
// Note: Log sinks added after the most recent call of SetUserPropertiesToLog
// can have different values.
func (l *LoggerSet) GetUserPropertiesToLog() *[]UserProperty { return l.userPropertiesToLog }

func (l *LoggerSet) Log(level LogLevel, message string, err error, ctx context.Context) {
	for _, logger := range l.loggers {
		logger.Log(level, message, err, ctx)
	}
}

func (l *LoggerSet) Logf(level LogLevel, err error, ctx context.Context, format string, args ...interface{}) {
	for _, logger := range l.loggers {
		logger.Logf(level, err, ctx, format, args...)
	}
}

func (l *LoggerSet) Logln(level LogLevel, err error, ctx context.Context, args ...interface{}) {
	for _, logger := range l.loggers {
		logger.Logln(level, err, ctx, args...)
	}
}

func (l *LoggerSet) Close(timeout time.Duration) error {
	errors := make(chan error)
	var wg sync.WaitGroup
	wg.Add(len(l.loggers))

	for _, logger := range l.loggers {
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
