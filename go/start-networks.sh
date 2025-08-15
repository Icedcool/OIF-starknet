#!/bin/bash

# Load environment variables from .env file
if [ -f ".env" ]; then
	export $(cat .env | grep -v '^#' | xargs)
	echo "📋 Loaded environment variables from .env"
fi

# Colors for each network (updated color scheme)
SEPOLIA_COLOR="\033[32m"        # Green
OPT_COLOR="\033[91m"            # Pastel Red
ARB_COLOR="\033[35m"            # Purple
BASE_COLOR="\033[38;5;27m"      # Royal Blue
STARKNET_COLOR="\033[38;5;208m" # Orange
RESET="\033[0m"                 # Reset

# Network IDs
SEPOLIA_ID="[SEP]"
OPT_ID="[OPT]"
ARB_ID="[ARB]"
BASE_ID="[BASE]"
STARKNET_ID="[STRK]"

echo "🚀 Starting All Network Forks (EVM + Starknet)"
echo "=============================================="
echo "💡 All networks will fork mainnet with real infrastructure"
echo "🛑 Use Ctrl+C to stop all networks"
echo ""

# Function to reset deployment state to fork block numbers
reset_deployment_state() {
	echo "🔄 Resetting deployment state to fork block numbers..."

	# Create the deployment state JSON with correct fork blocks
	cat >"deployment-state.json" <<EOF
{
  "networks": {
    "Sepolia": {
      "chainId": 11155111,
      "hyperlaneAddress": "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3",
      "orcaCoinAddress": "",
      "dogCoinAddress": "",
      "lastIndexedBlock": 8319000,
      "lastUpdated": "now"
    },
    "Optimism Sepolia": {
      "chainId": 11155420,
      "hyperlaneAddress": "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3",
      "orcaCoinAddress": "",
      "dogCoinAddress": "",
      "lastIndexedBlock": 27370000,
      "lastUpdated": "now"
    },
    "Arbitrum Sepolia": {
      "chainId": 421614,
      "hyperlaneAddress": "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3",
      "orcaCoinAddress": "",
      "dogCoinAddress": "",
      "lastIndexedBlock": 138020000,
      "lastUpdated": "now"
    },
    "Base Sepolia": {
      "chainId": 84532,
      "hyperlaneAddress": "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3",
      "orcaCoinAddress": "",
      "dogCoinAddress": "",
      "lastIndexedBlock": 25380000,
      "lastUpdated": "now"
    },
    "Starknet Sepolia": {
      "chainId": 23448591,
      "hyperlaneAddress": "",
      "orcaCoinAddress": "",
      "dogCoinAddress": "",
      "lastIndexedBlock": 1530000,
      "lastUpdated": "now"
    }
  }
}
EOF

	echo "✅ Deployment state reset to fork block numbers"
	echo "📝 Event listener will start from correct blocks"
}

# Function to start a testnet fork with color-coded logging
start_network() {
	local port=$1
	local chain_id=$2
	local color=$3
	local id=$4
	local testnet_name=$5

	# Choose RPC endpoint based on availability
	local rpc_url
	if [ -n "$ALCHEMY_API_KEY" ]; then
		case $testnet_name in
		"sepolia")
			rpc_url="https://eth-sepolia.g.alchemy.com/v2/${ALCHEMY_API_KEY}"
			;;
		"optimism-sepolia")
			rpc_url="https://opt-sepolia.g.alchemy.com/v2/${ALCHEMY_API_KEY}"
			;;
		"arbitrum-sepolia")
			rpc_url="https://arb-sepolia.g.alchemy.com/v2/${ALCHEMY_API_KEY}"
			;;
		"base-sepolia")
			rpc_url="https://base-sepolia.g.alchemy.com/v2/${ALCHEMY_API_KEY}"
			;;
		esac
		echo -e "${color}${id}${RESET} Using Alchemy RPC for ${testnet_name}"
	else
		case $testnet_name in
		"sepolia")
			rpc_url="https://rpc.sepolia.org"
			;;
		"optimism-sepolia")
			rpc_url="https://sepolia.optimism.io"
			;;
		"arbitrum-sepolia")
			rpc_url="https://sepolia-rollup.arbitrum.io/rpc"
			;;
		"base-sepolia")
			rpc_url="https://sepolia.base.org"
			;;
		esac
		echo -e "${color}${id}${RESET} Using public RPC for ${testnet_name}"
	fi

	# Fork from block when contract was last active to preserve state
	local fork_block
	case $testnet_name in
	"sepolia")
		fork_block=8319000 # After the last transaction
		;;
	"optimism-sepolia")
		fork_block=27370000 # After the last transaction
		;;
	"arbitrum-sepolia")
		fork_block=138020000 # After the last transactions
		;;
	"base-sepolia")
		fork_block=25380000 # After the last transaction
		;;
	*)
		fork_block=8319001
		;;
	esac
	echo -e "${color}${id}${RESET} Forking ${testnet_name} from block ${fork_block} (when contract was last used)"

	# Start anvil with testnet fork and pipe output through color filter
	anvil --port $port --chain-id $chain_id --fork-url "$rpc_url" --fork-block-number ${fork_block} 2>&1 | while IFS= read -r line; do
		echo -e "${color}${id}${RESET} $line"
	done &

	# Store the PID
	echo $! >"/tmp/anvil_$port.pid"

	echo -e "${color}${id}${RESET} ${testnet_name} fork started on port $port (Chain ID: $chain_id)"
}

# Function to start Starknet with Katana
start_starknet() {
	local port=$1
	local color=$2
	local id=$3

	echo -e "${color}${id}${RESET} Starting Starknet Sepolia fork with Katana..."

	# Check if katana is installed
	if ! command -v katana &>/dev/null; then
		echo -e "${color}${id}${RESET} ❌ Katana not found. Please install it first:"
		echo -e "${color}${id}${RESET}    curl -L https://github.com/dojoengine/dojo/releases/latest/download/katana-installer.sh | bash"
		echo -e "${color}${id}${RESET}    Or visit: https://book.dojoengine.org/toolchain/katana/installation"
		return 1
	fi

	# Choose RPC endpoint based on availability
	local rpc_url
	if [ -n "$ALCHEMY_API_KEY" ]; then
		rpc_url="https://starknet-sepolia.g.alchemy.com/starknet/version/rpc/v0_8/${ALCHEMY_API_KEY}"
		echo -e "${color}${id}${RESET} Using Alchemy RPC for Starknet Sepolia"
	else
		rpc_url="https://free-rpc.nethermind.io/starknet-sepolia-juno/"
		echo -e "${color}${id}${RESET} Using public RPC for Starknet Sepolia"
	fi

	# Start Katana with state forking
	katana --chain-id 23448591 --fork.provider "$rpc_url" --fork.block 1530000 2>&1 | while IFS= read -r line; do
		echo -e "${color}${id}${RESET} $line"
	done &

	# Store the PID
	echo $! >"/tmp/katana_$port.pid"

	echo -e "${color}${id}${RESET} Starknet Sepolia fork started on port $port (Chain ID: 23448591)"
}

# Function to stop all networks
cleanup() {
	echo ""
	echo "🛑 Stopping all networks..."

	# Kill all anvil processes
	for port in 8545 8546 8547 8548; do
		if [ -f "/tmp/anvil_$port.pid" ]; then
			pid=$(cat "/tmp/anvil_$port.pid")
			kill $pid 2>/dev/null || true
			rm -f "/tmp/anvil_$port.pid"
		fi
	done

	# Kill Katana process
	if [ -f "/tmp/katana_5050.pid" ]; then
		pid=$(cat "/tmp/katana_5050.pid")
		kill $pid 2>/dev/null || true
		rm -f "/tmp/katana_5050.pid"
	fi

	# Also kill any remaining anvil/katana processes
	pkill -f "anvil" 2>/dev/null || true
	pkill -f "katana" 2>/dev/null || true

	echo "✅ All networks stopped"
	exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

echo "🔧 Starting network forks..."
echo ""

# Reset deployment state to fork block numbers
reset_deployment_state

# Check if ALCHEMY_API_KEY is set
if [ -z "$ALCHEMY_API_KEY" ]; then
	echo "⚠️  ALCHEMY_API_KEY not set!"
	echo "💡 You'll be rate limited by the demo endpoint"
	echo "💡 Set ALCHEMY_API_KEY in your .env for full access"
	echo "💡 Or use alternative RPC endpoints (see script for options)"
	echo ""
	echo "🔗 Alternative RPC endpoints (free tiers):"
	echo "   • Sepolia: https://rpc.sepolia.org"
	echo "   • Optimism Sepolia: https://sepolia.optimism.io"
	echo "   • Arbitrum Sepolia: https://sepolia-rollup.arbitrum.io/rpc"
	echo "   • Base Sepolia: https://sepolia.base.org"
	echo "   • Starknet Sepolia: https://free-rpc.nethermind.io/starknet-sepolia-juno/"
	echo ""
fi

# Start all networks
start_network 8545 11155111 "$SEPOLIA_COLOR" "$SEPOLIA_ID" "sepolia"
start_network 8546 11155420 "$OPT_COLOR" "$OPT_ID" "optimism-sepolia"
start_network 8547 421614 "$ARB_COLOR" "$ARB_ID" "arbitrum-sepolia"
start_network 8548 84532 "$BASE_COLOR" "$BASE_ID" "base-sepolia"
start_starknet 5050 "$STARKNET_COLOR" "$STARKNET_ID"

echo ""
echo "⏳ Waiting for networks to be ready..."
sleep 3

echo ""
echo "🎉 All network forks are running!"
echo "================================"
echo -e "${SEPOLIA_COLOR}${SEPOLIA_ID}${RESET} Sepolia Fork             - http://localhost:8545 (Chain ID: 11155111)"
echo -e "${OPT_COLOR}${OPT_ID}${RESET} Optimism Sepolia Fork    - http://localhost:8546 (Chain ID: 11155420)"
echo -e "${ARB_COLOR}${ARB_ID}${RESET} Arbitrum Sepolia Fork    - http://localhost:8547 (Chain ID: 421614)"
echo -e "${BASE_COLOR}${BASE_ID}${RESET} Base Sepolia Fork        - http://localhost:8548 (Chain ID: 84532)"
echo -e "${STARKNET_COLOR}${STARKNET_ID}${RESET} Starknet Sepolia Fork   - http://localhost:5050 (Chain ID: 23448591)"
echo ""
echo "🚀 What you get for FREE on all forks:"
echo "   • Permit2 at 0x000000000022D473030F116dDEE9F6B43aC78BA3 (EVM)"
echo "   • USDC, WETH, and other real tokens (EVM)"
echo "   • Hyperlane Mailbox and infrastructure (EVM)"
echo "   • Real gas dynamics and market conditions (EVM)"
echo "   • Starknet state and contracts (Starknet)"
echo ""
echo "📦 Next steps:"
echo "   1. Fund accounts: make fund-accounts"
echo "   2. Deploy Hyperlane7683: make deploy-hyperlane"
echo "   3. Start solver: make run (will start from correct blocks)"
echo ""
echo "🔄 Or restart everything:"
echo "   make restart"
echo ""
echo "💡 Networks will continue logging here..."
echo "💡 Event listener will automatically start from fork blocks"
echo "🛑 Press Ctrl+C to stop all networks"
echo ""

# Wait for user to stop
echo "⏳ Networks running... (Press Ctrl+C to stop)"
wait
