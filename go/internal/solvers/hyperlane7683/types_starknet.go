package hyperlane7683

import (
	"math/big"
)

// TODO: Import proper Starknet types when available
// import (
//     "github.com/NethermindEth/starknet.go/types"
// )

// StarknetEvent represents a Starknet event
type StarknetEvent struct {
	FromAddress string   `json:"from_address"`
	Keys        []string `json:"keys"`
	Data        []string `json:"data"`
}

// StarknetTransaction represents a Starknet transaction
type StarknetTransaction struct {
	Type               string   `json:"type"`
	Version            string   `json:"version"`
	MaxFee             string   `json:"max_fee"`
	Signature          []string `json:"signature"`
	Nonce              string   `json:"nonce"`
	ContractAddress    string   `json:"contract_address"`
	EntryPointSelector string   `json:"entry_point_selector"`
	Calldata           []string `json:"calldata"`
}

// StarknetBlock represents a Starknet block
type StarknetBlock struct {
	BlockNumber   uint64   `json:"block_number"`
	BlockHash     string   `json:"block_hash"`
	ParentHash    string   `json:"parent_hash"`
	Timestamp     uint64   `json:"timestamp"`
	Transactions  []string `json:"transactions"`
	StateRoot     string   `json:"state_root"`
	GasPrice      string   `json:"gas_price"`
	SequencerAddr string   `json:"sequencer_address"`
}

// StarknetCallData represents the calldata for a Starknet contract call
type StarknetCallData struct {
	ContractAddress    string   `json:"contract_address"`
	EntryPointSelector string   `json:"entry_point_selector"`
	Calldata           []string `json:"calldata"`
}

// StarknetFeeEstimate represents gas and fee estimation for Starknet
type StarknetFeeEstimate struct {
	GasConsumed string `json:"gas_consumed"`
	GasPrice    string `json:"gas_price"`
	OverallFee  string `json:"overall_fee"`
}

// StarknetAccount represents a Starknet account
type StarknetAccount struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
	ClassHash string `json:"class_hash"`
}

// StarknetContract represents a Starknet contract
type StarknetContract struct {
	Address   string `json:"address"`
	ClassHash string `json:"class_hash"`
	Bytecode  []byte `json:"bytecode"`
	ABI       string `json:"abi"`
}

// TODO: Add more Starknet-specific types as needed
// - Event filters and queries
// - Transaction receipt types
// - Error response types
// - Network status types

// Helper functions for Starknet types

// StringToBigInt converts a Starknet string number to big.Int
func StringToBigInt(s string) (*big.Int, error) {
	// TODO: Implement proper conversion from Starknet hex string to big.Int
	// Starknet uses hex strings for numbers, need to handle this properly
	return big.NewInt(0), nil
}

// BigIntToString converts a big.Int to Starknet string format
func BigIntToString(b *big.Int) string {
	// TODO: Implement proper conversion from big.Int to Starknet hex string
	return "0x0"
}

// ValidateStarknetAddress validates a Starknet address format
func ValidateStarknetAddress(address string) bool {
	// TODO: Implement Starknet address validation
	// Starknet addresses are 64-character hex strings starting with 0x
	return len(address) == 66 && address[:2] == "0x"
}
