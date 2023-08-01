## Go-Vector-Logger

An easy way to use Loki for logs in Go.

### Usage

```go
package main

import (
  "github.com/scor2k/go-vector-logger"
)

var logger VectorLogger

func main() {
  hostname, _ := os.Hostname()
  logger.Init("test-app", "info", "127.0.0.1", 10100, hostname)

  logger.Debug("test debug message")
  logger.Info("test info message")
  logger.Error("test error message")
}
```
