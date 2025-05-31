# Changelog

## [0.8.0] - 2025-05-25

### Added
- TCP connection idle timeout management with automatic connection cleanup after 1 minute of inactivity
- Proactive connection management goroutine that periodically checks for idle connections
- Thread-safe connection handling with mutex protection
- `Close()` method for graceful shutdown of logger and background processes
- Example application (`example/basic_usage.go`) for testing and demonstration
- Local testing documentation and Vector configuration files

### Changed
- Refactored connection management from channel-based to direct synchronous approach
- Improved error handling and logging for connection failures
- Removed deprecated `TCPTimeout` option from `Options` struct
- Enhanced reconnection logic with better error recovery

### Fixed
- Module name corrected to `github.com/scor2k/go-vector-logger`
- Persistent connection handling rewritten for better reliability
- Connection timeout and reconnection edge cases

## [0.7.0] - 2025-05-21

### Added
- Persistent TCP connection with idle timeout and TCP connection timeout support.
- Graceful shutdown and draining of log messages.
