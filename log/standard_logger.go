package log

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// A standard go logger with log levels
type StandardLogger struct {
	minimumLevel LogLevel
	levelLoggers map[LogLevel]*log.Logger
}

func NewStandardLogger(minimumLevel LogLevel) *StandardLogger {
	// Create level loggers
	levelLoggers := make(map[LogLevel]*log.Logger)
	for _, logLevel := range LogLevels {
		levelLoggers[logLevel] = log.New(os.Stderr, fmt.Sprintf("%s: ", logLevel.String()), log.Ldate|log.Ltime|log.Lshortfile)
	}

	return &StandardLogger{
		minimumLevel: minimumLevel,
		levelLoggers: levelLoggers,
	}
}

func (l *StandardLogger) SetMinimumLevel(level LogLevel) {
	l.minimumLevel = level
}

func (l *StandardLogger) Log(level LogLevel, message string, err error, ctx context.Context) {
	if level >= l.minimumLevel {
		if err == nil {
			l.levelLoggers[level].Println(message)
		} else {
			l.levelLoggers[level].Printf("%s, %+v\n", message, err)
		}
	}
}

func (l *StandardLogger) Close(timeout time.Duration) error {
	// no-op
	return nil
}
