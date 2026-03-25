package flogger

import (
	"log"
	"log/slog"
	"os"
)

func InitLogger(logFile string) *slog.Logger {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Create a handler for writing logs to the file
	fileHandler := slog.NewTextHandler(file, nil)
	logger := slog.New(fileHandler)

	// Set this logger as the default
	slog.SetDefault(logger)

	return logger
}
