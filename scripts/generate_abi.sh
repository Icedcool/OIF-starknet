#!/bin/bash

# Add contracts here as "contract_name:abi_name"
contracts=(
	"Hyperlane7683:hyperlane7683"
)

echo "Running scarb build..."
cd ../cairo/ && scarb build && cd ../scripts

# Generate ABIs
for contract in "${contracts[@]}"; do
	IFS=':' read -r contract_name abi_name <<<"$contract"
	json_file="../cairo/target/dev/oif_starknet_${contract_name}.contract_class.json"
	abi_file="../ABI/${abi_name}.ts"

	npx abi-wan-kanabi --input "$json_file" --output "$abi_file"
done
