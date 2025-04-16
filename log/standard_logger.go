package log

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// A standard go logger with log levels
type StandardLogger struct {
	minimumLevel        LogLevel
	levelLoggers        map[LogLevel]*log.Logger
	userPropertiesToLog *[]UserProperty
}

func NewStandardLogger(minimumLevel LogLevel) *StandardLogger {
	// Create level loggers
	levelLoggers := make(map[LogLevel]*log.Logger)
	for _, logLevel := range LogLevels {
		levelLoggers[logLevel] = log.New(os.Stderr, fmt.Sprintf("%s: ", logLevel.String()), log.Ldate|log.Ltime)
	}

	return &StandardLogger{
		minimumLevel: minimumLevel,
		levelLoggers: levelLoggers,
	}
}

func (l *StandardLogger) SetMinimumLevel(level LogLevel) {
	l.minimumLevel = level
}

func (l *StandardLogger) GetMinimumLevel() LogLevel {
	return l.minimumLevel
}

func (l *StandardLogger) SetUserPropertiesToLog(userPropertiesToLog *[]UserProperty) {
	l.userPropertiesToLog = userPropertiesToLog
}

func (l *StandardLogger) GetUserPropertiesToLog() *[]UserProperty { return l.userPropertiesToLog }

func (l *StandardLogger) logInternal(level LogLevel, message string, err error, ctx context.Context) {
	if level >= l.minimumLevel {
		// Get the calling function skipping 4 frames so we can print the actual caller
		if _, file, line, ok := runtime.Caller(4); ok {
			lastSlash := strings.LastIndex(file, "/")
			fileName := file[lastSlash+1:]
			message = fmt.Sprintf("%s:%d: %s", fileName, line, message)
		}

		userProperties := GetUserPropertiesString(ctx, l.userPropertiesToLog)
		if userProperties != nil {
			message = fmt.Sprintf("%s (%s)", message, *userProperties)
		}

		if err == nil {
			l.levelLoggers[level].Println(message)
		} else {
			l.levelLoggers[level].Printf("%s, %+v\n", message, err)
		}
	}
}

func (l *StandardLogger) Log(level LogLevel, message string, err error, ctx context.Context) {
	if level >= l.minimumLevel {
		l.logInternal(level, message, err, ctx)
	}
}

func (l *StandardLogger) Logf(level LogLevel, err error, ctx context.Context, format string, args ...interface{}) {
	if level >= l.minimumLevel {
		l.logInternal(level, fmt.Sprintf(format, args...), err, ctx)
	}
}

func (l *StandardLogger) Logln(level LogLevel, err error, ctx context.Context, args ...interface{}) {
	if level >= l.minimumLevel {
		message := fmt.Sprintln(args...)

		// Remove the trailing newline from message as Log writes a newline
		if len(message) > 0 && message[len(message)-1] == '\n' {
			message = message[:len(message)-1]
		}

		l.logInternal(level, message, err, ctx)
	}
}

func (l *StandardLogger) Close(timeout time.Duration) error {
	// no-op
	return nil
}
