use core::hash::{HashStateExTrait, HashStateTrait};
use core::poseidon::PoseidonTrait;
use oif_starknet::libraries::utils::selector;
use openzeppelin_utils::cryptography::snip12::{SNIP12HashSpanImpl, StructHash};

/// Example witness

#[derive(Drop)]
pub struct MyWitness {
    pub a: u128,
    pub b: Beta,
    pub z: Zeta,
}

#[derive(Drop)]
pub struct Beta {
    pub b1: u128,
    pub b2: Span<felt252>,
}

#[derive(Drop)]
pub struct Zeta {
    pub z1: u8,
    pub z2: Span<felt252>,
}

/// Creating the witness type string

// Part of the type string for the MyWitness struct
// NOTE: `_MY_WITNESS_TYPE_STRING()` is the full witness type string
pub fn _MY_WITNESS_TYPE_STRING_PARTIAL() -> ByteArray {
    "\"My Witness\"(\"A\":\"u128\",\"B\":\"Beta\",\"Z\":\"Zeta\")"
}

// Beta & Zeta type strings
pub fn _BETA_TYPE_STRING() -> ByteArray {
    "\"Beta\"(\"B 1\":\"u128\",\"B 2\":\"felt*\")"
}
pub fn _ZETA_TYPE_STRING() -> ByteArray {
    "\"Zeta\"(\"Z 1\":\"u8\",\"Z 2\":\"felt*\")"
}

// Other type strings needed to create the witness type string (u256, TokenPermissions)
pub fn _U256_TYPE_STRING() -> ByteArray {
    "\"u256\"(\"low\":\"u128\",\"high\":\"u128\")"
}
pub fn _TOKEN_PERMISSIONS_TYPE_STRING() -> ByteArray {
    "\"Token Permissions\"(\"Token\":\"ContractAddress\",\"Amount\":\"u256\")"
}


// witness is struct hash of witness (so it is not called in the message hash function, as it is
// performed off chain)

// The outcome of this function is the full witness type string for the `MyWitness` struct
// NOTE:
// - The witness type string must include the TokenPermissions & u256 type strings after the witness
// type definition
// - If the witness type includes any reference types, they must be sorted
// alphabetically with TokenPermissions & u256
pub fn _MY_WITNESS_TYPE_STRING() -> ByteArray {
    format!(
        "\"witness\":\"My Witness\"){}{}{}{}{}",
        _MY_WITNESS_TYPE_STRING_PARTIAL(),
        _BETA_TYPE_STRING(),
        _TOKEN_PERMISSIONS_TYPE_STRING(),
        _U256_TYPE_STRING(),
        _ZETA_TYPE_STRING(),
    )
}


/// NOTE: MIGHT NOT BE NECESSARY

// Get struct hash for Span<felt252>
impl StructHashSpanFelt252 of StructHash<Span<felt252>> {
    fn hash_struct(self: @Span<felt252>) -> felt252 {
        let mut state = PoseidonTrait::new();
        for el in (*self) {
            state = state.update_with(*el);
        }
        state.finalize()
    }
}
// Get struct hash for Beta
pub impl BetaStructHash of StructHash<Beta> {
    fn hash_struct(self: @Beta) -> felt252 {
        PoseidonTrait::new()
            .update_with(selector(_BETA_TYPE_STRING()))
            .update_with(*self.b1)
            .update_with(self.b2.hash_struct())
            .finalize()
    }
}
// Get struct hash for Zeta
pub impl ZetaStructHash of StructHash<Zeta> {
    fn hash_struct(self: @Zeta) -> felt252 {
        PoseidonTrait::new()
            .update_with(selector(_ZETA_TYPE_STRING()))
            .update_with(*self.z1)
            .update_with(self.z2.hash_struct())
            .finalize()
    }
}
// Get struct hash for MyWitness
pub impl MyWitnessStructHash of StructHash<MyWitness> {
    fn hash_struct(self: @MyWitness) -> felt252 {
        PoseidonTrait::new()
            .update_with(selector(_MY_WITNESS_TYPE_STRING()))
            .update_with(*self.a)
            .update_with(self.b.hash_struct())
            .update_with(self.z.hash_struct())
            .finalize()
    }
}
