package main

import (
	"PoolHelper/src/cache/uniswap"
	"PoolHelper/src/multicall/generic"
	unipool "PoolHelper/src/pool/uniswap"
	"PoolHelper/src/structs/factory"
	"PoolHelper/src/structs/subscription"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"strings"
	"time"
)

///
/// Constants & Utils
///

const (
	Endpoint   = "wss://eth-mainnet.g.alchemy.com/v2/bruh"
	Timeout    = 20 * time.Second
	MaxTimeout = 30 * time.Second
	MaxRetries = 5
	CallCost   = 25_000
	MaxGas     = 30_000_000
)

var tokenList = []common.Address{
	common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"), // Tether USD (USDT)
	common.HexToAddress("0xb8c77482e45f1f44de1745f52c74426c631bdd52"), // Binance Coin (BNB)
	common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"), // USD Coin (USDC)
	common.HexToAddress("0xae7ab96520de3a18e5e111b5eaab095312d7fe84"), // Lido Staked Ether (STETH)
	common.HexToAddress("0x514910771af9ca656af840dff83e8264ecf986ca"), // Chainlink (LINK)
	common.HexToAddress("0x582d872a1b094fc48f5de31d3b73f2d9be47def1"), // The Open Network (TON)
	common.HexToAddress("0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0"), // Matic Network (MATIC)
	common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599"), // Wrapped Bitcoin (WBTC)
	common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"), // Dai (DAI)
	common.HexToAddress("0x95ad61b0a150d79219dcf64e1e6cc01f0b64c4ce"), // Shiba Inu (SHIB)
	common.HexToAddress("0x2af5d2ad76741191d15dfe7bf6ac92d4bd912ca3"), // LEO Token (LEO)
	common.HexToAddress("0x1f9840a85d5af5bf1d1762f925bdaddc4201f984"), // Uniswap (UNI)
	common.HexToAddress("0xe28b3b32b6c345a34ff64674606124dd5aceca30"), // Injective Protocol (INJ)
	common.HexToAddress("0x75231f58b43240c9718dd58b4967c5114342a86c"), // OKB (OKB)
	common.HexToAddress("0x5a98fcbea516cf06857215779fd812ca3bef1b32"), // Lido DAO (LDO)
	common.HexToAddress("0xc5f0f7b66764f6ec8c8dff7ba683102295e16409"), // First Digital USD (FDUSD)
	common.HexToAddress("0xf57e7e7c23978c3caec3c3548e3d615c346e79ff"), // Immutable X (IMX)
	common.HexToAddress("0xa0b73e1ff0b80914ab6fe0444e65848c4c34450b"), // Crypto.com Coin (CRO)
	common.HexToAddress("0x3c3a81e81dc49a522a592e7622a7e711c06bf354"), // Mantle (MNT)
	common.HexToAddress("0xa2e3356610840701bdf5611a53974510ae27e2e1"), // Wrapped Beacon ETH (WBETH)
	common.HexToAddress("0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2"), // Maker (MKR)
	common.HexToAddress("0x0000000000085d4780b73119b644ae5ecd22b376"), // TrueUSD (TUSD)
	common.HexToAddress("0x6de037ef9ad2725eb40118bb1702ebb27e4aeb24"), // Render Token (RNDR)
	common.HexToAddress("0xc944e90c64b2c07662a292be6244bdf05cda44a7"), // The Graph (GRT)
	common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393"), // Rocket Pool ETH (RETH)
	common.HexToAddress("0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9"), // Aave (AAVE)
	common.HexToAddress("0x4a220e6096b25eadb88358cb44068a3248254675"), // Quant (QNT)
	common.HexToAddress("0x667102bd3413bfeaa3dffb48fa8288819e480a88"), // Tokenize Xchange (TKX)
	common.HexToAddress("0x3845badade8e6dff049820680d1f14bd3903a5d0"), // The Sandbox (SAND)
	common.HexToAddress("0xbb0e17ef65f82ab018d8edd776e8dd940327b28b"), // Axie Infinity (AXS)
	common.HexToAddress("0xc011a73ee8576fb46f5e1c5751ca3b9fe0af2a6f"), // Synthetix Network Token (SNX)
	common.HexToAddress("0xf34960d9d60be18cc1d5afc1a6f012a723a28811"), // KuCoin Token (KCS)
	common.HexToAddress("0x3506424f91fd33084466f402d5d97f05f8e3b4af"), // Chiliz (CHZ)
	common.HexToAddress("0x62d0a8458ed7719fdaf978fe5929c6d342b0bfce"), // Beam (BEAM)
	common.HexToAddress("0x925206b8a707096ed26ae47c84747fe0bb734f59"), // WBT (WBT)
	common.HexToAddress("0x50d1c9771902476076ecfc8b2a83ad6b9355a4c9"), // FTX Token (FTT)
	common.HexToAddress("0x0f5d2fb29fb7d3cfee444a200298f468908cc942"), // Decentraland (MANA)
	common.HexToAddress("0x92d6c1e31e14520e676a687f0a93788b716beff5"), // dYdX (DYDX)
	common.HexToAddress("0x19de6b897ed14a376dda0fe53a5420d2ac828a28"), // Bitget Token (BGB)
	common.HexToAddress("0x5283d291dbcf85356a21ba090e6db59121208b44"), // Blur (BLUR)
	common.HexToAddress("0x0c356b7fd36a5357e5a017ef11887ba100c9ab76"), // Kava.io (KAVA)
	common.HexToAddress("0x3432b6a60d23ca0dfca7761b7ab56459d9c964d0"), // Frax Share (FXS)
	common.HexToAddress("0x15d4c048f83bd7e37d49ea4c83a07267ec4203da"), // Gala (GALA)
	common.HexToAddress("0x0c10bf8fcb7bf5412187a595ab97a3609160b5c6"), // Decentralized USD (USDD)
	common.HexToAddress("0x26b80fbfc01b71495f477d5237071242e0d959d7"), // Wrapped ROSE (wROSE)
	common.HexToAddress("0x5e8422345238f34275888049021821e8e08caa1f"), // Frax Ether (FRXETH)
	common.HexToAddress("0x853d955acef822db058eb8505911ed77f175b99e"), // Frax (FRAX)
	common.HexToAddress("0xd1d2eb1b1e90b638588728b4130137d262c87cae"), // Gala (GALA)
	common.HexToAddress("0x152649ea73beab28c5b49b26eb48f7ead6d4c898"), // PancakeSwap Token (Cake)

}

var v2Factories = []factory.Factory[any]{
	{
		Name:     "Uniswap V2",
		Address:  common.HexToAddress("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f"),
		InitHash: common.HexToHash("0x96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f"),
	},
	{
		Name:     "SushiSwap",
		Address:  common.HexToAddress("0xc0aee478e3658e2610c5f7a4a2e1777ce9e4f2ac"),
		InitHash: common.HexToHash("0xe18a34eb0e04b04f7a0ac29a6e80748dca96319b42c54d679cb821dca90c6303"),
	},
	{
		Name:     "FraxSwap",
		Address:  common.HexToAddress("0xC14d550632db8592D1243Edc8B95b0Ad06703867"),
		InitHash: common.HexToHash("0x4ce0b4ab368f39e4bd03ec712dfc405eb5a36cdb0294b3887b441cd1c743ced3"),
	},
}

var v3Factories = []factory.Factory[unipool.V3FeeType]{
	{
		Name:     "Uniswap V3",
		Address:  common.HexToAddress("0x1f98431c8ad98523631ae4a59f267346ea31f984"),
		InitHash: common.HexToHash("0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54"),
		FeeTypes: []unipool.V3FeeType{
			unipool.MIN,
			unipool.LOW,
			unipool.NORMAL,
			unipool.MAX,
		},
	},
}

func newCaller(c *ethclient.Client) *generic.MulticallContract {
	// load abi
	const rawABI = `[{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"aggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes[]","name":"returnData","type":"bytes[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"aggregate3","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"uint256","name":"value","type":"uint256"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3Value[]","name":"calls","type":"tuple[]"}],"name":"aggregate3Value","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"blockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"getBasefee","outputs":[{"internalType":"uint256","name":"basefee","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"name":"getBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getBlockNumber","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getChainId","outputs":[{"internalType":"uint256","name":"chainid","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockCoinbase","outputs":[{"internalType":"address","name":"coinbase","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockDifficulty","outputs":[{"internalType":"uint256","name":"difficulty","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockGasLimit","outputs":[{"internalType":"uint256","name":"gaslimit","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockTimestamp","outputs":[{"internalType":"uint256","name":"timestamp","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"addr","type":"address"}],"name":"getEthBalance","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getLastBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"tryAggregate","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"tryBlockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"}]`
	cAbi, err := abi.JSON(strings.NewReader(rawABI))
	if err != nil {
		panic(err)
	}

	return generic.NewCaller(
		common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
		CallCost,
		MaxGas,
		cAbi,
		c,
	)
}

///
/// Main
///

func main() {
	// connect to RPC client
	rpcClient, err := rpc.Dial(Endpoint)
	if err != nil {
		panic(err)
	}
	client := ethclient.NewClient(rpcClient)

	// create multicall
	m := newCaller(client)

	// create caches
	cV2 := uniswap.NewV2Cache()
	cV3 := uniswap.NewV3Cache()

	// get the latest block
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("=========================================")
	fmt.Println("=             Import Tokens             =")
	fmt.Println("=========================================")

	// import tokens for each cache
	importStart := time.Now()
	if err := cV2.ImportTokens(context.Background(), m, tokenList); err != nil {
		panic(err)
	}
	fmt.Printf("(V2) Imported %d tokens in %s\n", len(tokenList), time.Since(importStart))
	importStart = time.Now()
	if err := cV3.ImportTokens(context.Background(), m, tokenList); err != nil {
		panic(err)
	}
	fmt.Printf("(V3) Imported %d tokens in %s\n", len(tokenList), time.Since(importStart))
	fmt.Println()

	fmt.Println("=========================================")
	fmt.Println("=            Initialize Pools           =")
	fmt.Println("=========================================")

	// initialize pools for each cache
	oldCount := 0
	for _, f := range v2Factories {
		initStart := time.Now()
		if err := cV2.InitializePools(f); err != nil {
			panic(err)
		}
		fmt.Printf("(V2) Initialized %d pools for %s in %s\n", len(cV2.Pools())-oldCount, f.Name, time.Since(initStart))
		oldCount = len(cV2.Pools())
	}
	oldCount = 0
	for _, f := range v3Factories {
		initStart := time.Now()
		if err := cV3.InitializePools(f); err != nil {
			panic(err)
		}
		fmt.Printf("(V3) Initialized %d pools for %s in %s\n", len(cV2.Pools())-oldCount, f.Name, time.Since(initStart))
		oldCount = len(cV2.Pools())
	}
	fmt.Println("Total pools:", len(cV2.Pools())+len(cV3.Pools()))
	fmt.Println("Total V2 pools:", len(cV2.Pools()))
	fmt.Println("Total V3 pools:", len(cV3.Pools()))
	fmt.Println()

	fmt.Println("=========================================")
	fmt.Println("=             Sync Reserves             =")
	fmt.Println("=========================================")
	fmt.Printf("Syncing reserves for block %d\n", block.NumberU64())

	// sync reserves for each cache
	syncStart := time.Now()
	if err := cV2.SyncAll(context.Background(), m, block.NumberU64()); err != nil {
		panic(err)
	}
	fmt.Printf("(V2) Synced %d pools in %s\n", len(cV2.Pools()), time.Since(syncStart))
	syncStart = time.Now()
	if err := cV3.SyncAll(context.Background(), m, block.NumberU64()); err != nil {
		panic(err)
	}
	fmt.Printf("(V3) Synced %d pools in %s\n", len(cV3.Pools()), time.Since(syncStart))
	fmt.Println()

	fmt.Println("=========================================")
	fmt.Println("=          Subscribe to Blocks          =")
	fmt.Println("=========================================")

	// create subscription
	sub := subscription.NewBlockSubscription(rpcClient, Timeout, MaxTimeout, MaxRetries)
	if err = sub.Subscribe(context.Background()); err != nil {
		panic(err)
	}

	// listen for new blocks
	lastBlock := block.NumberU64()
	for {
		select {
		case _err := <-sub.Err():
			fmt.Println(fmt.Errorf("subscription error: %s", _err))
		case header := <-sub.Items():
			// check if the block number has changed
			if header.Item.Number.Uint64() == lastBlock {
				continue
			}

			// update the last block number
			lastBlock = header.Item.Number.Uint64()
			headerCtx := header.Context

			fmt.Println("")
			fmt.Println("=========================================")
			fmt.Println("Block:", lastBlock)

			// sync reserves for each cache
			syncStart = time.Now()
			if err := cV2.SyncAll(headerCtx, m, lastBlock); err != nil {
				if errors.Is(err, context.Canceled) {
					fmt.Println("block passed")
					continue
				}
				panic(err)
			}
			fmt.Printf("(V2) Synced %d pools in %s\n", len(cV2.Pools()), time.Since(syncStart))
			syncStart = time.Now()
			if err := cV3.SyncAll(context.Background(), m, lastBlock); err != nil {
				if errors.Is(err, context.Canceled) {
					fmt.Println("block passed")
					continue
				}
				panic(err)
			}
			fmt.Printf("(V3) Synced %d pools in %s\n", len(cV3.Pools()), time.Since(syncStart))
		}
	}
}
