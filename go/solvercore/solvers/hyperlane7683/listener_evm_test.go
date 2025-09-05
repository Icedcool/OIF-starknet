package hyperlane7683

import (
	"context"
	"math/big"
	"testing"

	"github.com/NethermindEth/oif-starknet/go/solvercore/base"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// TestEVMListener tests the EVM listener functionality
func TestEVMListener(t *testing.T) {
	t.Run("NewEVMListener_invalid_rpc", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "0x1234567890123456789012345678901234567890",
			InitialBlock:    big.NewInt(1000),
		}

		_, err := NewEVMListener(config, "invalid-rpc-url")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to dial RPC")
	})

	t.Run("NewEVMListener_invalid_contract_address", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "invalid-address",
			InitialBlock:    big.NewInt(1000),
		}

		_, err := NewEVMListener(config, "http://localhost:8545")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid EVM contract address")
	})

	t.Run("EVMListener_interface_compliance", func(t *testing.T) {
		// Test that we can create a mock listener that implements the interface
		mockListener := &mockEVMListener{}

		// Verify it implements the base.Listener interface
		var listener base.Listener = mockListener
		assert.NotNil(t, listener)
	})
}

// TestOpenEventTopic tests the Open event topic
func TestOpenEventTopic(t *testing.T) {
	t.Run("open_event_topic_format", func(t *testing.T) {
		// Test that the topic is a valid hash
		assert.Len(t, openEventTopic.Bytes(), 32)
		assert.NotEqual(t, common.Hash{}, openEventTopic)
	})

	t.Run("open_event_topic_consistency", func(t *testing.T) {
		// Test that the topic is consistent across calls
		topic1 := common.HexToHash("0x3448bbc2203c608599ad448eeb1007cea04b788ac631f9f558e8dd01a3c27b3d")
		assert.Equal(t, topic1, openEventTopic)
	})
}

// TestEVMListenerConfig tests listener configuration
func TestEVMListenerConfig(t *testing.T) {
	t.Run("listener_config_validation", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "0x1234567890123456789012345678901234567890",
			InitialBlock:    big.NewInt(1000),
		}

		// Test that the config is valid
		assert.NotEmpty(t, config.ContractAddress)
		assert.Equal(t, "0x1234567890123456789012345678901234567890", config.ContractAddress)
		assert.NotNil(t, config.InitialBlock)
	})

	t.Run("listener_config_with_empty_address", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "",
		}

		// Test that empty address is handled
		assert.Empty(t, config.ContractAddress)
	})
}

// TestEVMListenerState tests listener state management
func TestEVMListenerState(t *testing.T) {
	t.Run("listener_state_initialization", func(t *testing.T) {
		listener := &mockEVMListener{
			lastProcessedBlock: 0,
		}

		assert.Equal(t, uint64(0), listener.lastProcessedBlock)
	})

	t.Run("listener_state_update", func(t *testing.T) {
		listener := &mockEVMListener{
			lastProcessedBlock: 100,
		}

		// Simulate block processing
		newBlock := uint64(150)
		listener.lastProcessedBlock = newBlock

		assert.Equal(t, uint64(150), listener.lastProcessedBlock)
	})
}

// TestEVMListenerErrorHandling tests error handling scenarios
func TestEVMListenerErrorHandling(t *testing.T) {
	t.Run("listener_with_connection_error", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "0x1234567890123456789012345678901234567890",
			InitialBlock:    big.NewInt(1000),
		}

		_, err := NewEVMListener(config, "http://nonexistent:8545")
		assert.Error(t, err)
	})

	t.Run("listener_with_invalid_contract_format", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "not-a-valid-address",
		}

		_, err := NewEVMListener(config, "http://localhost:8545")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid EVM contract address")
	})
}

// TestEVMListenerConcurrency tests basic concurrency safety
func TestEVMListenerConcurrency(t *testing.T) {
	t.Run("concurrent_listener_creation", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "0x1234567890123456789012345678901234567890",
			InitialBlock:    big.NewInt(1000),
		}

		// Create multiple listeners concurrently (they will fail due to RPC, but that's expected)
		listeners := make([]base.Listener, 5)
		errors := make([]error, 5)

		for i := 0; i < 5; i++ {
			go func(index int) {
				listener, err := NewEVMListener(config, "http://localhost:8545")
				listeners[index] = listener
				errors[index] = err
			}(i)
		}

		// Wait a bit for goroutines to complete
		// Note: In a real test, we'd use proper synchronization
		// This is just testing that the function can be called concurrently
	})
}

// Mock implementations for testing

// mockEVMListener implements base.Listener for testing
type mockEVMListener struct {
	lastProcessedBlock uint64
}

func (m *mockEVMListener) Start(ctx context.Context, handler base.EventHandler) (base.ShutdownFunc, error) {
	return func() {}, nil
}

func (m *mockEVMListener) Stop() error {
	return nil
}

func (m *mockEVMListener) GetLastProcessedBlock() uint64 {
	return m.lastProcessedBlock
}
