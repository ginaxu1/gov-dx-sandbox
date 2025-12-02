package logger

import (
	"log/slog"
	"os"
)

// Log is the global logger instance.
var Log *slog.Logger

// Init Initializes the logger with the desired settings.
func Init() {
	// Example: JSON logs with Info level
	handler := slog.NewTextHandler(os.Stderr, nil)
	Log = slog.New(handler)

	Log.Info("Logger initialized")
}
