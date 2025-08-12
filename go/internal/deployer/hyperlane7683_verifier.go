package deployer

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Hyperlane7683Verifier handles Hyperlane7683 contract verification
// Note: We're using pre-deployed contracts like the TypeScript implementation
type Hyperlane7683Verifier struct {
	client *ethclient.Client
	auth   *bind.TransactOpts
}

// NewHyperlane7683Verifier creates a new Hyperlane7683Verifier
func NewHyperlane7683Verifier(client *ethclient.Client, auth *bind.TransactOpts) *Hyperlane7683Verifier {
	return &Hyperlane7683Verifier{
		client: client,
		auth:   auth,
	}
}

// GetPreDeployedAddress returns the pre-deployed Hyperlane7683 address
// This matches the TypeScript implementation which uses the same address across all networks
func (v *Hyperlane7683Verifier) GetPreDeployedAddress() common.Address {
	// Same address used by TypeScript implementation on all testnets
	return common.HexToAddress("0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")
}

// VerifyContractExists checks if the pre-deployed contract exists and is accessible
func (v *Hyperlane7683Verifier) VerifyContractExists() error {
	address := v.GetPreDeployedAddress()

	// Check if the contract exists by getting its code
	code, err := v.client.CodeAt(v.auth.Context, address, nil)
	if err != nil {
		return fmt.Errorf("failed to get contract code: %w", err)
	}

	if len(code) == 0 {
		return fmt.Errorf("no contract found at address %s", address.Hex())
	}

	log.Printf("✅ Verified pre-deployed Hyperlane7683 contract at %s", address.Hex())
	return nil
}
