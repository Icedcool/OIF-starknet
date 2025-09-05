package ethutil

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestERC20ABIAdditional tests additional ERC20 ABI validation
func TestERC20ABIAdditional(t *testing.T) {
	t.Run("ERC20ABI_methods_have_correct_signatures", func(t *testing.T) {
		abi, err := ParseABI(ERC20ABI)
		require.NoError(t, err)

		// Test balanceOf signature
		balanceOf := abi.Methods["balanceOf"]
		assert.Equal(t, "balanceOf", balanceOf.Name)
		assert.Equal(t, "view", balanceOf.StateMutability)
		assert.Len(t, balanceOf.Inputs, 1)
		assert.Len(t, balanceOf.Outputs, 1)

		// Test allowance signature
		allowance := abi.Methods["allowance"]
		assert.Equal(t, "allowance", allowance.Name)
		assert.Equal(t, "view", allowance.StateMutability)
		assert.Len(t, allowance.Inputs, 2)
		assert.Len(t, allowance.Outputs, 1)

		// Test transfer signature
		transfer := abi.Methods["transfer"]
		assert.Equal(t, "transfer", transfer.Name)
		assert.Equal(t, "nonpayable", transfer.StateMutability)
		assert.Len(t, transfer.Inputs, 2)
		assert.Len(t, transfer.Outputs, 1)

		// Test approve signature
		approve := abi.Methods["approve"]
		assert.Equal(t, "approve", approve.Name)
		assert.Equal(t, "nonpayable", approve.StateMutability)
		assert.Len(t, approve.Inputs, 2)
		assert.Len(t, approve.Outputs, 1)
	})
}

// TestParseABI tests the ABI parsing function
func TestParseABI(t *testing.T) {
	t.Run("valid_abi", func(t *testing.T) {
		validABI := `[{"inputs":[],"name":"test","outputs":[],"type":"function"}]`
		abi, err := ParseABI(validABI)
		require.NoError(t, err)
		assert.NotNil(t, abi)
	})

	t.Run("invalid_abi", func(t *testing.T) {
		invalidABI := `{"invalid": "json"}`
		_, err := ParseABI(invalidABI)
		assert.Error(t, err)
	})

	t.Run("empty_abi", func(t *testing.T) {
		_, err := ParseABI("")
		assert.Error(t, err)
	})

	t.Run("malformed_json", func(t *testing.T) {
		malformedJSON := `[{"inputs":[],"name":"test","outputs":[],"type":"function"` // Missing closing bracket
		_, err := ParseABI(malformedJSON)
		assert.Error(t, err)
	})
}

// TestAddressOperationsAdditional tests additional address-related utility functions
func TestAddressOperationsAdditional(t *testing.T) {
	t.Run("address_validation", func(t *testing.T) {
		validAddress := "0x1234567890123456789012345678901234567890"
		invalidAddress := "invalid-address"

		// Test valid address
		addr := common.HexToAddress(validAddress)
		assert.NotEqual(t, common.Address{}, addr)

		// Test invalid address
		addr = common.HexToAddress(invalidAddress)
		assert.Equal(t, common.Address{}, addr)
	})

	t.Run("address_comparison", func(t *testing.T) {
		addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
		addr2 := common.HexToAddress("0x1234567890123456789012345678901234567890")
		addr3 := common.HexToAddress("0x0987654321098765432109876543210987654321")

		assert.Equal(t, addr1, addr2)
		assert.NotEqual(t, addr1, addr3)
	})

	t.Run("address_string_conversion", func(t *testing.T) {
		addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
		addrStr := addr.Hex()
		assert.Equal(t, "0x1234567890123456789012345678901234567890", addrStr)
	})
}

// TestBigIntOperationsAdditional tests additional big.Int utility functions
func TestBigIntOperationsAdditional(t *testing.T) {
	t.Run("big_int_creation", func(t *testing.T) {
		// Test from int64
		val1 := big.NewInt(12345)
		assert.Equal(t, int64(12345), val1.Int64())

		// Test from string
		val2, ok := new(big.Int).SetString("12345678901234567890", 10)
		require.True(t, ok)
		assert.Equal(t, "12345678901234567890", val2.String())

		// Test from hex string
		val3, ok := new(big.Int).SetString("0x1234567890abcdef", 0)
		require.True(t, ok)
		assert.Equal(t, "1234567890abcdef", val3.Text(16))
	})

	t.Run("big_int_arithmetic", func(t *testing.T) {
		a := big.NewInt(100)
		b := big.NewInt(50)

		// Addition
		sum := new(big.Int).Add(a, b)
		assert.Equal(t, int64(150), sum.Int64())

		// Subtraction
		diff := new(big.Int).Sub(a, b)
		assert.Equal(t, int64(50), diff.Int64())

		// Multiplication
		product := new(big.Int).Mul(a, b)
		assert.Equal(t, int64(5000), product.Int64())

		// Division
		quotient := new(big.Int).Div(a, b)
		assert.Equal(t, int64(2), quotient.Int64())
	})

	t.Run("big_int_comparison", func(t *testing.T) {
		a := big.NewInt(100)
		b := big.NewInt(50)
		c := big.NewInt(100)

		assert.True(t, a.Cmp(b) > 0)  // a > b
		assert.True(t, b.Cmp(a) < 0)  // b < a
		assert.True(t, a.Cmp(c) == 0) // a == c
	})
}

// TestFunctionSignaturesAdditional tests additional function signature generation
func TestFunctionSignaturesAdditional(t *testing.T) {
	t.Run("function_signature_generation", func(t *testing.T) {
		// Test common ERC20 function signatures
		balanceOfSig := crypto.Keccak256Hash([]byte("balanceOf(address)"))
		allowanceSig := crypto.Keccak256Hash([]byte("allowance(address,address)"))
		transferSig := crypto.Keccak256Hash([]byte("transfer(address,uint256)"))
		approveSig := crypto.Keccak256Hash([]byte("approve(address,uint256)"))

		// These should be valid 32-byte hashes
		assert.Len(t, balanceOfSig.Bytes(), 32)
		assert.Len(t, allowanceSig.Bytes(), 32)
		assert.Len(t, transferSig.Bytes(), 32)
		assert.Len(t, approveSig.Bytes(), 32)

		// They should all be different
		assert.NotEqual(t, balanceOfSig, allowanceSig)
		assert.NotEqual(t, balanceOfSig, transferSig)
		assert.NotEqual(t, balanceOfSig, approveSig)
		assert.NotEqual(t, allowanceSig, transferSig)
		assert.NotEqual(t, allowanceSig, approveSig)
		assert.NotEqual(t, transferSig, approveSig)
	})

	t.Run("function_signature_consistency", func(t *testing.T) {
		// Test that the same function signature is generated consistently
		sig1 := crypto.Keccak256Hash([]byte("balanceOf(address)"))
		sig2 := crypto.Keccak256Hash([]byte("balanceOf(address)"))
		assert.Equal(t, sig1, sig2)
	})
}

// TestErrorHandlingAdditional tests additional error handling scenarios
func TestErrorHandlingAdditional(t *testing.T) {
	t.Run("error_wrapping", func(t *testing.T) {
		originalErr := assert.AnError
		wrappedErr := fmt.Errorf("failed to process: %w", originalErr)

		assert.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "failed to process")
		assert.Contains(t, wrappedErr.Error(), originalErr.Error())
	})

	t.Run("error_chain", func(t *testing.T) {
		// Test error unwrapping
		originalErr := assert.AnError
		wrappedErr := fmt.Errorf("level 2: %w", originalErr)
		doubleWrappedErr := fmt.Errorf("level 1: %w", wrappedErr)

		assert.Error(t, doubleWrappedErr)

		// Test error unwrapping
		unwrapped := doubleWrappedErr
		for i := 0; i < 2; i++ {
			unwrapped = errors.Unwrap(unwrapped)
		}
		assert.Equal(t, originalErr, unwrapped)
	})
}

// TestPrivateKeyOperationsAdditional tests additional private key operations
func TestPrivateKeyOperationsAdditional(t *testing.T) {
	t.Run("private_key_generation", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		assert.NotNil(t, privateKey)
	})

	t.Run("private_key_to_public_key", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		require.True(t, ok)
		assert.NotNil(t, publicKeyECDSA)
	})

	t.Run("private_key_to_address", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		require.True(t, ok)

		address := crypto.PubkeyToAddress(*publicKeyECDSA)
		assert.NotEqual(t, common.Address{}, address)
	})

	t.Run("private_key_serialization", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		// Test private key to bytes
		privateKeyBytes := crypto.FromECDSA(privateKey)
		assert.Len(t, privateKeyBytes, 32)

		// Test bytes to private key
		recoveredPrivateKey, err := crypto.ToECDSA(privateKeyBytes)
		require.NoError(t, err)
		assert.Equal(t, privateKey.D, recoveredPrivateKey.D)
	})
}

// TestUtilityFunctionsAdditional tests additional utility functions
func TestUtilityFunctionsAdditional(t *testing.T) {
	t.Run("string_utilities", func(t *testing.T) {
		// Test string trimming
		str := "  hello world  "
		trimmed := strings.TrimSpace(str)
		assert.Equal(t, "hello world", trimmed)

		// Test string prefix checking
		assert.True(t, strings.HasPrefix(str, "  "))
		assert.True(t, strings.HasSuffix(str, "  "))

		// Test string replacement
		replaced := strings.ReplaceAll(str, " ", "_")
		assert.Equal(t, "__hello_world__", replaced)
	})

	t.Run("number_utilities", func(t *testing.T) {
		// Test big.Int utilities
		val := big.NewInt(1000000000000000000) // 1 ETH in wei
		assert.True(t, val.Cmp(big.NewInt(0)) > 0)
		assert.True(t, val.Cmp(big.NewInt(1000000000000000001)) < 0)

		// Test conversion to string
		valStr := val.String()
		assert.Equal(t, "1000000000000000000", valStr)
	})
}

// Helper function to parse ABI (if not already defined)
func ParseABI(abiStr string) (*abi.ABI, error) {
	parsedABI, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, err
	}
	return &parsedABI, nil
}
