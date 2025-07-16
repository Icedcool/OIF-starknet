use starknet::ContractAddress;

/// Standard order struct to be signed by users, disseminated to fillers, and submitted to origin
/// settler contracts by fillers
#[derive(Serde, Drop)]
pub struct GaslessCrossChainOrder {
    origin_settler: ContractAddress,
    user: ContractAddress,
    nonce: felt252, //u256,
    origin_chain_id: u256,
    open_deadline: u64, //u32,
    fill_deadline: u64, //u32,
    order_data_type: felt252,
    order_data: ByteArray,
}

/// Standard order struct for user-opened orders, where the user is the one submitting the order
/// creation transaction
#[derive(Serde, Drop)]
pub struct OnchainCrossChainOrder {
    fill_deadline: u64, //u32,
    order_data_type: felt252,
    order_data: ByteArray,
}

/// An implementation-generic representation of an order intended for filler consumption
/// @dev Defines all requirements for filling an order by unbundling the implementation-specific
/// orderData.
/// @dev Intended to improve integration generalization by allowing fillers to compute the exact
/// input and output information of any order
#[derive(Serde, Drop)]
pub struct ResolvedCrossChainOrder {
    user: ContractAddress,
    originChainId: u256,
    open_deadline: u64,
    fill_deadline: u64,
    order_id: felt252,
    max_spent: Array<Output>,
    min_received: Array<Output>,
    fill_instructions: Array<FillInstruction>,
}

/// Tokens that must be received for a valid order fulfillment
#[derive(Serde, Drop)]
pub struct Output {
    token: felt252,
    amount: u256,
    recipient: felt252,
    chain_id: u256,
}

/// Instructions to parameterize each leg of the fill
#[derive(Serde, Drop)]
pub struct FillInstruction {
    destination_chain_id: u256,
    destination_settler: felt252,
    origin_data: Array<felt252>,
}

/// Standard interface for settlement contracts on the origin chain
#[starknet::interface]
pub trait IOriginSettler<TState> {
    /// Opens a gasless cross-chain order on behalf of a user.
    /// @dev To be called by the filler.
    /// @dev This method must emit the Open event
    ///
    /// Parameters:
    /// - `order`: The GaslessCrossChainOrder definition
    /// - `signature`: The user's signature over the order
    /// - `origin_filler_data`: Any filler-defined data required by the settler
    fn open_for(
        ref self: TState,
        order: GaslessCrossChainOrder,
        signature: Array<felt252>,
        origin_filler_data: ByteArray,
    );

    /// Opens a cross-chain order
    /// @dev To be called by the user
    /// @dev This method must emit the Open event
    ///
    /// Parameter:
    /// - `order`: The OnchainCrossChainOrder definition
    fn open(ref self: TState, order: OnchainCrossChainOrder);

    /// Resolves a specific GaslessCrossChainOrder into a generic ResolvedCrossChainOrder
    /// @dev Intended to improve standardized integration of various order types and settlement
    /// contracts
    ///
    /// Parameters:
    /// - ``order` The GaslessCrossChainOrder definition
    /// - `origin_filler_data` Any filler-defined data required by the settler
    ///
    /// Returns: ResolvedCrossChainOrder hydrated order data including the inputs and outputs of the
    /// order
    fn resolve_for(
        self: @TState, order: GaslessCrossChainOrder, origin_filler_data: Array<u8>,
    ) -> ResolvedCrossChainOrder;

    /// Resolves a specific OnchainCrossChainOrder into a generic ResolvedCrossChainOrder
    /// @dev Intended to improve standardized integration of various order types and settlement
    /// contracts
    ///
    /// Parameters:
    /// - `order`: The OnchainCrossChainOrder definition
    ///
    /// Returns: ResolvedCrossChainOrder hydrated order data including the inputs and outputs of the
    /// order
    fn resolve(self: @TState, order: OnchainCrossChainOrder) -> ResolvedCrossChainOrder;
}

/// Standard interface for settlement contracts on the destination chain
#[starknet::interface]
pub trait IDesitnationSettler<TState> {
    /// Fills a single leg of a particular order on the destination chain
    ///
    /// Parameters
    /// - `order_id`: Unique order identifier for this order
    /// - `origin_data`: Data emitted on the origin to parameterize the fill
    /// - `filler_data`: Data provided by the filler to inform the fill or express their preferences
    fn fill(ref self: TState, order_id: felt252, origin_data: ByteArray, filler_data: ByteArray);
}

