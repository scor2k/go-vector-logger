package main

import (
	"fmt"
	go_vector_logger "go-vector-logger"
)

func main() {

	log, err := go_vector_logger.New("test-app", "INFO", "127.0.0.1", 6000, go_vector_logger.Options{
		AlsoPrintMessages: true,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello, World!\n")

	log.Debug("test debug message")
	log.Info("test info message")
	log.Warn("test warning message")
	log.Error("test error message")
	log.Errorf("test error message with %s", "formatting")
	log.Fatalf("test fatal message with %s", "formatting")
	log.Fatal("test fatal message")
	log.FatalError(err)
	_ = log.Close() // won't be called because of the log.Fatal()
}
