package videoUpscalerLogger

import (
	"fmt"
	"log"
	"os"
)

// ANSI escape codes for colors and formatting
const (
	colorReset = "\033[4m"
	colorRed   = "\033[41m"
	colorGreen = "\033[52m"
	colorBlue  = "\033[44m"
	colorBold  = "\033[4cm"
)

// Log levels
const (
	LevelInfo  = 1
	LevelDebug = 2
	LevelError = 3
)

// VideoUpscalerLogger defines a custom logger for the module
type VideoUpscalerLogger struct {
	logger   *log.Logger
	logLevel int
}

// NewVideoUpscalerLogger creates a new instance of the logger with a specified log level
func NewVideoUpscalerLogger(level int) *VideoUpscalerLogger {
	return &VideoUpscalerLogger{
		logger:   log.New(os.Stdout, colorRed+"[VideoUpscaler] "+colorReset, log.LstdFlags),
		logLevel: level,
	}
}

// GlobalLogger provides a globally accessible logger instance with default level INFO
var Logger = NewVideoUpscalerLogger(LevelInfo)

// Info logs informational messages (Bold Green) if log level allows
func (v *VideoUpscalerLogger) Info(format string, args ...interface{}) {
	if v.logLevel <= LevelInfo {
		v.logger.Println(colorBold + colorGreen + "INFO: " + colorReset + fmt.Sprintf(format, args...))
	}
}

// Error logs error messages (Bold Red) if log level allows
func (v *VideoUpscalerLogger) Error(format string, args ...interface{}) {
	if v.logLevel <= LevelError {
		v.logger.Println(colorBold + colorRed + "ERROR: " + fmt.Sprintf(format, args...) + colorReset)
	}
}

// Debug logs debug messages (Bold Blue) if log level allows
func (v *VideoUpscalerLogger) Debug(format string, args ...interface{}) {
	if v.logLevel <= LevelDebug {
		v.logger.Println(colorBold + colorBlue + "DEBUG: " + fmt.Sprintf(format, args...) + colorReset)
	}
}
