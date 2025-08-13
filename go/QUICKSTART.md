# 🚀 Quick Start Guide

**Hyperlane7683 Intent Solver** in Go.

## 🔧 **Available Commands**

```bash
# Network Management
make start-networks      # Start all testnet forks
make kill-networks       # Stop all running networks

# Contract Verification
make verify-hyperlane    # Verify pre-deployed contracts exist

# Token Deployment
make `deploy-tokens`      # Deploy ERC20 tokens, fund accounts, and setup allowances

# Order Management
make open-basic-order   # Open a simple onchain order
make open-random-order  # Open a random onchain order (random origin, destination, in/out amounts, etc.)

# Solver Development
make build               # Build solver binary
make run                 # Run solver
make test                # Run tests
make clean               # Clean build artifacts
```

## 🚀 **Getting Started**

### **Step 1: Start Testnet Forks** 🌐

Open a **new terminal tab**:

```bash
cd go/
make start-networks
```

**Result:**

- 🟢 **Sepolia fork** on port 8545
- 🔵 **Optimism Sepolia fork** on port 8546
- 🟡 **Arbitrum Sepolia fork** on port 8547
- 🟣 **Base Sepolia fork** on port 8548

**Keep this terminal open** - you'll see logs from all networks.

### **Step 2: Verify Contracts** ✅

In **another terminal tab**:

```bash
cd go/
make verify-hyperlane
```

**Expected output:**

```bash
🎉 All pre-deployed contracts verified successfully!
💡 These contracts are ready to use for intent solving!
```

### **Step 3: Deploy Tokens** ✅

In **the same terminal tab**:

```bash
cd go/
make deploy-tokens
```

**Expected output:**

```bash
🎯 All networks configured!
   • OrcaCoin and DogCoin deployed to all networks
   • Test users funded with tokens
   • Allowances set for Hyperlane7683
   • Ready to open orders!
```

### **Step 4: Open some Orders** 📋

In **the same terminal tab**:

```bash
make open-basic-order
```

and/or:

```bash
make open-random-order
```

**Expected output(s):**

```bash
✅ Balance changes verified - order was actually opened!
🎉 Order execution completed!
```

### **Step 5: Run the Solver** 🧠

In the **same terminal**:

```bash
make build
make run
```

**Result:**

```
🙍 Intent Solver 📝
Starting...
Initializing solvers...
Initializing solver: hyperlane7683
Starting multi-network event listener...
...
```

Once the solver is running, it will first catch up on any missed Open events from the last indexed block to the latest block (for each network). Once historical events have been processed, the solver will start listening for new events on a schedule and process them. Note, for now, Open events are logged, the rest of the intent → fill pipeline is not yet implemented.

If the solver is stopped and restarted, it will automatically pickup where it left off and continue processing events from the last indexed block. The last indexed block is only reset when the networks are restarted (not the solver). When the networks are started/restarted, the last indexed block for each network is set to the block that the fork was started at.
