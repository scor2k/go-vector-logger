// Package go_vector_logger provides a logger that can write logs to stdout and send them to a remote Vector instance.
package go_vector_logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const (
	DEBUG string = "DEBUG"
	INFO         = "INFO"
	WARN         = "WARN"
	ERROR        = "ERROR"
	FATAL        = "FATAL"
)

// Options list different options you can optionally pass into New
type Options struct {
	Writer            io.Writer // Instead of over the network, write the log messages just to this writer
	AlsoPrintMessages bool      // In addition to the specific network, also log any messages to stdout
}

// VectorLogger represents a logger instance.
type VectorLogger struct {
	Application string // Application name.
	Level       string // Log level.
	VectorHost  string // Vector host.
	VectorPort  int64  // Vector port.
	Options     Options
}

func New(application string, level string, vectorHost string, vectorPort int64, options ...Options) (*VectorLogger, error) {
	var opts Options
	switch len(options) {
	case 0:
	case 1:
		opts = options[0]
	default:
		return nil, fmt.Errorf("Can only pass in one Options struct")
	}

	return &VectorLogger{
		Application: application,
		Level:       strings.ToUpper(level),
		VectorHost:  vectorHost,
		VectorPort:  vectorPort,
		Options:     opts,
	}, nil
}

// Message represents a log message.
type Message struct {
	Timestamp   string `json:"timestamp"`   // Log timestamp.
	Application string `json:"application"` // Application name.
	Level       string `json:"level"`       // Log level.
	Message     string `json:"message"`     // Log message.
}

// Init initializes the logger instance. This method is deprecated; use
// New() with a Options struct for more flexibility.
func (l *VectorLogger) Init(application string, level string, vectorHost string, vectorPort int64) {
	l.Application = application
	l.Level = strings.ToUpper(level)
	l.VectorHost = vectorHost
	l.VectorPort = vectorPort
	l.Options.AlsoPrintMessages = true
}

// Debugf logs a debug message with a formatted string.
func (l *VectorLogger) Debugf(format string, v ...interface{}) {
	if l.Level != DEBUG {
		return
	}
	l.sendMessage(fmt.Sprintf(format, v...), DEBUG)
}

// Debug logs a debug message.
func (l *VectorLogger) Debug(message string) {
	if l.Level != DEBUG {
		return
	}
	l.sendMessage(message, DEBUG)
}

// Infof logs an info message with a formatted string.
func (l *VectorLogger) Infof(format string, v ...interface{}) {
	if (l.Level == ERROR) || (l.Level == WARN) {
		return
	}
	l.sendMessage(fmt.Sprintf(format, v...), "INFO")
}

// Info logs an info message.
func (l *VectorLogger) Info(message string) {
	if (l.Level == ERROR) || (l.Level == WARN) {
		return
	}
	l.sendMessage(message, "INFO")
}

// Warnf logs an warning message with a formatted string.
func (l *VectorLogger) Warnf(format string, v ...interface{}) {
	if l.Level == ERROR {
		return
	}
	l.sendMessage(fmt.Sprintf(format, v...), WARN)
}

// Warn logs an warning message.
func (l *VectorLogger) Warn(message string) {
	if l.Level == ERROR {
		return
	}
	l.sendMessage(message, WARN)
}

// Errorf logs an error message with a formatted string.
func (l *VectorLogger) Errorf(format string, v ...interface{}) {
	l.sendMessage(fmt.Sprintf(format, v...), ERROR)
}

// Error logs an error message.
func (l *VectorLogger) Error(message string) {
	l.sendMessage(message, ERROR)
}

// Errorf logs an error message with a formatted string.
func (l *VectorLogger) Fatalf(format string, v ...interface{}) {
	l.sendMessage(fmt.Sprintf(format, v...), FATAL)
	os.Exit(1)
}

// Fatal logs an error message.
func (l *VectorLogger) Fatal(message string) {
	l.sendMessage(message, FATAL)
	os.Exit(1)
}

// send sends the log message to stdout and to a remote Vector instance.
func (l *VectorLogger) send(msg *Message) {
	// Write logs to the stdout with different (human-readable) format
	if l.Options.AlsoPrintMessages {
		_, _ = fmt.Fprintf(os.Stdout, "%23s | %5s | %s\n", msg.Timestamp, msg.Level, msg.Message)
	}

	dest := l.Options.Writer
	if dest == nil {
		// Setup network connection if the host is set
		if l.VectorHost == "" {
			return
		}

		// Send logs to the vector if the host is set
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", l.VectorHost, l.VectorPort))
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot send logs to vector on: %s:%d: %v\n", l.VectorHost, l.VectorPort, err)
			return
		}
		defer func(conn net.Conn) {
			err := conn.Close()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot close the connection to vector on: %s:%d: %v\n", l.VectorHost, l.VectorPort, err)
			}
		}(conn)
		dest = conn
	}

	// Convert the JSON object to bytes
	buf := new(bytes.Buffer)
	if errMarshal := json.NewEncoder(buf).Encode(msg); errMarshal != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot marshal log msg: %v\n", errMarshal)
		return
	}

	// Send the log bytes to the TCP socket
	if _, errSend := buf.WriteTo(dest); errSend != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot send data to vector: %v\n", errSend)
	}
}

// wrapper for sending a log message
func (l *VectorLogger) sendMessage(message string, level string) {
	newMessage := Message{
		Timestamp:   time.Now().UTC().Format("2006-01-02T15:04:05.00Z"),
		Application: l.Application,
		Level:       level,
		Message:     message,
	}
	l.send(&newMessage)
}
