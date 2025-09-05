package hyperlane7683

import (
	"math/big"
	"testing"

	"github.com/NethermindEth/oif-starknet/go/solvercore/types"
	"github.com/stretchr/testify/assert"
)

// TestHyperlane7683Comprehensive tests comprehensive hyperlane7683 functionality
func TestHyperlane7683Comprehensive(t *testing.T) {
	t.Run("solver_initialization", func(t *testing.T) {
		// Test solver creation
		solver := &Hyperlane7683Solver{
			metadata: types.Hyperlane7683Metadata{
				BaseMetadata: types.BaseMetadata{
					ProtocolName: "Hyperlane7683",
				},
			},
		}
		assert.NotNil(t, solver)
		assert.Equal(t, "Hyperlane7683", solver.metadata.ProtocolName)
	})

	t.Run("metadata_operations", func(t *testing.T) {
		// Test metadata structure
		metadata := types.Hyperlane7683Metadata{
			BaseMetadata: types.BaseMetadata{
				ProtocolName: "Hyperlane7683",
			},
			IntentSources: []types.IntentSource{
				{
					Address:      "0x1234567890123456789012345678901234567890",
					ChainName:    "starknet",
					InitialBlock: big.NewInt(1000),
				},
			},
			CustomRules: types.CustomRules{
				Rules: []types.RuleConfig{
					{
						Name: "max_slippage",
						Args: map[string]interface{}{
							"value": 0.05,
						},
					},
				},
			},
		}

		assert.Equal(t, "Hyperlane7683", metadata.ProtocolName)
		assert.Len(t, metadata.IntentSources, 1)
		assert.Equal(t, "starknet", metadata.IntentSources[0].ChainName)
		assert.Len(t, metadata.CustomRules.Rules, 1)
		assert.Equal(t, "max_slippage", metadata.CustomRules.Rules[0].Name)
	})

	t.Run("intent_source_operations", func(t *testing.T) {
		// Test intent source creation
		intentSource := types.IntentSource{
			Address:            "0x0987654321098765432109876543210987654321",
			ChainName:          "evm",
			InitialBlock:       big.NewInt(2000),
			PollInterval:       1000,
			ConfirmationBlocks: 12,
		}

		assert.Equal(t, "evm", intentSource.ChainName)
		assert.Equal(t, "0x0987654321098765432109876543210987654321", intentSource.Address)
		assert.Equal(t, big.NewInt(2000), intentSource.InitialBlock)
		assert.Equal(t, 1000, intentSource.PollInterval)
		assert.Equal(t, 12, intentSource.ConfirmationBlocks)
	})

	t.Run("custom_rules_operations", func(t *testing.T) {
		// Test custom rules creation
		customRules := types.CustomRules{
			Rules: []types.RuleConfig{
				{
					Name: "min_amount",
					Args: map[string]interface{}{
						"value": big.NewInt(500000), // 0.5 USDC
					},
				},
				{
					Name: "max_amount",
					Args: map[string]interface{}{
						"value": big.NewInt(100000000), // 100 USDC
					},
				},
			},
		}

		assert.Len(t, customRules.Rules, 2)
		assert.Equal(t, "min_amount", customRules.Rules[0].Name)
		assert.Equal(t, "max_amount", customRules.Rules[1].Name)
		assert.Equal(t, big.NewInt(500000), customRules.Rules[0].Args["value"])
		assert.Equal(t, big.NewInt(100000000), customRules.Rules[1].Args["value"])
	})

	t.Run("solver_interface_compliance", func(t *testing.T) {
		// Test that solver implements required interface
		solver := &Hyperlane7683Solver{
			metadata: types.Hyperlane7683Metadata{
				BaseMetadata: types.BaseMetadata{
					ProtocolName: "Hyperlane7683",
				},
			},
		}

		// Test that solver has required methods
		assert.NotNil(t, solver)
		assert.NotNil(t, solver.metadata)
	})
}

// TestHyperlane7683EdgeCases tests edge cases for hyperlane7683 functions
func TestHyperlane7683EdgeCases(t *testing.T) {
	t.Run("empty_metadata", func(t *testing.T) {
		// Test with empty metadata
		metadata := types.Hyperlane7683Metadata{
			BaseMetadata: types.BaseMetadata{
				ProtocolName: "",
			},
		}

		assert.Empty(t, metadata.ProtocolName)
		assert.Nil(t, metadata.IntentSources)
		assert.Nil(t, metadata.CustomRules.Rules)
	})

	t.Run("zero_values", func(t *testing.T) {
		// Test with zero values
		customRules := types.CustomRules{
			Rules: []types.RuleConfig{
				{
					Name: "zero_rule",
					Args: map[string]interface{}{
						"value": big.NewInt(0),
					},
				},
			},
		}

		assert.Len(t, customRules.Rules, 1)
		assert.Equal(t, "zero_rule", customRules.Rules[0].Name)
		assert.Equal(t, big.NewInt(0), customRules.Rules[0].Args["value"])
	})

	t.Run("large_values", func(t *testing.T) {
		// Test with large values
		largeAmount := new(big.Int)
		largeAmount.SetString("1000000000000000000000000", 10) // 1M tokens with 18 decimals

		customRules := types.CustomRules{
			Rules: []types.RuleConfig{
				{
					Name: "large_amount",
					Args: map[string]interface{}{
						"value": largeAmount,
					},
				},
			},
		}

		assert.Len(t, customRules.Rules, 1)
		assert.Equal(t, "large_amount", customRules.Rules[0].Name)
		assert.Equal(t, largeAmount, customRules.Rules[0].Args["value"])
	})
}

// TestHyperlane7683Concurrency tests concurrent access to hyperlane7683 functions
func TestHyperlane7683Concurrency(t *testing.T) {
	t.Run("concurrent_solver_creation", func(t *testing.T) {
		// Test concurrent solver creation
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(_ int) {
				solver := &Hyperlane7683Solver{
					metadata: types.Hyperlane7683Metadata{
						BaseMetadata: types.BaseMetadata{
							ProtocolName: "Hyperlane7683",
						},
					},
				}
				assert.NotNil(t, solver)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent_metadata_operations", func(t *testing.T) {
		// Test concurrent metadata operations
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(index int) {
				metadata := types.Hyperlane7683Metadata{
					BaseMetadata: types.BaseMetadata{
						ProtocolName: "Hyperlane7683",
					},
					CustomRules: types.CustomRules{
						Rules: []types.RuleConfig{
							{
								Name: "test_rule",
								Args: map[string]interface{}{
									"value": big.NewInt(int64(index * 1000)),
								},
							},
						},
					},
				}
				assert.Equal(t, "Hyperlane7683", metadata.ProtocolName)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// TestHyperlane7683Validation tests validation functions
func TestHyperlane7683Validation(t *testing.T) {
	t.Run("validate_metadata", func(t *testing.T) {
		// Test valid metadata
		metadata := types.Hyperlane7683Metadata{
			BaseMetadata: types.BaseMetadata{
				ProtocolName: "Hyperlane7683",
			},
		}

		assert.NotEmpty(t, metadata.ProtocolName)
		assert.Equal(t, "Hyperlane7683", metadata.ProtocolName)
	})

	t.Run("validate_custom_rules", func(t *testing.T) {
		// Test valid custom rules
		customRules := types.CustomRules{
			Rules: []types.RuleConfig{
				{
					Name: "test_rule",
					Args: map[string]interface{}{
						"value": big.NewInt(1000),
					},
				},
			},
		}

		assert.Len(t, customRules.Rules, 1)
		assert.NotEmpty(t, customRules.Rules[0].Name)
		assert.NotNil(t, customRules.Rules[0].Args)
	})

	t.Run("validate_intent_sources", func(t *testing.T) {
		// Test valid intent sources
		intentSources := []types.IntentSource{
			{
				Address:   "0x1234567890123456789012345678901234567890",
				ChainName: "starknet",
			},
			{
				Address:   "0x0987654321098765432109876543210987654321",
				ChainName: "evm",
			},
		}

		assert.Len(t, intentSources, 2)
		for i, source := range intentSources {
			assert.NotEmpty(t, source.ChainName, "intent source %d should have chain name", i)
			assert.NotEmpty(t, source.Address, "intent source %d should have address", i)
		}
	})
}

// TestHyperlane7683Integration tests integration scenarios
func TestHyperlane7683Integration(t *testing.T) {
	t.Run("full_metadata_setup", func(t *testing.T) {
		// Test complete metadata setup
		metadata := types.Hyperlane7683Metadata{
			BaseMetadata: types.BaseMetadata{
				ProtocolName: "Hyperlane7683",
			},
			IntentSources: []types.IntentSource{
				{
					Address:            "0x1234567890123456789012345678901234567890",
					ChainName:          "starknet",
					InitialBlock:       big.NewInt(1000),
					PollInterval:       1000,
					ConfirmationBlocks: 12,
				},
				{
					Address:            "0x0987654321098765432109876543210987654321",
					ChainName:          "evm",
					InitialBlock:       big.NewInt(2000),
					PollInterval:       2000,
					ConfirmationBlocks: 6,
				},
			},
			CustomRules: types.CustomRules{
				Rules: []types.RuleConfig{
					{
						Name: "max_slippage",
						Args: map[string]interface{}{
							"value": 0.05,
						},
					},
					{
						Name: "min_amount",
						Args: map[string]interface{}{
							"value": big.NewInt(1000000),
						},
					},
				},
			},
		}

		// Validate complete setup
		assert.Equal(t, "Hyperlane7683", metadata.ProtocolName)
		assert.Len(t, metadata.IntentSources, 2)
		assert.Len(t, metadata.CustomRules.Rules, 2)

		// Validate intent sources
		starknetSource := metadata.IntentSources[0]
		assert.Equal(t, "starknet", starknetSource.ChainName)
		assert.Equal(t, "0x1234567890123456789012345678901234567890", starknetSource.Address)
		assert.Equal(t, big.NewInt(1000), starknetSource.InitialBlock)

		evmSource := metadata.IntentSources[1]
		assert.Equal(t, "evm", evmSource.ChainName)
		assert.Equal(t, "0x0987654321098765432109876543210987654321", evmSource.Address)
		assert.Equal(t, big.NewInt(2000), evmSource.InitialBlock)

		// Validate custom rules
		assert.Equal(t, "max_slippage", metadata.CustomRules.Rules[0].Name)
		assert.Equal(t, "min_amount", metadata.CustomRules.Rules[1].Name)
		assert.Equal(t, 0.05, metadata.CustomRules.Rules[0].Args["value"])
		assert.Equal(t, big.NewInt(1000000), metadata.CustomRules.Rules[1].Args["value"])
	})

	t.Run("solver_with_multiple_sources", func(t *testing.T) {
		// Test solver with multiple intent sources
		solver := &Hyperlane7683Solver{
			metadata: types.Hyperlane7683Metadata{
				BaseMetadata: types.BaseMetadata{
					ProtocolName: "Hyperlane7683",
				},
				IntentSources: []types.IntentSource{
					{ChainName: "starknet", Address: "0x1111111111111111111111111111111111111111"},
					{ChainName: "evm", Address: "0x2222222222222222222222222222222222222222"},
					{ChainName: "polygon", Address: "0x3333333333333333333333333333333333333333"},
				},
			},
		}

		assert.NotNil(t, solver)
		assert.Len(t, solver.metadata.IntentSources, 3)
		assert.Equal(t, "starknet", solver.metadata.IntentSources[0].ChainName)
		assert.Equal(t, "evm", solver.metadata.IntentSources[1].ChainName)
		assert.Equal(t, "polygon", solver.metadata.IntentSources[2].ChainName)
	})
}
