package cache_test

import (
	"PoolHelper/src3/cache"
	"PoolHelper/src3/multicaller/generic"
	uniswapv2 "PoolHelper/src3/pool/uniswap"
	"PoolHelper/src3/token"
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"strings"
	"testing"
)

const publicRpc = "wss://ethereum.publicnode.com"

func TestFindPoolsV2_Count(t *testing.T) {
	initHash := "0x96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f"
	factory := common.HexToAddress("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f")
	_t := []common.Address{
		common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
		common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
		common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"),
		common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599"),
		common.HexToAddress("0x514910771af9ca656af840dff83e8264ecf986ca"),
	}

	m := cache.NewCache()
	for _, _t := range _t {
		err := m.AddToken(token.ERC20{
			Address:  _t,
			Decimals: big.NewInt(18),
		})
		if err != nil {
			panic(err)
		}
	}

	p := m.ImportV2Pools(factory, common.HexToHash(initHash))
	if len(p) != 10 {
		t.Errorf("wrong number of pools")
		t.Fatalf("expected %v, got %v", 10, len(p))
	}
}

func TestFindPoolsV3_Count(t *testing.T) {
	initHash := "0x96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f"
	factory := common.HexToAddress("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f")
	_t := []common.Address{
		common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
		common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
		common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"),
		common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599"),
		common.HexToAddress("0x514910771af9ca656af840dff83e8264ecf986ca"),
	}

	m := cache.NewCache()
	for _, _t := range _t {
		err := m.AddToken(token.ERC20{
			Address:  _t,
			Decimals: big.NewInt(18),
		})
		if err != nil {
			panic(err)
		}
	}

	feeTypes := []uniswapv2.FeeType{uniswapv2.MAX, uniswapv2.NORMAL, uniswapv2.LOW}

	p := m.ImportV3Pools(factory, common.HexToHash(initHash), feeTypes)
	if len(p) != 10*len(feeTypes) {
		t.Errorf("wrong number of pools")
		t.Fatalf("expected %v, got %v", 10*len(feeTypes), len(p))
	}
}

func TestInitializeTokens(t *testing.T) {
	// load raw abi
	const rawABI = `[{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"aggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes[]","name":"returnData","type":"bytes[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"aggregate3","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"uint256","name":"value","type":"uint256"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3Value[]","name":"calls","type":"tuple[]"}],"name":"aggregate3Value","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"blockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"getBasefee","outputs":[{"internalType":"uint256","name":"basefee","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"name":"getBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getBlockNumber","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getChainId","outputs":[{"internalType":"uint256","name":"chainid","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockCoinbase","outputs":[{"internalType":"address","name":"coinbase","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockDifficulty","outputs":[{"internalType":"uint256","name":"difficulty","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockGasLimit","outputs":[{"internalType":"uint256","name":"gaslimit","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockTimestamp","outputs":[{"internalType":"uint256","name":"timestamp","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"addr","type":"address"}],"name":"getEthBalance","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getLastBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"tryAggregate","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"tryBlockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"}]`
	cAbi, err := abi.JSON(strings.NewReader(rawABI))
	if err != nil {
		panic(err)
	}

	// client rpc
	client, err := ethclient.Dial(publicRpc)
	if err != nil {
		panic(err)
	}

	// multicaller
	caller := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")
	mcaller := generic.NewMulticall(
		caller,
		21_000,
		1_000_000,
		cAbi,
		client,
	)

	// newcache
	m := cache.NewCache()

	// to list
	_t := []common.Address{
		common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
		common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
		common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"),
		common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599"),
		common.HexToAddress("0x514910771af9ca656af840dff83e8264ecf986ca"),
	}

	// get latest block
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	// initialize tokens
	tokens, err := m.ImportTokens(context.Background(), _t, mcaller, block.Number().Uint64())
	if err != nil {
		panic(err)
	}

	// validate
	if len(tokens) != 5 {
		t.Errorf("wrong number of tokens")
		t.Fatalf("expected %v, got %v", 5, len(tokens))
	}

	for _, _token := range tokens {
		if _token.Symbol == "" {
			t.Errorf("_token symbol is empty")
			t.Fatalf("expected %v, got %v", false, _token.Symbol == "")
		}
		if _token.Decimals.Cmp(common.Big0) == 0 {
			t.Errorf("_token decimals is zero")
			t.Fatalf("expected %v, got %v", false, _token.Decimals.Cmp(common.Big0) == 0)
		}
		if _token.Name == "" {
			t.Errorf("_token name is empty")
			t.Fatalf("expected %v, got %v", false, _token.Name == "")
		}
		if _token.Address.Cmp(common.Address{}) == 0 {
			t.Errorf("_token address is empty")
			t.Fatalf("expected %v, got %v", false, _token.Address == common.Address{})
		}
	}

}

func TestInitializePools(t *testing.T) {
	// load raw abi
	const rawABI = `[{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"aggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes[]","name":"returnData","type":"bytes[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"aggregate3","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"uint256","name":"value","type":"uint256"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3Value[]","name":"calls","type":"tuple[]"}],"name":"aggregate3Value","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"blockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"getBasefee","outputs":[{"internalType":"uint256","name":"basefee","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"name":"getBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getBlockNumber","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getChainId","outputs":[{"internalType":"uint256","name":"chainid","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockCoinbase","outputs":[{"internalType":"address","name":"coinbase","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockDifficulty","outputs":[{"internalType":"uint256","name":"difficulty","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockGasLimit","outputs":[{"internalType":"uint256","name":"gaslimit","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockTimestamp","outputs":[{"internalType":"uint256","name":"timestamp","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"addr","type":"address"}],"name":"getEthBalance","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getLastBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"tryAggregate","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct MulticallContract.Call3[]","name":"calls","type":"tuple[]"}],"name":"tryBlockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct MulticallContract.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"}]`
	cAbi, err := abi.JSON(strings.NewReader(rawABI))
	if err != nil {
		panic(err)
	}

	// client rpc
	client, err := ethclient.Dial(publicRpc)
	if err != nil {
		panic(err)
	}

	// multicaller
	caller := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")
	mCaller := generic.NewMulticall(
		caller,
		21_000,
		21_000_000,
		cAbi,
		client,
	)

	// add tokens
	feeTiers := []uniswapv2.FeeType{uniswapv2.MAX, uniswapv2.NORMAL, uniswapv2.LOW, uniswapv2.MIN}
	_tokens := []common.Address{
		common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
		common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
		common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"),
		common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599"),
		common.HexToAddress("0x514910771af9ca656af840dff83e8264ecf986ca"),
	}

	m := cache.NewCache()
	for _, _t := range _tokens {
		err := m.AddToken(token.ERC20{
			Address:  _t,
			Decimals: big.NewInt(18),
		})
		if err != nil {
			panic(err)
		}
	}

	// get latest block
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	// initialize tokens
	_, err = m.ImportTokens(context.Background(), _tokens, mCaller, block.Number().Uint64())
	if err != nil {
		panic(err)
	}

	// add v3 pools
	initHashV3 := common.HexToHash("0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54")
	factoryV3 := common.HexToAddress("0x1f98431c8ad98523631ae4a59f267346ea31f984")
	m.ImportV3Pools(factoryV3, initHashV3, feeTiers)

	// add v2 pools
	initHashV2 := common.HexToHash("0x96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f")
	factoryV2 := common.HexToAddress("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f")
	m.ImportV2Pools(factoryV2, initHashV2)

	// initialize pools
	err = m.SyncAll(context.Background(), mCaller, block.Number().Uint64())
	if err != nil {
		panic(err)
	}

	// validate pool count
	if len(m.PoolsV2()) != 10 {
		t.Errorf("wrong number of pools")
		t.Fatalf("expected %v, got %v", 10, len(m.PoolsV2()))
	}
	if len(m.PoolsV3()) != 40 {
		t.Errorf("wrong number of pools")
		t.Fatalf("expected %v, got %v", 40, len(m.PoolsV3()))
	}

	// validate pool data
	for _, pool := range m.PoolsV2() {
		_p := pool.(*uniswapv2.V2Pool)
		_res, blockNum, timestamp := _p.State()
		if blockNum == 0 {
			t.Errorf("wrong block number")
			t.Fatalf("expected %v, got %v", false, blockNum == 0)
		}
		if timestamp == 0 {
			t.Errorf("wrong timestamp")
			t.Fatalf("expected %v, got %v", false, timestamp == 0)
		}
		if _res.Reserve0 == nil {
			t.Errorf("wrong reserve0")
			t.Fatalf("expected %v, got %v", false, _res.Reserve0 == nil)
		}
		if _res.Reserve1 == nil {
			t.Errorf("wrong reserve1")
			t.Fatalf("expected %v, got %v", false, _res.Reserve1 == nil)
		}
	}
	for _, pool := range m.PoolsV3() {
		_p := pool.(*uniswapv2.UniswapV3Pool)
		_s, blockNum, timestamp := _p.State()
		if blockNum == 0 {
			t.Errorf("wrong block number")
			t.Fatalf("expected %v, got %v", false, blockNum == 0)
		}
		if timestamp == 0 {
			t.Errorf("wrong timestamp")
			t.Fatalf("expected %v, got %v", false, timestamp == 0)
		}
		if _s.SqrtPriceX96 == nil {
			t.Errorf("wrong sqrt price")
			t.Fatalf("expected %v, got %v", false, _s.SqrtPriceX96 == nil)
		}
	}
}
