package main

import (
	"PoolHelper/src/cache"
	"PoolHelper/src/multicaller/generic"
	uniswap_v3 "PoolHelper/src/pool/uniswap-v3"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"strings"
	"time"
)

func main() {
	// load abi
	const rawABI = `[{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"aggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes[]","name":"returnData","type":"bytes[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"aggregate3","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"uint256","name":"value","type":"uint256"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3Value[]","name":"calls","type":"tuple[]"}],"name":"aggregate3Value","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"blockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"getBasefee","outputs":[{"internalType":"uint256","name":"basefee","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"name":"getBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getBlockNumber","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getChainId","outputs":[{"internalType":"uint256","name":"chainid","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockCoinbase","outputs":[{"internalType":"address","name":"coinbase","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockDifficulty","outputs":[{"internalType":"uint256","name":"difficulty","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockGasLimit","outputs":[{"internalType":"uint256","name":"gaslimit","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockTimestamp","outputs":[{"internalType":"uint256","name":"timestamp","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"addr","type":"address"}],"name":"getEthBalance","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getLastBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"tryAggregate","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"tryBlockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"}]`
	cAbi, err := abi.JSON(strings.NewReader(rawABI))
	if err != nil {
		panic(err)
	}

	// connect to RPC client
	client, err := ethclient.Dial("wss://ethereum.publicnode.com")

	// create multicall
	m := generic.NewMulticall(
		common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
		21_000,
		30_000_000,
		cAbi,
		client,
	)

	// create cache
	c := cache.NewCache()

	// get latest block
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	// import tokens
	_t := []common.Address{
		common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
		common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
		common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"),
		common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599"),
		common.HexToAddress("0x514910771af9ca656af840dff83e8264ecf986ca"),
		common.HexToAddress("0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"),
		common.HexToAddress("0x85f17cf997934a597031b2e18a9ab6ebd4b9f6a4"),
		common.HexToAddress("0xc5f0f7b66764F6ec8C8Dff7BA683102295E16409"),
		common.HexToAddress("0xf57e7e7c23978c3caec3c3548e3d615c346e79ff"),
		common.HexToAddress("0xB50721BCf8d664c30412Cfbc6cf7a15145234ad1"),
		common.HexToAddress("0xa0b73e1ff0b80914ab6fe0444e65848c4c34450b"),
	}
	_tokens, err := c.ImportTokens(context.Background(), _t, m, block.Number().Uint64())
	if err != nil {
		panic(err)
	}
	for _, token := range _tokens {
		fmt.Printf("Imported token: %s %v\n", token.Symbol, token.Address.Hex())
	}

	time.Sleep(3 * time.Second)

	// import v2 pools
	_pools := c.ImportV2Pools(
		common.HexToAddress("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f"),
		common.HexToHash("0x96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f"),
	)
	fmt.Println("")
	for _, p := range _pools {
		fmt.Printf("Imported V2 pool: %s\n", p.String())
	}

	// import v2 pools (sushiswap)
	_pools = c.ImportV2Pools(
		common.HexToAddress("0xc0aee478e3658e2610c5f7a4a2e1777ce9e4f2ac"),
		common.HexToHash("0xe18a34eb0e04b04f7a0ac29a6e80748dca96319b42c54d679cb821dca90c6303"),
	)
	fmt.Println("")
	for _, p := range _pools {
		fmt.Printf("Imported V2 pool: %s\n", p.String())
	}

	// import v3 pools
	_pools = c.ImportV3Pools(
		common.HexToAddress("0x1f98431c8ad98523631ae4a59f267346ea31f984"),
		common.HexToHash("0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54"),
		[]uniswap_v3.FeeType{uniswap_v3.NORMAL, uniswap_v3.LOW, uniswap_v3.MIN},
	)
	fmt.Println("")
	for _, p := range _pools {
		fmt.Printf("Imported V3 pool: %s\n", p.String())
	}

	// sync
	err, dur := c.SyncAll(context.Background(), m, block.Number().Uint64())
	if err != nil {
		panic(err)
	}

	// print v2 pool states
	fmt.Println("")
	for _, p := range c.PoolsV2() {
		s, _, _ := p.State()

		// skip empty pool
		if s.Reserve0.Sign() == 0 || s.Reserve1.Sign() == 0 {
			continue
		}

		_pair := p.Pair()
		fmt.Printf("V2 Pool: (%s) %s\n", _pair.String(), p.Address().Hex())
	}

	// print v3 pool states
	fmt.Println("")
	for _, p := range c.PoolsV3() {
		s, _, _ := p.State()

		// skip empty pool
		if s.SqrtPriceX96.Sign() == 0 {
			continue
		}

		_pair := p.Pair()
		fmt.Printf("V3 Pool: (%s) %s\n", _pair.String(), p.Address().Hex())
	}

	// print sync time
	fmt.Printf("SyncAll time: %v\n", dur)
	fmt.Printf("Synced cache to block %d\n", block.Number().Uint64())
	fmt.Printf("Pool count: %d\n", len(c.PoolsV2())+len(c.PoolsV3()))
	fmt.Printf("Token count: %d\n", len(c.Tokens()))

	// get pools
	var poolAddrs []common.Address
	pools := c.PoolsV2()
	for i := 0; i < 2 && i < len(pools); i++ {
		poolAddrs = append(poolAddrs, pools[i].Address())
	}

	// sync for the first two pairs
	err, dur = c.Sync(context.Background(), m, poolAddrs, block.Number().Uint64())
	if err != nil {
		panic(err)
	}

	// print sync time
	fmt.Println("Partial sync")
	fmt.Printf("Sync time: %v\n", dur)
	fmt.Printf("Synced cache to block %d\n", block.Number().Uint64())
}
