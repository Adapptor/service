package log

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

// A Sentry logger
type SentryLogger struct {
	minimumLevel LogLevel
}

// Create a new sentry logger
func NewSentryLogger(dsn string, debug bool, environment string, release string, minimumLevel LogLevel) (*SentryLogger, error) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Debug:       debug,
		Environment: environment,
		Release:     release,
		// Ensure stack traces are attached to messages as well as exceptions
		AttachStacktrace: true,
	}); err != nil {
		log.Printf("error initialisaing Sentry: %+v\n", err)
		return nil, err
	}

	return &SentryLogger{minimumLevel: minimumLevel}, nil
}

func (l *SentryLogger) SetMinimumLevel(level LogLevel) {
	l.minimumLevel = level
}

func (l *SentryLogger) Log(level LogLevel, message string, err error, ctx context.Context) {
	if level >= l.minimumLevel {
		switch level {
		case Trace, Debug:
			CaptureEvent(message, err, level, ctx)
		case Info:
			breadcrumb := sentry.Breadcrumb{
				Type:     level.String(),
				Category: "",
				Data:     nil,
				Message:  message,
			}
			sentry.AddBreadcrumb(&breadcrumb)
			CaptureEvent(message, err, level, ctx)
		case Warning, Error:
			CaptureEvent(message, err, level, ctx)
		}
	}
}

// Flushes the Sentry logger
func (l *SentryLogger) Close(timeout time.Duration) error {
	if !sentry.Flush(timeout) {
		return errors.New("failed to flush Sentry logger")
	}
	return nil
}

// CaptureEvent sends an event to Sentry
// If an error is present, it will be sent as an exception, otherwise
// it will be sent as a message
func CaptureEvent(message string, err error, level LogLevel, ctx context.Context) {
	var hub *sentry.Hub = sentry.CurrentHub()
	sentryLevel := GetSentryLevel(level)

	if ctx != nil {
		// Attempt to get the hub associated with the given context
		if tempHub := sentry.GetHubFromContext(ctx); tempHub != nil {
			hub = tempHub
		}
	}

	client := hub.Client()

	if client != nil {
		if err != nil {
			event := client.EventFromException(err, sentryLevel)
			hub.CaptureEvent(event)
		} else {
			event := client.EventFromMessage(message, sentryLevel)
			hub.CaptureEvent(event)
		}
	} else {
		log.Printf("%s: failed to find top-level Sentry client: %+v\n", Debug.String(), err)
		if err != nil {
			sentry.CaptureException(err)
		} else {
			sentry.CaptureMessage(message)
		}
	}
}

// GetSentryLevel Get the Sentry severity level correspodning to the given LogLevel
func GetSentryLevel(logLevel LogLevel) sentry.Level {
	switch logLevel {
	case Trace, Debug:
		return sentry.LevelDebug
	case Info:
		return sentry.LevelInfo
	case Warning:
		return sentry.LevelWarning
	case Error:
		return sentry.LevelError
	}
	return sentry.LevelInfo
}
