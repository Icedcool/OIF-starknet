package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetSolverState tests the GetSolverState function
func TestGetSolverState(t *testing.T) {
	t.Run("get_solver_state_basic", func(t *testing.T) {
		// Test basic solver state creation
		state, err := GetSolverState()
		require.NoError(t, err)
		assert.NotNil(t, state)
		assert.NotNil(t, state.Networks)
	})

	t.Run("get_solver_state_structure", func(t *testing.T) {
		// Test that solver state has expected structure
		state, err := GetSolverState()
		require.NoError(t, err)

		// Should have networks map
		assert.NotNil(t, state.Networks)

		// Should be able to add networks
		state.Networks["TestNetwork"] = SolverNetworkState{
			LastIndexedBlock: 1000,
			LastUpdated:      time.Now().Format(time.RFC3339),
		}
		assert.Equal(t, uint64(1000), state.Networks["TestNetwork"].LastIndexedBlock)
	})
}

// TestUpdateLastIndexedBlock tests the UpdateLastIndexedBlock function
func TestUpdateLastIndexedBlock(t *testing.T) {
	t.Run("update_last_indexed_block_basic", func(t *testing.T) {
		// Test basic update functionality
		err := UpdateLastIndexedBlock("Ethereum", 50000)
		require.NoError(t, err)

		// Verify the update by getting state
		state, err := GetSolverState()
		require.NoError(t, err)
		assert.Equal(t, uint64(50000), state.Networks["Ethereum"].LastIndexedBlock)
		assert.NotEmpty(t, state.Networks["Ethereum"].LastUpdated)
	})

	t.Run("update_last_indexed_block_multiple_networks", func(t *testing.T) {
		// Update blocks for multiple networks
		networks := []string{"Ethereum", "Base", "Starknet"}
		blocks := []uint64{10000, 20000, 30000}

		for i, network := range networks {
			err := UpdateLastIndexedBlock(network, blocks[i])
			require.NoError(t, err)
		}

		// Verify all updates
		state, err := GetSolverState()
		require.NoError(t, err)
		for i, network := range networks {
			assert.Equal(t, blocks[i], state.Networks[network].LastIndexedBlock)
		}
	})

	t.Run("update_last_indexed_block_zero_block", func(t *testing.T) {
		// Update block to zero for an existing network
		err := UpdateLastIndexedBlock("Ethereum", 0)
		require.NoError(t, err)

		// Verify the update
		state, err := GetSolverState()
		require.NoError(t, err)
		assert.Equal(t, uint64(0), state.Networks["Ethereum"].LastIndexedBlock)
	})
}

// TestSolverStateBasicOperations tests basic solver state operations
func TestSolverStateBasicOperations(t *testing.T) {
	t.Run("solver_state_creation", func(t *testing.T) {
		state := SolverState{
			Networks: map[string]SolverNetworkState{
				"Test": {
					LastIndexedBlock: 99999,
					LastUpdated:      "2023-12-31T23:59:59Z",
				},
			},
		}

		assert.NotNil(t, state.Networks)
		assert.Equal(t, uint64(99999), state.Networks["Test"].LastIndexedBlock)
		assert.Equal(t, "2023-12-31T23:59:59Z", state.Networks["Test"].LastUpdated)
	})

	t.Run("solver_state_network_operations", func(t *testing.T) {
		state := SolverState{
			Networks: make(map[string]SolverNetworkState),
		}

		// Add a network
		state.Networks["NewNetwork"] = SolverNetworkState{
			LastIndexedBlock: 5000,
			LastUpdated:      time.Now().Format(time.RFC3339),
		}

		assert.Contains(t, state.Networks, "NewNetwork")
		assert.Equal(t, uint64(5000), state.Networks["NewNetwork"].LastIndexedBlock)

		// Update the network
		state.Networks["NewNetwork"] = SolverNetworkState{
			LastIndexedBlock: 6000,
			LastUpdated:      time.Now().Format(time.RFC3339),
		}

		assert.Equal(t, uint64(6000), state.Networks["NewNetwork"].LastIndexedBlock)
	})
}

// TestSolverStateConcurrency tests concurrent access to solver state
func TestSolverStateConcurrency(t *testing.T) {
	t.Run("concurrent_solver_state_access", func(t *testing.T) {
		// Test concurrent updates on existing networks
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(index int) {
				network := "Ethereum"
				block := uint64(1000 + index)
				err := UpdateLastIndexedBlock(network, block)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify final state
		state, err := GetSolverState()
		require.NoError(t, err)
		assert.Contains(t, state.Networks, "Ethereum")
	})

	t.Run("concurrent_read_write", func(t *testing.T) {
		// Start writer goroutines
		writeDone := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func(index int) {
				network := "Base"
				block := uint64(2000 + index)
				err := UpdateLastIndexedBlock(network, block)
				assert.NoError(t, err)
				writeDone <- true
			}(i)
		}

		// Start reader goroutines
		readDone := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func() {
				state, err := GetSolverState()
				assert.NoError(t, err)
				assert.NotNil(t, state)
				readDone <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			<-writeDone
			<-readDone
		}
	})
}

// TestSolverStateEdgeCases tests edge cases for solver state
func TestSolverStateEdgeCases(t *testing.T) {
	t.Run("update_block_with_very_large_block_number", func(t *testing.T) {
		// Update with very large block number for existing network
		largeBlock := uint64(18446744073709551615) // Max uint64
		err := UpdateLastIndexedBlock("Starknet", largeBlock)
		require.NoError(t, err)

		// Verify the update
		state, err := GetSolverState()
		require.NoError(t, err)
		assert.Equal(t, largeBlock, state.Networks["Starknet"].LastIndexedBlock)
	})

	t.Run("update_block_with_zero_block_number", func(t *testing.T) {
		// Update with zero block number for existing network
		err := UpdateLastIndexedBlock("Optimism", 0)
		require.NoError(t, err)

		// Verify the update
		state, err := GetSolverState()
		require.NoError(t, err)
		assert.Equal(t, uint64(0), state.Networks["Optimism"].LastIndexedBlock)
	})

	t.Run("update_block_with_medium_block_number", func(t *testing.T) {
		// Update with medium block number for existing network
		err := UpdateLastIndexedBlock("Arbitrum", 50000)
		require.NoError(t, err)

		// Verify the update
		state, err := GetSolverState()
		require.NoError(t, err)
		assert.Equal(t, uint64(50000), state.Networks["Arbitrum"].LastIndexedBlock)
	})
}
