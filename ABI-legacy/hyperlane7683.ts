export const ABI = [
  {
    "type": "impl",
    "name": "OriginSettlerImpl",
    "interface_name": "oif_starknet::erc7683::interface::IOriginSettler"
  },
  {
    "type": "struct",
    "name": "alexandria_bytes::bytes::Bytes",
    "members": [
      {
        "name": "size",
        "type": "core::integer::u32"
      },
      {
        "name": "data",
        "type": "core::array::Array::<core::integer::u128>"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::erc7683::interface::GaslessCrossChainOrder",
    "members": [
      {
        "name": "origin_settler",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "user",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "nonce",
        "type": "core::felt252"
      },
      {
        "name": "origin_chain_id",
        "type": "core::integer::u32"
      },
      {
        "name": "open_deadline",
        "type": "core::integer::u64"
      },
      {
        "name": "fill_deadline",
        "type": "core::integer::u64"
      },
      {
        "name": "order_data_type",
        "type": "core::felt252"
      },
      {
        "name": "order_data",
        "type": "alexandria_bytes::bytes::Bytes"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::erc7683::interface::OnchainCrossChainOrder",
    "members": [
      {
        "name": "fill_deadline",
        "type": "core::integer::u64"
      },
      {
        "name": "order_data_type",
        "type": "core::felt252"
      },
      {
        "name": "order_data",
        "type": "alexandria_bytes::bytes::Bytes"
      }
    ]
  },
  {
    "type": "struct",
    "name": "core::integer::u256",
    "members": [
      {
        "name": "low",
        "type": "core::integer::u128"
      },
      {
        "name": "high",
        "type": "core::integer::u128"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::erc7683::interface::Output",
    "members": [
      {
        "name": "token",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "amount",
        "type": "core::integer::u256"
      },
      {
        "name": "recipient",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "chain_id",
        "type": "core::integer::u32"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::erc7683::interface::FillInstruction",
    "members": [
      {
        "name": "destination_chain_id",
        "type": "core::integer::u32"
      },
      {
        "name": "destination_settler",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "origin_data",
        "type": "alexandria_bytes::bytes::Bytes"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::erc7683::interface::ResolvedCrossChainOrder",
    "members": [
      {
        "name": "user",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "origin_chain_id",
        "type": "core::integer::u32"
      },
      {
        "name": "open_deadline",
        "type": "core::integer::u64"
      },
      {
        "name": "fill_deadline",
        "type": "core::integer::u64"
      },
      {
        "name": "order_id",
        "type": "core::integer::u256"
      },
      {
        "name": "max_spent",
        "type": "core::array::Array::<oif_starknet::erc7683::interface::Output>"
      },
      {
        "name": "min_received",
        "type": "core::array::Array::<oif_starknet::erc7683::interface::Output>"
      },
      {
        "name": "fill_instructions",
        "type": "core::array::Array::<oif_starknet::erc7683::interface::FillInstruction>"
      }
    ]
  },
  {
    "type": "interface",
    "name": "oif_starknet::erc7683::interface::IOriginSettler",
    "items": [
      {
        "type": "function",
        "name": "open_for",
        "inputs": [
          {
            "name": "order",
            "type": "oif_starknet::erc7683::interface::GaslessCrossChainOrder"
          },
          {
            "name": "signature",
            "type": "core::array::Array::<core::felt252>"
          },
          {
            "name": "origin_filler_data",
            "type": "alexandria_bytes::bytes::Bytes"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "open",
        "inputs": [
          {
            "name": "order",
            "type": "oif_starknet::erc7683::interface::OnchainCrossChainOrder"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "resolve_for",
        "inputs": [
          {
            "name": "order",
            "type": "oif_starknet::erc7683::interface::GaslessCrossChainOrder"
          },
          {
            "name": "origin_filler_data",
            "type": "alexandria_bytes::bytes::Bytes"
          }
        ],
        "outputs": [
          {
            "type": "oif_starknet::erc7683::interface::ResolvedCrossChainOrder"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "resolve",
        "inputs": [
          {
            "name": "order",
            "type": "oif_starknet::erc7683::interface::OnchainCrossChainOrder"
          }
        ],
        "outputs": [
          {
            "type": "oif_starknet::erc7683::interface::ResolvedCrossChainOrder"
          }
        ],
        "state_mutability": "view"
      }
    ]
  },
  {
    "type": "impl",
    "name": "DestinationSettlerImpl",
    "interface_name": "oif_starknet::erc7683::interface::IDestinationSettler"
  },
  {
    "type": "interface",
    "name": "oif_starknet::erc7683::interface::IDestinationSettler",
    "items": [
      {
        "type": "function",
        "name": "fill",
        "inputs": [
          {
            "name": "order_id",
            "type": "core::integer::u256"
          },
          {
            "name": "origin_data",
            "type": "alexandria_bytes::bytes::Bytes"
          },
          {
            "name": "filler_data",
            "type": "alexandria_bytes::bytes::Bytes"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      }
    ]
  },
  {
    "type": "impl",
    "name": "Base7683Extra",
    "interface_name": "oif_starknet::erc7683::interface::IERC7683Extra"
  },
  {
    "type": "enum",
    "name": "core::bool",
    "variants": [
      {
        "name": "False",
        "type": "()"
      },
      {
        "name": "True",
        "type": "()"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::erc7683::interface::FilledOrder",
    "members": [
      {
        "name": "origin_data",
        "type": "alexandria_bytes::bytes::Bytes"
      },
      {
        "name": "filler_data",
        "type": "alexandria_bytes::bytes::Bytes"
      }
    ]
  },
  {
    "type": "interface",
    "name": "oif_starknet::erc7683::interface::IERC7683Extra",
    "items": [
      {
        "type": "function",
        "name": "UNKNOWN",
        "inputs": [],
        "outputs": [
          {
            "type": "core::felt252"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "OPENED",
        "inputs": [],
        "outputs": [
          {
            "type": "core::felt252"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "FILLED",
        "inputs": [],
        "outputs": [
          {
            "type": "core::felt252"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "witness_hash",
        "inputs": [
          {
            "name": "resolved_order",
            "type": "oif_starknet::erc7683::interface::ResolvedCrossChainOrder"
          }
        ],
        "outputs": [
          {
            "type": "core::felt252"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "used_nonces",
        "inputs": [
          {
            "name": "user",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "nonce",
            "type": "core::felt252"
          }
        ],
        "outputs": [
          {
            "type": "core::bool"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "open_orders",
        "inputs": [
          {
            "name": "order_id",
            "type": "core::integer::u256"
          }
        ],
        "outputs": [
          {
            "type": "alexandria_bytes::bytes::Bytes"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "filled_orders",
        "inputs": [
          {
            "name": "order_id",
            "type": "core::integer::u256"
          }
        ],
        "outputs": [
          {
            "type": "oif_starknet::erc7683::interface::FilledOrder"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "order_status",
        "inputs": [
          {
            "name": "order_id",
            "type": "core::integer::u256"
          }
        ],
        "outputs": [
          {
            "type": "core::felt252"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "is_valid_nonce",
        "inputs": [
          {
            "name": "from",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "nonce",
            "type": "core::felt252"
          }
        ],
        "outputs": [
          {
            "type": "core::bool"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "settle",
        "inputs": [
          {
            "name": "order_ids",
            "type": "core::array::Array::<core::integer::u256>"
          },
          {
            "name": "value",
            "type": "core::integer::u256"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "refund_gasless_cross_chain_order",
        "inputs": [
          {
            "name": "orders",
            "type": "core::array::Array::<oif_starknet::erc7683::interface::GaslessCrossChainOrder>"
          },
          {
            "name": "value",
            "type": "core::integer::u256"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "refund_onchain_cross_chain_order",
        "inputs": [
          {
            "name": "orders",
            "type": "core::array::Array::<oif_starknet::erc7683::interface::OnchainCrossChainOrder>"
          },
          {
            "name": "value",
            "type": "core::integer::u256"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "invalidate_nonces",
        "inputs": [
          {
            "name": "nonce",
            "type": "core::felt252"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      }
    ]
  },
  {
    "type": "impl",
    "name": "BasicSwapExtraImpl",
    "interface_name": "oif_starknet::erc7683::interface::IBasicSwapExtra"
  },
  {
    "type": "interface",
    "name": "oif_starknet::erc7683::interface::IBasicSwapExtra",
    "items": [
      {
        "type": "function",
        "name": "SETTLED",
        "inputs": [],
        "outputs": [
          {
            "type": "core::felt252"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "REFUNDED",
        "inputs": [],
        "outputs": [
          {
            "type": "core::felt252"
          }
        ],
        "state_mutability": "view"
      }
    ]
  },
  {
    "type": "impl",
    "name": "OwnableImpl",
    "interface_name": "openzeppelin_access::ownable::interface::IOwnable"
  },
  {
    "type": "interface",
    "name": "openzeppelin_access::ownable::interface::IOwnable",
    "items": [
      {
        "type": "function",
        "name": "owner",
        "inputs": [],
        "outputs": [
          {
            "type": "core::starknet::contract_address::ContractAddress"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "transfer_ownership",
        "inputs": [
          {
            "name": "new_owner",
            "type": "core::starknet::contract_address::ContractAddress"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "renounce_ownership",
        "inputs": [],
        "outputs": [],
        "state_mutability": "external"
      }
    ]
  },
  {
    "type": "impl",
    "name": "MailboxClientImpl",
    "interface_name": "contracts::interfaces::IMailboxClient"
  },
  {
    "type": "interface",
    "name": "contracts::interfaces::IMailboxClient",
    "items": [
      {
        "type": "function",
        "name": "set_hook",
        "inputs": [
          {
            "name": "_hook",
            "type": "core::starknet::contract_address::ContractAddress"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "set_interchain_security_module",
        "inputs": [
          {
            "name": "_module",
            "type": "core::starknet::contract_address::ContractAddress"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "get_hook",
        "inputs": [],
        "outputs": [
          {
            "type": "core::starknet::contract_address::ContractAddress"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "get_local_domain",
        "inputs": [],
        "outputs": [
          {
            "type": "core::integer::u32"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "interchain_security_module",
        "inputs": [],
        "outputs": [
          {
            "type": "core::starknet::contract_address::ContractAddress"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "_is_latest_dispatched",
        "inputs": [
          {
            "name": "_id",
            "type": "core::integer::u256"
          }
        ],
        "outputs": [
          {
            "type": "core::bool"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "_is_delivered",
        "inputs": [
          {
            "name": "_id",
            "type": "core::integer::u256"
          }
        ],
        "outputs": [
          {
            "type": "core::bool"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "mailbox",
        "inputs": [],
        "outputs": [
          {
            "type": "core::starknet::contract_address::ContractAddress"
          }
        ],
        "state_mutability": "view"
      }
    ]
  },
  {
    "type": "impl",
    "name": "RouterImpl",
    "interface_name": "contracts::client::router_component::IRouter"
  },
  {
    "type": "interface",
    "name": "contracts::client::router_component::IRouter",
    "items": [
      {
        "type": "function",
        "name": "enroll_remote_router",
        "inputs": [
          {
            "name": "domain",
            "type": "core::integer::u32"
          },
          {
            "name": "router",
            "type": "core::integer::u256"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "enroll_remote_routers",
        "inputs": [
          {
            "name": "domains",
            "type": "core::array::Array::<core::integer::u32>"
          },
          {
            "name": "addresses",
            "type": "core::array::Array::<core::integer::u256>"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "unenroll_remote_router",
        "inputs": [
          {
            "name": "domain",
            "type": "core::integer::u32"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "unenroll_remote_routers",
        "inputs": [
          {
            "name": "domains",
            "type": "core::array::Array::<core::integer::u32>"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "handle",
        "inputs": [
          {
            "name": "origin",
            "type": "core::integer::u32"
          },
          {
            "name": "sender",
            "type": "core::integer::u256"
          },
          {
            "name": "message",
            "type": "alexandria_bytes::bytes::Bytes"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "domains",
        "inputs": [],
        "outputs": [
          {
            "type": "core::array::Array::<core::integer::u32>"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "routers",
        "inputs": [
          {
            "name": "domain",
            "type": "core::integer::u32"
          }
        ],
        "outputs": [
          {
            "type": "core::integer::u256"
          }
        ],
        "state_mutability": "view"
      }
    ]
  },
  {
    "type": "impl",
    "name": "GasRouterImpl",
    "interface_name": "contracts::client::gas_router_component::IGasRouter"
  },
  {
    "type": "struct",
    "name": "contracts::client::gas_router_component::GasRouterComponent::GasRouterConfig",
    "members": [
      {
        "name": "domain",
        "type": "core::integer::u32"
      },
      {
        "name": "gas",
        "type": "core::integer::u256"
      }
    ]
  },
  {
    "type": "enum",
    "name": "core::option::Option::<core::array::Array::<contracts::client::gas_router_component::GasRouterComponent::GasRouterConfig>>",
    "variants": [
      {
        "name": "Some",
        "type": "core::array::Array::<contracts::client::gas_router_component::GasRouterComponent::GasRouterConfig>"
      },
      {
        "name": "None",
        "type": "()"
      }
    ]
  },
  {
    "type": "enum",
    "name": "core::option::Option::<core::integer::u32>",
    "variants": [
      {
        "name": "Some",
        "type": "core::integer::u32"
      },
      {
        "name": "None",
        "type": "()"
      }
    ]
  },
  {
    "type": "enum",
    "name": "core::option::Option::<core::integer::u256>",
    "variants": [
      {
        "name": "Some",
        "type": "core::integer::u256"
      },
      {
        "name": "None",
        "type": "()"
      }
    ]
  },
  {
    "type": "interface",
    "name": "contracts::client::gas_router_component::IGasRouter",
    "items": [
      {
        "type": "function",
        "name": "set_destination_gas",
        "inputs": [
          {
            "name": "gas_configs",
            "type": "core::option::Option::<core::array::Array::<contracts::client::gas_router_component::GasRouterComponent::GasRouterConfig>>"
          },
          {
            "name": "domain",
            "type": "core::option::Option::<core::integer::u32>"
          },
          {
            "name": "gas",
            "type": "core::option::Option::<core::integer::u256>"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "destination_gas",
        "inputs": [
          {
            "name": "destination_domain",
            "type": "core::integer::u32"
          }
        ],
        "outputs": [
          {
            "type": "core::integer::u256"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "quote_gas_payment",
        "inputs": [
          {
            "name": "destination_domain",
            "type": "core::integer::u32"
          }
        ],
        "outputs": [
          {
            "type": "core::integer::u256"
          }
        ],
        "state_mutability": "view"
      }
    ]
  },
  {
    "type": "constructor",
    "name": "constructor",
    "inputs": [
      {
        "name": "permit2",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "mailbox",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "owner",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "hook",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "interchain_security_module",
        "type": "core::starknet::contract_address::ContractAddress"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::base7683::Base7683Component::Filled",
    "kind": "struct",
    "members": [
      {
        "name": "order_id",
        "type": "core::integer::u256",
        "kind": "data"
      },
      {
        "name": "origin_data",
        "type": "alexandria_bytes::bytes::Bytes",
        "kind": "data"
      },
      {
        "name": "filler_data",
        "type": "alexandria_bytes::bytes::Bytes",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::base7683::Base7683Component::Settle",
    "kind": "struct",
    "members": [
      {
        "name": "order_ids",
        "type": "core::array::Array::<core::integer::u256>",
        "kind": "data"
      },
      {
        "name": "orders_filler_data",
        "type": "core::array::Array::<alexandria_bytes::bytes::Bytes>",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::base7683::Base7683Component::Refund",
    "kind": "struct",
    "members": [
      {
        "name": "order_ids",
        "type": "core::array::Array::<core::integer::u256>",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::base7683::Base7683Component::NonceInvalidation",
    "kind": "struct",
    "members": [
      {
        "name": "owner",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "nonce",
        "type": "core::felt252",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::erc7683::interface::Open",
    "kind": "struct",
    "members": [
      {
        "name": "order_id",
        "type": "core::integer::u256",
        "kind": "key"
      },
      {
        "name": "resolved_order",
        "type": "oif_starknet::erc7683::interface::ResolvedCrossChainOrder",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::base7683::Base7683Component::Event",
    "kind": "enum",
    "variants": [
      {
        "name": "Filled",
        "type": "oif_starknet::base7683::Base7683Component::Filled",
        "kind": "nested"
      },
      {
        "name": "Settle",
        "type": "oif_starknet::base7683::Base7683Component::Settle",
        "kind": "nested"
      },
      {
        "name": "Refund",
        "type": "oif_starknet::base7683::Base7683Component::Refund",
        "kind": "nested"
      },
      {
        "name": "NonceInvalidation",
        "type": "oif_starknet::base7683::Base7683Component::NonceInvalidation",
        "kind": "nested"
      },
      {
        "name": "Open",
        "type": "oif_starknet::erc7683::interface::Open",
        "kind": "nested"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::basic_swap7683::BasicSwap7683Component::Settled",
    "kind": "struct",
    "members": [
      {
        "name": "order_id",
        "type": "core::integer::u256",
        "kind": "data"
      },
      {
        "name": "receiver",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::basic_swap7683::BasicSwap7683Component::Refunded",
    "kind": "struct",
    "members": [
      {
        "name": "order_id",
        "type": "core::integer::u256",
        "kind": "data"
      },
      {
        "name": "receiver",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::basic_swap7683::BasicSwap7683Component::Event",
    "kind": "enum",
    "variants": [
      {
        "name": "Settled",
        "type": "oif_starknet::basic_swap7683::BasicSwap7683Component::Settled",
        "kind": "nested"
      },
      {
        "name": "Refunded",
        "type": "oif_starknet::basic_swap7683::BasicSwap7683Component::Refunded",
        "kind": "nested"
      }
    ]
  },
  {
    "type": "event",
    "name": "openzeppelin_access::ownable::ownable::OwnableComponent::OwnershipTransferred",
    "kind": "struct",
    "members": [
      {
        "name": "previous_owner",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "new_owner",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      }
    ]
  },
  {
    "type": "event",
    "name": "openzeppelin_access::ownable::ownable::OwnableComponent::OwnershipTransferStarted",
    "kind": "struct",
    "members": [
      {
        "name": "previous_owner",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "new_owner",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      }
    ]
  },
  {
    "type": "event",
    "name": "openzeppelin_access::ownable::ownable::OwnableComponent::Event",
    "kind": "enum",
    "variants": [
      {
        "name": "OwnershipTransferred",
        "type": "openzeppelin_access::ownable::ownable::OwnableComponent::OwnershipTransferred",
        "kind": "nested"
      },
      {
        "name": "OwnershipTransferStarted",
        "type": "openzeppelin_access::ownable::ownable::OwnableComponent::OwnershipTransferStarted",
        "kind": "nested"
      }
    ]
  },
  {
    "type": "event",
    "name": "contracts::client::router_component::RouterComponent::Event",
    "kind": "enum",
    "variants": []
  },
  {
    "type": "event",
    "name": "contracts::client::gas_router_component::GasRouterComponent::Event",
    "kind": "enum",
    "variants": []
  },
  {
    "type": "event",
    "name": "contracts::client::mailboxclient_component::MailboxclientComponent::Event",
    "kind": "enum",
    "variants": []
  },
  {
    "type": "event",
    "name": "oif_starknet::hyperlane7683::Hyperlane7683::Event",
    "kind": "enum",
    "variants": [
      {
        "name": "Base7683Event",
        "type": "oif_starknet::base7683::Base7683Component::Event",
        "kind": "flat"
      },
      {
        "name": "BasicSwap7683Event",
        "type": "oif_starknet::basic_swap7683::BasicSwap7683Component::Event",
        "kind": "flat"
      },
      {
        "name": "OwnableEvent",
        "type": "openzeppelin_access::ownable::ownable::OwnableComponent::Event",
        "kind": "flat"
      },
      {
        "name": "RouterEvent",
        "type": "contracts::client::router_component::RouterComponent::Event",
        "kind": "flat"
      },
      {
        "name": "GasRouterEvent",
        "type": "contracts::client::gas_router_component::GasRouterComponent::Event",
        "kind": "flat"
      },
      {
        "name": "MailboxClientEvent",
        "type": "contracts::client::mailboxclient_component::MailboxclientComponent::Event",
        "kind": "flat"
      }
    ]
  }
] as const;
