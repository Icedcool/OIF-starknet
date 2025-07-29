use crate::mocks::mock_erc20::MockERC20;
use crate::mocks::interfaces::{IMintableDispatcher, IMintableDispatcherTrait};
use openzeppelin_token::erc20::interface::{IERC20Dispatcher, IERC20DispatcherTrait};
use starknet::{ContractAddress, ClassHash};
use snforge_std::signature::{KeyPair, KeyPairTrait};
use snforge_std::{ContractClassTrait, DeclareResultTrait, declare};
use openzeppelin_account::interface::AccountABIDispatcher;
use snforge_std::signature::stark_curve::{
    StarkCurveKeyPairImpl, StarkCurveSignerImpl, StarkCurveVerifierImpl,
};

pub fn ETH_ADDRESS() -> ContractAddress {
    0x049D36570D4e46f48e99674bd3fcc84644DdD6b96F7C741B1562B82f9e004dC7.try_into().unwrap()
}

pub fn deploy_eth() -> IERC20Dispatcher {
    let mock_erc20_contract = declare("MockERC20").unwrap().contract_class();
    let ctor_calldata: Array<felt252> = array!['Ethereum', 'ETH'];

    let (erc20_address, _) = mock_erc20_contract.deploy_at(@ctor_calldata, ETH_ADDRESS()).unwrap();
    IERC20Dispatcher { contract_address: erc20_address }
}

pub fn deploy_erc20(name: ByteArray, symbol: ByteArray) -> IERC20Dispatcher {
    let mock_erc20_contract = declare("MockERC20").unwrap().contract_class();
    let mut ctor_calldata: Array<felt252> = array![];
    name.serialize(ref ctor_calldata);
    symbol.serialize(ref ctor_calldata);

    let (erc20_address, _) = mock_erc20_contract.deploy(@ctor_calldata).unwrap();
    IERC20Dispatcher { contract_address: erc20_address }
}

pub fn deploy_permit2() -> ContractAddress {
    let mock_permit2_contract = declare("MockPermit2").unwrap().contract_class();
    let (mock_permit2_address, _) = mock_permit2_contract
        .deploy(@array![])
        .expect('mock permit2 deployment failed');

    mock_permit2_address
}

pub fn deal(token: ContractAddress, to: ContractAddress, amount: u256) {
    IMintableDispatcher { contract_address: token }.mint(to, amount);
}

pub fn deal_multiple(tokens: Array<ContractAddress>, tos: Array<ContractAddress>, amount: u256) {
    for token in tokens.span() {
        for to in tos.span() {
            deal(*token, *to, amount);
        };
    };
}


/// Accounts ///

#[derive(Drop, Copy)]
pub struct Account {
    pub account: AccountABIDispatcher,
    pub key_pair: KeyPair<felt252, felt252>,
}

pub fn generate_account() -> Account {
    let mock_account_contract = declare("MockAccount").unwrap().contract_class();
    let key_pair = KeyPairTrait::<felt252, felt252>::generate();
    let (account_address, _) = mock_account_contract.deploy(@array![key_pair.public_key]).unwrap();
    let account = AccountABIDispatcher { contract_address: account_address };
    Account { account, key_pair }
}

/// Utils ///

// Pop the earliest unpopped logged event for the contract as the requested type
// and checks there's no more data left on the event, preventing unaccounted params.
// Indexed event members are currently not supported, so they are ignored.
pub fn pop_log<T, +Drop<T>, impl TEvent: starknet::Event<T>>(
    address: ContractAddress,
) -> Option<T> {
    let (mut keys, mut data) = starknet::testing::pop_log_raw(starknet::get_contract_address())
        .expect('No logs found');
    //println!("keys: {:?}\ndata: {:?}", keys.clone(), data.clone());
    let ret = starknet::Event::deserialize(ref keys, ref data);
    assert(data.is_empty(), 'Event has extra data');
    ret
}

pub fn pop_log_raw(address: ContractAddress) -> (Span<felt252>, Span<felt252>) {
    let (mut keys, mut data) = starknet::testing::pop_log_raw(starknet::get_contract_address())
        .expect('No logs found');

    (keys, data)
}

