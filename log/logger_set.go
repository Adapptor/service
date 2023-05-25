package log

import (
	"context"
	"sync"
	"time"
)

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
	if logger != nil {
		l.loggers = append(l.loggers, logger)
	}
}

// Convenience function to set the minimum log level for all
// current log sinks.
// Log sinks added after this call will not be affected
func (l *LoggerSet) SetMinimumLevel(logLevel LogLevel) {
	for _, logger := range l.loggers {
		logger.SetMinimumLevel(logLevel)
	}
}

func (l *LoggerSet) Log(level LogLevel, message string, err error, ctx context.Context) {
	for _, logger := range l.loggers {
		logger.Log(level, message, err, ctx)
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
