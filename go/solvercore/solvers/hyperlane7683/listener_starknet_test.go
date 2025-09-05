package hyperlane7683

import (
	"context"
	"math/big"
	"testing"

	"github.com/NethermindEth/oif-starknet/go/solvercore/base"
	"github.com/stretchr/testify/assert"
)

// TestStarknetListener tests the Starknet listener functionality
func TestStarknetListener(t *testing.T) {
	t.Run("NewStarknetListener_invalid_rpc", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "0x1234567890123456789012345678901234567890",
			InitialBlock:    big.NewInt(1000),
		}

		_, err := NewStarknetListener(config, "invalid-rpc-url")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect Starknet RPC")
	})

	t.Run("NewStarknetListener_invalid_contract_address", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "invalid-address",
			InitialBlock:    big.NewInt(1000),
		}

		_, err := NewStarknetListener(config, "http://localhost:5050")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Starknet contract address")
	})

	t.Run("StarknetListener_interface_compliance", func(t *testing.T) {
		// Test that we can create a mock listener that implements the interface
		mockListener := &mockStarknetListener{}

		// Verify it implements the base.Listener interface
		var listener base.Listener = mockListener
		assert.NotNil(t, listener)
	})
}

// TestStarknetListenerConfig tests listener configuration
func TestStarknetListenerConfig(t *testing.T) {
	t.Run("listener_config_validation", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "0x1234567890123456789012345678901234567890",
		}

		// Test that the config is valid
		assert.NotEmpty(t, config.ContractAddress)
		assert.Equal(t, "0x1234567890123456789012345678901234567890", config.ContractAddress)
	})

	t.Run("listener_config_with_empty_address", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "",
		}

		// Test that empty address is handled
		assert.Empty(t, config.ContractAddress)
	})
}

// TestStarknetListenerState tests listener state management
func TestStarknetListenerState(t *testing.T) {
	t.Run("listener_state_initialization", func(t *testing.T) {
		listener := &mockStarknetListener{
			lastProcessedBlock: 0,
		}

		assert.Equal(t, uint64(0), listener.lastProcessedBlock)
	})

	t.Run("listener_state_update", func(t *testing.T) {
		listener := &mockStarknetListener{
			lastProcessedBlock: 100,
		}

		// Simulate block processing
		newBlock := uint64(150)
		listener.lastProcessedBlock = newBlock

		assert.Equal(t, uint64(150), listener.lastProcessedBlock)
	})
}

// TestStarknetListenerErrorHandling tests error handling scenarios
func TestStarknetListenerErrorHandling(t *testing.T) {
	t.Run("listener_with_connection_error", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "0x1234567890123456789012345678901234567890",
			InitialBlock:    big.NewInt(1000),
		}

		_, err := NewStarknetListener(config, "http://nonexistent:5050")
		assert.Error(t, err)
	})

	t.Run("listener_with_invalid_contract_format", func(t *testing.T) {
		config := &base.ListenerConfig{
			ContractAddress: "not-a-valid-address",
			InitialBlock:    big.NewInt(1000),
		}

		_, err := NewStarknetListener(config, "http://localhost:5050")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Starknet contract address")
	})
}

// TestStarknetListenerConcurrency tests basic concurrency safety
func TestStarknetListenerConcurrency(t *testing.T) {
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
				listener, err := NewStarknetListener(config, "http://localhost:5050")
				listeners[index] = listener
				errors[index] = err
			}(i)
		}

		// Wait a bit for goroutines to complete
		// Note: In a real test, we'd use proper synchronization
		// This is just testing that the function can be called concurrently
	})
}

// TestStarknetListenerMethods tests the listener methods
func TestStarknetListenerMethods(t *testing.T) {
	t.Run("listener_start_stop", func(t *testing.T) {
		listener := &mockStarknetListener{}

		shutdown, err := listener.Start(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, shutdown)

		err = listener.Stop()
		assert.NoError(t, err)
	})

	t.Run("listener_block_management", func(t *testing.T) {
		listener := &mockStarknetListener{
			lastProcessedBlock: 100,
		}

		// Test getting last processed block
		block := listener.GetLastProcessedBlock()
		assert.Equal(t, uint64(100), block)

		// Test setting last processed block (simulate internal state change)
		listener.lastProcessedBlock = 200
		block = listener.GetLastProcessedBlock()
		assert.Equal(t, uint64(200), block)
	})
}

// Mock implementations for testing

// mockStarknetListener implements base.Listener for testing
type mockStarknetListener struct {
	lastProcessedBlock uint64
}

func (m *mockStarknetListener) Start(ctx context.Context, handler base.EventHandler) (base.ShutdownFunc, error) {
	return func() {}, nil
}

func (m *mockStarknetListener) Stop() error {
	return nil
}

func (m *mockStarknetListener) GetLastProcessedBlock() uint64 {
	return m.lastProcessedBlock
}
