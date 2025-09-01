package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

// Init Initialize the logger with desired settings.
func Init() {
	// Example: JSON logs with Info level
	handler := slog.NewTextHandler(os.Stderr, nil)
	Log = slog.New(handler)

	Log.Info("Logger initialized")
}
