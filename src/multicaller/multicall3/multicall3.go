package multicall3

import (
	"PoolHelper/src/multicaller"
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
	"github.com/ethereum/go-ethereum/params"
	"math/big"
)

///
/// Call & Result
///

type Call3 struct {
	Target       common.Address
	AllowFailure bool
	CallData     []byte
}

type Result struct {
	success    bool
	returnData []byte
}

///
/// Multicall3
///

type Multicall3 struct {
	contract common.Address
	cAbi     abi.ABI
	client   multicaller.ClientDispatcher
}

func NewMulticall3(contract common.Address, cAbi abi.ABI, client multicaller.ClientDispatcher) *Multicall3 {
	return &Multicall3{
		contract: contract,
		cAbi:     cAbi,
		client:   client,
	}
}

func (m *Multicall3) multicall(ctx context.Context, calls []Call3, block uint64) ([]Result, error) {
	// encode calls
	callsData, err := m.cAbi.Pack("aggregate3", calls)
	if err != nil {
		return nil, err
	}

	// call the contract
	rawRes, err := m.client.CallContract(
		ctx,
		ethereum.CallMsg{
			To:   &m.contract,
			Data: callsData,
			Gas:  21_000 * params.GWei,
		},
		new(big.Int).SetUint64(block),
	)
	if err != nil {
		return nil, err
	}

	// decode results
	results := make([]Result, len(calls))
	err = m.cAbi.UnpackIntoInterface(&results, "aggregate3", rawRes)
	if err != nil {
		return nil, err
	}

	return results, nil
}

///
/// Implementation
///

func (m *Multicall3) FetchSlots(ctx context.Context, targets []common.Address, blockNumber uint64) ([]uniswapv3.Slot0, error) {
	// prepare calls
	calls := make([]Call3, len(targets))
	for i, target := range targets {
		calls[i] = Call3{
			Target:       target,
			AllowFailure: false,
			CallData:     crypto.Keccak256([]byte("slot0()"))[:4],
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
		if !result.success {
			return nil, err
		}
		if len(result.returnData) != 32*7 {
			return nil, errors.New(fmt.Sprintf("wrong return data length: %v", len(result.returnData)))
		}

		slot := uniswapv3.Slot0{
			SqrtPriceX96:               new(big.Int).SetBytes(result.returnData[0:32]),
			Tick:                       new(big.Int).SetBytes(result.returnData[32:64]),
			ObservationIndex:           binary.BigEndian.Uint16(result.returnData[64:66]),
			ObservationCardinality:     binary.BigEndian.Uint16(result.returnData[66:68]),
			ObservationCardinalityNext: binary.BigEndian.Uint16(result.returnData[68:70]),
			FeeProtocol:                result.returnData[70],
			Unlocked:                   result.returnData[71] != 0,
		}

		// store in the final result
		slots[i] = slot
	}

	return slots, nil
}

func (m *Multicall3) FetchTokens(ctx context.Context, targets []common.Address, blockNumber uint64) ([]token.Token, error) {
	// prepare calls
	calls := make([]Call3, len(targets)*3)
	for i, target := range targets {
		// decimals
		calls[i*3] = Call3{
			Target:       target,
			AllowFailure: false,
			CallData:     crypto.Keccak256([]byte("decimals()"))[:4],
		}

		// symbol
		calls[i*3+1] = Call3{
			Target:       target,
			AllowFailure: false,
			CallData:     crypto.Keccak256([]byte("symbol()"))[:4],
		}

		// name
		calls[i*3+2] = Call3{
			Target:       target,
			AllowFailure: false,
			CallData:     crypto.Keccak256([]byte("name()"))[:4],
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
		if !result.success {
			return nil, err
		}

		// validate data
		if len(result.returnData) < 68 {
			return nil, errors.New(fmt.Sprintf("invalid return data length for token info: %v", len(result.returnData)))
		}

		// decode data
		decimals := result.returnData[0]
		symbol := string(result.returnData[4:36])
		name := string(result.returnData[36:68])

		// trim null bytes
		symbol = symbol[:bytes.IndexByte(result.returnData[4:36], 0)]
		name = name[:bytes.IndexByte(result.returnData[36:68], 0)]

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
