package main

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/joho/godotenv"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
)

// Registers EVM routers and gas configs on Starknet Hyperlane using owner account

func main() {
	_ = godotenv.Load()
	networkName := "Starknet"
	netCfg, err := config.GetNetworkConfig(networkName)
	if err != nil {
		panic(err)
	}

	// Owner creds
	ownerAddr := mustEnv("STARKNET_DEPLOYER_ADDRESS")
	ownerPub := mustEnv("STARKNET_DEPLOYER_PUBLIC_KEY")
	ownerPriv := mustEnv("STARKNET_DEPLOYER_PRIVATE_KEY")

	// Starknet provider/account
	provider, err := rpc.NewProvider(netCfg.RPCURL)
	if err != nil {
		panic(err)
	}
	ownerAddrF, _ := utils.HexToFelt(ownerAddr)
	ks := account.NewMemKeystore()
	privBI, ok := new(big.Int).SetString(ownerPriv, 0)
	if !ok {
		panic("invalid STARKNET_DEPLOYER_PRIVATE_KEY")
	}
	ks.Put(ownerPub, privBI)
	acct, err := account.NewAccount(provider, ownerAddrF, ownerPub, ks, account.CairoV2)
	if err != nil {
		panic(err)
	}

	// Load Starknet Hyperlane address from deployment-state
	ds, err := deployer.GetDeploymentState()
	if err != nil {
		panic(err)
	}
	sn := ds.Networks[networkName]
	if sn.HyperlaneAddress == "" {
		panic("missing Starknet Hyperlane address in state")
	}
	hlAddrF, _ := utils.HexToFelt(sn.HyperlaneAddress)

	// Build arrays of EVM destinations and routers (bytes32 addresses)
	type evmEntry struct {
		domain uint32
		b32    [32]byte
	}
	var entries []evmEntry
	for name, cfg := range config.Networks {
		if name == networkName {
			continue
		}
		// Destination domain
		entries = append(entries, evmEntry{domain: uint32(cfg.HyperlaneDomain), b32: evmAddrToBytes32(cfg.HyperlaneAddress.Bytes())})
	}

	// Encode Arrays per Cairo ABI: len + elements
	calldata := make([]*felt.Felt, 0, 1+len(entries)+1+len(entries)*2)
	// destinations: Array<u32>
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(entries))))
	for _, e := range entries {
		calldata = append(calldata, utils.Uint64ToFelt(uint64(e.domain)))
	}
	// routers: Array<u256> (each as low, high felts)
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(entries))))
	for _, e := range entries {
		low, high := bytes32ToU256Felts(e.b32)
		calldata = append(calldata, low, high)
	}

	// enroll_remote_routers(uint32[] destinations, u256[] routers)
	call := rpc.InvokeFunctionCall{ContractAddress: hlAddrF, FunctionName: "enroll_remote_routers", CallData: calldata}
	tx, err := acct.BuildAndSendInvokeTxn(context.Background(), []rpc.InvokeFunctionCall{call}, nil)
	if err != nil {
		panic(fmt.Errorf("enroll_remote_routers failed: %w", err))
	}
	fmt.Printf("   ⛽ enroll_remote_routers tx: %s\n", tx.Hash.String())
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env: " + k)
	}
	return v
}

func feltFromBytes(b []byte) *felt.Felt {
	bi := new(big.Int).SetBytes(b)
	return utils.BigIntToFelt(bi)
}

func evmAddrToBytes32(addr20 []byte) (out [32]byte) { copy(out[12:], addr20); return }

func bytes32ToU256Felts(b32 [32]byte) (*felt.Felt, *felt.Felt) {
	// Split into high(16 bytes) and low(16 bytes)
	high := new(big.Int).SetBytes(b32[0:16])
	low := new(big.Int).SetBytes(b32[16:32])
	return utils.BigIntToFelt(low), utils.BigIntToFelt(high)
}
