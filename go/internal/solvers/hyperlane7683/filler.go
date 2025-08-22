package hyperlane7683

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	contracts "github.com/NethermindEth/oif-starknet/go/internal/contracts"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/internal/filler"
	"github.com/NethermindEth/oif-starknet/go/internal/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Helper functions for logging
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var UnknownOrderStatus = common.Hash{}
var ErrIntentAlreadyFilled = errors.New("intent already filled")

type Hyperlane7683Filler struct {
	*filler.BaseFillerImpl
	client   *ethclient.Client
	clients  map[uint64]*ethclient.Client
	signers  map[uint64]*bind.TransactOpts
	metadata types.Hyperlane7683Metadata
}

func NewHyperlane7683Filler(client *ethclient.Client) *Hyperlane7683Filler {
	metadata := types.Hyperlane7683Metadata{
		BaseMetadata:  types.BaseMetadata{ProtocolName: "Hyperlane7683"},
		IntentSources: []types.IntentSource{},
		CustomRules:   types.CustomRules{},
	}

	allowBlockLists := types.AllowBlockLists{AllowList: []types.AllowBlockListItem{}, BlockList: []types.AllowBlockListItem{}}

	return &Hyperlane7683Filler{
		BaseFillerImpl: filler.NewBaseFiller(allowBlockLists, metadata),
		client:         client,
		clients:        make(map[uint64]*ethclient.Client),
		signers:        make(map[uint64]*bind.TransactOpts),
		metadata:       metadata,
	}
}

// logDecodedOrderData attempts to abi-like decode origin_data bytes into fields matching OrderData
func logDecodedOrderData(originData []byte) {
	// Expect 13 head words (12 static fields + 1 offset) then tail (bytes)
	// Validate minimum length
	if len(originData) < 32*13 {
		fmt.Printf("   ⚠️  origin_data too short: %d bytes (expected >= %d)\n", len(originData), 32*13)
		return
	}
	getWord := func(i int) []byte { return originData[i*32 : (i+1)*32] }
	h := func(i int) string { return hex.EncodeToString(getWord(i)) }
	// Head layout per solidity struct OrderData
	// 0 sender, 1 recipient, 2 inputToken, 3 outputToken,
	// 4 amountIn, 5 amountOut, 6 senderNonce,
	// 7 originDomain, 8 destinationDomain, 9 destinationSettler,
	// 10 fillDeadline, 11 dataOffset, 12 (start of tail length)
	fmt.Printf("   🔎 Decoded OrderData (head):\n")
	fmt.Printf("     • sender: %s\n", h(0))
	fmt.Printf("     • recipient: %s\n", h(1))
	fmt.Printf("     • inputToken: %s\n", h(2))
	fmt.Printf("     • outputToken: %s\n", h(3))
	fmt.Printf("     • amountIn: %s\n", h(4))
	fmt.Printf("     • amountOut: %s\n", h(5))
	fmt.Printf("     • senderNonce: %s\n", h(6))
	fmt.Printf("     • originDomain: %s\n", h(7))
	fmt.Printf("     • destinationDomain: %s\n", h(8))
	fmt.Printf("     • destinationSettler: %s\n", h(9))
	fmt.Printf("     • fillDeadline: %s\n", h(10))
	fmt.Printf("     • dataOffset: %s\n", h(11))

	// Tail bytes length at offset dataOffset
	// Compute expected offset value
	dataOffset := new(big.Int).SetBytes(getWord(11)).Uint64()
	fmt.Printf("   🔎 Tail info: offset=%d (0x%x)\n", dataOffset, dataOffset)
	if int(dataOffset) > len(originData) {
		fmt.Printf("   ⚠️  dataOffset beyond origin_data length: %d > %d\n", dataOffset, len(originData))
		return
	}
	if len(originData) < int(dataOffset)+32 {
		fmt.Printf("   ⚠️  origin_data too short for tail length at offset: need %d, have %d\n", int(dataOffset)+32, len(originData))
		return
	}
	tailLen := new(big.Int).SetBytes(originData[dataOffset : dataOffset+32]).Uint64()
	fmt.Printf("     • tailLength (bytes.data length): %d\n", tailLen)
	if len(originData) < int(dataOffset+32+tailLen) {
		fmt.Printf("   ⚠️  origin_data too short for full tail: need %d, have %d\n", int(dataOffset+32+tailLen), len(originData))
		return
	}
	tail := originData[dataOffset+32 : dataOffset+32+tailLen]
	fmt.Printf("     • tail (hex): %s\n", hex.EncodeToString(tail))
}

func (f *Hyperlane7683Filler) ProcessIntent(ctx context.Context, args types.ParsedArgs, originChainName string, blockNumber uint64) error {
	fmt.Printf("🔵 Processing Intent: %s-%s on chain %s (block %d)\n", f.metadata.ProtocolName, args.OrderID, originChainName, blockNumber)
	intent, err := f.PrepareIntent(ctx, args)
	if err != nil {
		return err
	}
	if !intent.Success {
		return nil
	}
	if err := f.Fill(ctx, args, intent.Data, originChainName, blockNumber); err != nil {
		return fmt.Errorf("fill execution failed: %w", err)
	}
	if err := f.SettleOrder(ctx, args, intent.Data, originChainName); err != nil {
		return fmt.Errorf("order settlement failed: %w", err)
	}
	return nil
}

func (f *Hyperlane7683Filler) Fill(ctx context.Context, args types.ParsedArgs, data types.IntentData, originChainName string, blockNumber uint64) error {
	fmt.Printf("🔵 Filling Intent: %s-%s on chain %s (block %d)\n", f.metadata.ProtocolName, args.OrderID, originChainName, blockNumber)
	fmt.Printf("   Fill Instructions: %d instructions\n", len(data.FillInstructions))
	fmt.Printf("   Max Spent: %d outputs\n", len(data.MaxSpent))

	for i, instruction := range data.FillInstructions {
		fmt.Printf("   📦 Instruction %d: Chain %s, Settler %s\n", i+1, instruction.DestinationChainID.String(), instruction.DestinationSettler.Hex())

		// Route to Starknet filler if destination is Starknet
		if isStarknetChainID(instruction.DestinationChainID) {
			sf, err := f.getStarknetFillerForChain(instruction.DestinationChainID)
			if err != nil {
				return fmt.Errorf("failed to init Starknet filler for %s: %w", instruction.DestinationChainID.String(), err)
			}
			fmt.Printf("   🧪 OriginData Debug: len=%d hex=%s\n", len(instruction.OriginData), common.Bytes2Hex(instruction.OriginData))
			if err := sf.Fill(ctx, args.OrderID, instruction.OriginData); err != nil {
				return fmt.Errorf("starknet fill failed: %w", err)
			}
			continue
		}

		client, err := f.getClientForChain(instruction.DestinationChainID)
		if err != nil {
			return fmt.Errorf("failed to get client for chain %s: %w", instruction.DestinationChainID.String(), err)
		}
		signer, err := f.getSignerForChain(instruction.DestinationChainID)
		if err != nil {
			return fmt.Errorf("failed to get signer for chain %s: %w", instruction.DestinationChainID.String(), err)
		}

		if i < len(data.MaxSpent) {
			maxSpent := data.MaxSpent[i]
			fmt.Printf("   💰 MaxSpent[%d]: Token=%s, Amount=%s, Recipient=%s, ChainID=%s\n", i, maxSpent.Token.Hex(), maxSpent.Amount.String(), maxSpent.Recipient.Hex(), maxSpent.ChainID.String())
		}

		orderIdBytes := common.FromHex(args.OrderID)
		var orderIdArr [32]byte
		copy(orderIdArr[:], orderIdBytes)

		// Use the original order ID - no reversal needed
		// The order ID should be consistent across both chains
		fmt.Printf("     • orderID: %s\n", hex.EncodeToString(orderIdBytes))

		originDataBytes := instruction.OriginData

		// Verbose decode of origin_data to compare EVM->EVM vs SN->EVM
		logDecodedOrderData(originDataBytes)

		// CRITICAL FIX: The contract expects bytes _fillerData, not address
		// We need to pass filler data as bytes, not as a padded address
		// For now, pass empty bytes as filler data (this matches the working EVM→EVM order)
		var fillerDataBytes []byte
		// TODO: In the future, we might want to encode actual filler data here

		// Use the proper Go bindings like the rest of the codebase
		// The issue is likely in our parameter handling, not the bindings
		fmt.Printf("   🔧 Using proper Go bindings for fill function\n")

		// Get the contract instance
		contract, err := contracts.NewHyperlane7683(instruction.DestinationSettler, client)
		if err != nil {
			return fmt.Errorf("failed to bind contract at %s: %w", instruction.DestinationSettler.Hex(), err)
		}

		// Add comprehensive logging for EVM fills
		fmt.Printf("   🧪 EVMFill Debug\n")
		fmt.Printf("     • orderID: %s\n", args.OrderID)
		fmt.Printf("     • orderID bytes: %s\n", hex.EncodeToString(orderIdBytes))
		fmt.Printf("     • orderID padded: %s\n", hex.EncodeToString(orderIdArr[:]))
		fmt.Printf("     • origin_data_len: %d bytes\n", len(originDataBytes))
		fmt.Printf("     • origin_data_hex: %s\n", hex.EncodeToString(originDataBytes))
		fmt.Printf("     • filler_address: %s\n", signer.From.Hex())
		fmt.Printf("     • filler_data_bytes: %s\n", hex.EncodeToString(fillerDataBytes))

		// Calculate keccak256 hash of origin_data
		hash := crypto.Keccak256(originDataBytes)

		// Reverse endianness to match Cairo's u256_reverse_endian
		reversedHash := make([]byte, 32)
		for i := 0; i < 16; i++ {
			// Swap bytes to reverse endianness (Cairo uses u128_byte_reverse on both low and high parts)
			reversedHash[i] = hash[31-i]
			reversedHash[31-i] = hash[i]
		}

		fmt.Printf("     • keccak256(origin_data): 0x%s\n", hex.EncodeToString(hash))
		fmt.Printf("     • keccak256_reversed_endian: 0x%s\n", hex.EncodeToString(reversedHash))

		fmt.Printf("   🔄 Executing fill call to contract %s on chain %s\n", instruction.DestinationSettler.Hex(), instruction.DestinationChainID.String())

		// CRITICAL: Log the exact calldata being sent to the EVM contract
		fmt.Printf("   📋 Fill Transaction Calldata:\n")
		fmt.Printf("     • orderID (32 bytes): %s\n", hex.EncodeToString(orderIdArr[:]))
		fmt.Printf("     • origin_data (%d bytes): %s\n", len(originDataBytes), hex.EncodeToString(originDataBytes))
		fmt.Printf("     • filler_data (%d bytes): %s\n", len(fillerDataBytes), hex.EncodeToString(fillerDataBytes))

		// Log the exact function call details
		fmt.Printf("   🔍 Function Call Details:\n")
		fmt.Printf("     • Function selector: 0x82e2c43f (fill function)\n")
		fmt.Printf("     • ABI encoding: orderID + origin_data + filler_data\n")
		fmt.Printf("     • Expected total length: %d bytes\n", 32+len(originDataBytes)+len(fillerDataBytes))

		// REMOVED: Manual calldata construction that was interfering with Go bindings
		// The Go bindings will handle proper ABI encoding automatically

		// Debug the orderID array specifically
		fmt.Printf("   🔍 OrderID Array Debug:\n")
		fmt.Printf("     • orderIdArr type: %T, length: %d\n", orderIdArr, len(orderIdArr))
		fmt.Printf("     • orderIdArr hex: %s\n", hex.EncodeToString(orderIdArr[:]))
		fmt.Printf("     • orderIdArr first 4 bytes: %s\n", hex.EncodeToString(orderIdArr[:4]))
		fmt.Printf("     • orderIdArr last 4 bytes: %s\n", hex.EncodeToString(orderIdArr[28:]))

		// Verify the orderID is exactly 32 bytes and properly formatted
		if len(orderIdArr) != 32 {
			fmt.Printf("   🚨 ERROR: orderIdArr length is %d, expected 32!\n", len(orderIdArr))
		}

		// CRITICAL: Log right before the Fill call to catch any data corruption
		fmt.Printf("   🚨 About to call contract.Fill() with:\n")
		fmt.Printf("     • orderIdArr type: %T, length: %d\n", orderIdArr, len(orderIdArr))
		fmt.Printf("     • originDataBytes type: %T, length: %d\n", originDataBytes, len(originDataBytes))
		fmt.Printf("     • fillerDataBytes type: %T, length: %d\n", fillerDataBytes, len(fillerDataBytes))

		// Get gas price for transaction
		if gp, gpErr := client.SuggestGasPrice(ctx); gpErr == nil {
			signer.GasPrice = gp
		}

		fmt.Printf("   🚨 Using proper Go bindings for contract.Fill()...\n")

		// DETAILED FIELD LOGGING: Print each field being passed to contract.Fill()
		fmt.Printf("   📝 Contract.Fill() Parameters:\n")
		fmt.Printf("     • Parameter 1 - signer.From: %s\n", signer.From.Hex())
		fmt.Printf("     • Parameter 2 - orderIdArr ([32]byte): %x\n", orderIdArr)
		fmt.Printf("     • Parameter 3 - originDataBytes ([]byte): length=%d\n", len(originDataBytes))
		fmt.Printf("       └─ First 64 bytes: %x\n", originDataBytes[:min(64, len(originDataBytes))])
		fmt.Printf("       └─ Last 64 bytes: %x\n", originDataBytes[max(0, len(originDataBytes)-64):])
		fmt.Printf("     • Parameter 4 - fillerDataBytes ([]byte): length=%d\n", len(fillerDataBytes))
		if len(fillerDataBytes) > 0 {
			fmt.Printf("       └─ Content: %x\n", fillerDataBytes)
		} else {
			fmt.Printf("       └─ Content: (empty)\n")
		}

		// Additional contract details
		fmt.Printf("   🏗️  Contract Details:\n")
		fmt.Printf("     • Contract Address: %s\n", instruction.DestinationSettler.Hex())
		fmt.Printf("     • Chain ID: %s\n", instruction.DestinationChainID.String())
		fmt.Printf("     • Gas Price: %s wei\n", signer.GasPrice.String())

		// Use the contract.Fill() method like the rest of the codebase
		// This ensures proper ABI encoding and parameter handling
		// Attempt to retrieve revert reason with eth_call using ABI-packed calldata
		//originDataBytes[31] = 0xaa
		tx, err := contract.Fill(signer, orderIdArr, originDataBytes, fillerDataBytes)
		if err != nil {
			// Attempt to retrieve revert reason with eth_call using ABI-packed calldata
			if contracts.Hyperlane7683ABI != "" {
				if parsed, perr := abi.JSON(strings.NewReader(contracts.Hyperlane7683ABI)); perr == nil {
					if callData, packErr := parsed.Pack("fill", orderIdArr, originDataBytes, fillerDataBytes); packErr == nil {
						if _, callErr := client.CallContract(ctx, ethereum.CallMsg{From: signer.From, To: &instruction.DestinationSettler, Data: callData}, nil); callErr != nil {
							fmt.Printf("   ❌ eth_call revert while packing fill: %v\n", callErr)
						}
					}
				}
			}
			return fmt.Errorf("failed to send fill tx: %w, ", err)
		}

		// CRITICAL: Log the actual transaction input data to see what was sent
		fmt.Printf("   📊 Actual Transaction Data Sent:\n")
		fmt.Printf("     • Transaction input data length: %d bytes\n", len(tx.Data()))
		fmt.Printf("     • Transaction input data: %s\n", hex.EncodeToString(tx.Data()))
		fmt.Printf("     • Transaction hash: %s\n", tx.Hash().Hex())
		receipt, err := bind.WaitMined(ctx, client, tx)
		if err != nil {
			return fmt.Errorf("failed to wait for fill confirmation: %w", err)
		}
		if receipt.Status == 0 {
			return fmt.Errorf("fill transaction failed at block %d", receipt.BlockNumber)
		}
		fmt.Printf("   ✅ Fill transaction confirmed at block %d\n", receipt.BlockNumber)

		// Check order status after fill to verify it changed
		if err := f.checkOrderStatusAfterFill(ctx, client, instruction.DestinationSettler, args.OrderID); err != nil {
			fmt.Printf("   ⚠️  Order status check failed: %v\n", err)
		}
	}
	fmt.Printf("   🎉 All fill instructions processed\n")
	return nil
}

// getOrderStatus gets the current status of an order from the contract
func (f *Hyperlane7683Filler) getOrderStatus(ctx context.Context, client *ethclient.Client, settlerAddr common.Address, orderID *big.Int) (string, error) {
	// orderStatus(bytes32) function selector: 0x3d4b4f7f
	statusSelector := []byte{0x3d, 0x4b, 0x4f, 0x7f}

	// Pack the orderID parameter (32 bytes)
	param := make([]byte, 32)
	orderIDBytes := orderID.Bytes()
	copy(param[32-len(orderIDBytes):], orderIDBytes)

	callData := append(statusSelector, param...)
	result, err := client.CallContract(ctx, ethereum.CallMsg{To: &settlerAddr, Data: callData}, nil)
	if err != nil {
		return "", fmt.Errorf("orderStatus call failed: %w", err)
	}

	if len(result) < 32 {
		return "", fmt.Errorf("invalid orderStatus result length: %d", len(result))
	}

	return common.BytesToHash(result).Hex(), nil
}

func (f *Hyperlane7683Filler) SettleOrder(ctx context.Context, args types.ParsedArgs, data types.IntentData, originChainName string) error {
	return nil
}

func (f *Hyperlane7683Filler) AddDefaultRules() {
	f.AddRule(f.filterByTokenAndAmount)
	f.AddRule(f.intentNotFilled)
}

func (f *Hyperlane7683Filler) filterByTokenAndAmount(args types.ParsedArgs, _ *filler.FillerContext) error {
	return nil
}

func (f *Hyperlane7683Filler) intentNotFilled(args types.ParsedArgs, _ *filler.FillerContext) error {
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found")
	}
	first := args.ResolvedOrder.FillInstructions[0]
	if isStarknetChainID(first.DestinationChainID) {
		return nil
	}
	settlerAddr := first.DestinationSettler
	client, err := f.getClientForChain(first.DestinationChainID)
	if err != nil {
		return fmt.Errorf("intentNotFilled: failed to get client for chain %s: %w", first.DestinationChainID.String(), err)
	}
	orderStatusABI := `[{"type":"function","name":"orderStatus","inputs":[{"type":"bytes32","name":"orderId"}],"outputs":[{"type":"bytes32","name":""}],"stateMutability":"view"}]`
	parsedABI, err := abi.JSON(strings.NewReader(orderStatusABI))
	if err != nil {
		return fmt.Errorf("intentNotFilled: failed to parse orderStatus ABI: %w", err)
	}
	var orderIdArr [32]byte
	copy(orderIdArr[:], common.LeftPadBytes(common.FromHex(args.OrderID), 32))
	callData, err := parsedABI.Pack("orderStatus", orderIdArr)
	if err != nil {
		return fmt.Errorf("intentNotFilled: failed to pack orderStatus call: %w", err)
	}
	result, err := client.CallContract(context.Background(), ethereum.CallMsg{To: &settlerAddr, Data: callData}, nil)
	if err != nil {
		return fmt.Errorf("intentNotFilled: failed to call orderStatus on %s: %w", settlerAddr.Hex(), err)
	}
	if len(result) < 32 {
		return fmt.Errorf("intentNotFilled: invalid orderStatus result length=%d", len(result))
	}
	orderStatus := common.BytesToHash(result[:32])
	if orderStatus != UnknownOrderStatus {
		fmt.Printf("   ⏩ Skipping EVM fill: order status=%s (already processed)\n", orderStatus.Hex())
		return ErrIntentAlreadyFilled
	}
	return nil
}

// checkOrderStatusAfterFill verifies that the order status changed after the fill
func (f *Hyperlane7683Filler) checkOrderStatusAfterFill(ctx context.Context, client *ethclient.Client, settlerAddr common.Address, orderIDHex string) error {
	orderStatusABI := `[{"type":"function","name":"orderStatus","inputs":[{"type":"bytes32","name":"orderId"}],"outputs":[{"type":"bytes32","name":""}],"stateMutability":"view"}]`
	parsedABI, err := abi.JSON(strings.NewReader(orderStatusABI))
	if err != nil {
		return fmt.Errorf("failed to parse orderStatus ABI: %w", err)
	}

	var orderIdArr [32]byte
	copy(orderIdArr[:], common.LeftPadBytes(common.FromHex(orderIDHex), 32))
	callData, err := parsedABI.Pack("orderStatus", orderIdArr)
	if err != nil {
		return fmt.Errorf("failed to pack orderStatus call: %w", err)
	}

	result, err := client.CallContract(ctx, ethereum.CallMsg{To: &settlerAddr, Data: callData}, nil)
	if err != nil {
		return fmt.Errorf("failed to call orderStatus: %w", err)
	}
	if len(result) < 32 {
		return fmt.Errorf("invalid orderStatus result length=%d", len(result))
	}

	orderStatus := common.BytesToHash(result[:32])
	fmt.Printf("   🔍 Post-fill order status: %s\n", orderStatus.Hex())

	if orderStatus == UnknownOrderStatus {
		fmt.Printf("   ⚠️  Order status still UNKNOWN after fill - fill may not have worked!\n")
	} else {
		fmt.Printf("   ✅ Order status changed to %s - fill successful!\n", orderStatus.Hex())
	}
	return nil
}

func (f *Hyperlane7683Filler) getClientForChain(chainID *big.Int) (*ethclient.Client, error) {
	chainIDUint := chainID.Uint64()
	if client, ok := f.clients[chainIDUint]; ok {
		return client, nil
	}
	rpcURL, err := config.GetRPCURLByChainID(chainIDUint)
	if err != nil {
		return nil, fmt.Errorf("failed to get RPC URL for chain %d: %w", chainIDUint, err)
	}
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to chain %d at %s: %w", chainIDUint, rpcURL, err)
	}
	f.clients[chainIDUint] = client
	return client, nil
}

func (f *Hyperlane7683Filler) getSignerForChain(chainID *big.Int) (*bind.TransactOpts, error) {
	chainIDUint := chainID.Uint64()
	if signer, ok := f.signers[chainIDUint]; ok {
		return signer, nil
	}
	solverPrivateKey := os.Getenv("SOLVER_PRIVATE_KEY")
	if solverPrivateKey == "" {
		return nil, fmt.Errorf("SOLVER_PRIVATE_KEY environment variable not set")
	}
	pk, err := crypto.HexToECDSA(strings.TrimPrefix(solverPrivateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse solver private key: %w", err)
	}
	from := crypto.PubkeyToAddress(pk.PublicKey)
	signer := bind.NewKeyedTransactor(pk)
	signer.From = from
	f.signers[chainIDUint] = signer
	return signer, nil
}

// Helpers for chain routing
func isStarknetChainID(chainID *big.Int) bool {
	return chainID.Uint64() == config.Networks["Starknet Sepolia"].ChainID
}

func (f *Hyperlane7683Filler) getStarknetFillerForChain(chainID *big.Int) (*StarknetFiller, error) {
	if !isStarknetChainID(chainID) {
		return nil, fmt.Errorf("not a starknet chain: %s", chainID.String())
	}
	rpcURL := config.Networks["Starknet Sepolia"].RPCURL
	// Load Hyperlane address from centralized deployment state
	ds, derr := deployer.GetDeploymentState()
	if derr != nil {
		return nil, fmt.Errorf("failed to load deployment state: %w", derr)
	}
	sn := ds.Networks["Starknet Sepolia"]
	return NewStarknetFiller(rpcURL, sn.HyperlaneAddress)
}
