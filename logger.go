// Package go_vector_logger provides a logger that can write logs to stdout and send them to a remote Vector instance.
package go_vector_logger

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// VectorLogger represents a logger instance.
type VectorLogger struct {
	Application string // Application name.
	Level       string // Log level.
	VectorHost  string // Vector host.
	VectorPort  int64  // Vector port.
}

// Message represents a log message.
type Message struct {
	Timestamp   string `json:"timestamp"`   // Log timestamp.
	Application string `json:"application"` // Application name.
	Level       string `json:"level"`       // Log level.
	Message     string `json:"message"`     // Log message.
}

// Init initializes the logger instance.
func (l *VectorLogger) Init(application string, level string, vectorHost string, vectorPort int64) {
	l.Application = application
	l.Level = level
	l.VectorHost = vectorHost
	l.VectorPort = vectorPort
}

// Debugf logs a debug message with a formatted string.
func (l *VectorLogger) Debugf(format string, v ...interface{}) {
	if strings.ToUpper(l.Level) != "DEBUG" {
		return
	}
	l.sendMessage(fmt.Sprintf(format, v...), "DEBUG")
}

// Debug logs a debug message.
func (l *VectorLogger) Debug(message string) {
	if strings.ToUpper(l.Level) != "DEBUG" {
		return
	}
	l.sendMessage(message, "DEBUG")
}

// Infof logs an info message with a formatted string.
func (l *VectorLogger) Infof(format string, v ...interface{}) {
	if (strings.ToUpper(l.Level) == "ERROR") || (strings.ToUpper(l.Level) == "WARN") {
		return
	}
	l.sendMessage(fmt.Sprintf(format, v...), "INFO")
}

// Info logs an info message.
func (l *VectorLogger) Info(message string) {
	if (strings.ToUpper(l.Level) == "ERROR") || (strings.ToUpper(l.Level) == "WARN") {
		return
	}
	l.sendMessage(message, "INFO")
}

// Warnf logs an warning message with a formatted string.
func (l *VectorLogger) Warnf(format string, v ...interface{}) {
	if strings.ToUpper(l.Level) == "ERROR" {
		return
	}
	l.sendMessage(fmt.Sprintf(format, v...), "WARN")
}

// Warn logs an warning message.
func (l *VectorLogger) Warn(message string) {
	if strings.ToUpper(l.Level) == "ERROR" {
		return
	}
	l.sendMessage(message, "WARN")
}

// Errorf logs an error message with a formatted string.
func (l *VectorLogger) Errorf(format string, v ...interface{}) {
	l.sendMessage(fmt.Sprintf(format, v...), "ERROR")
}

// Error logs an error message.
func (l *VectorLogger) Error(message string) {
	l.sendMessage(message, "ERROR")
}

// send sends the log message to stdout and to a remote Vector instance.
func (l *VectorLogger) send(msg *Message) {
	// Write logs to the stdout with different (human-readable) format
	_, _ = fmt.Fprintf(os.Stdout, "%23s | %5s | %s\n", msg.Timestamp, msg.Level, msg.Message)

	if l.VectorHost == "" {
		return
	}

	// Send logs to the vector if the host is set
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", l.VectorHost, l.VectorPort))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot send logs to vector on: %s:%d\n", l.VectorHost, l.VectorPort)
		return
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot close the connection to vector on: %s:%d\n", l.VectorHost, l.VectorPort)
		}
	}(conn)

	// Convert the JSON object to bytes
	logBytes, errMarshal := json.Marshal(msg)
	if errMarshal != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot marshal log msg: %v\n", errMarshal)
		return
	}
	// Send the log bytes to the TCP socket
	_, errSend := conn.Write(logBytes)
	if errSend != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot send data to vector: %v\n", errSend)
	}
}

// wrapper for sending a log message
func (l *VectorLogger) sendMessage(message string, level string) {
	newMessage := Message{
		Timestamp:   time.Now().UTC().Format("2006-01-02T15:04:05.99Z"),
		Application: l.Application,
		Level:       strings.ToUpper(level),
		Message:     message,
	}
	l.send(&newMessage)
}
