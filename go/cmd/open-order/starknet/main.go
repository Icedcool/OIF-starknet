package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/joho/godotenv"
)

// Network configuration - will be loaded from deployment state
var networks []NetworkConfig

// NetworkConfig represents a single network configuration
type NetworkConfig struct {
	name             string
	url              string
	chainID          uint64
	hyperlaneAddress string
	orcaCoinAddress  string
	dogCoinAddress   string
}

// loadNetworks loads network configuration from deployment state
func loadNetworks() error {
	// For now, hardcode the Starknet network and load token addresses from files
	// TODO: Integrate with centralized config system
	networks = []NetworkConfig{
		{
			name:             "Starknet Sepolia",
			url:              "http://localhost:5050",
			chainID:          23448591,
			hyperlaneAddress: "",
			orcaCoinAddress:  "",
			dogCoinAddress:   "",
		},
	}

	// Load token addresses from deployment files
	for i, network := range networks {
		if network.name == "Starknet Sepolia" {
			// Load Hyperlane7683 address
			if hyperlaneAddr, err := loadHyperlaneAddress(); err == nil {
				networks[i].hyperlaneAddress = hyperlaneAddr
				fmt.Printf("   🔍 Loaded %s Hyperlane7683: %s\n", network.name, hyperlaneAddr)
			}

			// Load token addresses
			if tokens, err := loadTokenAddresses(); err == nil {
				for _, token := range tokens {
					if token.Name == "OrcaCoin" {
						networks[i].orcaCoinAddress = token.Address
						fmt.Printf("   🔍 Loaded %s OrcaCoin: %s\n", network.name, token.Address)
					} else if token.Name == "DogCoin" {
						networks[i].dogCoinAddress = token.Address
						fmt.Printf("   🔍 Loaded %s DogCoin: %s\n", network.name, token.Address)
					}
				}
			}
		}
	}

	return nil
}

// loadHyperlaneAddress loads the Hyperlane7683 address from deployment file
func loadHyperlaneAddress() (string, error) {
	// Try multiple possible paths
	paths := []string{
		"state/network_state/starknet-sepolia-deployment.json",
		"../state/network_state/starknet-sepolia-deployment.json",
		"../../state/network_state/starknet-sepolia-deployment.json",
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			fmt.Printf("   🔍 Loaded Hyperlane address from: %s\n", path)
			var deployment struct {
				DeployedAddress string `json:"deployedAddress"`
			}
			if err := json.Unmarshal(data, &deployment); err != nil {
				continue
			}
			return deployment.DeployedAddress, nil
		}
	}

	return "", fmt.Errorf("could not find Hyperlane deployment file in any of the expected paths")
}

// loadTokenAddresses loads token addresses from deployment file
func loadTokenAddresses() ([]TokenInfo, error) {
	// Try multiple possible paths
	paths := []string{
		"state/network_state/starknet-sepolia-mock-erc20-deployment.json",
		"../state/network_state/starknet-sepolia-mock-erc20-deployment.json",
		"../../state/network_state/starknet-sepolia-mock-erc20-deployment.json",
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			fmt.Printf("   🔍 Loaded token addresses from: %s\n", path)
			var deployment struct {
				Tokens []TokenInfo `json:"tokens"`
			}
			if err := json.Unmarshal(data, &deployment); err != nil {
				continue
			}
			return deployment.Tokens, nil
		}
	}

	return nil, fmt.Errorf("could not find token deployment file in any of the expected paths")
}

// TokenInfo represents token deployment information
type TokenInfo struct {
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Address   string `json:"address"`
	ClassHash string `json:"classHash"`
}

// Test user configuration
var testUsers = []struct {
	name       string
	privateKey string
	address    string
}{
	{"Alice", "SN_ALICE_PRIVATE_KEY", "0x13d9ee239f33fea4f8785b9e3870ade909e20a9599ae7cd62c1c292b73af1b7"},
	{"Bob", "SN_BOB_PRIVATE_KEY", "0x17cc6ca902ed4e8baa8463a7009ff18cc294fa85a94b4ce6ac30a9ebd6057c7"},
	{"Solver", "SN_SOLVER_PRIVATE_KEY", "0x2af9427c5a277474c079a1283c880ee8a6f0f8fbf73ce969c08d88befec1bba"},
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
	OpenDeadline     uint64 // Changed from uint32 to uint64
	FillDeadline     uint64 // Changed from uint32 to uint64
}

// OrderData struct matching the Cairo OrderData
type OrderData struct {
	Sender             *felt.Felt
	Recipient          *felt.Felt
	InputToken         *felt.Felt
	OutputToken        *felt.Felt
	AmountIn           *big.Int // Changed from *felt.Felt to *big.Int for u256 splitting
	AmountOut          *big.Int // Changed from *felt.Felt to *big.Int for u256 splitting
	SenderNonce        *felt.Felt
	OriginDomain       uint32
	DestinationDomain  uint32
	DestinationSettler *felt.Felt
	OpenDeadline       uint64       // Added missing field
	FillDeadline       uint64       // Changed from uint32 to uint64
	Data               []*felt.Felt // Empty for basic orders
}

// OnchainCrossChainOrder struct matching the Cairo interface
type OnchainCrossChainOrder struct {
	FillDeadline  uint64 // Changed from uint32 to uint64
	OrderDataType *felt.Felt
	OrderData     []*felt.Felt
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables")
	}

	// Load network configuration
	if err := loadNetworks(); err != nil {
		fmt.Printf("❌ Failed to load networks: %v\n", err)
		os.Exit(1)
	}

	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: open-starknet-order <command>")
		fmt.Println("Commands:")
		fmt.Println("  basic     - Open a basic hardcoded test order")
		fmt.Println("  random    - Open a randomly generated order")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "basic":
		openBasicOrder()
	case "random":
		openRandomOrder()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func openBasicOrder() {
	fmt.Println("🚀 Opening Basic Starknet Test Order...")

	// Hardcoded basic order from Starknet to EVM
	order := OrderConfig{
		OriginChain:      "Starknet Sepolia",
		DestinationChain: "Base Sepolia", // Example EVM destination
		InputToken:       "OrcaCoin",
		OutputToken:      "DogCoin",
		InputAmount:      new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 1000 tokens
		OutputAmount:     new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 1000 tokens
		User:             "Alice",
		OpenDeadline:     uint64(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint64(time.Now().Add(24 * time.Hour).Unix()),
	}

	executeOrder(order)
}

func openRandomOrder() {
	fmt.Println("🎲 Opening Random Starknet Test Order...")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// For now, always use Starknet as origin since this is the Starknet order tool
	originChain := "Starknet Sepolia"

	// Random destination (could be any EVM chain)
	evmDestinations := []string{"Base Sepolia", "Optimism Sepolia", "Arbitrum Sepolia"}
	destIdx := rand.Intn(len(evmDestinations))
	destinationChain := evmDestinations[destIdx]

	// Random user
	userIdx := rand.Intn(len(testUsers))

	// Random amounts (100-10000 tokens)
	inputAmount := rand.Intn(9901) + 100  // 100-10000
	outputAmount := rand.Intn(9901) + 100 // 100-10000

	order := OrderConfig{
		OriginChain:      originChain,
		DestinationChain: destinationChain,
		InputToken:       "OrcaCoin",
		OutputToken:      "DogCoin",
		InputAmount:      new(big.Int).Mul(big.NewInt(int64(inputAmount)), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		OutputAmount:     new(big.Int).Mul(big.NewInt(int64(outputAmount)), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		User:             testUsers[userIdx].name,
		OpenDeadline:     uint64(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint64(time.Now().Add(24 * time.Hour).Unix()),
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

	// Find origin network (should be Starknet)
	var originNetwork *NetworkConfig
	for _, network := range networks {
		if network.name == order.OriginChain {
			originNetwork = &network
			break
		}
	}

	if originNetwork == nil {
		fmt.Printf("❌ Origin network not found: %s\n", order.OriginChain)
		os.Exit(1)
	}

	// Connect to Starknet RPC
	client, err := rpc.NewProvider(originNetwork.url)
	if err != nil {
		fmt.Printf("❌ Failed to connect to %s: %v\n", order.OriginChain, err)
		os.Exit(1)
	}

	// Get user private key and public key
	userKey := os.Getenv(fmt.Sprintf("SN_%s_PRIVATE_KEY", strings.ToUpper(order.User)))
	userPublicKey := os.Getenv(fmt.Sprintf("SN_%s_PUBLIC_KEY", strings.ToUpper(order.User)))
	if userKey == "" || userPublicKey == "" {
		fmt.Printf("❌ Missing credentials for user: %s\n", order.User)
		os.Exit(1)
	}

	// Get user address
	var userAddr string
	for _, user := range testUsers {
		if user.name == order.User {
			userAddr = user.address
			break
		}
	}

	fmt.Printf("   🔗 Connected to %s (Chain ID: %d)\n", order.OriginChain, originNetwork.chainID)
	fmt.Printf("   👤 User: %s (%s)\n", order.User, userAddr)

	// For now, use hardcoded domains (these should come from config)
	originDomain := uint32(23448591)      // Starknet Sepolia domain
	destinationDomain := uint32(11155420) // Base Sepolia domain

	fmt.Printf("   🔎 Origin Domain: %d\n", originDomain)
	fmt.Printf("   🔎 Destination Domain: %d\n", destinationDomain)

	// Preflight: check balances and allowances
	inputToken := originNetwork.orcaCoinAddress
	owner := userAddr
	spender := originNetwork.hyperlaneAddress

	// Get initial balances
	initialUserBalance, err := getTokenBalanceFromRPC(client, inputToken, owner, "OrcaCoin")
	if err == nil {
		fmt.Printf("   🔍 Initial InputToken balance(owner): %s\n", formatTokenAmount(initialUserBalance))
	} else {
		fmt.Printf("   ⚠️  Could not read initial balance: %v\n", err)
	}

	// Check allowance
	allowance, err := getTokenAllowanceFromRPC(client, inputToken, owner, spender, "OrcaCoin")
	if err == nil {
		fmt.Printf("   🔍 InputToken allowance(owner→hyperlane): %s\n", formatTokenAmount(allowance))
	} else {
		fmt.Printf("   ⚠️  Could not read allowance: %v\n", err)
	}

	// Store initial balance for comparison
	initialBalance := initialUserBalance

	// Generate a random nonce for the order
	senderNonce := big.NewInt(time.Now().UnixNano())

	// Build the order data
	orderData := buildOrderData(order, originNetwork, destinationDomain, senderNonce)

		// Build the OnchainCrossChainOrder
	crossChainOrder := OnchainCrossChainOrder{
		FillDeadline:  order.FillDeadline,
		OrderDataType: getOrderDataTypeHash(),
		OrderData:     encodeOrderData(orderData),
	}

		// Use generated bindings for open()
	fmt.Printf("   📝 Calling open() function...\n")

	// Create user account for transaction signing
	userAddrFelt, err := utils.HexToFelt(userAddr)
	if err != nil {
		fmt.Printf("❌ Failed to convert user address to felt: %v\n", err)
		os.Exit(1)
	}

	// Initialize user's keystore
	userKs := account.NewMemKeystore()
	userPrivKeyBI, ok := new(big.Int).SetString(userKey, 0)
	if !ok {
		fmt.Printf("❌ Failed to convert private key for %s: %v\n", order.User, err)
		os.Exit(1)
	}
	userKs.Put(userPublicKey, userPrivKeyBI)

	// Create user account (Cairo v2)
	userAccnt, err := account.NewAccount(client, userAddrFelt, userPublicKey, userKs, account.CairoV2)
	if err != nil {
		fmt.Printf("❌ Failed to create account for %s: %v\n", order.User, err)
		os.Exit(1)
	}

	// Get Hyperlane7683 contract address
	hyperlaneAddrFelt, err := utils.HexToFelt(originNetwork.hyperlaneAddress)
	if err != nil {
		fmt.Printf("❌ Failed to convert Hyperlane7683 address to felt: %v\n", err)
		os.Exit(1)
	}

		fmt.Println("\n📝 Now attempting to open the order...")

		// Call the open function
	fmt.Printf("   📝 Sending open transaction...\n")
	
	// Build the transaction
	// The open function takes: OnchainCrossChainOrder { fill_deadline: u64, order_data_type: felt252, order_data: Bytes }
	// Cairo automatically serializes the struct, so we just pass the fields in order
	calldata := []*felt.Felt{
		// OnchainCrossChainOrder fields in order:
		utils.Uint64ToFelt(crossChainOrder.FillDeadline), // fill_deadline: u64
		crossChainOrder.OrderDataType,                    // order_data_type: felt252
		// order_data: Bytes - send the Bytes struct directly
	}
	// Add the orderData (which is already a Bytes struct with size, data_len, and data array)
	calldata = append(calldata, crossChainOrder.OrderData...)

	tx, err := userAccnt.BuildAndSendInvokeTxn(
		context.Background(),
		[]rpc.InvokeFunctionCall{{
			ContractAddress: hyperlaneAddrFelt, // Call on Hyperlane7683 contract, not user account
			FunctionName:    "open",
			CallData:        calldata,
		}},
		nil,
	)
	if err != nil {
		fmt.Printf("❌ Failed to send open transaction: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   🚀 Transaction sent: %s\n", tx.Hash.String())
	fmt.Printf("   ⏳ Waiting for confirmation...\n")

	// Wait for transaction receipt
	_, err = userAccnt.WaitForTransactionReceipt(context.Background(), tx.Hash, time.Second)
	if err != nil {
		fmt.Printf("❌ Failed to wait for transaction confirmation: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   ✅ Order opened successfully!\n")
	fmt.Printf("   🎯 Order ID: %s\n", calculateOrderId(orderData))

	// Verify that balances actually changed as expected
	fmt.Printf("   🔍 Verifying balance changes...\n")
	if err := verifyBalanceChangesFromRPC(client, inputToken, owner, spender, initialBalance, order.InputAmount); err != nil {
		fmt.Printf("   ⚠️  Balance verification failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Balance changes verified!\n")
	}

	fmt.Printf("\n🎉 Order execution completed!\n")
}

func buildOrderData(order OrderConfig, originNetwork *NetworkConfig, destinationDomain uint32, senderNonce *big.Int) OrderData {
	// Get the actual user address for the specified user
	var userAddr string
	for _, user := range testUsers {
		if user.name == order.User {
			userAddr = user.address
			break
		}
	}

	// Convert addresses to felt
	userAddrFelt, _ := utils.HexToFelt(userAddr)
	inputTokenFelt, _ := utils.HexToFelt(originNetwork.orcaCoinAddress)
	outputTokenFelt, _ := utils.HexToFelt(originNetwork.dogCoinAddress)
	destSettlerFelt, _ := utils.HexToFelt(originNetwork.hyperlaneAddress)

	// Convert amounts to big.Int (they're already big.Int in the order config)
	// senderNonceFelt := utils.BigIntToFelt(senderNonce)

	return OrderData{
		Sender:             userAddrFelt,
		Recipient:          userAddrFelt,
		InputToken:         inputTokenFelt,
		OutputToken:        outputTokenFelt,
		AmountIn:           order.InputAmount,  // Already *big.Int
		AmountOut:          order.OutputAmount, // Already *big.Int
		SenderNonce:        utils.BigIntToFelt(senderNonce),
		OriginDomain:       uint32(23448591), // Starknet Sepolia domain
		DestinationDomain:  destinationDomain,
		DestinationSettler: destSettlerFelt,
		OpenDeadline:       order.OpenDeadline,
		FillDeadline:       order.FillDeadline,
		Data:               []*felt.Felt{}, // Empty for basic orders
	}
}

func getOrderDataTypeHash() *felt.Felt {
	// This should match the Cairo contract's OrderEncoder::ORDER_DATA_TYPE_HASH
	// This tells the contract how to parse the orderData using OrderEncoder::decode

	// The actual hash from the Cairo contract's OrderEncoder
	// This matches the type string: "Order Data"(...)
	hash, err := utils.HexToFelt("0x3ED8862ABBF6BBE28E01F529E75203031B5A7475E38592F6BDAC6469409A7E8")
	if err != nil {
		// This should never fail for a hardcoded hex string, but handle it gracefully
		panic(fmt.Sprintf("Failed to convert type hash: %v", err))
	}
	return hash
}

func encodeOrderData(orderData OrderData) []*felt.Felt {
	// Encode as Cairo Bytes packed big-endian stream following OrderEncoder::encode

	// helpers
	leftPad := func(src []byte, size int) []byte {
		if len(src) >= size {
			return src[len(src)-size:]
		}
		out := make([]byte, size)
		copy(out[size-len(src):], src)
		return out
	}

	feltToBytes32 := func(f *felt.Felt) []byte {
		b := f.Bytes() // [32]byte
		return b[:]
	}

	bigIntToBytes32 := func(n *big.Int) []byte {
		if n == nil {
			return make([]byte, 32)
		}
		return leftPad(n.Bytes(), 32)
	}

	u32ToBytes := func(v uint32) []byte {
		b := make([]byte, 4)
		b[0] = byte(v >> 24)
		b[1] = byte(v >> 16)
		b[2] = byte(v >> 8)
		b[3] = byte(v)
		return b
	}

	u64ToBytes := func(v uint64) []byte {
		b := make([]byte, 8)
		b[0] = byte(v >> 56)
		b[1] = byte(v >> 48)
		b[2] = byte(v >> 40)
		b[3] = byte(v >> 32)
		b[4] = byte(v >> 24)
		b[5] = byte(v >> 16)
		b[6] = byte(v >> 8)
		b[7] = byte(v)
		return b
	}

	// build raw packed bytes
	raw := make([]byte, 0, 272)
	// sender
	raw = append(raw, feltToBytes32(orderData.Sender)...)
	// recipient
	raw = append(raw, feltToBytes32(orderData.Recipient)...)
	// input_token
	raw = append(raw, feltToBytes32(orderData.InputToken)...)
	// output_token
	raw = append(raw, feltToBytes32(orderData.OutputToken)...)
	// amount_in
	raw = append(raw, bigIntToBytes32(orderData.AmountIn)...)
	// amount_out
	raw = append(raw, bigIntToBytes32(orderData.AmountOut)...)
	// sender_nonce
	raw = append(raw, feltToBytes32(orderData.SenderNonce)...)
	// origin_domain (u32)
	raw = append(raw, u32ToBytes(orderData.OriginDomain)...)
	// destination_domain (u32)
	raw = append(raw, u32ToBytes(orderData.DestinationDomain)...)
	// destination_settler
	raw = append(raw, feltToBytes32(orderData.DestinationSettler)...)
	// fill_deadline (u64)
	raw = append(raw, u64ToBytes(orderData.FillDeadline)...)
	// data (empty)

	// split into u128 words (16 bytes)
	numElements := (len(raw) + 15) / 16
	words := make([]*felt.Felt, 0, numElements)
	for i := 0; i < len(raw); i += 16 {
		end := i + 16
		if end > len(raw) {
			chunk := make([]byte, 16)
			copy(chunk, raw[i:])
			words = append(words, utils.BigIntToFelt(new(big.Int).SetBytes(chunk)))
		} else {
			words = append(words, utils.BigIntToFelt(new(big.Int).SetBytes(raw[i:end])))
		}
	}

	// Bytes serialization: size, data length, then words
	bytesStruct := make([]*felt.Felt, 0, 2+len(words))
	bytesStruct = append(bytesStruct, utils.Uint64ToFelt(uint64(len(raw))))
	bytesStruct = append(bytesStruct, utils.Uint64ToFelt(uint64(len(words))))
	bytesStruct = append(bytesStruct, words...)

	return bytesStruct
}

// toU256 converts a big.Int to low and high felt values for u256 representation
func toU256(num *big.Int) (low, high *felt.Felt) {
	// Create a mask for the lower 128 bits
	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))

	// Extract low and high parts
	lowBigInt := new(big.Int).And(num, mask)
	highBigInt := new(big.Int).Rsh(num, 128)

	// Convert to felt
	lowFelt := utils.BigIntToFelt(lowBigInt)
	highFelt := utils.BigIntToFelt(highBigInt)

	return lowFelt, highFelt
}

func calculateOrderId(orderData OrderData) string {
	// Generate a simple order ID for now
	// In production, this should match the contract's order ID generation
	return fmt.Sprintf("sn_order_%d", time.Now().UnixNano())
}

// getTokenBalance gets the balance of a token for a specific address
func getTokenBalance(accnt *account.Account, tokenAddress, userAddress, tokenName string) (*big.Int, error) {
	// Convert addresses to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	userAddrFelt, err := utils.HexToFelt(userAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid user address: %w", err)
	}

	// Build the balanceOf function call
	balanceCall := rpc.FunctionCall{
		ContractAddress:    tokenAddrFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("balanceOf"),
		Calldata:           []*felt.Felt{userAddrFelt},
	}

	// Call the contract to get balance
	resp, err := accnt.Provider.Call(context.Background(), balanceCall, rpc.WithBlockTag("latest"))
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no response from balanceOf call")
	}

	// Convert felt response to big.Int
	balanceFelt := resp[0]
	balanceBigInt := utils.FeltToBigInt(balanceFelt)

	return balanceBigInt, nil
}

// getTokenAllowance gets the allowance of a token for a specific spender
func getTokenAllowance(accnt *account.Account, tokenAddress, ownerAddress, spenderAddress, tokenName string) (*big.Int, error) {
	// Convert addresses to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	ownerAddrFelt, err := utils.HexToFelt(ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid owner address: %w", err)
	}

	spenderAddrFelt, err := utils.HexToFelt(spenderAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid spender address: %w", err)
	}

	// Build the allowance function call
	allowanceCall := rpc.FunctionCall{
		ContractAddress:    tokenAddrFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("allowance"),
		Calldata:           []*felt.Felt{ownerAddrFelt, spenderAddrFelt},
	}

	// Call the contract to get allowance
	resp, err := accnt.Provider.Call(context.Background(), allowanceCall, rpc.WithBlockTag("latest"))
	if err != nil {
		return nil, fmt.Errorf("invalid owner address: %w", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no response from allowance call")
	}

	// Convert felt response to big.Int
	allowanceFelt := resp[0]
	allowanceBigInt := utils.FeltToBigInt(allowanceFelt)

	return allowanceBigInt, nil
}

// getTokenBalanceFromRPC gets the balance of a token for a specific address using RPC
func getTokenBalanceFromRPC(client rpc.RpcProvider, tokenAddress, userAddress, tokenName string) (*big.Int, error) {
	// Convert addresses to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	userAddrFelt, err := utils.HexToFelt(userAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid user address: %w", err)
	}

	// Build the balanceOf function call
	balanceCall := rpc.FunctionCall{
		ContractAddress:    tokenAddrFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("balanceOf"),
		Calldata:           []*felt.Felt{userAddrFelt},
	}

	// Call the contract to get balance
	resp, err := client.Call(context.Background(), balanceCall, rpc.WithBlockTag("latest"))
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no response from balanceOf call")
	}

	// Convert felt response to big.Int
	balanceFelt := resp[0]
	balanceBigInt := utils.FeltToBigInt(balanceFelt)

	return balanceBigInt, nil
}

// getTokenAllowanceFromRPC gets the allowance of a token for a specific spender using RPC
func getTokenAllowanceFromRPC(client rpc.RpcProvider, tokenAddress, ownerAddress, spenderAddress, tokenName string) (*big.Int, error) {
	// Convert addresses to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	ownerAddrFelt, err := utils.HexToFelt(ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid owner address: %w", err)
	}

	spenderAddrFelt, err := utils.HexToFelt(spenderAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid spender address: %w", err)
	}

	// Build the allowance function call
	allowanceCall := rpc.FunctionCall{
		ContractAddress:    tokenAddrFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("allowance"),
		Calldata:           []*felt.Felt{ownerAddrFelt, spenderAddrFelt},
	}

	// Call the contract to get allowance
	resp, err := client.Call(context.Background(), allowanceCall, rpc.WithBlockTag("latest"))
	if err != nil {
		return nil, fmt.Errorf("failed to call allowance: %w", err)
	}

	if len(resp) < 2 {
		return nil, fmt.Errorf("insufficient response from allowance call: expected 2 values for u256, got %d", len(resp))
	}

	// For u256, the response should be [low, high] where:
	// - low contains the first 128 bits
	// - high contains the remaining 128 bits
	lowFelt := resp[0]
	highFelt := resp[1]

	// Convert low and high felts to big.Ints
	lowBigInt := utils.FeltToBigInt(lowFelt)
	highBigInt := utils.FeltToBigInt(highFelt)

	// Combine low and high into a single u256 value
	// high << 128 + low
	shiftedHigh := new(big.Int).Lsh(highBigInt, 128)
	totalAllowance := new(big.Int).Add(shiftedHigh, lowBigInt)

	return totalAllowance, nil
}

// verifyBalanceChanges verifies that opening an order actually transferred tokens
func verifyBalanceChanges(accnt *account.Account, tokenAddress, userAddress, hyperlaneAddress string, initialBalance *big.Int, expectedTransferAmount *big.Int) error {
	// Wait a moment for the transaction to be fully processed
	time.Sleep(2 * time.Second)

	// Get final balance
	finalUserBalance, err := getTokenBalance(accnt, tokenAddress, userAddress, "OrcaCoin")
	if err != nil {
		return fmt.Errorf("failed to get final user balance: %w", err)
	}

	// Calculate actual change
	userBalanceChange := new(big.Int).Sub(initialBalance, finalUserBalance)

	// Print balance changes
	fmt.Printf("     💰 User balance change: %s → %s (Δ: %s)\n",
		formatTokenAmount(initialBalance),
		formatTokenAmount(finalUserBalance),
		formatTokenAmount(userBalanceChange))

	// Verify the change matches expectations
	if userBalanceChange.Cmp(expectedTransferAmount) != 0 {
		return fmt.Errorf("user balance decreased by %s, expected %s",
			formatTokenAmount(userBalanceChange),
			formatTokenAmount(expectedTransferAmount))
	}

	return nil
}

// verifyBalanceChangesFromRPC verifies that opening an order actually transferred tokens using RPC
func verifyBalanceChangesFromRPC(client rpc.RpcProvider, tokenAddress, userAddress, hyperlaneAddress string, initialBalance *big.Int, expectedTransferAmount *big.Int) error {
	// Wait a moment for the transaction to be fully processed
	time.Sleep(2 * time.Second)

	// Get final balance
	finalUserBalance, err := getTokenBalanceFromRPC(client, tokenAddress, userAddress, "OrcaCoin")
	if err != nil {
		return fmt.Errorf("failed to get final user balance: %w", err)
	}

	// Calculate actual change
	userBalanceChange := new(big.Int).Sub(initialBalance, finalUserBalance)

	// Print balance changes
	fmt.Printf("     💰 User balance change: %s → %s (Δ: %s)\n",
		formatTokenAmount(initialBalance),
		formatTokenAmount(finalUserBalance),
		formatTokenAmount(userBalanceChange))

	// Verify the change matches expectations
	if userBalanceChange.Cmp(expectedTransferAmount) != 0 {
		return fmt.Errorf("user balance decreased by %s, expected %s",
			formatTokenAmount(userBalanceChange),
			formatTokenAmount(expectedTransferAmount))
	}

	return nil
}

// formatTokenAmount formats a token amount for display (converts from wei to tokens)
func formatTokenAmount(amount *big.Int) string {
	// Convert from wei (18 decimals) to tokens
	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	tokens := new(big.Float).Quo(new(big.Float).SetInt(amount), new(big.Float).SetInt(decimals))
	return tokens.Text('f', 0) + " tokens"
}
