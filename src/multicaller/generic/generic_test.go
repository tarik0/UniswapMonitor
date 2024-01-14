package generic_test

import (
	"PoolHelper/src/multicaller/generic"
	uniswapv2 "PoolHelper/src/pool/uniswap-v2"
	uniswapv3 "PoolHelper/src/pool/uniswap-v3"
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
	"testing"
)

const rawABI = `[{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall.Call[]","name":"calls","type":"tuple[]"}],"name":"aggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes[]","name":"returnData","type":"bytes[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall.Call[]","name":"calls","type":"tuple[]"}],"name":"aggregate3","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"uint256","name":"value","type":"uint256"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall.Call3Value[]","name":"calls","type":"tuple[]"}],"name":"aggregate3Value","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall.Call[]","name":"calls","type":"tuple[]"}],"name":"blockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"getBasefee","outputs":[{"internalType":"uint256","name":"basefee","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"name":"getBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getBlockNumber","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getChainId","outputs":[{"internalType":"uint256","name":"chainid","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockCoinbase","outputs":[{"internalType":"address","name":"coinbase","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockDifficulty","outputs":[{"internalType":"uint256","name":"difficulty","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockGasLimit","outputs":[{"internalType":"uint256","name":"gaslimit","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockTimestamp","outputs":[{"internalType":"uint256","name":"timestamp","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"addr","type":"address"}],"name":"getEthBalance","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getLastBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall.Call[]","name":"calls","type":"tuple[]"}],"name":"tryAggregate","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall.Call[]","name":"calls","type":"tuple[]"}],"name":"tryBlockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"}]`

type MockedClient struct {
	returnData []byte
}

func (m *MockedClient) CallContract(_ context.Context, _ ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	return m.returnData, nil
}

func TestFetchReserves(t *testing.T) {
	// return struct type
	uint256Type, _ := abi.NewType("uint256", "", nil)
	returnDataType, _ := abi.NewType("bytes[]", "", nil)
	uint112Type, _ := abi.NewType("uint112", "", nil)
	uint32Type, _ := abi.NewType("uint32", "", nil)

	// create mocked mockedReserves
	mockedReserves := []uniswapv2.Reserves{
		{
			Reserve0: big.NewInt(1),
			Reserve1: big.NewInt(2),
		},
		{
			Reserve0: big.NewInt(3),
			Reserve1: big.NewInt(4),
		},
		{
			Reserve0: big.NewInt(5),
			Reserve1: big.NewInt(6),
		},
	}

	// create mocked result
	returnData := make([][]byte, 0)
	for _, reserve := range mockedReserves {
		// pack the mockedReserves
		packedRes, err := abi.Arguments{
			{
				Name: "reserve0",
				Type: uint112Type,
			},
			{
				Name: "reserve1",
				Type: uint112Type,
			},
			{
				Name: "blockTimestampLast",
				Type: uint32Type,
			},
		}.Pack(reserve.Reserve0, reserve.Reserve1, uint32(123))
		if err != nil {
			t.Fatal(err)
		}

		// append
		returnData = append(returnData, packedRes)
	}

	// pack the result
	packed, err := abi.Arguments{
		{
			Name: "blockNumber",
			Type: uint256Type,
		},
		{
			Name: "returnData",
			Type: returnDataType,
		},
	}.Pack(big.NewInt(123), returnData)
	if err != nil {
		t.Fatal(err)
	}

	// mocked client
	client := &MockedClient{
		returnData: packed,
	}

	// load abi
	cAbi, err := abi.JSON(strings.NewReader(rawABI))
	if err != nil {
		t.Fatal(err)
	}

	// multicaller
	m := generic.NewMulticall(
		common.BigToAddress(big.NewInt(1)),
		21_000,
		30_000_000,
		cAbi,
		client,
	)

	// fetch mockedReserves
	reserves, err := m.FetchReserves(context.Background(), []common.Address{
		common.BigToAddress(big.NewInt(1)),
		common.BigToAddress(big.NewInt(2)),
		common.BigToAddress(big.NewInt(3)),
	}, 123)
	if err != nil {
		t.Fatal(err)
	}

	// check the result
	if reserves[0].Reserve0.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("expected %v", big.NewInt(1))
		t.Fatalf("got %v", reserves[0].Reserve0)
	}
	if reserves[0].Reserve1.Cmp(big.NewInt(2)) != 0 {
		t.Errorf("expected %v", big.NewInt(2))
		t.Fatalf("got %v", reserves[0].Reserve1)
	}
}

func TestFetchSlots(t *testing.T) {
	// mocked slot0
	mockedSlots := []uniswapv3.Slot0{
		{
			SqrtPriceX96:               big.NewInt(1),
			Tick:                       big.NewInt(2),
			ObservationIndex:           big.NewInt(3),
			ObservationCardinality:     big.NewInt(4),
			ObservationCardinalityNext: big.NewInt(5),
			FeeProtocol:                big.NewInt(6),
			Unlocked:                   true,
		},
		{
			SqrtPriceX96:               big.NewInt(7),
			Tick:                       big.NewInt(8),
			ObservationIndex:           big.NewInt(9),
			ObservationCardinality:     big.NewInt(10),
			ObservationCardinalityNext: big.NewInt(11),
			FeeProtocol:                big.NewInt(12),
			Unlocked:                   true,
		},
		{
			SqrtPriceX96:               big.NewInt(13),
			Tick:                       big.NewInt(14),
			ObservationIndex:           big.NewInt(15),
			ObservationCardinality:     big.NewInt(16),
			ObservationCardinalityNext: big.NewInt(17),
			FeeProtocol:                big.NewInt(18),
			Unlocked:                   true,
		},
	}

	// create mocked result
	uint256, _ := abi.NewType("uint256", "", nil)
	returnData := make([][]byte, 0)
	for _, slot := range mockedSlots {
		tmp := 0
		if slot.Unlocked {
			tmp = 1
		}

		// pack the mockedReserves
		packedRes, err := abi.Arguments{
			{
				Name: "sqrtPriceX96",
				Type: uint256,
			},
			{
				Name: "tick",
				Type: uint256,
			},
			{
				Name: "observationIndex",
				Type: uint256,
			},
			{
				Name: "observationCardinality",
				Type: uint256,
			},
			{
				Name: "observationCardinalityNext",
				Type: uint256,
			},
			{
				Name: "feeProtocol",
				Type: uint256,
			},
			{
				Name: "unlocked",
				Type: uint256,
			},
		}.Pack(
			slot.SqrtPriceX96,
			slot.Tick,
			slot.ObservationIndex,
			slot.ObservationCardinality,
			slot.ObservationCardinalityNext,
			slot.FeeProtocol,
			big.NewInt(int64(tmp)),
		)
		if err != nil {
			t.Fatal(err)
		}

		// append
		returnData = append(returnData, packedRes)
	}

	// pack the result
	returnDataType, _ := abi.NewType("bytes[]", "", nil)
	packed, err := abi.Arguments{
		{
			Name: "blockNumber",
			Type: uint256,
		},
		{
			Name: "returnData",
			Type: returnDataType,
		},
	}.Pack(big.NewInt(123), returnData)
	if err != nil {
		t.Fatal(err)
	}

	// mocked client
	client := &MockedClient{
		returnData: packed,
	}

	// load abi
	cAbi, err := abi.JSON(strings.NewReader(rawABI))
	if err != nil {
		t.Fatal(err)
	}

	// multicaller
	m := generic.NewMulticall(
		common.BigToAddress(big.NewInt(1)),
		21_000,
		30_000_000,
		cAbi,
		client,
	)

	// fetch slots
	slots, err := m.FetchSlots(context.Background(), []common.Address{
		common.BigToAddress(big.NewInt(1)),
		common.BigToAddress(big.NewInt(2)),
		common.BigToAddress(big.NewInt(3)),
	}, 123)
	if err != nil {
		t.Fatal(err)
	}

	// check the result
	if len(slots) != 3 {
		t.Errorf("expected %v", 3)
		t.Fatalf("got %v", len(slots))
	}
	if slots[0].SqrtPriceX96.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("expected %v", big.NewInt(1))
		t.Fatalf("got %v", slots[0].SqrtPriceX96)
	}
	if slots[0].Tick.Cmp(big.NewInt(2)) != 0 {
		t.Errorf("expected %v", big.NewInt(2))
		t.Fatalf("got %v", slots[0].Tick)
	}
	if slots[0].ObservationIndex.Cmp(big.NewInt(3)) != 0 {
		t.Errorf("expected %v", big.NewInt(3))
		t.Fatalf("got %v", slots[0].ObservationIndex)
	}
}

func TestFetchTokens(t *testing.T) {
	// mocked tokens
	mockedTokens := []common.Address{
		common.BigToAddress(big.NewInt(1)),
		common.BigToAddress(big.NewInt(2)),
		common.BigToAddress(big.NewInt(3)),
	}

	// types
	uint256Type, _ := abi.NewType("uint256", "", nil)
	returnDataType, _ := abi.NewType("bytes[]", "", nil)
	stringType, _ := abi.NewType("string", "", nil)
	uint8Type, _ := abi.NewType("uint8", "", nil)

	// create mocked result
	returnData := make([][]byte, 0)
	for range mockedTokens {
		// pack name
		packedName, err := abi.Arguments{
			{
				Name: "name",
				Type: stringType,
			},
		}.Pack("name")
		if err != nil {
			t.Fatal(err)
		}

		// pack symbol
		packedSymbol, err := abi.Arguments{
			{
				Name: "symbol",
				Type: stringType,
			},
		}.Pack("symbol")

		// pack decimals
		packedDecimals, err := abi.Arguments{
			{
				Name: "decimals",
				Type: uint8Type,
			},
		}.Pack(uint8(18))

		// append
		returnData = append(returnData, packedDecimals, packedSymbol, packedName)
	}

	// pack the result
	packed, err := abi.Arguments{
		{
			Name: "blockNumber",
			Type: uint256Type,
		},
		{
			Name: "returnData",
			Type: returnDataType,
		},
	}.Pack(big.NewInt(123), returnData)
	if err != nil {
		t.Fatal(err)
	}

	// mocked client
	client := &MockedClient{
		returnData: packed,
	}

	// load abi
	cAbi, err := abi.JSON(strings.NewReader(rawABI))
	if err != nil {
		t.Fatal(err)
	}

	// multicaller
	m := generic.NewMulticall(
		common.BigToAddress(big.NewInt(1)),
		21_000,
		30_000_000,
		cAbi,
		client,
	)

	// fetch tokens
	tokens, err := m.FetchTokens(context.Background(), []common.Address{
		common.BigToAddress(big.NewInt(1)),
		common.BigToAddress(big.NewInt(2)),
		common.BigToAddress(big.NewInt(3)),
	}, 123)
	if err != nil {
		t.Fatal(err)
	}

	// check the result
	if len(tokens) != 3 {
		t.Errorf("expected %v", 3)
		t.Fatalf("got %v", len(tokens))
	}
	if tokens[0].Name != "name" {
		t.Errorf("expected %v", "name")
		t.Fatalf("got %v", tokens[0].Name)
	}
	if tokens[0].Symbol != "symbol" {
		t.Errorf("expected %v", "symbol")
		t.Fatalf("got %v", tokens[0].Symbol)
	}
	if tokens[0].Decimals.Cmp(big.NewInt(18)) != 0 {
		t.Errorf("expected %v", 18)
		t.Fatalf("got %v", tokens[0].Decimals)
	}
}
