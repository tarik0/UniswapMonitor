package generic_test

import (
	"PoolHelper/src3/multicaller/generic"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func TestMulticallContract_SplitCalls(t *testing.T) {
	callCost, maxGas := uint64(1), uint64(10)
	m := generic.NewMulticall(common.BigToAddress(big.NewInt(0)), callCost, maxGas, abi.ABI{}, nil)

	var calls []generic.Call3
	for i := 0; i < 101; i++ {
		calls = append(calls, generic.Call3{
			Target:       common.BigToAddress(big.NewInt(int64(i))),
			CallData:     []byte{},
			AllowFailure: false,
		})
	}

	chunks := m.SplitCalls(calls)
	if len(chunks) != 11 {
		t.Errorf("wrong number of chunks")
	}
	if len(chunks[len(chunks)-1]) != 1 {
		t.Errorf("wrong number of calls in last chunk")
	}
}
