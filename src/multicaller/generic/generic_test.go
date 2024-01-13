package generic_test

import (
	"PoolHelper/src/multicaller/generic"
	uniswapv2 "PoolHelper/src/pool/uniswap-v2"
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
	"testing"
)

const rawABI = `[{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall3.Call[]","name":"calls","type":"tuple[]"}],"name":"aggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes[]","name":"returnData","type":"bytes[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall3.Call[]","name":"calls","type":"tuple[]"}],"name":"aggregate3","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall3.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bool","name":"allowFailure","type":"bool"},{"internalType":"uint256","name":"value","type":"uint256"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall3.Call3Value[]","name":"calls","type":"tuple[]"}],"name":"aggregate3Value","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall3.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall3.Call[]","name":"calls","type":"tuple[]"}],"name":"blockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall3.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"getBasefee","outputs":[{"internalType":"uint256","name":"basefee","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"name":"getBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getBlockNumber","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getChainId","outputs":[{"internalType":"uint256","name":"chainid","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockCoinbase","outputs":[{"internalType":"address","name":"coinbase","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockDifficulty","outputs":[{"internalType":"uint256","name":"difficulty","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockGasLimit","outputs":[{"internalType":"uint256","name":"gaslimit","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentBlockTimestamp","outputs":[{"internalType":"uint256","name":"timestamp","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"addr","type":"address"}],"name":"getEthBalance","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getLastBlockHash","outputs":[{"internalType":"bytes32","name":"blockHash","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall3.Call[]","name":"calls","type":"tuple[]"}],"name":"tryAggregate","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall3.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall3.Call[]","name":"calls","type":"tuple[]"}],"name":"tryBlockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall3.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"}]`

type MockedClient struct {
	returnData []byte
}

func (m *MockedClient) CallContract(_ context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
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

}

func TestFetchTokens(t *testing.T) {

}
