## Go-Vector-Logger

An easy way to use [Vector](https://vector.dev) for logs in Go.

### Usage

```go
package main

import (
  "github.com/scor2k/go-vector-logger"
)

var logger VectorLogger

func main() {
  logger.Init("test-app", "info", "127.0.0.1", 10100)

  logger.Debug("test debug message")
  logger.Info("test info message")
  logger.Warn("test warning message")
  logger.Error("test error message")
  logger.Errorf("test error message with %s", "formatting")
}
```
