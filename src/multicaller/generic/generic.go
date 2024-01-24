package generic

import (
	"PoolHelper/src/multicaller"
	uniswapv2 "PoolHelper/src/pool/uniswap-v2"
	uniswapv3 "PoolHelper/src/pool/uniswap-v3"
	"PoolHelper/src/token"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	ReturnData [][]byte
}

///
/// Multicall
///

type Multicall struct {
	contract common.Address
	callCost uint64
	maxGas   uint64
	cAbi     abi.ABI
	client   multicaller.ClientDispatcher
}

func NewMulticall(contract common.Address, callCost uint64, maxGas uint64, cAbi abi.ABI, client multicaller.ClientDispatcher) *Multicall {
	return &Multicall{
		contract: contract,
		callCost: callCost,
		maxGas:   maxGas,
		cAbi:     cAbi,
		client:   client,
	}
}

func (m *Multicall) splitCalls(calls []Call3) [][]Call3 {
	callChunks := make([][]Call3, 0)
	callChunk := make([]Call3, 0)
	callChunkGas := uint64(0)

	for _, call := range calls {
		// check if adding this call would exceed the max gas limit for the chunk
		if callChunkGas+m.callCost > m.maxGas {
			// if it exceeds, start a new chunk
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

func (m *Multicall) multicall(ctx context.Context, calls []Call3, block uint64) (Result, error) {
	// split calls into chunks
	callChunks := m.splitCalls(calls)
	result := Result{
		Block: block,
	}

	for _, callChunk := range callChunks {
		// encode calls
		callsData, err := m.cAbi.Pack("aggregate3", callChunk)
		if err != nil {
			return result, err
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
			return result, err
		}

		// decode results
		inter, err := m.cAbi.Unpack("aggregate3", rawRes)
		if err != nil {
			return result, err
		}

		// validate return data
		res := inter[0].([]struct {
			Success    bool   "json:\"success\""
			ReturnData []byte "json:\"returnData\""
		})
		if len(res) != len(callChunk) {
			return result, errors.New(fmt.Sprintf("return data length mismatch: %v != %v", len(res), len(callChunk)))
		}

		// merge results
		for _, r := range res {
			result.ReturnData = append(result.ReturnData, r.ReturnData)
		}
	}

	return result, nil
}

///
/// Implementation
///

func (m *Multicall) FetchSlots(ctx context.Context, targets []common.Address, blockNumber uint64) ([]uniswapv3.Slot0, error) {
	// prepare calls
	calls := make([]Call3, len(targets))
	for i, target := range targets {
		calls[i] = Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("slot0()"))[:4],
			AllowFailure: true,
		}
	}

	// call
	result, err := m.multicall(ctx, calls, blockNumber)
	if err != nil {
		return nil, err
	}

	// decode results
	slots := make([]uniswapv3.Slot0, len(targets))
	for i, returnData := range result.ReturnData {
		// pair doesn't exist.
		if len(returnData) == 0 {
			slots[i] = uniswapv3.Slot0{
				SqrtPriceX96:               big.NewInt(0),
				Tick:                       big.NewInt(0),
				ObservationIndex:           big.NewInt(0),
				ObservationCardinality:     big.NewInt(0),
				ObservationCardinalityNext: big.NewInt(0),
				FeeProtocol:                big.NewInt(0),
				Unlocked:                   false,
			}
			continue
		}

		if len(returnData) != 224 {
			return nil, errors.New(fmt.Sprintf("wrong return data length: %v", len(returnData)))
		}

		slot := uniswapv3.Slot0{
			SqrtPriceX96:               new(big.Int).SetBytes(returnData[0:32]),
			Tick:                       new(big.Int).SetBytes(returnData[32:64]),
			ObservationIndex:           new(big.Int).SetBytes(returnData[64:96]),
			ObservationCardinality:     new(big.Int).SetBytes(returnData[96:128]),
			ObservationCardinalityNext: new(big.Int).SetBytes(returnData[128:160]),
			FeeProtocol:                new(big.Int).SetBytes(returnData[160:192]),
			Unlocked:                   returnData[223] != 0,
		}

		// store in the final result
		slots[i] = slot
	}

	return slots, nil
}

func (m *Multicall) FetchReserves(ctx context.Context, targets []common.Address, blockNumber uint64) ([]uniswapv2.Reserves, error) {
	// prepare calls
	calls := make([]Call3, len(targets))
	for i, target := range targets {
		calls[i] = Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("getReserves()"))[:4],
			AllowFailure: true,
		}
	}

	// call
	results, err := m.multicall(ctx, calls, blockNumber)
	if err != nil {
		return nil, err
	}

	// decode results
	reserves := make([]uniswapv2.Reserves, len(targets))
	for i, result := range results.ReturnData {
		if len(result) == 0 {
			reserves[i] = uniswapv2.Reserves{
				Reserve0: big.NewInt(0),
				Reserve1: big.NewInt(0),
			}
			continue
		}

		if len(result) != 32*3 {
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

func (m *Multicall) FetchTokens(ctx context.Context, targets []common.Address, blockNumber uint64) ([]token.ERC20, error) {
	// prepare calls
	calls := make([]Call3, 0)
	for _, target := range targets {
		// decimals
		calls = append(calls, Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("decimals()"))[:4],
			AllowFailure: false,
		})

		// symbol
		calls = append(calls, Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("symbol()"))[:4],
			AllowFailure: false,
		})

		// name
		calls = append(calls, Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("name()"))[:4],
			AllowFailure: false,
		})
	}

	// call
	results, err := m.multicall(ctx, calls, blockNumber)
	if err != nil {
		return nil, err
	}

	// decode results
	tokens := make([]token.ERC20, 0)
	for i := 0; i < len(results.ReturnData); i += 3 {
		// validate data
		if len(results.ReturnData[i]) != 32 {
			return nil, errors.New(fmt.Sprintf("invalid return data length for decimals: %v", len(results.ReturnData[i])))
		}
		if len(results.ReturnData[i+1]) < 32 {
			return nil, errors.New(fmt.Sprintf("invalid return data length for symbol: %v", len(results.ReturnData[i+1])))
		}
		if len(results.ReturnData[i+2]) < 32 {
			return nil, errors.New(fmt.Sprintf("invalid return data length for name: %v", len(results.ReturnData[i+2])))
		}

		// decode data
		stringType, _ := abi.NewType("string", "", nil)
		args := abi.Arguments{
			{
				Name: "str",
				Type: stringType,
			},
		}
		decimals := new(big.Int).SetBytes(results.ReturnData[i])
		symbol, err := args.Unpack(results.ReturnData[i+1])
		name, err := args.Unpack(results.ReturnData[i+2])
		if err != nil {
			return nil, err
		}

		// create token info
		tokenInfo := token.ERC20{
			Address:  targets[i/3],
			Decimals: decimals,
			Symbol:   symbol[0].(string),
			Name:     name[0].(string),
		}

		// store in the final result
		tokens = append(tokens, tokenInfo)
	}

	return tokens, nil
}
