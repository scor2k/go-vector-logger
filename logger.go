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
	Application string   // Application name.
	Level       string   // Log level.
	VectorHost  string   // Vector host.
	VectorPort  int64    // Vector port.
	Options     Options  // Options for the logger
	conn        net.Conn // Persistent TCP connection
	lastActivityTime time.Time // Timestamp of the last communication.
	TimeoutDuration time.Duration // Duration after which an inactive connection should be considered timed out.
	mu          sync.Mutex // For ensuring thread-safe access to conn and lastActivityTime.
	stopChan    chan struct{} // Channel to signal the connection management goroutine to stop.
	wg          sync.WaitGroup // For waiting for the connection management goroutine to exit.
}

// establishConnection creates a TCP connection to the Vector instance.
func establishConnection(host string, port int64) (net.Conn, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		return nil, fmt.Errorf("cannot establish connection to the TCP endpoint on: %s:%d: %v", host, port, err)
	}
	return conn, nil
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

	logger := &VectorLogger{
		Application: application,
		Level:       strings.ToUpper(level),
		VectorHost:  vectorHost,
		VectorPort:  vectorPort,
		Options:     opts,
	}

	logger.TimeoutDuration = 1 * time.Minute

	// Establish persistent TCP connection if needed
	if opts.Writer == nil && vectorHost != "" {
		conn, err := establishConnection(vectorHost, vectorPort)
		if err != nil {
			return nil, err
		}
		logger.conn = conn
		logger.lastActivityTime = time.Now()
	}

	logger.stopChan = make(chan struct{})

	if opts.Writer == nil && logger.VectorHost != "" && logger.conn != nil {
		l.wg.Add(1)
		go logger.manageConnection()
	}

	return logger, nil
}

// manageConnection is a background goroutine that proactively closes idle connections.
func (l *VectorLogger) manageConnection() {
	defer l.wg.Done()
	// Set ticker to a fraction of the timeoutDuration, e.g., timeoutDuration / 2, but not less than a minimum (e.g., 5s)
	// For this implementation, we'll use a fixed 10 seconds as specified.
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	fmt.Printf("Starting connection manager for %s:%d\n", l.VectorHost, l.VectorPort) // For debugging

	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			if l.conn != nil && time.Since(l.lastActivityTime) > l.TimeoutDuration {
				fmt.Printf("Proactively closing idle Vector connection to %s:%d\n", l.VectorHost, l.VectorPort)
				l.conn.Close()
				l.conn = nil
			}
			l.mu.Unlock()
		case <-l.stopChan:
			fmt.Printf("Stopping connection manager for %s:%d\n", l.VectorHost, l.VectorPort) // For debugging
			return
		}
	}
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

// Fatal logs an error message.
func (l *VectorLogger) FatalError(message error) {
	l.sendMessage(message.Error(), FATAL)
	os.Exit(1)
}

// send sends the log message to stdout and to a remote Vector instance.
func (l *VectorLogger) send(msg *Message) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Write logs to the stdout with different (human-readable) format
	if l.Options.AlsoPrintMessages {
		_, _ = fmt.Fprintf(os.Stdout, "%23s | %5s | %s\n", msg.Timestamp, msg.Level, msg.Message)
	}

	var dest io.Writer = l.Options.Writer

	if dest == nil && l.VectorHost != "" {
		// Network connection logic
		if l.conn == nil {
			// Try to establish connection if it doesn't exist
			conn, err := establishConnection(l.VectorHost, l.VectorPort)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[ERROR] initial connection failed: %v\n", err)
				return
			}
			l.conn = conn
			l.lastActivityTime = time.Now()
		} else if time.Since(l.lastActivityTime) > l.TimeoutDuration {
			// Connection timed out, try to re-establish
			if l.conn != nil {
				l.conn.Close()
				l.conn = nil // Explicitly set to nil after closing
			}
			conn, err := establishConnection(l.VectorHost, l.VectorPort)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[ERROR] timed out connection re-establishment failed: %v\n", err)
				l.conn = nil // Ensure conn is nil if re-establishment fails
				return
			}
			l.conn = conn
			l.lastActivityTime = time.Now()
		}

		if l.conn == nil { // If connection is still nil after attempts, return
			_, _ = fmt.Fprintf(os.Stderr, "[ERROR] no valid network connection available\n")
			return
		}
		dest = l.conn
	} else if dest == nil && l.VectorHost == "" {
		// No writer and no vector host configured, nothing to do.
		return
	}


	// Convert the JSON object to bytes
	buf := new(bytes.Buffer)
	if errMarshal := json.NewEncoder(buf).Encode(msg); errMarshal != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot marshal log msg: %v\n", errMarshal)
		return
	}

	// Send the log bytes
	if _, errSend := buf.WriteTo(dest); errSend != nil {
		// Send failed
		if l.Options.Writer == nil && l.VectorHost != "" { // Check if it was a network send
			// Network send failed, attempt to reconnect and send again
			if l.conn != nil {
				l.conn.Close()
				l.conn = nil
			}

			conn, err := establishConnection(l.VectorHost, l.VectorPort)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[ERROR] re-connection after send failure failed: %v\n", err)
				l.conn = nil // Ensure conn is nil
				return
			}
			l.conn = conn
			l.lastActivityTime = time.Now()
			dest = l.conn // Update dest to the new connection

			// Retry sending
			// Re-encode to a new buffer, as the previous buffer might have been partially written or its state is uncertain.
			retryBuf := new(bytes.Buffer)
			if errMarshalRetry := json.NewEncoder(retryBuf).Encode(msg); errMarshalRetry != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot marshal log msg for retry: %v\n", errMarshalRetry)
				return
			}
			if _, errSendAgain := retryBuf.WriteTo(dest); errSendAgain != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot send data to the TCP endpoint after re-connection: %v\n", errSendAgain)
				// Even if this send fails, we keep the new connection for future attempts.
				// But we should probably close it and set to nil if this fails too, to force re-establishment next time.
				if l.conn != nil {
					l.conn.Close()
					l.conn = nil
				}
				return
			}
			// If second send is successful, update lastActivityTime
			l.lastActivityTime = time.Now()
		} else {
			// Send failed on a non-network writer (e.g. custom io.Writer)
			_, _ = fmt.Fprintf(os.Stderr, "[ERROR] failed to write to custom writer: %v\n", errSend)
		}
		return // Return after handling send error
	}

	// If send was successful and it was a network send, update lastActivityTime
	if dest == l.conn && l.conn != nil { // Check l.conn != nil for safety, though dest == l.conn implies it
		l.lastActivityTime = time.Now()
	}
}

func (l *VectorLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Signal the connection management goroutine to stop
	if l.stopChan != nil {
		close(l.stopChan)
		l.stopChan = nil
	}

	// Wait for the connection management goroutine to finish
	l.wg.Wait()

	l.Options.Writer = nil
	if l.conn != nil {
		err := l.conn.Close()
		l.conn = nil // Set conn to nil after closing
		return err
	}
	return nil
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
