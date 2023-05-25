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

	// Logger interface
	minimumLevel LogLevel
}

const DropLog = logging.Severity(-1)

// Map of log levels to Stackdriver log levels.
var logLevelToStackDriverSeverity = map[LogLevel]logging.Severity{
	Trace:   DropLog,
	Debug:   logging.Debug,
	Info:    logging.Info,
	Warning: logging.Warning,
	Error:   logging.Error,
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

func (sw *StackdriverWriter) Flush() {
	sw.Logger.Flush()
}

func (sw *StackdriverWriter) SetMinimumLevel(level LogLevel) {
	sw.minimumLevel = level
}

func (sw *StackdriverWriter) Log(level LogLevel, message string, err error, ctx context.Context) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	if level >= sw.minimumLevel {
		if err == nil {
			sw.Logger.Log(logging.Entry{Severity: logLevelToStackDriverSeverity[level], Payload: fmt.Sprintf("%v", message)})
		} else {
			sw.Logger.Log(logging.Entry{Severity: logLevelToStackDriverSeverity[level], Payload: fmt.Sprintf("%v, %+v", message, err)})
		}
	}
}

// Closes the client and flushes the buffer to the Stackdriver Logging
// service.
func (sw *StackdriverWriter) Close() error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	return sw.Client.Close()
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
func (sw *StackdriverWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

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
		sw.Logger.Log(logging.Entry{Severity: logLevel, Payload: logText})
	}
	return n, nil
}
