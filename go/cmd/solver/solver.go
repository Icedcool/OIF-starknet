package solver

// Solver package - contains the main solver logic
// This allows the solver to be imported and run from the main CLI

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/NethermindEth/oif-starknet/go/solvercore"
	"github.com/NethermindEth/oif-starknet/go/solvercore/config"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

// Custom formatter that outputs only the message
type cleanFormatter struct{}

func (f *cleanFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return append([]byte(entry.Message), '\n'), nil
}

// RunSolver runs the main solver application
func RunSolver() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize networks from centralized config after .env is loaded
	config.InitializeNetworks()

	// Set up clean logging
	logrus.SetFormatter(&cleanFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logrus.Info("🔄 Shutdown signal received, stopping solver...")
		cancel()
	}()

	// Initialize solver manager
	solverManager := solvercore.NewSolverManager(cfg)

	// Start the solver
	logrus.Info("🚀 Starting OIF Starknet Solver...")
	logrus.Info("   📊 Monitoring networks:", strings.Join(config.GetNetworkNames(), ", "))
	logrus.Info("   ⏰ Poll interval: 1000ms (default)")
	logrus.Info("   🛑 Press Ctrl+C to stop")

	if err := solverManager.Start(ctx); err != nil {
		logrus.Fatalf("Solver failed: %v", err)
	}

	logrus.Info("✅ Solver stopped gracefully")
}

// TestConnection tests the connection to all configured networks
func TestConnection() {
	_, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	config.InitializeNetworks()

	logrus.Info("🔍 Testing network connections...")

	for _, networkName := range config.GetNetworkNames() {
		networkConfig, exists := config.Networks[networkName]
		if !exists {
			logrus.Warnf("   ⚠️  Network %s not found in config", networkName)
			continue
		}

		logrus.Infof("   🔗 Testing %s (%s)...", networkName, networkConfig.RPCURL)

		// Test connection (works for both EVM and Starknet)
		client, err := ethclient.Dial(networkConfig.RPCURL)
		if err != nil {
			logrus.Errorf("   ❌ Failed to connect to %s: %v", networkName, err)
			continue
		}

		chainID, err := client.ChainID(context.Background())
		if err != nil {
			logrus.Errorf("   ❌ Failed to get chain ID for %s: %v", networkName, err)
			client.Close() // Close immediately on error
			continue
		}

		logrus.Infof("   ✅ %s connected (Chain ID: %s)", networkName, chainID.String())
		client.Close() // Close immediately after successful test
	}

	logrus.Info("🎉 Network connection test completed")
}
