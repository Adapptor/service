package log

import "strings"

// A log severity level
type LogLevel int

const (
	Trace LogLevel = iota
	Debug
	// Info-level logs indicate normal service lifecycle events, 'valuable' events for audit, etc.
	Info
	// Warning-level logs indicate events that are unexpected but tolerable, which may need further investigation
	Warning
	// Error-level logs indicate errors that should be investigated
	Error
	// Fatal-level logs indicate errors that terminated the service and should be investigated
	Fatal
)

// All log levels
var LogLevels = [...]LogLevel{Trace, Debug, Info, Warning, Error, Fatal}
var LogLevelStrings = [...]string{"TRACE", "DEBUG", "INFO", "WARNING", "ERROR", "FATAL"}

func (l LogLevel) String() string {
	return LogLevelStrings[l]
}

// GetLogger Get the LogLevel that matches the given log level string.
// Defaults to Info.
func GetLogLevel(level string) LogLevel {

	switch strings.ToUpper(level) {
	case "TRACE":
		return Trace
	case "DEBUG":
		return Debug
	case "INFO":
		return Info
	case "WARNING":
		return Warning
	case "ERROR":
		return Error
	case "FATAL":
		return Fatal
	default:
		return Info
	}
}
