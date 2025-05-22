## Go-Vector-Logger

An easy way to use [Vector](https://vector.dev) for logs in Go.

See [Changelog.md](./Changelog.md) for release history.

### Usage

```go
package main

import (
  "github.com/scor2k/go-vector-logger"
)

func main() {
  log, err := go_vector_logger.New("test-app", "INFO", "127.0.0.1", 10100)
  if err != nil {
    panic(err)
  }

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

#### Options

You can pass additional options to `New()`:

```go
type Options struct {
  Writer            io.Writer     // Instead of over the network, write the log messages just to this writer
  AlsoPrintMessages bool          // In addition to the specific network/writer, also log any messages to stdout
  TCPTimeout        time.Duration // Timeout for TCP connection and write. If zero, defaults to 1 second.
}
```

Example with options:

```go
log, err := go_vector_logger.New(
  "test-app",
  "INFO",
  "127.0.0.1",
  10100,
  go_vector_logger.Options{
    AlsoPrintMessages: true,
  },
)
if err != nil {
  panic(err)
}
```
