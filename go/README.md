# Hyperlane7683 Solver - Go Implementation

This (Golang) solver is an extension to BootNodeDev's Hyperlane7683 (Typescript) solver adding support for Starknet. This codebase should be used as a reference for protocols to implement or extend.

## Overview

The solver listens for `Open` events from Hyperlane7683 contracts on Starknet and multiple EVM chains, then fills the intents based on configurable rules.

## Architecture

```js
go/
├── bin/                              # Built binaries
├── cmd/                              # CLI entry points
│   ├── open-order/                   # Create orders (EVM & Starknet)
│   │   ├── evm/                      # EVM order creation utilities
│   │   └── starknet/                 # Starknet order creation utilities
│   ├── setup-forks/                  # Setup local testnet forks
│   │   ├── evm/                      # EVM fork setup (Anvil)
│   │   └── starknet/                 # Starknet fork setup (Katana)
│   └── solver/                       # Main solver binary
├── solvercore/                       # Core solver logic
│   ├── base/                         # Core solver interfaces (listener & solver)
│   │   ├── listener.go               # Base listener interface
│   │   └── solver.go                 # Base solver interface
│   ├── config/                       # Configuration management
│   │   ├── config.go                 # Solver configuration
│   │   ├── networks.go               # Multi-chain network configs
│   │   └── solver_state.go           # Tracks last indexed blocks per network
│   ├── contracts/                    # Go bindings/EVM code for contract interactions & deployments
│   │   ├── erc20_contract.go         # ERC20 contract byte code and ABI
│   │   └── hyperlane7683.go          # Hyperlane7683 contract bindings (EVM)
│   ├── logutil/                      # Terminal logging utilities
│   │   ├── logutil.go                # Basic logging utilities
│   │   └── solver_logger.go          # Enhanced solver logging with cross-chain context
│   ├── solvers/                      # Solver implementations
│   │   └── hyperlane7683/            # Hyperlane7683 solver
│   │       ├── chain_handler.go      # Wrapper interface for solvers
│   │       ├── hyperlane_evm.go      # EVM chain handler (fill and settle orders on EVM chains)
│   │       ├── hyperlane_starknet.go # Starknet chain handler (fill and settle orders on Starknet)
│   │       ├── listener_evm.go       # EVM Open event listener
│   │       ├── listener_starknet.go  # Starknet Open event listener
│   │       ├── rules.go              # Intent validation rules (balance checks, profitability, allowlists)
│   │       └── solver.go             # Main Hyperlane7683 solver orchestration
│   ├── types/                        # Cross-chain data structures
│   │   ├── address_utils.go          # Address conversion utilities
│   │   └── types.go                  # Core type definitions
│   └── solver_manager.go             # Solver orchestration & lifecycle
├── pkg/                              # Public utilities
│   ├── ethutil/                      # Ethereum utilities (signing, gas, ERC20)
│   └── starknetutil/                 # Starknet utilities (address conversion, ERC20 operations)
├── state/                            # Persistent state storage
├── Makefile                          # Build & deployment automation
└── go.mod                            # Go module dependencies
```

### Key Design Patterns

#### 1. **Interface-Based Multi-Chain Architecture**

- `Listener` interface enables any blockchain to plug into the system
- `BaseFiller` interface provides a common intent processing pipeline
- Chain-specific implementations handle translation between common types and native operations

#### 2. **Translation Layer Strategy**

The system uses **multiple translation layers** for maximum extensibility:

**Level 1: Chain Events → Common Format**

```
EVM Open Event → ParsedArgs
Starknet Open Event → ParsedArgs
XYZ Chain Event → ParsedArgs (easy to add)
```

**Level 2: Common Format → Chain Operations**

```
ParsedArgs → EVM Fill Transaction (hyperlane_evm.go)
ParsedArgs → Starknet Fill Transaction (hyperlane_starknet.go)
ParsedArgs → XYZ Fill Transaction (hyperlane_xyz.go - easy to add)
```

#### 3. **Concurrent Multi-Network Processing**

- Each network runs its own goroutine-based listener
- All events flow through a unified handler for consistent processing
- Context-based cancellation enables graceful shutdown across all networks

#### 4. **Extensibility for New VMs**

To add support for a new blockchain (e.g., Solana):

1. **Create listener**: `listener_solana.go` implementing `Listener`
2. **Create operations**: `hyperlane_solana.go` with Solana-specific fill logic
3. **Update routing**: Add Solana case in `solver.go` destination routing
4. **Add config**: Network configuration in `solvercore/config/networks.go`

#### **Context-Based Lifecycle Management**

```go
ctx, cancel := context.WithCancel(context.Background())
// All goroutines respect this context for graceful shutdown
```

#### **Coordinated Goroutine Management**

```go
sm.shutdownWg.Add(1)
go func() {
    defer sm.shutdownWg.Done()
    <-sm.ctx.Done()
    shutdownFunc()  // Clean shutdown per network
}()
```

#### **Multi-Network Concurrent Event Processing**

- Each blockchain network runs in its own goroutine
- Events from all chains feed through the same `EventHandler` function
- Maintains **order integrity** while enabling **parallel processing**
- No blocking between networks - if one network is slow, others continue processing

## 🚀 Current Status

**🎉 (Local Sepolia) solves all 3 order types on local forks**: Requires spoofing a call to each EVM Hyperlane7683 contract to register the Starknet domain

**🎉 (Live Sepolia) solves 2/3 order types on live Sepolia:**: Awaiting Hyperlane contract to register Starknet domain

### Recent Improvements

- **✅ Enhanced Logging**: Cross-chain context with colored network tags (`[BASE] → [STRK]`)
- **✅ Rules System**: Pluggable validation with balance checks, profitability analysis, and allow/block lists
- **✅ Performance Optimizations**: `uint256` library for efficient 256-bit arithmetic
- **✅ Single Binary Architecture**: Unified CLI with `tools` subcommands for development
- **✅ Clean Package Structure**: Renamed `internal/` to `solvercore/` for better clarity

## Quick Start

1. Install dependencies:

   ```bash
   go mod tidy
   ```

2. Configure your environment:

   ```bash
   cp example.env .env
   # Edit .env with your configuration
   ```

3. Run the solver:
   ```bash
   make run
   ```

## Configuration

The solver uses environment variables to manage:

- RPC endpoints for different chains
- Private keys for transaction signing
- Contract addresses
- Operational parameters (polling intervals, gas limits, starting block numbers, etc.)

## Running on Local Forks

Besides an Alchemy API key, the `example.env` file has all of the values needed to run the solver locally on forks of Sepolia networks, just make sure you copy them over to a `.env` file. Make sure you have katana and anvil installed before continuing. For an efficient setup, it is recommended that you open 3 terminals and move each of them to the `go/` directory.

In the first terminal, run the following command to make sure the state file is clean and the binaries are built:

```bash
make rebuild
```

After this is finished, run the following command to start local (Sepolia) forks of Ethereum, Optimism, Arbitrum, Base, and Starknet. You can leave this terminal running and watch transaction logs come in.

```bash
make start-networks
```

In the second terminal, run this command to deploy a mock ERC-20 token onto each network, fund the accounts on each network, and register the Starknet domain on each EVM Hyperlane7683 contract:

```bash
make setup-forks
```

Once this is finished, start the solver by running:

```bash
make run
```

We will use the third terminal to create orders. There are 3 order commands to choose from for each of the different order types. Run these at will (they also work on live Sepolia if you adjust your `.env` accordingly):

```bash
make open-random-evm-order    # Opens a random order from one EVM chain to another

make open-random-evm-sn-order # Opens a random order from an EVM chain to Starknet

make open-random-sn-order     # Opens a random order from Starknet to an EVM chain

```

After an order is opened, you will see a single txn on the origin chain (this is the order being opened, allowances are set to infinity for Alice, so there is no need to set it again here). Right after this, you will see the solver detect the event and start processing it. Most of the time, this will lead to 3 transactions on the destination chain (approving the hyperlane contract to spend the solver's tokens, filling the order, then settling it). In some cases, you may see only 2 transactions (fill & settle) if the destination chain solver already has enough allowance to cover the fill amount (i.e solver was interrupted after approving but before filling).

> ⚡Note: The full life-cycle of an order is as follows: 1) Opened on origin; Alice locks input tokens into the origin chain's hyperlane contract. 2) Fill on destination; solver sends output tokens to Alice's destination chain wallet, routed through the destination chain's hyperlane contract. 3) Settle on destination; prevents orders from being filled twice, triggers dispatch. 4) Then finally, the last step is to wait for the Hyperlane protocol to dispatch the settlement to the origin; releasing the locked input tokens to the solver on the origin chain. This last step is not handled by the solver.

> ⚡Note: The current (live) Sepolia contracts allow for the solver to fully process EVM->EVM and EVM->Starknet orders; however, in order for Starknet->EVM orders to be fully processed, the Hyperlane team needs to register the Starknet domain on the EVM Hyperlane7683 contracts. Until this is done, the solver will fill the order and then skip the settlement txn. This is not an issue on local forks as the `setup-forks` command registers the Starknet domain on each EVM Hyperlane7683 contract by spoofing the caller. This functionality is dictated by the `.env` var `FORKING`. If set to `true`, the solver will assume it can call `settle` on the destination EVM chain for Starknet->EVM orders. If set to `false`, it will skip the settlement step for these orders.

### Extra

The solver is capable of back-filling events. You can try ending the solver process (terminal 2) and open some orders. Once you run the solver again, it will pick up the orders that were opened while it was offline and then continue polling live.

If you stop the network forks process (terminal 1), each chain's state is lost. You will need to run the networks again, then run `make setup-forks` to redeploy the contracts, fund the accounts, and re-setup the forked contract states.

The `SOLVER_START_BLOCK` for each network (in the `.env`) becomes the value of the `LAST_INDEXED_BLOCK` for each network when the `state/solver_state.json` file is created. This file updates as the solver runs and (along with the latest block number) it determines the block range for event listener's next poll. If this file does not exist or is empty, it is initialized using the values in the `.env`.

- A value of 0 for the `LAST_INDEXED_BLOCK` tells the solver to not backfill any events for this network and to only listen for new ones. Any other number will be used as the lower bound for backfilling. This works as expected when we set the `SOLVER_START_BLOCK` to 0 and there is no state file created yet (it will initialize it with the values from the `.env`). That is, if a chain's `LAST_INDEXED_BLOCK` is set to 0, then the first time the solver runs (or first time it runs with a clean/empty state file), we will skip backfilling. If we stop and start the solver again, it will no longer have a 0 for `LAST_INDEXED_BLOCK`, so it will start backfilling from the last index. If you wish to jump back to the current block and skip backfilling again (without needing to restart the networks), run `make clean || make rebuild`. This will remove the state file so that it can be re-initialized from the `.env` values next time the solver is started.

## Extending

This implementation is designed to be easily extensible:

- Support new chains in `solvercore/config/networks.go` & `solvercore/solvers/hyperlane7683/`
- Add new solvers (Eco, Polymer) in `solvercore/solvers/`

## 🚀 Areas for Extension

### Test Coverage Improvements
- **`solvercore/contracts`** (0.1% coverage) - Auto-generated contract bindings with minimal testable code
- **`solvercore/logutil`** (26.8% coverage) - Logging utilities and structured logging
- **`solvercore/config`** (11.7% coverage) - Configuration management and environment handling
- **`solvercore/solvers/hyperlane7683`** (4.2% coverage) - Core solver logic and rules engine

### Integration Testing
- Complete order lifecycle testing (opening → filling → settling)
- Cross-chain operations and Hyperlane message passing
- Error scenarios and network failure handling
- Performance testing with high-frequency orders

### Documentation
- Comprehensive test strategy documentation
- CI/CD integration guidelines
- Coverage expectations and thresholds

## License

Apache-2.0
