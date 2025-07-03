use core::hash::{HashStateExTrait, HashStateTrait};
use core::poseidon::PoseidonTrait;
use oif_starknet::libraries::utils::selector;
use openzeppelin_utils::cryptography::snip12::{
    OffchainMessageHash, SNIP12HashSpanImpl, SNIP12Metadata, StarknetDomain, StructHash,
};
use starknet::{ContractAddress, get_caller_address, get_contract_address, get_tx_info};

// Example witness
#[derive(Drop)]
pub struct ExampleWitness {
    pub a: u128,
    pub b: Beta,
    pub z: Zeta,
}

#[derive(Drop)]
pub struct Beta {
    pub bb: u128,
    pub bbb: Span<felt252>,
}

#[derive(Drop)]
pub struct Zeta {
    pub zz: u8,
    pub zzz: Span<felt252>,
}

// ExampleWitness (partial) type string
// NOTE: This is a sub-string of the full witness type string, see `_EXAMPLE_WITNESS_TYPE_STRING()`
// for full witness type string
pub fn _EXAMPLE_WITNESS_TYPE_STRING_PARTIAL() -> ByteArray {
    "\"ExampleWitness\"(\"a\":\"u128\",\"b\":\"Beta\",\"z\":\"Zeta\")"
}

// Beta type string
pub fn _BETA_TYPE_STRING() -> ByteArray {
    "\"Beta\"(\"bb\":\"u128\",\"bbb\":\"felt*\")"
}

// Zeta type string
pub fn _ZETA_TYPE_STRING() -> ByteArray {
    "\"Zeta\"(\"zz\":\"u8\",\"zzz\":\"felt*\")"
}

// Example of the u256 type string
pub fn _U256_TYPE_STRING() -> ByteArray {
    "\"u256\"(\"low\":\"u128\",\"high\":\"u128\")"
}

// Example o the TokenPermissions (partial) type string
// Normally the u256 type string is attached to the end of this type string,
// but because snip12/eip712 requires alphabetical sorting for all reference type strings, we
// break it up to use as a single sub-string in the below example.
pub fn _TOKEN_PERMISSIONS_TYPE_STRING_PARTIAL() -> ByteArray {
    "\"TokenPermissions\"(\"token\":\"ContractAddress\",\"amount\":\"u256\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")"
}

// NOTE: Here is an example of a witness type string using the above `ExampleWitness` struct.
// The output of the function is the full witness type string to be passed into the contracts.
// - The witness type string must begin with:
// ```"witness":"YourWitnessTypeName")"YourWitnessTypeName"(...<witness type fields>...)```
// - The witness type string must include the TokenPermissions & u256 type strings after the witness
// type definition
// - If the witness type includes any reference types, they must be sorted
// alphabetically with TokenPermissions & u256
pub fn _EXAMPLE_WITNESS_TYPE_STRING() -> ByteArray {
    format!(
        "\"witness\":\"ExampleWitness\"){}{}{}{}{}",
        _EXAMPLE_WITNESS_TYPE_STRING_PARTIAL(),
        _BETA_TYPE_STRING(),
        _TOKEN_PERMISSIONS_TYPE_STRING_PARTIAL(),
        _U256_TYPE_STRING(),
        _ZETA_TYPE_STRING(),
    )
}

impl StructHashSpanFelt252 of StructHash<Span<felt252>> {
    fn hash_struct(self: @Span<felt252>) -> felt252 {
        let mut state = PoseidonTrait::new();
        for el in (*self) {
            state = state.update_with(*el);
        }
        state.finalize()
    }
}

pub impl BetaStructHash of StructHash<Beta> {
    fn hash_struct(self: @Beta) -> felt252 {
        PoseidonTrait::new()
            .update_with(selector(_BETA_TYPE_STRING()))
            .update_with(*self.bb)
            .update_with(self.bbb.hash_struct())
            .finalize()
    }
}

pub impl ZetaStructHash of StructHash<Zeta> {
    fn hash_struct(self: @Zeta) -> felt252 {
        PoseidonTrait::new()
            .update_with(selector(_ZETA_TYPE_STRING()))
            .update_with(*self.zz)
            .update_with(self.zzz.hash_struct())
            .finalize()
    }
}

pub impl ExampleWitnessStructHash of StructHash<ExampleWitness> {
    fn hash_struct(self: @ExampleWitness) -> felt252 {
        PoseidonTrait::new()
            .update_with(selector(_EXAMPLE_WITNESS_TYPE_STRING()))
            .update_with(*self.a)
            .update_with(self.b.hash_struct())
            .update_with(self.z.hash_struct())
            .finalize()
    }
}
