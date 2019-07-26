//  Adapptor helpers for writing web services
package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type ServerType int

const (
	Production ServerType = iota
	Staging
	Development
	LiveTest
	UAT
)

func (s ServerType) String() string {
	return [...]string{"Production", "Staging", "Development", "LiveTest", "UAT"}[s]
}

type LogType int

const (
	Trace LogType = iota
	Debug
	Info
	Warning
	Error
)

func (l LogType) String() string {
	return [...]string{"Trace", "Debug", "Info", "Warning", "Error"}[l]
}

type Logs struct {
	Trace   *log.Logger
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func LogDump(logger *log.Logger, obj interface{}) {
	js, _ := json.Marshal(obj)
	logger.Printf("%v", string(js))
}

type LogLevelInfo struct {
	Logger **log.Logger
	Tag    string
}

var Log *Logs
var logLevelMap map[LogType]LogLevelInfo

func SetupLog(logfile string, minLogLevel LogType) *Logs {
	return SetupLogWriters(logfile, []io.Writer{}, minLogLevel)
}

func SetupStackdriverLog(logfile string, cfg BaseConfig, minLogLevel LogType) (*Logs, error) {
	stackdriverWriter, err := NewStackdriverWriter(cfg)
	if err != nil {
		return nil, err
	}

	extraLogWriters := []io.Writer{stackdriverWriter}
	return SetupLogWriters(logfile, extraLogWriters, minLogLevel), nil
}

func SetupLogWriters(logfile string, extraLogWriters []io.Writer, minLogLevel LogType) *Logs {
	if Log == nil {
		//  initialize with stderr logger
		Log = setupTempLog()
	}

	logLevelMap = map[LogType]LogLevelInfo{
		Trace:   {&Log.Trace, "TRACE"},
		Debug:   {&Log.Debug, "DEBUG"},
		Info:    {&Log.Info, "INFO"},
		Warning: {&Log.Warning, "WARNING"},
		Error:   {&Log.Error, "ERROR"},
	}

	fileLogWriter := &lumberjack.Logger{
		Filename:   logfile,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	}

	//  Combine log writer streams
	logWriters := []io.Writer{os.Stderr, fileLogWriter}
	logWriters = append(logWriters, extraLogWriters...)
	logWriter := io.MultiWriter(logWriters...)

	for _, logType := range []LogType{Trace, Debug, Info, Warning, Error} {
		var levelWriter *io.Writer
		if logType >= minLogLevel {
			levelWriter = &logWriter
		} else {
			levelWriter = &ioutil.Discard
		}

		logLevelInfo := logLevelMap[logType]
		var logger **log.Logger = logLevelInfo.Logger
		*logger = log.New(*levelWriter, fmt.Sprintf("%v: ", logLevelInfo.Tag), log.Ldate|log.Ltime|log.Lshortfile)
	}

	return Log
}

func setupTempLog() *Logs {
	//  Set up a temporary logger to stdout while loading config
	tempLogger := log.New(os.Stderr, "INIT: ", log.Ldate|log.Ltime|log.Lshortfile)

	logs := Logs{
		Trace:   tempLogger,
		Debug:   tempLogger,
		Info:    tempLogger,
		Warning: tempLogger,
		Error:   tempLogger,
	}

	return &logs
}

// GetLogger Get the LogType that matches the given log level string.
// Defaults to Info.
func GetLogType(logLevel string) LogType {

	switch strings.ToLower(logLevel) {
	case "trace":
		return Trace
	case "debug":
		return Debug
	case "info":
		return Info
	case "warning":
		return Warning
	case "error":
		return Error
	default:
		return Info
	}
}

type Map map[string]interface{}

func (m *Map) GetString(key string) string {
	var vstr string

	if value, ok := (*m)[key]; ok {
		vstr, _ = value.(string)
	}

	return vstr
}

func ReadJsonMap(reader io.Reader) (Map, error) {
	decoder := json.NewDecoder(reader)
	var body Map
	err := decoder.Decode(&body)
	return body, err
}

type MapError struct {
	s string
}

var Nil = MapError{s: "service.MapNil"}
var Remove = MapError{s: "service.MapRemove"}

func (m *Map) UpdatePath(path string, updateValue interface{}) (interface{}, error) {
	return m.traversePath(path, updateValue, true)
}

func (m *Map) UpdateExistingPath(path string, updateValue interface{}) (interface{}, error) {
	return m.traversePath(path, updateValue, false)
}

func (m *Map) QueryPath(path string) (interface{}, error) {
	return m.traversePath(path, nil, false)
}

// Traverse the given slash separated path and update if a value is provided,
// returning the updated value.  Otherwise return the value at the given path.
//
// updateValue can be a valid JSON value (map, string, number, bool). An
// updateValue of nil requests a query, Nil requests a value of nil, Remove
// requests a deletion
//
// If createPaths is true, any missing path components are initialized as
// empty maps
func (m *Map) traversePath(path string, updateValue interface{}, createPaths bool) (interface{}, error) {
	path = strings.TrimPrefix(path, "/")
	components := strings.Split(path, "/")

	Log.Debug.Printf("Traversing path %v with value %v", path, updateValue)

	if len(components) == 1 && components[0] == "" && updateValue != nil {
		//  Complete replacement of root map, updateValue must be a generic map
		*m = Map(updateValue.(map[string]interface{}))
		return m, nil
	}

	var lastIndex = len(components) - 1

	ref := *m
	var child interface{} = nil

	for i, component := range components {
		var ok bool

		if component == "" {
			return nil, fmt.Errorf("Empty component encountered in path %v", path)
		}

		isUpdate := updateValue != nil

		if i == lastIndex && isUpdate {
			Log.Debug.Printf("Updating component %v value %+v", component, updateValue)

			var jsonUpdateValue = updateValue
			if updateValue == Nil {
				jsonUpdateValue = nil
			} else if updateValue == Remove {
				delete(ref, component)
				return Remove, nil
			}

			ref[component] = jsonUpdateValue
			return ref[component], nil
		} else {
			child, ok = ref[component]
			//  Error if this child is not found
			if !ok {
				if createPaths && isUpdate {
					Log.Debug.Printf("Creating path for component %v", component)
					newPath := map[string]interface{}{}
					ref[component] = newPath
					ref = newPath
					continue
				} else {
					return nil, fmt.Errorf("Child component %v of path %v not found", component, path)
				}
			}

			if i == lastIndex && !isUpdate {
				//  Return the queried value
				Log.Debug.Printf("Returning query value %+v", child)
				return child, nil
			}

			//  Keep going - child must be a map to enable further traversal
			ref, ok = child.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("Child component %v of path %v is not a map", component, path)
			}
		}
	}

	//  XXX Shouldn't get here
	return nil, fmt.Errorf("Unexpected return from TraversePath %v", path)
}

func WriteJsonResponse(w http.ResponseWriter, obj interface{}) {
	js, _ := json.Marshal(obj)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func WriteHttpError(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintf(w, "%d error : %v", code, error)
}

func WriteJsonStringResponse(w http.ResponseWriter, statusCode int, str string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(str))
}

func WriteBytes(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func PerthLocation() *time.Location {
	awst, err := time.LoadLocation("Australia/Perth")
	if err != nil {
		Log.Error.Fatal("PerthLocation() unable to find Perth timezone in local timezone db")
	}

	return awst
}

func PerthNow() time.Time {
	return time.Now().In(PerthLocation())
}

func ParseProtoTime(timeStr string) (time.Time, error) {
	date, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		return date, err
	}

	date = time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		date.Hour(),
		date.Minute(),
		date.Second(),
		date.Nanosecond(),
		PerthLocation(),
	)

	return date, nil
}

//  Log contents of a reader
func LogReader(logType LogType, reader io.Reader, prefix string) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	logger := *(logLevelMap[logType].Logger)
	logger.Println(prefix, buf.String())
}

func NewHttpClientTimeout(timeout time.Duration) http.Client {
	return NewHttpClient(timeout, false)
}

func NewHttpClient(timeout time.Duration, insecureSkipVerify bool) http.Client {
	transport := http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, timeout)
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
	}

	client := http.Client{
		Transport: &transport,
		Timeout:   timeout,
	}

	return client
}

//  Integer min/max
func Min(x, y int32) int32 {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int32) int32 {
	if x > y {
		return x
	}
	return y
}

func MapFromStringSlice(strs []string) map[string]bool {
	result := make(map[string]bool)
	for _, s := range strs {
		result[s] = true
	}
	return result
}
