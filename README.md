# PoolHelper

PoolHelper is a Go application designed to interact with Ethereum blockchain data, specifically focusing on Uniswap and SushiSwap decentralized exchanges (DEXs). It aims to import tokens, initialize and sync pool reserves, and subscribe to new blocks for real-time updates. The application leverages Ethereum's smart contracts and multicall capabilities for efficient data retrieval and processing.

## Features

- **Token Importing**: Bulk import of multiple ERC-20 tokens into Uniswap V2 and V3, and SushiSwap pools.
- **Pool Initialization**: Initialize liquidity pools from specified DEX factories.
- **Reserve Synchronization**: Sync the reserves of each pool to get the latest state, helpful for obtaining the most recent liquidity and price data.
- **Block Subscription**: Listen for new blocks and update pool reserves in real-time, ensuring data remains current.

## Requirements

- Go (version 1.15 or higher)
- Ethereum Node or RPC service (e.g., Infura, Alchemy)

## Setup

1. **Ethereum RPC Endpoint**: Modify the `Endpoint` constant in the main.go file to your Ethereum node or RPC service URL.
2. **Token List and Factory Addresses**: Review and adjust the `tokenList`, `v2Factories`, and `v3Factories` slices to include the tokens and factory contracts you are interested in.
3. **Build the Project**:
    ```bash
    go build -o poolhelper .
    ```

## Usage

Run the compiled binary to start the application:

```bash
./poolhelper
