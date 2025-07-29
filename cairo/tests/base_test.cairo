//#[starknet::contract]
//pub trait IBaseTest<TState> {
//    fn users(self: @TState) -> Array<starknet::ContractAddress>;
//}

#[starknet::contract]
pub mod BaseTest {
    use permit2::libraries::utils::selector;
    use snforge_std::signature::SignerTrait;
    use snforge_std::signature::stark_curve::{
        StarkCurveKeyPairImpl, StarkCurveSignerImpl, StarkCurveVerifierImpl,
    };
    use core::hash::{HashStateExTrait, HashStateTrait};
    use core::poseidon::PoseidonTrait;
    use openzeppelin_utils::cryptography::snip12::{
        SNIP12HashSpanImpl, SNIP12Metadata, StarknetDomain, StructHash,
    };

    use alexandria_bytes::{Bytes, BytesTrait};
    use starknet::{ContractAddress, get_contract_address};
    use permit2::interfaces::{
        signature_transfer::{PermitTransferFrom, PermitBatchTransferFrom, TokenPermissions},
        permit2::{IPermit2Dispatcher, IPermit2DispatcherTrait},
    };
    use permit2::snip12_utils::permits::{TokenPermissionsStructHash, U256StructHash};
    use oif_starknet::erc7683::interface::{
        Open, GaslessCrossChainOrder, OnchainCrossChainOrder, ResolvedCrossChainOrder,
    };
    use oif_starknet::base7683::Base7683Component;
    use openzeppelin_token::erc20::interface::{IERC20Dispatcher, IERC20DispatcherTrait};

    use crate::common::{
        ETH_ADDRESS, pop_log, pop_log_raw, generate_account, deploy_eth, deploy_erc20,
        deploy_permit2, deal, deal_multiple,
    };
    use starknet::storage::{
        Map, StoragePathEntry, StoragePointerReadAccess, StoragePointerWriteAccess,
    };
    use snforge_std::signature::{KeyPair, KeyPairTrait};
    use crate::common::Account;

    use snforge_std::{start_cheat_caller_address, stop_cheat_caller_address, replace_bytecode};

    const E18: u256 = 10_000_000_000_000_000_000;
    const origin: u32 = 1;
    const destination: u32 = 2;
    const amount: u256 = 100;


    #[storage]
    struct Storage {
        DOMAIN_SEPARATOR: felt252,
        permit2: ContractAddress,
        kaka: ContractAddress,
        karp: ContractAddress,
        veg: ContractAddress,
        kaka_key_pair: (felt252, felt252),
        karp_key_pair: (felt252, felt252),
        veg_key_pair: (felt252, felt252),
        counter_part: ContractAddress,
        input_token: ContractAddress, // erc20 dis.
        output_token: ContractAddress,
        balance_id: Map<ContractAddress, u256>,
    }


    #[constructor]
    fn constructor(ref self: ContractState) {
        let kaka = generate_account();
        let karp = generate_account();
        let veg = generate_account();
        let counter_part: ContractAddress = 'counterpart'.try_into().unwrap();

        let eth = deploy_eth();
        let input_token = deploy_erc20("Input Token", "IN");
        let output_token = deploy_erc20("Output Token", "OUT");

        let permit2 = deploy_permit2();
        let DOMAIN_SEPARATOR = IPermit2Dispatcher { contract_address: permit2 }.DOMAIN_SEPARATOR();
        self.DOMAIN_SEPARATOR.write(DOMAIN_SEPARATOR);

        deal_multiple(
            array![
                input_token.contract_address,
                output_token.contract_address,
                eth.contract_address,
                eth.contract_address,
            ],
            array![
                kaka.account.contract_address,
                karp.account.contract_address,
                veg.account.contract_address,
            ],
            1_000_000_000 * E18,
        );

        self.balance_id.entry(kaka.account.contract_address).write(0);
        self.balance_id.entry(karp.account.contract_address).write(1);
        self.balance_id.entry(veg.account.contract_address).write(2);
        self.balance_id.entry(counter_part).write(3);

        self.kaka.write(kaka.account.contract_address);
        self.kaka_key_pair.write((kaka.key_pair.secret_key, kaka.key_pair.public_key));

        self.karp.write(kaka.account.contract_address);
        self.karp_key_pair.write((karp.key_pair.secret_key, karp.key_pair.public_key));

        self.veg.write(kaka.account.contract_address);
        self.veg_key_pair.write((veg.key_pair.secret_key, veg.key_pair.public_key));

        self.counter_part.write(counter_part);
    }


    #[generate_trait]
    pub impl InternalImpl<TState> of InternalTrait<TState> {
        fn _prepare_onchain_order(
            self: @ContractState, order_data: Bytes, fill_deadline: u64, order_data_type: felt252,
        ) -> OnchainCrossChainOrder {
            OnchainCrossChainOrder { order_data, fill_deadline, order_data_type }
        }

        fn _prepare_gasless_order(
            self: @ContractState,
            origin_settler: ContractAddress,
            user: ContractAddress,
            origin_chain_id: u32,
            order_data: Bytes,
            nonce: felt252,
            open_deadline: u64,
            fill_deadline: u64,
            order_data_type: felt252,
        ) -> GaslessCrossChainOrder {
            GaslessCrossChainOrder {
                origin_settler,
                user,
                origin_chain_id,
                order_data,
                nonce,
                open_deadline,
                fill_deadline,
                order_data_type,
            }
        }

        fn _get_order_id_from_logs(
            self: @ContractState, order: OnchainCrossChainOrder, signature: Bytes,
        ) -> (u256, ResolvedCrossChainOrder) {
            let (keys, mut data) = pop_log_raw(get_contract_address());

            let resolved_order: ResolvedCrossChainOrder = Serde::deserialize(ref data).unwrap();
            let order_id: u256 = u256 {
                low: (*keys.at(1)).try_into().unwrap(), high: (*keys.at(2)).try_into().unwrap(),
            };

            (order_id, resolved_order)
        }

        fn _token_balances(self: @ContractState, token: IERC20Dispatcher) -> Array<u256> {
            array![
                token.balance_of(self.kaka.read()),
                token.balance_of(self.karp.read()),
                token.balance_of(self.veg.read()),
                token.balance_of(self.counter_part.read()),
            ]
        }
        fn _balances(self: @ContractState) -> Array<u256> {
            let eth = IERC20Dispatcher { contract_address: ETH_ADDRESS() };
            array![
                eth.balance_of(self.kaka.read()),
                eth.balance_of(self.karp.read()),
                eth.balance_of(self.veg.read()),
                eth.balance_of(self.counter_part.read()),
            ]
        }

        const _TOKEN_PERMISSIONS_TYPE_HASH: felt252 = selector!(
            "\"Token Permissions\"(\"Token\":\"ContractAddress\",\"Amount\":\"u256\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")\"",
        );

        const FULL_WITNESS_TYPE_HASH: felt252 = selector!(
            "\"Permit Witness Transfer From\"(\"Permitted\":\"Token Permissions\",\"Spender\":\"ContractAddress\",\"Nonce\":\"felt\",\"Deadline\":\"u256\",\"Witness\":\"Resolved Cross Chain Order\")\"Bytes\"(\"Size\":\"u128\",\"Data\":\"u128*\")\"Fill Instruction\"(\"Destination Chain ID\":\"u128\",\"Destination Settler\":\"ContractAddress\",\"Origin Data\":\"Bytes\")\"Resolved Cross Chain Order\"(\"User\":\"ContractAddress\",\"Origin Chain ID\":\"u128\",\"Open Deadline\":\"timestamp\",\"Fill Deadline\":\"timestamp\",\"Order ID\":\"u256\",\"Max Spent\":\"Output*\",\"Min Received\":\"Output*\",\"Fill Instructions\":\"Fill Instruction*\")\"Output\"(\"Token\":\"ContractAddress\",\"Amount\":\"u256\",\"Recipient\":\"ContractAddress\",\"Chain ID\":\"u128\")\"Token Permissions\"(\"Token\":\"ContractAddress\",\"Amount\":\"u256\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")",
        );

        const FULL_WITNESS_BATCH_TYPE_HASH: felt252 = selector!(
            "\"Permit Witness Batch Transfer From\"(\"Permitted\":\"Token Permissions*\",\"Spender\":\"ContractAddress\",\"Nonce\":\"felt\",\"Deadline\":\"u256\",\"Witness\":\"Resolved Cross Chain Order\")\"Bytes\"(\"Size\":\"u128\",\"Data\":\"u128*\")\"Fill Instruction\"(\"Destination Chain ID\":\"u128\",\"Destination Settler\":\"ContractAddress\",\"Origin Data\":\"Bytes\")\"Resolved Cross Chain Order\"(\"User\":\"ContractAddress\",\"Origin Chain ID\":\"u128\",\"Open Deadline\":\"timestamp\",\"Fill Deadline\":\"timestamp\",\"Order ID\":\"u256\",\"Max Spent\":\"Output*\",\"Min Received\":\"Output*\",\"Fill Instructions\":\"Fill Instruction*\")\"Output\"(\"Token\":\"ContractAddress\",\"Amount\":\"u256\",\"Recipient\":\"ContractAddress\",\"Chain ID\":\"u128\")\"Token Permissions\"(\"Token\":\"ContractAddress\",\"Amount\":\"u256\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")",
        );

        //fn _get_permit_witness_signature(
        //    self: @ContractState,
        //    spender: ContractAddress,
        //    permit: PermitTransferFrom,
        //    signer: Account,
        //    type_hash: felt252,
        //    witness: felt252,
        //    domain_separator: felt252,
        //) -> (felt252, felt252) {
        //    let hashed_permit = PoseidonTrait::new()
        //        .update_with(Self::FULL_WITNESS_TYPE_HASH)
        //        .update_with(permit.permitted.hash_struct())
        //        .update_with(starknet::get_caller_address())
        //        .update_with(permit.nonce)
        //        .update_with(permit.deadline.hash_struct())
        //        .update_with(witness)
        //        .finalize();

        //    let msg_hash = PoseidonTrait::new()
        //        .update_with(domain_separator)
        //        .update_with(signer.account.contract_address)
        //        .update_with(hashed_permit)
        //        .finalize();

        //    signer.key_pair.sign(msg_hash).unwrap()
        //}

        fn _get_permit_batch_witness_signature(
            self: @ContractState,
            signer: Account,
            spender: ContractAddress,
            permit: PermitBatchTransferFrom,
            type_hash: felt252,
            witness: felt252,
            domain_separator: felt252,
        ) -> (felt252, felt252) {
            let mut hashed_permissions: Array<felt252> = array![];
            for permission in permit.permitted {
                hashed_permissions.append(permission.hash_struct());
            };

            let hashed_permit = PoseidonTrait::new()
                .update_with(type_hash)
                .update_with(hashed_permissions.span())
                .update_with(starknet::get_caller_address())
                .update_with(permit.nonce)
                .update_with(permit.deadline.hash_struct())
                .update_with(witness)
                .finalize();

            let msg_hash = PoseidonTrait::new()
                .update_with(domain_separator)
                .update_with(signer.account.contract_address)
                .update_with(hashed_permit)
                .finalize();

            signer.key_pair.sign(msg_hash).unwrap()
        }

        fn _default_erc20_permit_multiple(
            self: @ContractState,
            tokens: Array<ContractAddress>,
            nonce: felt252,
            amount: u256,
            deadline: u256,
        ) -> PermitBatchTransferFrom {
            let mut permitted: Array<TokenPermissions> = array![];
            for token in tokens.span() {
                permitted.append(TokenPermissions { token: *token, amount });
            };

            PermitBatchTransferFrom { permitted: permitted.span(), nonce, deadline }
        }

        fn _get_signature(
            self: @ContractState,
            signer: Account,
            spender: ContractAddress,
            witness: felt252,
            token: ContractAddress,
            nonce: felt252,
            amount: u256,
            deadline: u256,
        ) -> (felt252, felt252) {
            let permit = Self::_default_erc20_permit_multiple(
                self, array![token], nonce, amount, deadline,
            );

            Self::_get_permit_batch_witness_signature(
                self,
                signer,
                spender,
                permit,
                Self::FULL_WITNESS_BATCH_TYPE_HASH,
                witness,
                self.DOMAIN_SEPARATOR.read(),
            )
        }

        fn _assert_resolved_order(
            self: @ContractState,
            resolved_order: ResolvedCrossChainOrder,
            order_data: Bytes,
            user: ContractAddress,
            fill_deadline: u64,
            open_deadline: u64,
            to: ContractAddress,
            destination_settler: ContractAddress,
            origin_chain_id: u32,
            input_token: ContractAddress,
            output_token: ContractAddress,
        ) {
            assert_eq!(resolved_order.max_spent.len(), 1);
            assert_eq!(*resolved_order.max_spent.at(0).token, output_token);
            assert_eq!(*resolved_order.max_spent.at(0).amount, amount);
            assert_eq!(*resolved_order.max_spent.at(0).recipient, to);
            assert_eq!(*resolved_order.max_spent.at(0).chain_id, destination);
            assert_eq!(resolved_order.min_received.len(), 1);
            assert_eq!(*resolved_order.min_received.at(0).token, input_token);
            assert_eq!(*resolved_order.min_received.at(0).amount, amount);
            assert_eq!(*resolved_order.min_received.at(0).recipient, 0.try_into().unwrap());
            assert_eq!(*resolved_order.min_received.at(0).chain_id, origin);
            assert_eq!(resolved_order.fill_instructions.len(), 1);
            assert_eq!(*resolved_order.fill_instructions.at(0).destination_chain_id, destination);
            assert_eq!(
                *resolved_order.fill_instructions.at(0).destination_settler, destination_settler,
            );
            assert(
                resolved_order.fill_instructions.at(0).origin_data == @order_data,
                'Fill instructions do not match',
            );
            assert_eq!(resolved_order.user, user);
            assert_eq!(resolved_order.origin_chain_id, origin_chain_id);
            assert_eq!(resolved_order.open_deadline, open_deadline);
            assert_eq!(resolved_order.fill_deadline, fill_deadline);
        }
    }
}
