## Go-Vector-Logger

An easy way to use [Vector](https://vector.dev) for logs in Go.

### Usage

```go
package main

import (
  "github.com/scor2k/go-vector-logger"
)

var log VectorLogger

func main() {
  log := go_vector_logger.New("test-app", "INFO", "127.0.0.1", 10100)

  log.Debug("test debug message")
  log.Info("test info message")
  log.Warn("test warning message")
  log.Error("test error message")
  log.Errorf("test error message with %s", "formatting")
  log.Fatal("test error message")
  log.Fatalf("test error message with %s", "formatting")
  log.FatalError(err)
}
```

Also, you can pass in aditional options to `New()`:

```
type Options struct {
  Writer            io.Writer // Instead of over the network, write the log messages just to this writer
  AlsoPrintMessages bool      // In addition to the specific network/writer, also log any messages to stdout
}

[...]
log := go_vector_logger.New(
  "test-app",
  "INFO",
  "127.0.0.1",
  10100,
  Options{
    AlsoPrintMessages: true,
  },
)
