package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// Network configuration - will be loaded from deployment state
var networks []NetworkConfig

// NetworkConfig represents a single network configuration
type NetworkConfig struct {
	name             string
	url              string
	chainID          uint64
	hyperlaneAddress common.Address
	orcaCoinAddress  common.Address
	dogCoinAddress   common.Address
}

// loadNetworks loads network configuration from deployment state
func loadNetworks() error {
	state, err := deployer.GetDeploymentState()
	if err != nil {
		return fmt.Errorf("failed to load deployment state: %w", err)
	}

	networks = []NetworkConfig{
		{"Sepolia", "http://localhost:8545", 11156380, common.HexToAddress("0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3"), common.Address{}, common.Address{}},
		{"Optimism Sepolia", "http://localhost:8546", 11156391, common.HexToAddress("0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3"), common.Address{}, common.Address{}},
		{"Arbitrum Sepolia", "http://localhost:8547", 421614, common.HexToAddress("0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3"), common.Address{}, common.Address{}},
		{"Base Sepolia", "http://localhost:8548", 84532, common.HexToAddress("0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3"), common.Address{}, common.Address{}},
	}

	// Update with actual deployed addresses from state
	for i, network := range networks {
		if stateNetwork, exists := state.Networks[network.name]; exists {
			if stateNetwork.OrcaCoinAddress != "" {
				networks[i].orcaCoinAddress = common.HexToAddress(stateNetwork.OrcaCoinAddress)
				fmt.Printf("   🔍 Loaded %s OrcaCoin: %s\n", network.name, stateNetwork.OrcaCoinAddress)
			}
			if stateNetwork.DogCoinAddress != "" {
				networks[i].dogCoinAddress = common.HexToAddress(stateNetwork.DogCoinAddress)
				fmt.Printf("   🔍 Loaded %s DogCoin: %s\n", network.name, stateNetwork.DogCoinAddress)
			}
		} else {
			fmt.Printf("   ⚠️  No deployment state found for network: %s\n", network.name)
		}
	}

	return nil
}

// getHyperlaneDomain returns the Hyperlane domain ID for a given network
// These are the exact values our forked networks are using as chain IDs
func getHyperlaneDomain(networkName string) uint32 {
	switch networkName {
	case "Sepolia":
		return 11156380 // 0xaa37dc - matches actual contract (from fill event)
	case "Optimism Sepolia":
		return 11156391 // 0xaa36a7 - matches actual contract (from fill event)
	case "Arbitrum Sepolia":
		return 421614 // 0x66eee - matches our fork
	case "Base Sepolia":
		return 84532 // 0x14a34 - matches actual contract (from router enrollment)
	default:
		return 0
	}
}

// Test user configuration
var testUsers = []struct {
	name       string
	privateKey string
	address    string
}{
	{"Alice", "ALICE_PRIVATE_KEY", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"},
	{"Bob", "BOB_PRIVATE_KEY", "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"},
	{"Charlie", "CHARLIE_PRIVATE_KEY", "0x90F79bf6EB2c4f870365E785982E1f101E93b906"},
}

// Order configuration
type OrderConfig struct {
	OriginChain      string
	DestinationChain string
	InputToken       string
	OutputToken      string
	InputAmount      *big.Int
	OutputAmount     *big.Int
	User             string
	OpenDeadline     uint32
	FillDeadline     uint32
}

// OrderData struct matching the Solidity OrderData
type OrderData struct {
	Sender             [32]byte
	Recipient          [32]byte
	InputToken         [32]byte
	OutputToken        [32]byte
	AmountIn           *big.Int
	AmountOut          *big.Int
	SenderNonce        *big.Int
	OriginDomain       uint32
	DestinationDomain  uint32
	DestinationSettler [32]byte
	FillDeadline       uint32
	Data               []byte
}

// OnchainCrossChainOrder struct matching the Solidity interface
type OnchainCrossChainOrder struct {
	FillDeadline  uint32
	OrderDataType [32]byte
	OrderData     []byte
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Load network configuration from deployment state
	if err := loadNetworks(); err != nil {
		log.Fatalf("Failed to load networks: %v", err)
	}

	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: open-order <command>")
		fmt.Println("Commands:")
		fmt.Println("  basic     - Open a basic hardcoded test order")
		fmt.Println("  random    - Open a randomly generated order")
		fmt.Println("  batch     - Open multiple random orders (future)")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "basic":
		openBasicOrder()
	case "random":
		openRandomOrder()
	case "batch":
		fmt.Println("Batch order opening not yet implemented")
		os.Exit(1)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func openBasicOrder() {
	fmt.Println("🚀 Opening Basic Test Order...")

	// Hardcoded basic order
	order := OrderConfig{
		OriginChain:      "Sepolia",
		DestinationChain: "Optimism Sepolia",
		InputToken:       "OrcaCoin",
		OutputToken:      "DogCoin",
		InputAmount:      new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 1000 tokens
		OutputAmount:     new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 1000 tokens
		User:             "Alice",
		OpenDeadline:     uint32(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint32(time.Now().Add(24 * time.Hour).Unix()),
	}

	executeOrder(order)
}

func openRandomOrder() {
	fmt.Println("🎲 Opening Random Test Order...")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Random origin and destination chains
	originIdx := rand.Intn(len(networks))
	destIdx := rand.Intn(len(networks))
	for destIdx == originIdx {
		destIdx = rand.Intn(len(networks))
	}

	// Random user
	userIdx := rand.Intn(len(testUsers))

	// Random amounts (100-10000 tokens)
	inputAmount := rand.Intn(9901) + 100  // 100-10000
	outputAmount := rand.Intn(9901) + 100 // 100-10000

	order := OrderConfig{
		OriginChain:      networks[originIdx].name,
		DestinationChain: networks[destIdx].name,
		InputToken:       "OrcaCoin",
		OutputToken:      "DogCoin",
		InputAmount:      new(big.Int).Mul(big.NewInt(int64(inputAmount)), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		OutputAmount:     new(big.Int).Mul(big.NewInt(int64(outputAmount)), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		User:             testUsers[userIdx].name,
		OpenDeadline:     uint32(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint32(time.Now().Add(24 * time.Hour).Unix()),
	}

	fmt.Printf("🎯 Random Order Generated:\n")
	fmt.Printf("   Origin: %s\n", order.OriginChain)
	fmt.Printf("   Destination: %s\n", order.DestinationChain)
	fmt.Printf("   User: %s\n", order.User)
	inputFloat := new(big.Float).Quo(new(big.Float).SetInt(order.InputAmount), new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	outputFloat := new(big.Float).Quo(new(big.Float).SetInt(order.OutputAmount), new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	fmt.Printf("   Input: %s OrcaCoins\n", inputFloat.Text('f', 0))
	fmt.Printf("   Output: %s DogCoins\n", outputFloat.Text('f', 0))

	executeOrder(order)
}

func executeOrder(order OrderConfig) {
	fmt.Printf("\n📋 Executing Order: %s → %s\n", order.OriginChain, order.DestinationChain)

	// Find origin network
	var originNetwork *NetworkConfig

	for _, network := range networks {
		if network.name == order.OriginChain {
			originNetwork = &network
			break
		}
	}

	if originNetwork == nil {
		log.Fatalf("Origin network not found: %s", order.OriginChain)
	}

	// Connect to origin network
	client, err := ethclient.Dial(originNetwork.url)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", order.OriginChain, err)
	}
	defer client.Close()

	// Get user private key
	userKey := os.Getenv(fmt.Sprintf("%s_PRIVATE_KEY", strings.ToUpper(order.User)))
	if userKey == "" {
		log.Fatalf("Private key not found for user: %s", order.User)
	}

	// Parse private key
	userKey = strings.TrimPrefix(userKey, "0x")
	privateKey, err := crypto.HexToECDSA(userKey)
	if err != nil {
		log.Fatalf("Failed to parse private key for %s: %v", order.User, err)
	}

	// Create auth
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(int64(originNetwork.chainID)))
	if err != nil {
		log.Fatalf("Failed to create auth: %v", err)
	}

	// Get current gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}
	auth.GasPrice = gasPrice

	fmt.Printf("   🔗 Connected to %s (Chain ID: %d)\n", order.OriginChain, originNetwork.chainID)
	fmt.Printf("   👤 User: %s (%s)\n", order.User, crypto.PubkeyToAddress(privateKey.PublicKey).Hex())

	// Find destination network
	var destinationNetwork *NetworkConfig

	for _, network := range networks {
		if network.name == order.DestinationChain {
			destinationNetwork = &network
			break
		}
	}

	if destinationNetwork == nil {
		log.Fatalf("Destination network not found: %s", order.DestinationChain)
	}

    // Read localDomain from the origin Hyperlane contract to guarantee it matches on-chain
	localDomain, err := getLocalDomain(client, originNetwork.hyperlaneAddress)
	if err != nil {
		log.Fatalf("Failed to read localDomain from origin contract: %v", err)
	}
	fmt.Printf("   🔎 localDomain (on-chain): %d\n", localDomain)

    // Preflight: balances and allowances on origin for input token
    inputToken := originNetwork.orcaCoinAddress
    owner := auth.From
    spender := originNetwork.hyperlaneAddress
    bal, err := getErc20Balance(client, inputToken, owner)
    if err == nil {
        fmt.Printf("   🔍 InputToken balance(owner): %s\n", bal.String())
    } else {
        fmt.Printf("   ⚠️  Could not read balance: %v\n", err)
    }
    allowance, err := getErc20Allowance(client, inputToken, owner, spender)
    if err == nil {
        fmt.Printf("   🔍 InputToken allowance(owner→settler): %s\n", allowance.String())
    } else {
        fmt.Printf("   ⚠️  Could not read allowance: %v\n", err)
    }

	// Pick a fresh senderNonce recognized by the contract to avoid InvalidNonce
	senderNonce, err := pickValidSenderNonce(client, originNetwork.hyperlaneAddress, auth.From)
	if err != nil {
		log.Fatalf("Failed to pick a valid sender nonce: %v", err)
	}

	// Build the order data
	orderData := buildOrderData(order, originNetwork, destinationNetwork, localDomain, senderNonce)

	// Debug: Print the order data we're building
	fmt.Printf("   📝 Order data details:\n")
	fmt.Printf("      Sender: %x\n", orderData.Sender)
	fmt.Printf("      Recipient: %x\n", orderData.Recipient)
	fmt.Printf("      InputToken: %x\n", orderData.InputToken)
	fmt.Printf("      OutputToken: %x\n", orderData.OutputToken)
	fmt.Printf("      AmountIn: %s\n", orderData.AmountIn.String())
	fmt.Printf("      AmountOut: %s\n", orderData.AmountOut.String())
	fmt.Printf("      OriginDomain: %d (Hyperlane domain for %s)\n", orderData.OriginDomain, originNetwork.name)
	fmt.Printf("      DestinationDomain: %d (Hyperlane domain for %s)\n", orderData.DestinationDomain, destinationNetwork.name)

	// Build the OnchainCrossChainOrder
	crossChainOrder := OnchainCrossChainOrder{
		FillDeadline:  order.FillDeadline,
		OrderDataType: getOrderDataTypeHash(),
		OrderData:     encodeOrderData(orderData),
	}

	// Debug: Print the cross-chain order details
	fmt.Printf("   📝 Cross-chain order details:\n")
	fmt.Printf("      FillDeadline: %d\n", crossChainOrder.FillDeadline)
	fmt.Printf("      OrderDataType: %x\n", crossChainOrder.OrderDataType)
	fmt.Printf("      OrderData length: %d bytes\n", len(crossChainOrder.OrderData))

	// Build the open() function call with the CORRECT selector
	fmt.Printf("   📝 Order data prepared\n")
	openSelector := []byte{0xe9, 0x17, 0xa9, 0x62} // CORRECT open() selector from valid transaction

	// The open() function expects: open(uint32 fillDeadline, bytes32 orderDataType, bytes calldata orderData)
	// NOT the OnchainCrossChainOrder struct
	openCalldata := encodeOpenFunctionCallDirect(openSelector, crossChainOrder.FillDeadline, crossChainOrder.OrderDataType, crossChainOrder.OrderData)

	fmt.Printf("   🔑 Function selector: 0x%x (CORRECT open function)\n", openSelector)
	fmt.Printf("   📦 Calldata length: %d bytes\n", len(openCalldata))

	// Send the open transaction
	tx, err := sendOpenTransaction(client, auth, originNetwork.hyperlaneAddress, openCalldata)
	if err != nil {
		log.Fatalf("Failed to send open transaction: %v", err)
	}

	fmt.Printf("   🚀 Transaction sent: %s\n", tx.Hash().Hex())
	fmt.Printf("   ⏳ Waiting for confirmation...\n")

	// Wait for transaction confirmation
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("Failed to wait for transaction confirmation: %v", err)
	}

	if receipt.Status == 1 {
		fmt.Printf("   ✅ Order opened successfully!\n")
		fmt.Printf("   📊 Gas used: %d\n", receipt.GasUsed)
		fmt.Printf("   🎯 Order ID: %s\n", calculateOrderId(orderData))
	} else {
		fmt.Printf("   ❌ Order opening failed\n")
		fmt.Printf("   🔍 Transaction hash: %s\n", tx.Hash().Hex())
		fmt.Printf("   📊 Gas used: %d\n", receipt.GasUsed)

		// Try to get more details about the failure
		fmt.Printf("   🔍 Checking transaction details...\n")
		txDetails, _, err := client.TransactionByHash(context.Background(), tx.Hash())
		if err != nil {
			fmt.Printf("   ❌ Could not retrieve transaction details: %v\n", err)
		} else {
			fmt.Printf("   📝 Transaction data: 0x%x\n", txDetails.Data())
		}

        // Simulate to get revert data and decode any custom errors
        fmt.Printf("   🔍 Simulating call to decode revert...\n")
        if err := simulateAndDecodeRevert(client, originNetwork.hyperlaneAddress, auth.From, openCalldata); err != nil {
            fmt.Printf("   ❌ Failed to simulate/decode revert: %v\n", err)
        }
	}

	fmt.Printf("\n🎉 Order execution completed!\n")
}

// Minimal ERC20 helpers
func getErc20Balance(client *ethclient.Client, token common.Address, owner common.Address) (*big.Int, error) {
    abiStr := `[{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`
    p, err := abi.JSON(strings.NewReader(abiStr))
    if err != nil { return nil, err }
    data, err := p.Pack("balanceOf", owner)
    if err != nil { return nil, err }
    msg := ethereum.CallMsg{To: &token, Data: data}
    res, err := client.CallContract(context.Background(), msg, nil)
    if err != nil { return nil, err }
    var out *big.Int
    if err := p.UnpackIntoInterface(&out, "balanceOf", res); err != nil { return nil, err }
    return out, nil
}

func getErc20Allowance(client *ethclient.Client, token common.Address, owner, spender common.Address) (*big.Int, error) {
    abiStr := `[{"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"address","name":"","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`
    p, err := abi.JSON(strings.NewReader(abiStr))
    if err != nil { return nil, err }
    data, err := p.Pack("allowance", owner, spender)
    if err != nil { return nil, err }
    msg := ethereum.CallMsg{To: &token, Data: data}
    res, err := client.CallContract(context.Background(), msg, nil)
    if err != nil { return nil, err }
    var out *big.Int
    if err := p.UnpackIntoInterface(&out, "allowance", res); err != nil { return nil, err }
    return out, nil
}

// simulateAndDecodeRevert runs an eth_call with the same calldata and decodes common custom errors
func simulateAndDecodeRevert(client *ethclient.Client, to, from common.Address, data []byte) error {
    msg := ethereum.CallMsg{To: &to, From: from, Data: data, Gas: 2_000_000}
    _, err := client.CallContract(context.Background(), msg, nil)
    if err == nil {
        fmt.Printf("   ✅ eth_call succeeded (unexpected for a failing tx)\n")
        return nil
    }
    // Expecting a "execution reverted"-style error; extract data if present
    // go-ethereum attaches the revert data to the error string; we can do a second call with a custom tracer,
    // or parse with the built-in helper. Simpler: do CallContract with a big gas and then fetch debug via error string.
    // As a fallback, attempt to decode if the provider returns the revert data hex in err.Error().

    errStr := err.Error()
    // Look for "data: 0x..." substring
    idx := strings.Index(errStr, "data: 0x")
    if idx == -1 {
        fmt.Printf("   ⚠️  No revert data found in error string: %s\n", errStr)
        return nil
    }
    hexData := errStr[idx+6:]
    // Trim trailing context
    if sp := strings.Index(hexData, "]"); sp != -1 {
        hexData = hexData[:sp]
    }
    hexData = strings.TrimSpace(hexData)
    if strings.HasSuffix(hexData, ",") {
        hexData = hexData[:len(hexData)-1]
    }
    if strings.HasPrefix(hexData, "0x") {
        hexData = hexData[2:]
    }
    revertData, decErr := hex.DecodeString(hexData)
    if decErr != nil || len(revertData) < 4 {
        fmt.Printf("   ⚠️  Could not parse revert data from error: %v\n", decErr)
        return nil
    }

    // Known custom errors
    errorABI := `[
        {"type":"error","name":"InvalidOrderType","inputs":[{"type":"bytes32","name":"orderType"}]},
        {"type":"error","name":"InvalidOriginDomain","inputs":[{"type":"uint32","name":"originDomain"}]},
        {"type":"error","name":"InvalidOrderId","inputs":[]},
        {"type":"error","name":"OrderFillExpired","inputs":[]},
        {"type":"error","name":"InvalidOrderDomain","inputs":[]},
        {"type":"error","name":"InvalidDomain","inputs":[]},
        {"type":"error","name":"InvalidSender","inputs":[]},
        {"type":"error","name":"InvalidNonce","inputs":[]},
        {"type":"error","name":"InvalidNativeAmount","inputs":[]}
    ]`
    parsed, err := abi.JSON(strings.NewReader(errorABI))
    if err != nil {
        return err
    }

    sel := revertData[:4]
    for name, def := range parsed.Errors {
        if bytes.Equal(sel, def.ID.Bytes()[:4]) {
            // Decode args
            vals, decErr := def.Inputs.Unpack(revertData[4:])
            if decErr != nil {
                fmt.Printf("   🔍 Revert: %s (failed to unpack args: %v)\n", name, decErr)
            } else {
                fmt.Printf("   🔍 Revert: %s %v\n", name, vals)
            }
            return nil
        }
    }
    fmt.Printf("   ⚠️  Unknown revert selector: 0x%x\n", sel)
    return nil
}

func buildOrderData(order OrderConfig, originNetwork *NetworkConfig, destinationNetwork *NetworkConfig, originDomain uint32, senderNonce *big.Int) OrderData {
	// Get the actual user address for the specified user
	var userAddr common.Address
	for _, user := range testUsers {
		if user.name == order.User {
			userAddr = common.HexToAddress(user.address)
			break
		}
	}

	// For now, use the same address as recipient (user receives their own tokens back)
	recipientAddr := userAddr

	// Use the actual deployed token addresses
	// Input token from origin network, output token from destination network
	inputTokenAddr := originNetwork.orcaCoinAddress
	outputTokenAddr := destinationNetwork.dogCoinAddress // Correct: Use destination network's DogCoin

	// Convert destination settler to bytes32
	destSettler := destinationNetwork.hyperlaneAddress

	// Convert addresses to bytes32 (left-pad with zeros)
	var userBytes, recipientBytes, inputTokenBytes, outputTokenBytes, destSettlerBytes [32]byte
	copy(userBytes[12:], userAddr.Bytes()) // Address is 20 bytes, pad to 32
	copy(recipientBytes[12:], recipientAddr.Bytes())
	copy(inputTokenBytes[12:], inputTokenAddr.Bytes())
	copy(outputTokenBytes[12:], outputTokenAddr.Bytes())
	copy(destSettlerBytes[12:], destSettler.Bytes())

	return OrderData{
		Sender:             userBytes,
		Recipient:          recipientBytes,
		InputToken:         inputTokenBytes,
		OutputToken:        outputTokenBytes,
		AmountIn:           order.InputAmount,
		AmountOut:          order.OutputAmount,
		SenderNonce:        senderNonce,
		OriginDomain:       originDomain,
		DestinationDomain:  getHyperlaneDomain(destinationNetwork.name),
		DestinationSettler: destSettlerBytes,
		FillDeadline:       order.FillDeadline,
		Data:               []byte{}, // Empty for basic orders
	}
}

// getLocalDomain reads the `localDomain()` from the Hyperlane7683 contract on the connected chain
func getLocalDomain(client *ethclient.Client, contractAddress common.Address) (uint32, error) {
	abiStr := `[{"inputs":[],"name":"localDomain","outputs":[{"internalType":"uint32","name":"","type":"uint32"}],"stateMutability":"view","type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return 0, err
	}

	data, err := parsedABI.Pack("localDomain")
	if err != nil {
		return 0, err
	}

	msg := ethereum.CallMsg{To: &contractAddress, Data: data}
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return 0, err
	}

	var domain uint32
	if err := parsedABI.UnpackIntoInterface(&domain, "localDomain", result); err != nil {
		return 0, err
	}

	return domain, nil
}

// isValidNonce calls the contract to check whether a nonce is usable for a given address
func isValidNonce(client *ethclient.Client, contractAddress, from common.Address, nonce *big.Int) (bool, error) {
	abiStr := `[{"inputs":[{"internalType":"address","name":"_from","type":"address"},{"internalType":"uint256","name":"_nonce","type":"uint256"}],"name":"isValidNonce","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return false, err
	}

	data, err := parsedABI.Pack("isValidNonce", from, nonce)
	if err != nil {
		return false, err
	}

	msg := ethereum.CallMsg{To: &contractAddress, Data: data}
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return false, err
	}

	var valid bool
	if err := parsedABI.UnpackIntoInterface(&valid, "isValidNonce", result); err != nil {
		return false, err
	}
	return valid, nil
}

// pickValidSenderNonce finds a nonce that the contract reports as valid for the provided sender
func pickValidSenderNonce(client *ethclient.Client, contractAddress, from common.Address) (*big.Int, error) {
	// Start with a pseudo-random seed and probe upward
	seed := time.Now().Unix() % 1_000_000
	if seed < 1 {
		seed = 1
	}
	nonce := big.NewInt(seed)
	for i := 0; i < 1000; i++ {
		valid, err := isValidNonce(client, contractAddress, from, nonce)
		if err != nil {
			return nil, err
		}
		if valid {
			return new(big.Int).Set(nonce), nil
		}
		nonce = new(big.Int).Add(nonce, big.NewInt(1))
	}
	return nil, fmt.Errorf("could not find a valid sender nonce after 1000 attempts starting from %d", seed)
}

func getOrderDataTypeHash() [32]byte {
	// This should match OrderEncoder.orderDataType() from Solidity EXACTLY
	// Including field names and spacing
	orderDataType := "OrderData(bytes32 sender,bytes32 recipient,bytes32 inputToken,bytes32 outputToken,uint256 amountIn,uint256 amountOut,uint256 senderNonce,uint32 originDomain,uint32 destinationDomain,bytes32 destinationSettler,uint32 fillDeadline,bytes data)"
	hash := crypto.Keccak256Hash([]byte(orderDataType))

	// Debug: Print the hash we're generating
	fmt.Printf("   🔍 Generated orderDataType hash: %x\n", hash)
	fmt.Printf("   🔍 Expected hash from contract: 08d75650babf4de09c9273d48ef647876057ed91d4323f8a2e3ebc2cd8a63b5e\n")

	return hash
}

func encodeOrderData(orderData OrderData) []byte {
	// Pack a single tuple that matches Solidity's OrderData and abi.encode(order)
	tupleT, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "sender", Type: "bytes32"},
		{Name: "recipient", Type: "bytes32"},
		{Name: "inputToken", Type: "bytes32"},
		{Name: "outputToken", Type: "bytes32"},
		{Name: "amountIn", Type: "uint256"},
		{Name: "amountOut", Type: "uint256"},
		{Name: "senderNonce", Type: "uint256"},
		{Name: "originDomain", Type: "uint32"},
		{Name: "destinationDomain", Type: "uint32"},
		{Name: "destinationSettler", Type: "bytes32"},
		{Name: "fillDeadline", Type: "uint32"},
		{Name: "data", Type: "bytes"},
	})
	if err != nil {
		log.Fatalf("Failed to define OrderData tuple type: %v", err)
	}

	args := abi.Arguments{{Type: tupleT}}

	encoded, err := args.Pack(orderData)
	if err != nil {
		log.Fatalf("Failed to ABI-pack OrderData tuple: %v", err)
	}

	fmt.Printf("   ✅ OrderData ABI encoding successful\n")
	return encoded
}

func encodeOrderDataManual(orderData OrderData) []byte {
	// Manual encoding following the exact ABI format for tuples
	// Based on the TypeScript typechain interface

	encoded := make([]byte, 0)

	// For a tuple with dynamic bytes field, we need:
	// 1. Offset to the bytes data field (32 bytes)
	// 2. All static fields in order (each 32 bytes)
	// 3. The dynamic bytes field (length + data)

	// Calculate offset to the bytes data field
	// From valid transaction: offset = 0x60 (96 bytes)
	// This means: 32 (offset) + 64 (static fields) = 96 bytes
	offset := big.NewInt(96)
	offsetBytes := make([]byte, 32)
	offset.FillBytes(offsetBytes)
	encoded = append(encoded, offsetBytes...)

	// Add all static fields (each padded to 32 bytes)
	encoded = append(encoded, orderData.Sender[:]...)
	encoded = append(encoded, orderData.Recipient[:]...)
	encoded = append(encoded, orderData.InputToken[:]...)
	encoded = append(encoded, orderData.OutputToken[:]...)

	// uint256 fields (32 bytes each)
	amountInBytes := make([]byte, 32)
	orderData.AmountIn.FillBytes(amountInBytes)
	encoded = append(encoded, amountInBytes...)

	amountOutBytes := make([]byte, 32)
	orderData.AmountOut.FillBytes(amountOutBytes)
	encoded = append(encoded, amountOutBytes...)

	senderNonceBytes := make([]byte, 32)
	orderData.SenderNonce.FillBytes(senderNonceBytes)
	encoded = append(encoded, senderNonceBytes...)

	// uint32 fields (padded to 32 bytes)
	originDomainBytes := make([]byte, 32)
	big.NewInt(int64(orderData.OriginDomain)).FillBytes(originDomainBytes)
	encoded = append(encoded, originDomainBytes...)

	destDomainBytes := make([]byte, 32)
	big.NewInt(int64(orderData.DestinationDomain)).FillBytes(destDomainBytes)
	encoded = append(encoded, destDomainBytes...)

	encoded = append(encoded, orderData.DestinationSettler[:]...)

	fillDeadlineBytes := make([]byte, 32)
	big.NewInt(int64(orderData.FillDeadline)).FillBytes(fillDeadlineBytes)
	encoded = append(encoded, fillDeadlineBytes...)

	// Now add the bytes data (length + data)
	dataLenBytes := make([]byte, 32)
	big.NewInt(int64(len(orderData.Data))).FillBytes(dataLenBytes)
	encoded = append(encoded, dataLenBytes...)
	encoded = append(encoded, orderData.Data...)

	return encoded
}

func encodeOpenFunctionCall(selector []byte, order OnchainCrossChainOrder) []byte {
	// Encode the function call: open((uint32,bytes32,bytes))
	// The function signature is: open((uint32,bytes32,bytes))
	// We need to encode the tuple parameter properly

	calldata := make([]byte, 0)
	calldata = append(calldata, selector...)

	// For the tuple parameter, we need:
	// 1. Offset to the tuple data (32 bytes)
	// 2. The tuple data itself

	// Offset is 32 bytes (the size of the offset field)
	offset := big.NewInt(32)
	offsetBytes := make([]byte, 32)
	offset.FillBytes(offsetBytes)
	calldata = append(calldata, offsetBytes...)

	// Now add the tuple data
	calldata = append(calldata, order.OrderData...)

	return calldata
}

func encodeOpenFunctionCallDirect(selector []byte, fillDeadline uint32, orderDataType [32]byte, orderData []byte) []byte {
	// Expert Go approach: Use the exact ABI from the compiled contract
	// Function signature: open((uint32,bytes32,bytes)) - tuple parameter

	// Use the exact ABI from the compiled Hyperlane7683 contract
	functionABI := `[{"type":"function","name":"open","inputs":[{"type":"tuple","components":[{"type":"uint32","name":"fillDeadline"},{"type":"bytes32","name":"orderDataType"},{"type":"bytes","name":"orderData"}]}],"outputs":[],"stateMutability":"payable"}]`

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(functionABI))
	if err != nil {
		fmt.Printf("   ⚠️  Function ABI parsing failed: %v\n", err)
		// Fallback to manual encoding
		return encodeOpenFunctionCallDirectManual(selector, fillDeadline, orderDataType, orderData)
	}

	// Create the struct that matches the ABI exactly
	orderParam := OnchainCrossChainOrder{
		FillDeadline:  fillDeadline,
		OrderDataType: orderDataType,
		OrderData:     orderData,
	}

	// Use abi.Pack to encode the function call - this handles everything automatically
	encoded, err := parsedABI.Pack("open", orderParam)
	if err != nil {
		fmt.Printf("   ⚠️  Function encoding failed: %v\n", err)
		// Fallback to manual encoding
		return encodeOpenFunctionCallDirectManual(selector, fillDeadline, orderDataType, orderData)
	}

	fmt.Printf("   ✅ Function ABI encoding successful using abi.Pack\n")

	// abi.Pack() already includes the function selector, so we don't need to add it manually
	return encoded
}

func encodeOpenFunctionCallDirectManual(selector []byte, fillDeadline uint32, orderDataType [32]byte, orderData []byte) []byte {
	// Manual encoding fallback for the open() function
	// Function signature: open(uint32 fillDeadline, bytes32 orderDataType, bytes calldata orderData)

	calldata := make([]byte, 0)
	calldata = append(calldata, selector...)

	// Encode fillDeadline (uint32) - 4 bytes
	deadlineBytes := make([]byte, 4)
	big.NewInt(int64(fillDeadline)).FillBytes(deadlineBytes)
	calldata = append(calldata, deadlineBytes...)

	// Encode orderDataType (bytes32) - 32 bytes
	calldata = append(calldata, orderDataType[:]...)

	// Encode orderData (bytes) - offset + length + data
	// Offset to the data (32 bytes)
	offset := big.NewInt(68) // 4 + 32 + 32
	offsetBytes := make([]byte, 32)
	offset.FillBytes(offsetBytes)
	calldata = append(calldata, offsetBytes...)

	// Data length (32 bytes)
	dataLen := big.NewInt(int64(len(orderData)))
	dataLenBytes := make([]byte, 32)
	dataLen.FillBytes(dataLenBytes)
	calldata = append(calldata, dataLenBytes...)

	// The actual data
	calldata = append(calldata, orderData...)

	return calldata
}

func sendOpenTransaction(client *ethclient.Client, auth *bind.TransactOpts, contractAddr common.Address, calldata []byte) (*types.Transaction, error) {
	// Get current nonce
	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	// Create transaction
    tx := types.NewTransaction(
		nonce,
		contractAddr,
		big.NewInt(0),  // No ETH value
        uint64(2_000_000), // Gas limit - bumped for debugging
		auth.GasPrice,
		calldata,
	)

	// Sign transaction
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %v", err)
	}

	return signedTx, nil
}

func calculateOrderId(orderData OrderData) string {
	// Match Solidity: keccak256(abi.encode(order))
	encoded := encodeOrderData(orderData)
	hash := crypto.Keccak256Hash(encoded)
	return hash.Hex()
}
