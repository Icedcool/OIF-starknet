# 🚀 Quick Start Guide

**Hyperlane7683 Intent Solver** in Go.

## 🎯 **What We're Building**

A Go-based intent solver that:

- **Listens to Hyperlane7683 contracts** on multiple testnet forks
- **Processes cross-chain orders** automatically
- **Uses pre-deployed contracts** (no deployment needed)
- **Supports both EVM and Starknet** (Starknet coming soon)

## 🏗️ **Architecture Overview**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Testnet       │    │   Local Fork    │    │   Go Solver     │
│   (Sepolia)     │───▶│   (Port 8545)   │───▶│   (Event       │
│                 │    │                 │    │    Listener)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

**Benefits of forking:**

- ✅ **Real contracts** (Permit2, Hyperlane7683, etc.)
- ✅ **Real token addresses** (USDC, WETH, etc.)
- ✅ **Real gas dynamics** (but you control the network)
- ✅ **No deployment** - everything already exists

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

```
✅ Verified pre-deployed Hyperlane7683 contract at 0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3
🎉 All pre-deployed contracts verified successfully!
```

### **Step 3: Run the Solver** 🧠

In the **same terminal** as Step 2:

```bash
make build
make run
```

**Result:**

- 🔍 Solver starts listening to all testnet forks
- 📡 Connects to Hyperlane7683 contracts
- ⚡ Begins processing intents (currently mock data/lifecycle)

### **Step 4: Open an Order** 📋

**TODO: This step will be implemented next**

We'll add the ability to:

- Open a cross-chain order with a make command

### **Step 5...: Fill Orders** 📋

**TODO: This step will be implemented after**

We'll add the ability to:

- Fill orders observed from events
- Watch the solver process them in real-time
- See the complete intent → fill pipeline

## 🔧 **Available Commands**

```bash
# Network Management
make start-networks      # Start all testnet forks
make kill-networks       # Stop all running networks

# Contract Verification
make verify-hyperlane    # Verify pre-deployed contracts exist

# Solver Development
make build               # Build solver binary
make run                 # Run solver
make test                # Run tests
make clean               # Clean build artifacts
```

## 🎮 **Current Status**

### **✅ Working Features:**

- **Testnet forking** - All 4 networks running locally
- **Contract verification** - Hyperlane7683 accessible on all forks
- **Solver framework** - Mock intent processing pipeline
- **Multi-chain support** - Solver connects to all networks

### **🚧 Coming Next:**

- **Order opening** - Create test orders to fill
- **Real event listening** - Listen to actual Hyperlane7683 events
- **Intent processing** - Process real cross-chain orders
- **Transaction execution** - Actually fill intents on-chain
