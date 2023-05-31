package log

import (
	"context"
	"errors"
	"testing"
	"time"
)

type LoggerThatClosesWithError struct {
	closeErrorMessage string
}

func (l *LoggerThatClosesWithError) SetMinimumLevel(logLevel LogLevel) {}
func (l *LoggerThatClosesWithError) GetMinimumLevel() LogLevel {
	return Debug
}
func (l *LoggerThatClosesWithError) SetUserPropertiesToLog(userPropertiesToLog *[]UserProperty) {}
func (l *LoggerThatClosesWithError) GetUserPropertiesToLog() *[]UserProperty                    { return nil }
func (l *LoggerThatClosesWithError) Log(level LogLevel, message string, err error, ctx context.Context) {
}
func (l *LoggerThatClosesWithError) Logf(level LogLevel, err error, ctx context.Context, format string, args ...interface{}) {
}
func (l *LoggerThatClosesWithError) Logln(level LogLevel, err error, ctx context.Context, args ...interface{}) {
}
func (l *LoggerThatClosesWithError) Close(timeout time.Duration) error {
	// Sleep for half the timeout, then error
	time.Sleep(timeout / 2)
	return errors.New(l.closeErrorMessage)
}

// Test LoggerSet.Close reports a logger error on Close
func TestLogSetClose(t *testing.T) {
	closeErrorMessage := "LoggerThatWontClose failed to close"
	loggerSet := NewLoggerSet(Info)
	loggerSet.AddLogger(&LoggerThatClosesWithError{
		closeErrorMessage: closeErrorMessage,
	})
	loggerSet.AddLogger(NewStandardLogger(Info))

	err := loggerSet.Close(time.Second)

	if err == nil || err.Error() != closeErrorMessage {
		t.Errorf("LoggerSet.Close failed to close with error: %s", closeErrorMessage)
	}
}
