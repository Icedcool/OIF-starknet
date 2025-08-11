package deployer

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ForkVerificationConfig holds configuration for forked network verification
type ForkVerificationConfig struct {
	RPCURL      string
	PrivateKey  string
	ChainName   string
	ChainID     int64
}

// ForkVerificationManager manages verification of pre-deployed contracts on forked networks
type ForkVerificationManager struct {
	configs []ForkVerificationConfig
}

// NewForkVerificationManager creates a new ForkVerificationManager
func NewForkVerificationManager(configs []ForkVerificationConfig) *ForkVerificationManager {
	return &ForkVerificationManager{
		configs: configs,
	}
}

// VerifyPreDeployedContracts verifies that pre-deployed contracts exist on all forked networks
func (m *ForkVerificationManager) VerifyPreDeployedContracts() error {
	log.Printf("🔍 Verifying pre-deployed contracts on all forked networks...")

	for _, config := range m.configs {
		log.Printf("📡 Checking %s (Chain ID: %d)...", config.ChainName, config.ChainID)

		// Connect to the forked network
		client, err := ethclient.Dial(config.RPCURL)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", config.ChainName, err)
		}
		defer client.Close()

		// Parse private key (remove 0x prefix if present)
		privateKeyHex := config.PrivateKey
		if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
			privateKeyHex = privateKeyHex[2:]
		}
		
		privateKey, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return fmt.Errorf("failed to parse private key for %s: %w", config.ChainName, err)
		}

		// Get chain ID
		chainID, err := client.ChainID(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get chain ID for %s: %w", config.ChainName, err)
		}

		// Create auth
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			return fmt.Errorf("failed to create auth for %s: %w", config.ChainName, err)
		}

		// Create verifier and verify contract
		verifier := NewHyperlane7683Verifier(client, auth)
		if err := verifier.VerifyContractExists(); err != nil {
			return fmt.Errorf("failed to verify contract on %s: %w", config.ChainName, err)
		}

		log.Printf("✅ %s: Contract verified successfully", config.ChainName)
	}

	log.Printf("🎉 All pre-deployed contracts verified successfully!")
	return nil
}

// GetContractAddresses returns the pre-deployed contract addresses for all networks
func (m *ForkVerificationManager) GetContractAddresses() map[string]common.Address {
	addresses := make(map[string]common.Address)
	
	// All networks use the same pre-deployed address
	contractAddress := common.HexToAddress("0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")
	
	for _, config := range m.configs {
		addresses[config.ChainName] = contractAddress
	}
	
	return addresses
}
