package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	level      Level
	format     string
	output     io.Writer
	timeFormat string
}

type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
}

var DefaultLogger *Logger

func Init(level, format, output, timeFormat string) {
	var logLevel Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		logLevel = DEBUG
	case "INFO":
		logLevel = INFO
	case "WARN", "WARNING":
		logLevel = WARN
	case "ERROR":
		logLevel = ERROR
	case "FATAL":
		logLevel = FATAL
	default:
		logLevel = INFO
	}

	var writer io.Writer
	switch strings.ToLower(output) {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		if file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			writer = file
		} else {
			writer = os.Stdout
		}
	}

	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	DefaultLogger = &Logger{
		level:      logLevel,
		format:     format,
		output:     writer,
		timeFormat: timeFormat,
	}
}

func (l *Logger) log(level Level, message string, data map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().Format(l.timeFormat),
		Level:     level.String(),
		Message:   message,
		Data:      data,
	}

	// Add file and line info for ERROR and FATAL levels
	if level >= ERROR {
		_, file, line, ok := runtime.Caller(3)
		if ok {
			entry.File = file
			entry.Line = line
		}
	}

	var output string
	if l.format == "json" {
		jsonBytes, _ := json.Marshal(entry)
		output = string(jsonBytes)
	} else {
		output = l.formatText(entry)
	}

	fmt.Fprintln(l.output, output)

	if level == FATAL {
		os.Exit(1)
	}
}

func (l *Logger) formatText(entry LogEntry) string {
	var parts []string
	parts = append(parts, entry.Timestamp)
	parts = append(parts, fmt.Sprintf("[%s]", entry.Level))
	parts = append(parts, entry.Message)

	if len(entry.Data) > 0 {
		dataStr, _ := json.Marshal(entry.Data)
		parts = append(parts, string(dataStr))
	}

	if entry.File != "" {
		parts = append(parts, fmt.Sprintf("(%s:%d)", entry.File, entry.Line))
	}

	return strings.Join(parts, " ")
}

// Public logging functions
func Debug(message string, data ...map[string]interface{}) {
	var d map[string]interface{}
	if len(data) > 0 {
		d = data[0]
	}
	DefaultLogger.log(DEBUG, message, d)
}

func Info(message string, data ...map[string]interface{}) {
	var d map[string]interface{}
	if len(data) > 0 {
		d = data[0]
	}
	DefaultLogger.log(INFO, message, d)
}

func Warn(message string, data ...map[string]interface{}) {
	var d map[string]interface{}
	if len(data) > 0 {
		d = data[0]
	}
	DefaultLogger.log(WARN, message, d)
}

func Error(message string, data ...map[string]interface{}) {
	var d map[string]interface{}
	if len(data) > 0 {
		d = data[0]
	}
	DefaultLogger.log(ERROR, message, d)
}

func Fatal(message string, data ...map[string]interface{}) {
	var d map[string]interface{}
	if len(data) > 0 {
		d = data[0]
	}
	DefaultLogger.log(FATAL, message, d)
}

// Structured logging with context
func WithField(key string, value interface{}) map[string]interface{} {
	return map[string]interface{}{key: value}
}

func WithFields(fields map[string]interface{}) map[string]interface{} {
	return fields
}

// Standard library compatibility
func Println(v ...interface{}) {
	Info(fmt.Sprint(v...))
}

func Printf(format string, v ...interface{}) {
	Info(fmt.Sprintf(format, v...))
}

// Setup for compatibility with standard log package
func SetupStandardLogger() {
	log.SetOutput(&logWriter{})
	log.SetFlags(0)
}

type logWriter struct{}

func (w *logWriter) Write(p []byte) (n int, err error) {
	Info(strings.TrimSpace(string(p)))
	return len(p), nil
}
