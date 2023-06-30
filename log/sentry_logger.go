package log

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

// A Sentry logger
type SentryLogger struct {
	minimumLevel        LogLevel
	userPropertiesToLog *[]UserProperty
}

// NewSentryLogger creates a new Sentry logger
//
// dsn: Sentry DSN
// debug: whether to log Sentry SDK debug messages
// environment: the service environment (e.g., dev, staging, production)
// release: the service release version (e.g., v1.2.3)
// tags: custom tags to add to all events (e.g., "service" => "acme api")
// minimumLevel: minimum log level
// userPropertiesToLog: user properties to log (e.g., userId, email)
func NewSentryLogger(dsn string, debug bool, environment string, release string, tags *map[string]string, minimumLevel LogLevel) (*SentryLogger, error) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Debug:       debug,
		Environment: environment,
		Release:     release,
		// Ensure stack traces are attached to messages as well as exceptions
		AttachStacktrace: true,
	}); err != nil {
		log.Printf("error initialising Sentry: %+v\n", err)
		return nil, err
	}

	if tags != nil {
		// Add custom tags to the Sentry scope; these will be reported with every event
		for key, value := range *tags {
			sentry.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTag(key, value)
			})
		}
	}

	return &SentryLogger{minimumLevel: minimumLevel}, nil
}

func (l *SentryLogger) SetMinimumLevel(level LogLevel) {
	l.minimumLevel = level
}

func (l *SentryLogger) GetMinimumLevel() LogLevel {
	return l.minimumLevel
}

func (l *SentryLogger) SetUserPropertiesToLog(userPropertiesToLog *[]UserProperty) {
	l.userPropertiesToLog = userPropertiesToLog
}

func (l *SentryLogger) GetUserPropertiesToLog() *[]UserProperty { return l.userPropertiesToLog }

// Log levels below `Warning` are added as breadcrumbs, unless they fall below the configured minimum level.
func (l *SentryLogger) Log(level LogLevel, message string, err error, ctx context.Context) {
	if level >= l.minimumLevel {
		switch level {
		case Trace, Debug, Info:
			breadcrumb := sentry.Breadcrumb{
				Type:     level.String(),
				Category: "",
				Data:     nil,
				Message:  message,
			}
			sentry.AddBreadcrumb(&breadcrumb)
		case Warning, Error, Fatal:
			l.CaptureEvent(message, err, level, ctx)
		}
	}
}

func (l *SentryLogger) Logf(level LogLevel, err error, ctx context.Context, format string, args ...interface{}) {
	if level >= l.minimumLevel {
		l.Log(level, fmt.Sprintf(format, args...), err, ctx)
	}
}

func (l *SentryLogger) Logln(level LogLevel, err error, ctx context.Context, args ...interface{}) {
	if level >= l.minimumLevel {
		l.Log(level, fmt.Sprintln(args...), err, ctx)
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
//
// If the provided context includes user information, it will be associated
// with this event.
func (l *SentryLogger) CaptureEvent(message string, err error, level LogLevel, ctx context.Context) {
	var hub *sentry.Hub = sentry.CurrentHub()
	sentryLevel := GetSentryLevel(level)

	if ctx != nil {
		// Attempt to get the hub associated with the given context
		if tempHub := sentry.GetHubFromContext(ctx); tempHub != nil {
			hub = tempHub
		}
	}

	client := hub.Client()
	sentryUser := getUser(ctx, l.userPropertiesToLog)

	if client != nil {
		if err != nil {
			event := client.EventFromException(err, sentryLevel)
			if sentryUser != nil {
				hub.WithScope(func(s *sentry.Scope) {
					s.SetUser(*sentryUser)
					hub.CaptureEvent(event)
				})
			} else {
				hub.CaptureEvent(event)
			}
		} else {
			event := client.EventFromMessage(message, sentryLevel)
			if sentryUser != nil {
				hub.WithScope(func(s *sentry.Scope) {
					s.SetUser(*sentryUser)
					hub.CaptureEvent(event)
				})
			} else {
				hub.CaptureEvent(event)
			}
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
	case Fatal:
		return sentry.LevelFatal
	}
	return sentry.LevelInfo
}

// Examines the supplied context for user properties that can be
// associated with the log event, and returns a Sentry user, or
// nil if no user properties are found.
func getUser(ctx context.Context, userPropertiesToLog *[]UserProperty) *sentry.User {
	haveUserToLog := false
	var sentryUser sentry.User

	userPropertiesMap := GetUserPropertiesMap(ctx)

	if userPropertiesMap != nil {
		if id, ok := (*userPropertiesMap)[UserPropertyId]; ok && ContainsUserProperty(*userPropertiesToLog, UserPropertyId) {
			sentryUser.ID = id
			haveUserToLog = true
		}
		if email, ok := (*userPropertiesMap)[UserPropertyEmail]; ok && ContainsUserProperty(*userPropertiesToLog, UserPropertyEmail) {
			sentryUser.Email = email
			haveUserToLog = true
		}
		if userName, ok := (*userPropertiesMap)[UserPropertyName]; ok && ContainsUserProperty(*userPropertiesToLog, UserPropertyName) {
			sentryUser.Username = userName
			haveUserToLog = true
		}
	}

	if haveUserToLog {
		return &sentryUser
	} else {
		return nil
	}
}
