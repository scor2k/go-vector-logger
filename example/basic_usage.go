package main

import (
	"fmt"
	go_vector_logger "go-vector-logger"
	"time"
)

// update go.mod to run the tests
// module go-vector-logger

func main() {

	log, err := go_vector_logger.New("test-app", "INFO", "127.0.0.1", 6000, go_vector_logger.Options{
		AlsoPrintMessages: false,
	})
	if err != nil {
		panic(err)
	}

	for i := range 1000 {
		log.Infof("Iteration %d", i)
		time.Sleep(2 * time.Millisecond)
		if i > 0 && i%100 == 0 {
			fmt.Println("Sleep for 5 seconds to test idle functionality")
			time.Sleep(5 * time.Second)
			fmt.Println("Next")
		}
	}

	fmt.Printf("10k log messages sent")

	log.Debug("test debug message")
	log.Info("test info message")
	log.Warn("test warning message")
	_ = log.Close() // test how re-connect is work
	log.Error("test error message")
	log.Errorf("test error message with %s", "formatting")
	log.Fatalf("test fatal message with %s", "formatting")
	log.Fatal("test fatal message")
	log.FatalError(err)
}
