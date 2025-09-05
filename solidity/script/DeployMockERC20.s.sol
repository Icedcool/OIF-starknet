// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "../src/MockERC20.sol";

/**
 * @title DeployMockERC20
 * @dev Deploy MockERC20 contract using Forge with exact compiler settings
 * Usage: forge script script/DeployMockERC20.s.sol:DeployMockERC20 --chain <chain_id> --broadcast --verify
 */
contract DeployMockERC20 is Script {
    function run() external {
        string memory deployerPrivateKeyStr = vm.envString("DEPLOYER_PRIVATE_KEY");
        uint256 deployerPrivateKey;
        
        // Handle private key with or without 0x prefix
        if (bytes(deployerPrivateKeyStr).length > 2 && 
            bytes(deployerPrivateKeyStr)[0] == "0" && 
            bytes(deployerPrivateKeyStr)[1] == "x") {
            deployerPrivateKey = vm.parseUint(deployerPrivateKeyStr);
        } else {
            deployerPrivateKey = vm.parseUint(string(abi.encodePacked("0x", deployerPrivateKeyStr)));
        }
        
        vm.startBroadcast(deployerPrivateKey);
        
        // Deploy MockERC20
        MockERC20 mockERC20 = new MockERC20();
        
        vm.stopBroadcast();
        
        console.log("MockERC20 deployed at:", address(mockERC20));
        console.log("Name:", mockERC20.name());
        console.log("Symbol:", mockERC20.symbol());
        console.log("Decimals:", mockERC20.decimals());
        
        // Show compiler settings being used
        console.log("\n=== Compiler Settings ===");
        console.log("These settings will be used for verification:");
        console.log("- Check foundry.toml for exact configuration");
        console.log("- Solidity version: specified in foundry.toml");
        console.log("- Optimizer: specified in foundry.toml");
        console.log("- Optimizer runs: specified in foundry.toml");
    }
}
