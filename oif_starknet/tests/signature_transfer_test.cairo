use oif_starknet::libraries::bitmap::BitmapPackingTrait;
use oif_starknet::libraries::permit_hash::{
    OffchainMessageHashWitnessTrait, PermitBatchStructHash, PermitSingleStructHash,
    StructHashPermitBatchTransferFrom, StructHashPermitTransferFrom,
    StructHashWitnessPermitBatchTransferFrom, StructHashWitnessPermitTransferFrom,
    TokenPermissionsStructHash,
};
use oif_starknet::mocks::mock_erc20::{IMintableDispatcher, IMintableDispatcherTrait};
use oif_starknet::mocks::mock_types::{Beta, ExampleWitness, Zeta, _EXAMPLE_WITNESS_TYPE_STRING};
use oif_starknet::permit2::permit2::Permit2::SNIP12MetadataImpl;
use oif_starknet::permit2::signature_transfer::interface::{
    ISignatureTransferDispatcherTrait, PermitBatchTransferFrom, PermitTransferFrom,
    SignatureTransferDetails, TokenPermissions,
};
use openzeppelin_token::erc20::interface::IERC20DispatcherTrait;
use openzeppelin_utils::cryptography::snip12::OffchainMessageHash;
use snforge_std::signature::SignerTrait;
use snforge_std::signature::stark_curve::StarkCurveSignerImpl;
use snforge_std::{
    EventSpyAssertionsTrait, spy_events, start_cheat_block_timestamp, start_cheat_caller_address,
    start_cheat_caller_address_global, stop_cheat_block_timestamp, stop_cheat_caller_address,
    stop_cheat_caller_address_global,
};
use starknet::get_block_timestamp;
use crate::common::E18;
use crate::setup::setupST as setup;


pub const DEFAULT_AMOUNT: u256 = E18;


#[test]
fn test_permit_transfer_from() {
    let setup = setup();
    let nonce = 0;
    let token_permission = TokenPermissions {
        token: setup.token0.contract_address, amount: 10 * E18,
    };
    let permit = PermitTransferFrom {
        permitted: token_permission, nonce, deadline: (get_block_timestamp() + 100).into(),
    };
    let transfer_details = SignatureTransferDetails {
        to: setup.to.account.contract_address, requested_amount: DEFAULT_AMOUNT,
    };

    // Hashing uses the caller's address, so we must mock it here
    start_cheat_caller_address_global(setup.bystander);
    let message_hash = permit.get_message_hash(setup.from.account.contract_address);
    stop_cheat_caller_address_global();
    // Sign the message hash
    let (r, s) = setup.from.key_pair.sign(message_hash).unwrap();
    let signature = array![r, s];

    let start_balance_from = setup.token0.balance_of(setup.from.account.contract_address);
    let start_balance_to = setup.token0.balance_of(setup.to.account.contract_address);

    // Bystander calls `permit_transfer_from`
    start_cheat_caller_address(setup.permit2.contract_address, setup.bystander);
    setup
        .permit2
        .permit_transfer_from(
            permit, transfer_details, setup.from.account.contract_address, signature,
        );
    stop_cheat_caller_address(setup.permit2.contract_address);

    let end_balance_from = setup.token0.balance_of(setup.from.account.contract_address);
    let end_balance_to = setup.token0.balance_of(setup.to.account.contract_address);

    assert_eq!(end_balance_from, start_balance_from - DEFAULT_AMOUNT);
    assert_eq!(end_balance_to, start_balance_to + DEFAULT_AMOUNT);
}

#[test]
fn test_permit_batch_transfer_from() {
    let setup = setup();
    let nonce = 0;
    let tokens = array![setup.token0.contract_address, setup.token1.contract_address];
    let token_permissions: Span<TokenPermissions> = tokens
        .clone()
        .into_iter()
        .map(|token| TokenPermissions { token, amount: 10 * E18 })
        .collect::<Array<_>>()
        .span();
    let permit = PermitBatchTransferFrom {
        permitted: token_permissions, nonce, deadline: (get_block_timestamp() + 100).into(),
    };
    let transfer_details = array![
        SignatureTransferDetails {
            to: setup.to.account.contract_address, requested_amount: DEFAULT_AMOUNT,
        },
        SignatureTransferDetails {
            to: setup.to.account.contract_address, requested_amount: DEFAULT_AMOUNT,
        },
    ]
        .span();

    // Hashing uses the caller's address, so we must mock it here
    start_cheat_caller_address_global(setup.bystander);
    let message_hash = permit.get_message_hash(setup.from.account.contract_address);
    stop_cheat_caller_address_global();
    // Sign the message hash
    let (r, s) = setup.from.key_pair.sign(message_hash).unwrap();
    let signature = array![r, s];

    let start_balance_from0 = setup.token0.balance_of(setup.from.account.contract_address);
    let start_balance_to0 = setup.token0.balance_of(setup.to.account.contract_address);
    let start_balance_from1 = setup.token1.balance_of(setup.from.account.contract_address);
    let start_balance_to1 = setup.token1.balance_of(setup.to.account.contract_address);

    // Bystander calls `permit_batch_transfer_from`
    start_cheat_caller_address(setup.permit2.contract_address, setup.bystander);
    setup
        .permit2
        .permit_batch_transfer_from(
            permit, transfer_details, setup.from.account.contract_address, signature,
        );
    stop_cheat_caller_address(setup.permit2.contract_address);

    let end_balance_from0 = setup.token0.balance_of(setup.from.account.contract_address);
    let end_balance_to0 = setup.token0.balance_of(setup.to.account.contract_address);
    let end_balance_from1 = setup.token1.balance_of(setup.from.account.contract_address);
    let end_balance_to1 = setup.token1.balance_of(setup.to.account.contract_address);

    assert_eq!(end_balance_from0, start_balance_from0 - DEFAULT_AMOUNT);
    assert_eq!(end_balance_to0, start_balance_to0 + DEFAULT_AMOUNT);
    assert_eq!(end_balance_from1, start_balance_from1 - DEFAULT_AMOUNT);
    assert_eq!(end_balance_to1, start_balance_to1 + DEFAULT_AMOUNT);
}


#[test]
#[should_panic(expected: 'Nonce already invalidated')]
fn test_should_panic_permit_transfer_from_invalid_nonce() {
    let setup = setup();
    let nonce = 0;
    let token_permission = TokenPermissions {
        token: setup.token0.contract_address, amount: 10 * E18,
    };
    let transfer_details = SignatureTransferDetails {
        to: setup.to.account.contract_address, requested_amount: DEFAULT_AMOUNT,
    };
    let permit = PermitTransferFrom {
        permitted: token_permission, nonce, deadline: (get_block_timestamp() + 100).into(),
    };

    // Hashing uses the caller's address, so we must mock it here
    start_cheat_caller_address_global(setup.bystander);
    let message_hash = permit.get_message_hash(setup.from.account.contract_address);
    stop_cheat_caller_address_global();
    // Sign the message hash
    let (r, s) = setup.from.key_pair.sign(message_hash).unwrap();
    let signature = array![r, s];

    // Bystander calls `permit_transfer_from`
    start_cheat_caller_address(setup.permit2.contract_address, setup.bystander);
    setup
        .permit2
        .permit_transfer_from(
            permit, transfer_details, setup.from.account.contract_address, signature.clone(),
        );
    // Bystander tries to call `permit_transfer_from` again with the same nonce
    setup
        .permit2
        .permit_transfer_from(
            permit, transfer_details, setup.from.account.contract_address, signature,
        );
    stop_cheat_caller_address(setup.permit2.contract_address);
}

#[test]
#[should_panic(expected: 'Nonce already invalidated')]
fn test_should_panic_permit_batch_transfer_from_invalid_nonce() {
    let setup = setup();
    let nonce = 0;
    let tokens = array![setup.token0.contract_address, setup.token1.contract_address];
    let token_permissions: Span<TokenPermissions> = tokens
        .clone()
        .into_iter()
        .map(|token| TokenPermissions { token, amount: 10 * E18 })
        .collect::<Array<_>>()
        .span();
    let permit = PermitBatchTransferFrom {
        permitted: token_permissions, nonce, deadline: (get_block_timestamp() + 100).into(),
    };
    let transfer_details = array![
        SignatureTransferDetails {
            to: setup.to.account.contract_address, requested_amount: DEFAULT_AMOUNT,
        },
        SignatureTransferDetails {
            to: setup.to.account.contract_address, requested_amount: DEFAULT_AMOUNT,
        },
    ]
        .span();

    // Hashing uses the caller's address, so we must mock it here
    start_cheat_caller_address_global(setup.bystander);
    let message_hash = permit.get_message_hash(setup.from.account.contract_address);
    stop_cheat_caller_address_global();
    // Sign the message hash
    let (r, s) = setup.from.key_pair.sign(message_hash).unwrap();
    let signature = array![r, s];

    // Bystander calls `permit_transfer_from`
    start_cheat_caller_address(setup.permit2.contract_address, setup.bystander);
    setup
        .permit2
        .permit_batch_transfer_from(
            permit, transfer_details, setup.from.account.contract_address, signature.clone(),
        );
    // Bystander tries to call `permit_transfer_from` again with the same nonce
    setup
        .permit2
        .permit_batch_transfer_from(
            permit, transfer_details, setup.from.account.contract_address, signature,
        );
    stop_cheat_caller_address(setup.permit2.contract_address);
}

#[test]
#[fuzzer]
fn test_permit_transfer_from_random_nonce_and_amount(mut nonce: felt252, mut amount: u256) {
    let setup = setup();
    let (nonce_space, bit_pos) = BitmapPackingTrait::unpack_nonce(nonce);
    // Limit nonce to only valid bit_pos's
    nonce = BitmapPackingTrait::pack_nonce(nonce_space, bit_pos % 251);
    // Limit amount to <= 1000 * E18
    //amount = amount % (99 * E18);
    let token_permission = TokenPermissions { token: setup.token0.contract_address, amount };
    let permit = PermitTransferFrom {
        permitted: token_permission, nonce, deadline: (get_block_timestamp() + 100).into(),
    };
    let transfer_details = SignatureTransferDetails {
        to: setup.to.account.contract_address, requested_amount: amount,
    };

    // Hashing uses the caller's address, so we must mock it here
    start_cheat_caller_address_global(setup.bystander);
    let message_hash = permit.get_message_hash(setup.from.account.contract_address);
    stop_cheat_caller_address_global();
    // Sign the message hash
    let (r, s) = setup.from.key_pair.sign(message_hash).unwrap();
    let signature = array![r, s];

    // Owner tops up the `from` account with tokens
    start_cheat_caller_address(setup.token0.contract_address, setup.owner);
    IMintableDispatcher { contract_address: setup.token0.contract_address }
        .mint(setup.from.account.contract_address, amount);
    stop_cheat_caller_address(setup.token0.contract_address);

    let start_balance_from = setup.token0.balance_of(setup.from.account.contract_address);
    let start_balance_to = setup.token0.balance_of(setup.to.account.contract_address);

    // Bystander calls `permit_transfer_from`
    start_cheat_caller_address(setup.permit2.contract_address, setup.bystander);
    setup
        .permit2
        .permit_transfer_from(
            permit, transfer_details, setup.from.account.contract_address, signature,
        );
    stop_cheat_caller_address(setup.permit2.contract_address);

    let end_balance_from = setup.token0.balance_of(setup.from.account.contract_address);
    let end_balance_to = setup.token0.balance_of(setup.to.account.contract_address);

    assert_eq!(end_balance_to, start_balance_to + amount);
    assert_eq!(end_balance_from, start_balance_from - amount);
}

#[test]
#[fuzzer]
fn test_permit_transfer_spend_less_than_full(mut nonce: felt252, amount: u256) {
    let setup = setup();
    let (nonce_space, bit_pos) = BitmapPackingTrait::unpack_nonce(nonce);
    nonce = BitmapPackingTrait::pack_nonce(nonce_space, bit_pos % 251);
    let amount_to_spend = amount / 2;
    let token_permission = TokenPermissions { token: setup.token0.contract_address, amount };
    let permit = PermitTransferFrom {
        permitted: token_permission, nonce, deadline: (get_block_timestamp() + 100).into(),
    };
    let transfer_details = SignatureTransferDetails {
        to: setup.to.account.contract_address, requested_amount: amount_to_spend,
    };

    // Hashing uses the caller's address, so we must mock it here
    start_cheat_caller_address_global(setup.bystander);
    let message_hash = permit.get_message_hash(setup.from.account.contract_address);
    stop_cheat_caller_address_global();
    // Sign the message hash
    let (r, s) = setup.from.key_pair.sign(message_hash).unwrap();
    let signature = array![r, s];

    // Owner tops up the `from` account with tokens
    start_cheat_caller_address(setup.token0.contract_address, setup.owner);
    IMintableDispatcher { contract_address: setup.token0.contract_address }
        .mint(setup.from.account.contract_address, amount);
    stop_cheat_caller_address(setup.token0.contract_address);

    let start_balance_from = setup.token0.balance_of(setup.from.account.contract_address);
    let start_balance_to = setup.token0.balance_of(setup.to.account.contract_address);

    // Bystander calls `permit_transfer_from`
    start_cheat_caller_address(setup.permit2.contract_address, setup.bystander);
    setup
        .permit2
        .permit_transfer_from(
            permit, transfer_details, setup.from.account.contract_address, signature,
        );
    stop_cheat_caller_address(setup.permit2.contract_address);

    let end_balance_from = setup.token0.balance_of(setup.from.account.contract_address);
    let end_balance_to = setup.token0.balance_of(setup.to.account.contract_address);
    assert_eq!(end_balance_from, start_balance_from - amount_to_spend);
    assert_eq!(end_balance_to, start_balance_to + amount_to_spend);
}

#[test]
fn test_permit_batch_tranfer_from_multi_permit_single_transfer() {
    let setup = setup();

    let nonce = 0;
    let tokens = array![setup.token0.contract_address, setup.token1.contract_address];
    let token_permissions = tokens
        .clone()
        .into_iter()
        .map(|token| TokenPermissions { token, amount: DEFAULT_AMOUNT })
        .collect::<Array<_>>()
        .span();
    let transfer_details = array![
        // Transfer 0 tokens
        SignatureTransferDetails { to: setup.to.account.contract_address, requested_amount: 0 },
        // Transer some tokens
        SignatureTransferDetails {
            to: setup.to.account.contract_address, requested_amount: DEFAULT_AMOUNT,
        },
    ]
        .span();
    let permit = PermitBatchTransferFrom {
        permitted: token_permissions, nonce, deadline: (get_block_timestamp() + 100).into(),
    };

    // Hashing uses the caller's address, so we must mock it here
    start_cheat_caller_address_global(setup.bystander);
    let message_hash = permit.get_message_hash(setup.from.account.contract_address);
    stop_cheat_caller_address_global();
    // Sign the message hash
    let (r, s) = setup.from.key_pair.sign(message_hash).unwrap();
    let signature = array![r, s];

    let start_balance_from0 = setup.token0.balance_of(setup.from.account.contract_address);
    let start_balance_to0 = setup.token0.balance_of(setup.to.account.contract_address);
    let start_balance_from1 = setup.token1.balance_of(setup.from.account.contract_address);
    let start_balance_to1 = setup.token1.balance_of(setup.to.account.contract_address);

    // Bystander calls `permit_batch_transfer_from`
    start_cheat_caller_address(setup.permit2.contract_address, setup.bystander);
    setup
        .permit2
        .permit_batch_transfer_from(
            permit, transfer_details, setup.from.account.contract_address, signature,
        );
    stop_cheat_caller_address(setup.permit2.contract_address);

    let end_balance_from0 = setup.token0.balance_of(setup.from.account.contract_address);
    let end_balance_to0 = setup.token0.balance_of(setup.to.account.contract_address);
    let end_balance_from1 = setup.token1.balance_of(setup.from.account.contract_address);
    let end_balance_to1 = setup.token1.balance_of(setup.to.account.contract_address);

    assert_eq!(end_balance_from0, start_balance_from0);
    assert_eq!(end_balance_to0, start_balance_to0);
    assert_eq!(end_balance_from1, start_balance_from1 - DEFAULT_AMOUNT);
    assert_eq!(end_balance_to1, start_balance_to1 + DEFAULT_AMOUNT);
}

#[test]
fn test_permit_witness_transfer_from() {
    let setup = setup();
    let nonce = 0;
    let token_permission = TokenPermissions {
        token: setup.token0.contract_address, amount: 10 * E18,
    };
    let transfer_details = SignatureTransferDetails {
        to: setup.to.account.contract_address, requested_amount: DEFAULT_AMOUNT,
    };
    let permit = PermitTransferFrom {
        permitted: token_permission, nonce, deadline: (get_block_timestamp() + 100).into(),
    };

    // Create a witness
    let witness = ExampleWitness {
        a: 1, b: Beta { bb: 2, bbb: array![].span() }, z: Zeta { zz: 3, zzz: array![].span() },
    }
        .hash_struct();

    // Hashing uses the caller's address, so we must mock it here
    start_cheat_caller_address_global(setup.bystander);
    let message_hash = permit
        .get_message_hash_with_witness(
            setup.from.account.contract_address, witness, _EXAMPLE_WITNESS_TYPE_STRING(),
        );
    stop_cheat_caller_address_global();
    // Sign the message hash
    let (r, s) = setup.from.key_pair.sign(message_hash).unwrap();
    let signature = array![r, s];

    let start_balance_from = setup.token0.balance_of(setup.from.account.contract_address);
    let start_balance_to = setup.token0.balance_of(setup.to.account.contract_address);

    // Bystander calls `permit_transfer_from`
    start_cheat_caller_address(setup.permit2.contract_address, setup.bystander);
    setup
        .permit2
        .permit_witness_transfer_from(
            permit,
            transfer_details,
            setup.from.account.contract_address,
            witness,
            _EXAMPLE_WITNESS_TYPE_STRING(),
            signature,
        );
    stop_cheat_caller_address(setup.permit2.contract_address);
    let end_balance_from = setup.token0.balance_of(setup.from.account.contract_address);
    let end_balance_to = setup.token0.balance_of(setup.to.account.contract_address);

    assert_eq!(end_balance_from, start_balance_from - DEFAULT_AMOUNT);
    assert_eq!(end_balance_to, start_balance_to + DEFAULT_AMOUNT);
}

#[test]
fn test_permit_witness_batch_transfer_from() {}


// LEFT OFF HERE:
// test invalid spender (single, batch, witness, witness batch) fails
// test invalid witness (hash, type string) fails (single, batch)
// test struct hashes/message hashes match starknet.js values (use go for this ?) (do this after all
// other tests are checked/added ?)
// ask chat gpt for test inspiration, no need to write them for me yet
// move on to allowance transfer tests

#[test]
#[ignore]
fn test_correct_witness_type_hashes() {
    assert(true, '');
}

