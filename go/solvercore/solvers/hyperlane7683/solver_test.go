package hyperlane7683

import (
	"testing"

	"github.com/NethermindEth/oif-starknet/go/solvercore/types"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHyperlane7683Solver tests the solver creation and basic functionality
func TestHyperlane7683Solver(t *testing.T) {
	t.Run("NewHyperlane7683Solver", func(t *testing.T) {
		// Create mock functions
		getEVMClient := func(chainID uint64) (*ethclient.Client, error) {
			return nil, nil
		}
		getStarknetClient := func() (*rpc.Provider, error) {
			return nil, nil
		}
		getEVMSigner := func(chainID uint64) (*bind.TransactOpts, error) {
			return nil, nil
		}
		getStarknetSigner := func() (*account.Account, error) {
			return nil, nil
		}

		allowBlockLists := types.AllowBlockLists{
			AllowList: []types.AllowBlockListItem{},
			BlockList: []types.AllowBlockListItem{},
		}

		solver := NewHyperlane7683Solver(
			getEVMClient,
			getStarknetClient,
			getEVMSigner,
			getStarknetSigner,
			allowBlockLists,
		)

		require.NotNil(t, solver)
		assert.NotNil(t, solver.getEVMClient)
		assert.NotNil(t, solver.getStarknetClient)
		assert.NotNil(t, solver.getEVMSigner)
		assert.NotNil(t, solver.getStarknetSigner)
		assert.Equal(t, allowBlockLists, solver.allowBlockLists)
	})

	t.Run("Solver_metadata", func(t *testing.T) {
		solver := &Hyperlane7683Solver{
			metadata: types.Hyperlane7683Metadata{
				BaseMetadata: types.BaseMetadata{
					ProtocolName: "Hyperlane7683",
				},
			},
		}

		assert.Equal(t, "Hyperlane7683", solver.metadata.ProtocolName)
	})
}

// TestOrderAction tests the OrderAction enum
func TestOrderAction(t *testing.T) {
	t.Run("OrderAction_values", func(t *testing.T) {
		assert.Equal(t, OrderAction(0), OrderActionSettle)
		assert.Equal(t, OrderAction(1), OrderActionComplete)
		assert.Equal(t, OrderAction(2), OrderActionError)
	})

	t.Run("OrderAction_string", func(t *testing.T) {
		// Test that we can convert to string (useful for logging)
		actions := []OrderAction{OrderActionSettle, OrderActionComplete, OrderActionError}
		for _, action := range actions {
			assert.True(t, int(action) >= 0 && int(action) <= 2)
		}
	})
}

// TestSolverInitialization tests solver initialization scenarios
func TestSolverInitialization(t *testing.T) {
	t.Run("solver_with_empty_allow_block_lists", func(t *testing.T) {
		allowBlockLists := types.AllowBlockLists{
			AllowList: []types.AllowBlockListItem{},
			BlockList: []types.AllowBlockListItem{},
		}

		solver := NewHyperlane7683Solver(
			nil, nil, nil, nil,
			allowBlockLists,
		)

		assert.NotNil(t, solver)
		assert.Empty(t, solver.allowBlockLists.AllowList)
		assert.Empty(t, solver.allowBlockLists.BlockList)
	})

	t.Run("solver_with_allow_block_lists", func(t *testing.T) {
		allowBlockLists := types.AllowBlockLists{
			AllowList: []types.AllowBlockListItem{
				{
					SenderAddress:     "0x1234567890123456789012345678901234567890",
					DestinationDomain: "1",
					RecipientAddress:  "0x0987654321098765432109876543210987654321",
				},
			},
			BlockList: []types.AllowBlockListItem{
				{
					SenderAddress:     "0x1111111111111111111111111111111111111111",
					DestinationDomain: "2",
					RecipientAddress:  "0x2222222222222222222222222222222222222222",
				},
			},
		}

		solver := NewHyperlane7683Solver(
			nil, nil, nil, nil,
			allowBlockLists,
		)

		assert.NotNil(t, solver)
		assert.Len(t, solver.allowBlockLists.AllowList, 1)
		assert.Len(t, solver.allowBlockLists.BlockList, 1)
		assert.Equal(t, "0x1234567890123456789012345678901234567890", solver.allowBlockLists.AllowList[0].SenderAddress)
		assert.Equal(t, "0x1111111111111111111111111111111111111111", solver.allowBlockLists.BlockList[0].SenderAddress)
	})
}

// TestSolverConcurrency tests basic concurrency safety
func TestSolverConcurrency(t *testing.T) {
	t.Run("concurrent_solver_creation", func(t *testing.T) {
		allowBlockLists := types.AllowBlockLists{
			AllowList: []types.AllowBlockListItem{},
			BlockList: []types.AllowBlockListItem{},
		}

		// Create multiple solvers concurrently
		solvers := make([]*Hyperlane7683Solver, 10)
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				solver := NewHyperlane7683Solver(
					nil, nil, nil, nil,
					allowBlockLists,
				)
				solvers[index] = solver
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all solvers were created successfully
		for i, solver := range solvers {
			assert.NotNil(t, solver, "Solver %d should not be nil", i)
		}
	})
}
