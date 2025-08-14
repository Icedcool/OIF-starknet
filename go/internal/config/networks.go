package config

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// NetworkConfig represents a single network configuration
type NetworkConfig struct {
	Name             string
	RPCURL           string
	ChainID          uint64
	HyperlaneAddress common.Address
	HyperlaneDomain  uint32
	ForkStartBlock   uint64
}

// Networks contains all network configurations
var Networks = map[string]NetworkConfig{
	"Sepolia": {
		Name:             "Sepolia",
		RPCURL:           "http://localhost:8545",
		ChainID:          11155111,
		HyperlaneAddress: common.HexToAddress("0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3"),
		HyperlaneDomain:  11156380, // 0xaa37dc - matches actual contract (from fill event)
		ForkStartBlock:   8319000,
	},
	"Optimism Sepolia": {
		Name:             "Optimism Sepolia",
		RPCURL:           "http://localhost:8546",
		ChainID:          11155420,
		HyperlaneAddress: common.HexToAddress("0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3"),
		HyperlaneDomain:  11155420, // 0xaa36a7 - matches actual contract (from fill event)
		ForkStartBlock:   27370000,
	},
	"Arbitrum Sepolia": {
		Name:             "Arbitrum Sepolia",
		RPCURL:           "http://localhost:8547",
		ChainID:          421614,
		HyperlaneAddress: common.HexToAddress("0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3"),
		HyperlaneDomain:  421614,
		ForkStartBlock:   138020000,
	},
	"Base Sepolia": {
		Name:             "Base Sepolia",
		RPCURL:           "http://localhost:8548",
		ChainID:          84532,
		HyperlaneAddress: common.HexToAddress("0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3"),
		HyperlaneDomain:  84532, // 0x14a34 - matches actual contract (from router enrollment)
		ForkStartBlock:   25380000,
	},
}

// GetNetworkConfig returns the configuration for a given network name
func GetNetworkConfig(networkName string) (NetworkConfig, error) {
	if config, exists := Networks[networkName]; exists {
		return config, nil
	}
	return NetworkConfig{}, fmt.Errorf("network not found: %s", networkName)
}

// GetRPCURL returns the RPC URL for a given network name
func GetRPCURL(networkName string) (string, error) {
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return "", err
	}
	return config.RPCURL, nil
}

// GetChainID returns the chain ID for a given network name
func GetChainID(networkName string) (uint64, error) {
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return 0, err
	}
	return config.ChainID, nil
}

// GetHyperlaneAddress returns the Hyperlane contract address for a given network name
func GetHyperlaneAddress(networkName string) (common.Address, error) {
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return common.Address{}, err
	}
	return config.HyperlaneAddress, nil
}

// GetHyperlaneDomain returns the Hyperlane domain ID for a given network name
func GetHyperlaneDomain(networkName string) (uint32, error) {
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return 0, err
	}
	return config.HyperlaneDomain, nil
}

// GetForkStartBlock returns the fork start block for a given network name
func GetForkStartBlock(networkName string) (uint64, error) {
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return 0, err
	}
	return config.ForkStartBlock, nil
}

// GetNetworkNames returns all available network names
func GetNetworkNames() []string {
	names := make([]string, 0, len(Networks))
	for name := range Networks {
		names = append(names, name)
	}
	return names
}

// ValidateNetworkName checks if a network name is valid
func ValidateNetworkName(networkName string) bool {
	_, exists := Networks[networkName]
	return exists
}

// GetDefaultNetwork returns the default network (Base Sepolia)
func GetDefaultNetwork() NetworkConfig {
	return Networks["Base Sepolia"]
}

// GetDefaultRPCURL returns the default RPC URL
func GetDefaultRPCURL() string {
	return Networks["Base Sepolia"].RPCURL
}
