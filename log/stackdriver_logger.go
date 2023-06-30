package log

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"cloud.google.com/go/logging"
	"google.golang.org/api/option"
)

// ensure we always implement io.WriteCloser
var _ io.WriteCloser = (*StackdriverWriter)(nil)

type StackdriverWriter struct {
	Client *logging.Client
	Logger *logging.Logger
	mu     sync.Mutex

	minimumLevel        LogLevel
	userPropertiesToLog *[]UserProperty
}

const DropLog = logging.Severity(-1)

// Map of log levels to Stackdriver log levels.
var logLevelToStackDriverSeverity = map[LogLevel]logging.Severity{
	Trace:   DropLog,
	Debug:   logging.Debug,
	Info:    logging.Info,
	Warning: logging.Warning,
	Error:   logging.Error,
	Fatal:   logging.Critical,
}

func NewStackdriverWriter(configName string, googleLogName string, googleProject string, opts ...option.ClientOption) (*StackdriverWriter, error) {

	if len(googleLogName) < 1 {
		return nil, errors.New("google log name not configured")
	}

	if len(googleProject) < 1 {
		return nil, errors.New("google project name not configured")
	}

	ctx := context.Background()

	logName := fmt.Sprintf("%v-%v", googleLogName, configName)
	// Creates a client.
	client, err := logging.NewClient(ctx, googleProject, opts...)
	if err != nil {
		return nil, err
	}

	return &StackdriverWriter{
		Client: client,
		Logger: client.Logger(logName),
		mu:     sync.Mutex{},
	}, nil
}

func (l *StackdriverWriter) Flush() {
	l.Logger.Flush()
}

func (l *StackdriverWriter) SetMinimumLevel(level LogLevel) {
	l.minimumLevel = level
}

func (l *StackdriverWriter) GetMinimumLevel() LogLevel {
	return l.minimumLevel
}

func (l *StackdriverWriter) SetUserPropertiesToLog(userPropertiesToLog *[]UserProperty) {
	l.userPropertiesToLog = userPropertiesToLog
}

func (l *StackdriverWriter) GetUserPropertiesToLog() *[]UserProperty { return l.userPropertiesToLog }

func (l *StackdriverWriter) Log(level LogLevel, message string, err error, ctx context.Context) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level >= l.minimumLevel {
		userProperties := GetUserPropertiesString(ctx, l.userPropertiesToLog)
		if userProperties != nil {
			message = fmt.Sprintf("%s (%s)", message, *userProperties)
		}

		if err == nil {
			l.Logger.Log(logging.Entry{Severity: logLevelToStackDriverSeverity[level], Payload: fmt.Sprintf("%v", message)})
		} else {
			l.Logger.Log(logging.Entry{Severity: logLevelToStackDriverSeverity[level], Payload: fmt.Sprintf("%v, %+v", message, err)})
		}
	}
}

func (l *StackdriverWriter) Logf(level LogLevel, err error, ctx context.Context, format string, args ...interface{}) {
	if level >= l.minimumLevel {
		l.Log(level, fmt.Sprintf(format, args...), err, ctx)
	}
}

func (l *StackdriverWriter) Logln(level LogLevel, err error, ctx context.Context, args ...interface{}) {
	if level >= l.minimumLevel {
		l.Log(level, fmt.Sprintln(args...), err, ctx)
	}
}

// Closes the client and flushes the buffer to the Stackdriver Logging
// service.
func (l *StackdriverWriter) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.Client.Close()
}

// Deprecated
var severityMap = map[string]logging.Severity{
	"TRACE":   DropLog,
	"DEBUG":   logging.Debug,
	"INFO":    logging.Info,
	"WARNING": logging.Warning,
	"ERROR":   logging.Error,
}

// Deprecated: use Log(...)
func (l *StackdriverWriter) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	logText := string(p[:])
	logLevel := logging.Default
	for prefix, severity := range severityMap {
		if strings.HasPrefix(logText, prefix+":") {
			logLevel = severity
			break
		}
	}

	// Adds an entry to the log buffer.
	if logLevel > DropLog {
		l.Logger.Log(logging.Entry{Severity: logLevel, Payload: logText})
	}
	return n, nil
}
