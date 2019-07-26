package service

import (
	"cloud.google.com/go/logging"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

// ensure we always implement io.WriteCloser
var _ io.WriteCloser = (*StackdriverWriter)(nil)

type StackdriverWriter struct {
	Client *logging.Client
	Logger *logging.Logger
	mu     sync.Mutex
}

const kDropLog = logging.Severity(-1)

var severityMap = map[string]logging.Severity{
	"TRACE":   kDropLog,
	"DEBUG":   logging.Debug,
	"INFO":    logging.Info,
	"WARNING": logging.Warning,
	"ERROR":   logging.Error,
}

func NewStackdriverWriter(cfg BaseConfig) (*StackdriverWriter, error) {
	if len(cfg.Google.LogName) < 1 {
		return nil, errors.New("Google log name not configured")
	}

	if len(cfg.Google.Project) < 1 {
		return nil, errors.New("Google project name not configured")
	}

	ctx := context.Background()

	logName := fmt.Sprintf("%v-%v", cfg.Google.LogName, cfg.ConfigName)
	// Creates a client.
	client, err := logging.NewClient(ctx, cfg.Google.Project)
	if err != nil {
		return nil, err
	}

	return &StackdriverWriter{
		Client: client,
		Logger: client.Logger(logName),
		mu:     sync.Mutex{},
	}, nil
}

func (sw *StackdriverWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	logText := string(p[:len(p)])
	logLevel := logging.Default
	for prefix, severity := range severityMap {
		if strings.HasPrefix(logText, prefix+":") {
			logLevel = severity
			break
		}
	}

	// Adds an entry to the log buffer.
	if logLevel > kDropLog {
		sw.Logger.Log(logging.Entry{Severity: logLevel, Payload: logText})
	}
	return n, nil
}

func (sw *StackdriverWriter) Close() error {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	// Closes the client and flushes the buffer to the Stackdriver Logging
	// service.
	return sw.Client.Close()
}

func (sw *StackdriverWriter) Flush() {
	sw.Logger.Flush()
}
