package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/NethermindEth/oif-starknet/go/internal"
	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/sirupsen/logrus"
)

// Custom formatter that outputs only the message
type cleanFormatter struct{}

func (f *cleanFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return append([]byte(entry.Message), '\n'), nil
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	logger := setupLogger(cfg)
	logger.Info("🙍 Intent Solver 📝")
	logger.Info("Starting...")

	// Create solver manager
	solverManager := internal.NewSolverManager(cfg, logger)

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize solvers
	if err := solverManager.InitializeSolvers(); err != nil {
		logger.Fatalf("Failed to initialize solvers: %v", err)
	}

	logger.Info("All solvers initialized successfully")

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Received shutdown signal, shutting down...")

	// Shutdown gracefully
	solverManager.Shutdown()
	logger.Info("Solver shutdown complete")
}

// setupLogger configures the logger based on configuration
func setupLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Warnf("Invalid log level %s, using info: %v", cfg.LogLevel, err)
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if cfg.LogFormat == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		// Custom formatter that outputs only the message text
		logger.SetFormatter(&cleanFormatter{})
	}

	return logger
}
