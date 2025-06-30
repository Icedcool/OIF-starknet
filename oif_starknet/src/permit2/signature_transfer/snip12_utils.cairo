use core::hash::{HashStateExTrait, HashStateTrait};
use core::keccak::compute_keccak_byte_array;
use core::poseidon::PoseidonTrait;
use oif_starknet::libraries::utils::selector;
use oif_starknet::permit2::signature_transfer::interface::{
    PermitBatchTransferFrom, PermitTransferFrom, TokenPermissions,
};
use openzeppelin_utils::cryptography::snip12::{
    SNIP12HashSpanImpl, StarknetDomain, StructHash, StructHashStarknetDomainImpl,
};
use starknet::{ContractAddress, get_tx_info};

/// Utils (move for re-use) ///

// @dev need to setup domain config/init/etc
pub fn DOMAIN() -> StarknetDomain {
    StarknetDomain {
        name: 'dApp name', version: '1', chain_id: get_tx_info().unbox().chain_id, revision: 1,
    }
}

/// SNIP-12 TYPE_HASHES ///
/// - u256
/// - TokenPermissions
/// - PermitTransferFrom
/// - PermitBatchTransferFrom
/// - PermitWitnessTransferFrom
/// - PermitWitnessBatchTransferFrom

pub const U256_TYPE_HASH: felt252 = selector!("\"u256\"(\"low\":\"u128\",\"high\":\"u128\")");

pub const TOKEN_PERMISSIONS_TYPEHASH: felt252 = selector!(
    "\"TokenPermissions\"(\"token\":\"ContractAddress\",\"amount\":\"u256\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")",
);

pub const PERMIT_TRANSFER_FROM_TYPEHASH: felt252 = selector!(
    "\"PermitTransferFrom\"(\"permitted\":\"TokenPermissions\",\"spender\":\"ContractAddress\",\"nonce\":\"felt\",\"deadline\":\"u256\")\"TokenPermissions\"(\"token\":\"ContractAddress\",\"amount\":\"u256\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")",
);

pub const PERMIT_BATCH_TRANSFER_FROM_TYPEHASH: felt252 = selector!(
    "\"PermitBatchTransferFrom\"(\"permitted\":\"TokenPermissions*\",\"spender\":\"ContractAddress\",\"nonce\":\"felt\",\"deadline\":\"u256\")\"TokenPermissions\"(\"token\":\"ContractAddress\",\"amount\":\"u256\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")",
);

pub fn _PERMIT_WITNESS_TRANSFER_FROM_TYPEHASH_STUB() -> ByteArray {
    "\"PermitWitnessTransferFrom\"(\"permitted\":\"TokenPermissions\",\"spender\":\"ContractAddress\",\"nonce\":\"felt\",\"deadline\":\"u256\")"
}

pub fn _PERMIT_WITNESS_BATCH_TRANSFER_FROM_TYPEHASH_STUB() -> ByteArray {
    "\"PermitWitnessTransferFrom\"(\"permitted\":\"TokenPermissions*\",\"spender\":\"ContractAddress\",\"nonce\":\"felt\",\"deadline\":\"u256\")"
}

pub fn _PERMIT_WITNESS_TRANSFER_FROM_TYPEHASH(witness_type_string: ByteArray) -> felt252 {
    let stub = _PERMIT_WITNESS_TRANSFER_FROM_TYPEHASH_STUB();
    selector(format!("{stub}{witness_type_string}"))
}

pub fn _PERMIT_BATCH_WITNESS_TRANSFER_FROM_TYPEHASH(witness_type_string: ByteArray) -> felt252 {
    let stub = _PERMIT_WITNESS_BATCH_TRANSFER_FROM_TYPEHASH_STUB();
    selector(format!("{stub}{witness_type_string}"))
}

/// HASHING STRUCTS ///
/// - u256
/// - TokenPermissions
/// - PermitTransferFrom
/// - PermitBatchTransferFrom
/// - PermitWitnessTransferFrom
/// - PermitWitnessBatchTransferFrom

pub impl U256StructHash of StructHash<u256> {
    fn hash_struct(self: @u256) -> felt252 {
        PoseidonTrait::new().update_with(U256_TYPE_HASH).update_with(*self).finalize()
    }
}

pub impl TokenPermissionsStructHash of StructHash<TokenPermissions> {
    fn hash_struct(self: @TokenPermissions) -> felt252 {
        PoseidonTrait::new().update_with(TOKEN_PERMISSIONS_TYPEHASH).update_with(*self).finalize()
    }
}

pub impl StructHashPermitTransferFrom of StructHash<PermitTransferFrom> {
    fn hash_struct(self: @PermitTransferFrom) -> felt252 {
        PoseidonTrait::new()
            .update_with(PERMIT_TRANSFER_FROM_TYPEHASH)
            .update_with(self.permitted.hash_struct())
            .update_with(starknet::get_caller_address())
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .finalize()
    }
}

pub impl StructHashPermitBatchTransferFrom of StructHash<PermitBatchTransferFrom> {
    fn hash_struct(self: @PermitBatchTransferFrom) -> felt252 {
        let hashed_permissions = self
            .permitted
            .into_iter()
            .map(|permission| permission.hash_struct())
            .collect::<Array<felt252>>()
            .span();

        PoseidonTrait::new()
            .update_with(PERMIT_BATCH_TRANSFER_FROM_TYPEHASH)
            .update_with(hashed_permissions)
            .update_with(starknet::get_caller_address())
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .finalize()
    }
}

pub trait StructHashWitnessTrait<T> {
    fn hash_with_witness(self: @T, witness: felt252, witness_type_string: ByteArray) -> felt252;
}

impl StructHashWitnessPermitTransferFrom of StructHashWitnessTrait<PermitTransferFrom> {
    fn hash_with_witness(
        self: @PermitTransferFrom, witness: felt252, witness_type_string: ByteArray,
    ) -> felt252 {
        PoseidonTrait::new()
            .update_with(_PERMIT_WITNESS_TRANSFER_FROM_TYPEHASH(witness_type_string))
            .update_with(self.permitted.hash_struct())
            .update_with(starknet::get_caller_address())
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .update_with(witness)
            .finalize()
    }
}

impl StructHashWitnessPermitBatchTransferFrom of StructHashWitnessTrait<PermitBatchTransferFrom> {
    fn hash_with_witness(
        self: @PermitBatchTransferFrom, witness: felt252, witness_type_string: ByteArray,
    ) -> felt252 {
        let hashed_permissions = self
            .permitted
            .into_iter()
            .map(|permission| permission.hash_struct())
            .collect::<Array<felt252>>()
            .span();

        PoseidonTrait::new()
            .update_with(_PERMIT_BATCH_WITNESS_TRANSFER_FROM_TYPEHASH(witness_type_string))
            .update_with(hashed_permissions)
            .update_with(starknet::get_caller_address())
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .update_with(witness)
            .finalize()
    }
}

/// HASHING OFFCHAIN MESSAGES ///
/// - PermitTransferFrom (+ PermitWitnessTransferFrom)
/// - PermitBatchTransferFrom (+ PermitWitnessBatchTransferFrom)
// @dev unused ?

#[derive(Drop, Copy)]
pub struct PermitTransferFromMessage {
    pub owner: ContractAddress,
    pub permitted: TokenPermissions,
    pub spender: ContractAddress,
    pub nonce: felt252,
    pub deadline: u256,
}

#[derive(Drop, Copy)]
pub struct PermitBatchTransferFromMessage {
    pub owner: ContractAddress,
    pub permitted: Span<TokenPermissions>,
    pub spender: ContractAddress,
    pub nonce: felt252,
    pub deadline: u256,
}

pub impl OffChainMessageHashPermitTransferFrom of StructHash<PermitTransferFromMessage> {
    fn hash_struct(self: @PermitTransferFromMessage) -> felt252 {
        PoseidonTrait::new()
            // Domain
            .update_with('StarkNetDomain')
            .update_with(DOMAIN())
            // Account
            .update_with(*self.owner)
            // Message
            .update_with(TOKEN_PERMISSIONS_TYPEHASH)
            .update_with(self.permitted.hash_struct())
            .update_with(*self.spender)
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .finalize()
    }
}

pub impl OffChainMessageHashPermitBatchTransferFrom of StructHash<PermitBatchTransferFromMessage> {
    fn hash_struct(self: @PermitBatchTransferFromMessage) -> felt252 {
        let hashed_permissions = self
            .permitted
            .into_iter()
            .map(|permission| permission.hash_struct())
            .collect::<Array<felt252>>()
            .span();

        PoseidonTrait::new()
            // Domain
            .update_with('StarkNetDomain')
            .update_with(DOMAIN())
            // Account
            .update_with(*self.owner)
            // Message
            .update_with(PERMIT_BATCH_TRANSFER_FROM_TYPEHASH)
            .update_with(hashed_permissions)
            .update_with(*self.spender)
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .finalize()
    }
}

pub impl OffChainMessageHashWitnessPermitTransferFrom of StructHashWitnessTrait<
    PermitTransferFromMessage,
> {
    fn hash_with_witness(
        self: @PermitTransferFromMessage, witness: felt252, witness_type_string: ByteArray,
    ) -> felt252 {
        PoseidonTrait::new()
            // Domain
            .update_with('StarkNetDomain')
            .update_with(DOMAIN())
            // Account
            .update_with(*self.owner)
            // Message
            .update_with(_PERMIT_WITNESS_TRANSFER_FROM_TYPEHASH(witness_type_string))
            .update_with(self.permitted.hash_struct())
            .update_with(*self.spender)
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .update_with(witness)
            .finalize()
    }
}

pub impl OffChainMessageHashWitnessPermitBatchTransferFrom of StructHashWitnessTrait<
    PermitBatchTransferFromMessage,
> {
    fn hash_with_witness(
        self: @PermitBatchTransferFromMessage, witness: felt252, witness_type_string: ByteArray,
    ) -> felt252 {
        let hashed_permissions = self
            .permitted
            .into_iter()
            .map(|permission| permission.hash_struct())
            .collect::<Array<felt252>>()
            .span();

        PoseidonTrait::new()
            // Domain
            .update_with('StarkNetDomain')
            .update_with(DOMAIN())
            // Account
            .update_with(*self.owner)
            // Message
            .update_with(_PERMIT_BATCH_WITNESS_TRANSFER_FROM_TYPEHASH(witness_type_string))
            .update_with(hashed_permissions)
            .update_with(*self.spender)
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .update_with(witness)
            .finalize()
    }
}
