package generic

import (
	"PoolHelper/src/multicall"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

///
/// Call3
///

type Call3 struct {
	Target       common.Address
	CallData     []byte
	AllowFailure bool
}

type Result struct {
	Block      uint64
	ReturnData []byte
}

type Multicall multicall.Multicaller[Call3, Result]

///
/// MulticallContract
///

type MulticallContract struct {
	contract common.Address
	callCost uint64
	maxGas   uint64
	cAbi     abi.ABI
	client   multicall.ClientDispatcher
}

func NewCaller(contract common.Address, callCost uint64, maxGas uint64, cAbi abi.ABI, client multicall.ClientDispatcher) *MulticallContract {
	return &MulticallContract{
		contract: contract,
		callCost: callCost,
		maxGas:   maxGas,
		cAbi:     cAbi,
		client:   client,
	}
}

func (m *MulticallContract) SplitCalls(calls []Call3) [][]Call3 {
	callChunks := make([][]Call3, 0)
	callChunk := make([]Call3, 0)
	callChunkGas := uint64(0)

	for _, call := range calls {
		// check if adding this call would exceed the max gas limit for the chunk
		if callChunkGas+m.callCost > m.maxGas {
			// if it exceeds, start a newcache chunk
			callChunks = append(callChunks, callChunk)
			callChunk = make([]Call3, 0)
			callChunkGas = 0
		}

		// add call to the current chunk and update the gas estimate
		callChunk = append(callChunk, call)
		callChunkGas += m.callCost
	}

	// add the last chunk if it contains any calls
	if len(callChunk) > 0 {
		callChunks = append(callChunks, callChunk)
	}

	return callChunks
}

func (m *MulticallContract) Aggregate(ctx context.Context, calls []Call3, block uint64) ([]Result, error) {
	// split calls into chunks
	callChunks := m.SplitCalls(calls)
	results := make([]Result, 0)

	for _, callChunk := range callChunks {
		// encode calls
		callsData, err := m.cAbi.Pack("aggregate3", callChunk)
		if err != nil {
			return nil, err
		}

		// set block number
		var callBlock *big.Int
		if block != 0 {
			callBlock = new(big.Int).SetUint64(block)
		}

		// call the contract
		rawRes, err := m.client.CallContract(
			ctx,
			ethereum.CallMsg{
				To:   &m.contract,
				Data: callsData,
				Gas:  m.maxGas,
			},
			callBlock,
		)
		if err != nil {
			return nil, err
		}

		// decode results
		inter, err := m.cAbi.Unpack("aggregate3", rawRes)
		if err != nil {
			return nil, err
		}

		// validate return data
		res := inter[0].([]struct {
			Success    bool   "json:\"success\""
			ReturnData []byte "json:\"returnData\""
		})
		if len(res) != len(callChunk) {
			return nil, errors.New(fmt.Sprintf("return data length mismatch: %v != %v", len(res), len(callChunk)))
		}

		// merge results
		for _, r := range res {
			results = append(results, Result{
				Block:      block,
				ReturnData: r.ReturnData,
			})
		}
	}

	return results, nil
}
