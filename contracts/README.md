## Foundry

**Foundry is a blazing fast, portable and modular toolkit for Ethereum application development written in Rust.**

Foundry consists of:

- **Forge**: Ethereum testing framework (like Truffle, Hardhat and DappTools).
- **Cast**: Swiss army knife for interacting with EVM smart contracts, sending transactions and getting chain data.
- **Anvil**: Local Ethereum node, akin to Ganache, Hardhat Network.
- **Chisel**: Fast, utilitarian, and verbose solidity REPL.

## Documentation

https://book.getfoundry.sh/

## Usage

### Build

```shell
# auth first: gh auth login
$ forge build --extra-output-files abi
```

### Test

```shell
$ forge test
$ forge test --match-path test/Permit2Vault.t.sol
```

### Format

```shell
$ forge fmt
```

### Gas Snapshots

```shell
$ forge snapshot
```


## 合约
### 部署
```
forge script --broadcast \
--rpc-url <RPC-URL> \
--private-key <PRIVATE_KEY> \
--sig 'run()' \
script/deployParameters/Deploy<network>.s.sol:Deploy<network>
```
### 升级
```
forge script --broadcast \
--rpc-url <RPC-URL> \
--private-key <PRIVATE_KEY> \
--sig 'run()' \
script/deployParameters/Deploy<network>.s.sol:Upgrade<network>
```

### example
```shell
anvil --fork-url https://eth-mainnet.g.alchemy.com/v2/0aoAtW5IQvhhwLgW4wFQFbW7eM4czhOb --fork-block-number 22608355 --port 8546
anvil --fork-url https://arb-mainnet.g.alchemy.com/v2/0aoAtW5IQvhhwLgW4wFQFbW7eM4czhOb --fork-block-number 342773642 --port 8546

# deploy
forge script --broadcast \
--rpc-url http://localhost:8546 \
--private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
--sig 'run()' \
script/deployParameters/DeployBase.s.sol:DeployBase

# upgrade
forge script --broadcast \
--rpc-url http://localhost:8546 \
--private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
--sig 'run()' \
script/deployParameters/DeployBase.s.sol:UpgradeBase

# tenderly verify
forge verify-contract \
  --num-of-optimizations 20000 \
  --compiler-version v0.8.23 \
  0x0b36229371E14CfE9e6745C6B5b17E2394202bB9 \
  SettlementContract \
  --show-standard-json-input \
  > standard-json-input.json
```

### forge inspect FlashLoanLiquidation abi       

╭-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------╮
| Type        | Signature                                                                         | Selector                                                           |
+======================================================================================================================================================================+
| event       | Initialized(uint64)                                                               | 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2 |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| event       | Liquidation(address,address,uint256,address,uint256,uint256)                      | 0x433a748a5cc0601caaac2885db271290bee092018cdcfb93e7652eba112f598f |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| event       | OwnershipTransferred(address,address)                                             | 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0 |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| event       | Upgraded(address)                                                                 | 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| error       | AddressEmptyCode(address)                                                         | 0x9996b315                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| error       | ERC1967InvalidImplementation(address)                                             | 0x4c9c8ce3                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| error       | ERC1967NonPayable()                                                               | 0xb398979f                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| error       | FailedCall()                                                                      | 0xd6bda275                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| error       | InvalidInitialization()                                                           | 0xf92ee8a9                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| error       | NotInitializing()                                                                 | 0xd7e6bcf8                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| error       | OwnableInvalidOwner(address)                                                      | 0x1e4fbdf7                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| error       | OwnableUnauthorizedAccount(address)                                               | 0x118cdaa7                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| error       | UUPSUnauthorizedCallContext()                                                     | 0xe07c8dba                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| error       | UUPSUnsupportedProxiableUUID(bytes32)                                             | 0xaa1d49a4                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | ADDRESSES_PROVIDER() view returns (IPoolAddressesProvider)                        | 0x0542975c                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | POOL() view returns (IPool)                                                       | 0x7535d246                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | UPGRADE_INTERFACE_VERSION() view returns (string)                                 | 0xad3cb1cc                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | aave_v3_pool() view returns (IPool)                                               | 0xf5f8bbbc                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | dex() view returns (IDex)                                                         | 0x692058c2                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | executeLiquidation(address,address,address,uint256) nonpayable                    | 0x05c3786d                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | executeOperation(address,uint256,uint256,address,bytes) nonpayable returns (bool) | 0x1b11d0ff                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | initialize(address,address,address) nonpayable                                    | 0xc0c53b8b                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | owner() view returns (address)                                                    | 0x8da5cb5b                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | proxiableUUID() view returns (bytes32)                                            | 0x52d1902d                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | renounceOwnership() nonpayable                                                    | 0x715018a6                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | transferOwnership(address) nonpayable                                             | 0xf2fde38b                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | upgradeToAndCall(address,bytes) payable                                           | 0x4f1ef286                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | usdc() view returns (address)                                                     | 0x3e413bee                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| function    | withdrawToken(address,uint256) nonpayable                                         | 0x9e281a98                                                         |
|-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------|
| constructor | constructor() nonpayable                                                          |                                                                    |
╰-------------+-----------------------------------------------------------------------------------+--------------------------------------------------------------------╯


## Prod
### deploy
```
export ETH_RPC_URL=Your URL
export PRIVATE_KEY=Your Private key

<!-- eth -->
forge script --broadcast \
--rpc-url $ETH_RPC_URL \
--private-key $PRIVATE_KEY \
--sig 'run()' \
script/deployParameters/DeployETH.s.sol:DeployETH

<!-- base -->
forge script --broadcast \
--rpc-url $BASE_RPC_URL \
--private-key $PRIVATE_KEY \
--sig 'run()' \
script/deployParameters/DeployBase.s.sol:DeployBase

<!-- op -->
forge script --broadcast \
--rpc-url $OP_RPC_URL \
--private-key $PRIVATE_KEY \
--sig 'run()' \
script/deployParameters/DeployOP.s.sol:DeployOP

<!-- arb -->
forge script --broadcast \
--rpc-url $ARB_RPC_URL \
--private-key $PRIVATE_KEY \
--sig 'run()' \
script/deployParameters/DeployARB.s.sol:DeployARB

<!-- avax -->
forge script --broadcast \
--rpc-url $AVAX_RPC_URL \
--private-key $PRIVATE_KEY \
--sig 'run()' \
script/deployParameters/DeployAVAX.s.sol:DeployAVAX
```

### upgrade
```
export ETH_RPC_URL=Your URL
export PRIVATE_KEY=Your Private key

<!-- eth -->
forge script --broadcast \
--rpc-url $ETH_RPC_URL \
--private-key $PRIVATE_KEY \
--sig 'run()' \
script/deployParameters/DeployETH.s.sol:UpgradeETH

<!-- base -->
forge script --broadcast \
--rpc-url $BASE_RPC_URL \
--private-key $PRIVATE_KEY \
--sig 'run()' \
script/deployParameters/DeployBase.s.sol:DeployBase

<!-- op -->
forge script --broadcast \
--rpc-url $OP_RPC_URL \
--private-key $PRIVATE_KEY \
--sig 'run()' \
script/deployParameters/DeployOP.s.sol:DeployOP

<!-- arb -->
forge script --broadcast \
--rpc-url $ARB_RPC_URL \
--private-key $PRIVATE_KEY \
--sig 'run()' \
script/deployParameters/DeployARB.s.sol:DeployARB

<!-- avax -->
forge script --broadcast \
--rpc-url $AVAX_RPC_URL \
--private-key $PRIVATE_KEY \
--sig 'run()' \
script/deployParameters/DeployAVAX.s.sol:DeployAVAX
```

