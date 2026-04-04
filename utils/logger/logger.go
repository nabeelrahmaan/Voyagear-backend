package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Log structurally acts as the application's global mechanism
var Log *logrus.Logger

// InitLogger maps output dynamically to a localized file while enforcing explicit terminal behaviors
func InitLogger() *logrus.Logger {
	Log = logrus.New()

	// TextFormatter keeps it readable for you, while cleanly enforcing timestamp blocks
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true, // Force vibrant output in powershell
	})

	// Ensure the physical `logs` environment exists natively before hooking
	if err := os.MkdirAll("logs", 0755); err != nil {
		logrus.Fatalf("Failed to construct explicit logs configuration directory: %v", err)
	}

	// Open `server.log` explicitly appending into it (or generating if missing)
	logFile, err := os.OpenFile("logs/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("Failed to initialize server.log binding: %v", err)
	}

	// Route structural logs symmetrically across the Native console AND the physical text file
	Log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	// Track explicit INFO, WARNING, and ERROR configurations broadly
	Log.SetLevel(logrus.InfoLevel)

	Log.Info("Server Logging Architecture System Actively Initialized")

	return Log
}
