package generic

import (
	"PoolHelper/src/multicaller"
	uniswapv2 "PoolHelper/src/pool/uniswap-v2"
	uniswapv3 "PoolHelper/src/pool/uniswap-v3"
	"PoolHelper/src/token"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

///
/// Call
///

type Call struct {
	Target   common.Address
	CallData []byte
}

type Result []byte

///
/// Multicall3
///

type Multicall3 struct {
	contract common.Address
	callCost uint64
	maxGas   uint64
	cAbi     abi.ABI
	client   multicaller.ClientDispatcher
}

func NewMulticall(contract common.Address, callCost uint64, maxGas uint64, cAbi abi.ABI, client multicaller.ClientDispatcher) *Multicall3 {
	return &Multicall3{
		contract: contract,
		callCost: callCost,
		maxGas:   maxGas,
		cAbi:     cAbi,
		client:   client,
	}
}

func (m *Multicall3) splitCalls(calls []Call) [][]Call {
	callChunks := make([][]Call, 0)
	callChunk := make([]Call, 0)
	callChunkGas := uint64(0)

	for _, call := range calls {
		// check if adding this call would exceed the max gas limit for the chunk
		if callChunkGas+m.callCost > m.maxGas {
			// if it exceeds, start a new chunk
			callChunks = append(callChunks, callChunk)
			callChunk = make([]Call, 0)
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

func (m *Multicall3) multicall(ctx context.Context, calls []Call, block uint64) ([]Result, error) {
	// split calls into chunks
	callChunks := m.splitCalls(calls)
	results := make([]Result, len(calls))

	for _, callChunk := range callChunks {
		// encode calls
		callsData, err := m.cAbi.Pack("aggregate", callChunk)
		if err != nil {
			return nil, err
		}

		// call the contract
		rawRes, err := m.client.CallContract(
			ctx,
			ethereum.CallMsg{
				To:   &m.contract,
				Data: callsData,
				Gas:  m.maxGas,
			},
			new(big.Int).SetUint64(block),
		)
		if err != nil {
			return nil, err
		}

		// decode results
		fmt.Println(common.Bytes2Hex(rawRes))
		_results := make([]Result, len(calls))
		err = m.cAbi.UnpackIntoInterface(&_results, "aggregate", rawRes)
		if err != nil {
			return nil, err
		}

		// merge results
		results = append(results, _results...)
	}

	return results, nil
}

///
/// Implementation
///

func (m *Multicall3) FetchSlots(ctx context.Context, targets []common.Address, blockNumber uint64) ([]uniswapv3.Slot0, error) {
	// prepare calls
	calls := make([]Call, len(targets))
	for i, target := range targets {
		calls[i] = Call{
			Target:   target,
			CallData: crypto.Keccak256([]byte("slot0()"))[:4],
		}
	}

	// call
	results, err := m.multicall(ctx, calls, blockNumber)
	if err != nil {
		return nil, err
	}

	// decode results
	slots := make([]uniswapv3.Slot0, len(targets))
	for i, result := range results {
		if len(result) != 32*7 {
			return nil, errors.New(fmt.Sprintf("wrong return data length: %v", len(result)))
		}

		slot := uniswapv3.Slot0{
			SqrtPriceX96:               new(big.Int).SetBytes(result[0:32]),
			Tick:                       new(big.Int).SetBytes(result[32:64]),
			ObservationIndex:           binary.BigEndian.Uint16(result[64:66]),
			ObservationCardinality:     binary.BigEndian.Uint16(result[66:68]),
			ObservationCardinalityNext: binary.BigEndian.Uint16(result[68:70]),
			FeeProtocol:                result[70],
			Unlocked:                   result[71] != 0,
		}

		// store in the final result
		slots[i] = slot
	}

	return slots, nil
}

func (m *Multicall3) FetchReserves(ctx context.Context, targets []common.Address, blockNumber uint64) ([]uniswapv2.Reserves, error) {
	// prepare calls
	calls := make([]Call, len(targets))
	for i, target := range targets {
		calls[i] = Call{
			Target:   target,
			CallData: crypto.Keccak256([]byte("getReserves()"))[:4],
		}
	}

	// call
	results, err := m.multicall(ctx, calls, blockNumber)
	if err != nil {
		return nil, err
	}

	// decode results
	reserves := make([]uniswapv2.Reserves, len(targets))
	for i, result := range results {
		if len(result) != 32*2 {
			return nil, errors.New(fmt.Sprintf("wrong return data length: %v", len(result)))
		}

		reserve0 := new(big.Int).SetBytes(result[0:32])
		reserve1 := new(big.Int).SetBytes(result[32:64])

		// store in the final result
		reserves[i] = uniswapv2.Reserves{
			Reserve0: reserve0,
			Reserve1: reserve1,
		}
	}

	return reserves, nil
}

func (m *Multicall3) FetchTokens(ctx context.Context, targets []common.Address, blockNumber uint64) ([]token.Token, error) {
	// prepare calls
	calls := make([]Call, len(targets)*3)
	for i, target := range targets {
		// decimals
		calls[i*3] = Call{
			Target:   target,
			CallData: crypto.Keccak256([]byte("decimals()"))[:4],
		}

		// symbol
		calls[i*3+1] = Call{
			Target:   target,
			CallData: crypto.Keccak256([]byte("symbol()"))[:4],
		}

		// name
		calls[i*3+2] = Call{
			Target:   target,
			CallData: crypto.Keccak256([]byte("name()"))[:4],
		}
	}

	// call
	results, err := m.multicall(ctx, calls, blockNumber)
	if err != nil {
		return nil, err
	}

	// decode results
	tokens := make([]token.Token, len(targets))
	for i, result := range results {
		// validate data
		if len(result) < 68 {
			return nil, errors.New(fmt.Sprintf("invalid return data length for token info: %v", len(result)))
		}

		// decode data
		decimals := result[0]
		symbol := string(result[4:36])
		name := string(result[36:68])

		// trim null bytes
		symbol = symbol[:bytes.IndexByte(result[4:36], 0)]
		name = name[:bytes.IndexByte(result[36:68], 0)]

		// create token info
		tokenInfo := token.Token{
			Decimals: decimals,
			Symbol:   symbol,
			Name:     name,
		}

		// store in the final result
		tokens[i] = tokenInfo
	}

	return tokens, nil
}
