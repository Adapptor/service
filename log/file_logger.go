package log

import (
	"context"
	"fmt"
	"log"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// A rolling file logger
type FileLogger struct {
	minimumLevel     LogLevel
	lumberjackLogger *lumberjack.Logger
	levelLoggers     map[LogLevel]*log.Logger
}

func NewFileLogger(filename string, minimumLevel LogLevel, maximumSizeMegabytes int, maximumRetainedLogFilesCount int, maximumRetainedLogFilesAgeDays int) *FileLogger {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maximumSizeMegabytes,
		MaxBackups: maximumRetainedLogFilesCount,
		MaxAge:     maximumRetainedLogFilesAgeDays,
	}

	levelLoggers := make(map[LogLevel]*log.Logger)
	for _, logLevel := range LogLevels {
		levelLoggers[logLevel] = log.New(lumberjackLogger, fmt.Sprintf("%s: ", logLevel.String()), log.Ldate|log.Ltime|log.Lshortfile)
	}

	return &FileLogger{
		minimumLevel: minimumLevel,
		// Wrap the lumberjack logger with go standard loggers to get decorations consistent with the StandardLogger
		levelLoggers:     levelLoggers,
		lumberjackLogger: lumberjackLogger,
	}
}

func (l *FileLogger) Log(level LogLevel, message string, err error, ctx context.Context) {
	if level >= l.minimumLevel {
		if err == nil {
			l.levelLoggers[level].Println(message)
		} else {
			l.levelLoggers[level].Printf("%s, %+v\n", message, err)
		}
	}
}

func (l *FileLogger) SetMinimumLevel(level LogLevel) {
	l.minimumLevel = level
}

func (l *FileLogger) Close(timeout time.Duration) error {
	return l.lumberjackLogger.Close()
}
