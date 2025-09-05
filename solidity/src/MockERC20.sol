// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title MockERC20
 * @dev A simple ERC20 token with minting capability for testing purposes
 * This is used in local forks to fund test accounts with tokens
 */
contract MockERC20 is ERC20 {
    uint8 private _decimals;

    constructor() ERC20("MockERC20", "MOCK") {
        _decimals = 18;
    }

    /**
     * @dev Returns the number of decimals used to get its user representation.
     */
    function decimals() public view virtual override returns (uint8) {
        return _decimals;
    }

    /**
     * @dev Mint tokens to a specific address
     * @param to The address to mint tokens to
     * @param amount The amount of tokens to mint
     */
    function mint(address to, uint256 amount) public {
        _mint(to, amount);
    }

    ///**
    // * @dev Mint tokens to multiple addresses
    // * @param recipients Array of addresses to mint tokens to
    // * @param amounts Array of amounts to mint to each recipient
    // */
    //function mintBatch(address[] calldata recipients, uint256[] calldata amounts) public {
    //    require(recipients.length == amounts.length, "MockERC20: arrays length mismatch");
    //    
    //    for (uint256 i = 0; i < recipients.length; i++) {
    //        _mint(recipients[i], amounts[i]);
    //    }
    //}
}
