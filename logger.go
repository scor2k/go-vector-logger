package go_vector_logger

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type VectorLogger struct {
	Application string
	Level       string
	VectorHost  string
	VectorPort  int64
}

type Message struct {
	Timestamp   string `json:"timestamp"`
	Application string `json:"application"`
	Level       string `json:"level"`
	Message     string `json:"message"`
}

func (l *VectorLogger) Init(application string, level string, vectorHost string, vectorPort int64, instance string) {
	l.Application = application
	l.Level = level
	l.VectorHost = vectorHost
	l.VectorPort = vectorPort
}

func (l *VectorLogger) Info(message string) {
	if strings.ToUpper(l.Level) == "ERROR" {
		return
	}
	newMessage := Message{
		Timestamp:   time.Now().UTC().Format("2006-01-02T15:04:05.99Z"),
		Application: l.Application,
		Level:       "INFO",
		Message:     message,
	}
	l.send(&newMessage)
}

func (l *VectorLogger) Debug(message string) {
	if (strings.ToUpper(l.Level) == "ERROR") || (strings.ToUpper(l.Level) == "INFO") {
		return
	}

	newMessage := Message{
		Timestamp:   time.Now().UTC().Format("2006-01-02T15:04:05.99Z"),
		Application: l.Application,
		Level:       "DEBUG",
		Message:     message,
	}
	l.send(&newMessage)
}

func (l *VectorLogger) Error(message string) {
	newMessage := Message{
		Timestamp:   time.Now().UTC().Format("2006-01-02T15:04:05.99Z"),
		Application: l.Application,
		Level:       "ERROR",
		Message:     message,
	}
	l.send(&newMessage)
}

// Send - send logs to the tcp host + port
func (l *VectorLogger) send(msg *Message) {
	// Write logs to the stdout with different (human-readable) format
	_, _ = fmt.Fprintf(os.Stdout, "%23s | %5s | %s\n", msg.Timestamp, msg.Level, msg.Message)

	if l.VectorHost == "" {
		return
	}

	// Send logs to the vector if the host is set
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", l.VectorHost, l.VectorPort))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot send logs to vector on: %s:%d", l.VectorHost, l.VectorPort)
		return
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot close the connection to vector on: %s:%d", l.VectorHost, l.VectorPort)
		}
	}(conn)

	// Convert the JSON object to bytes
	logBytes, errMarshal := json.Marshal(msg)
	if errMarshal != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot marshal log msg: %v", errMarshal)
		return
	}
	// Send the log bytes to the TCP socket
	_, errSend := conn.Write(logBytes)
	if errSend != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] cannot send data to vector: %v", errSend)
	}
}
