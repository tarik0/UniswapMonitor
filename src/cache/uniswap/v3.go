package uniswap

import (
	"PoolHelper/src/multicall/generic"
	"PoolHelper/src/pool"
	"PoolHelper/src/pool/uniswap"
	"PoolHelper/src/structs/factory"
	"PoolHelper/src/structs/pair"
	"PoolHelper/src/structs/token"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"strings"
	"sync"
	"unicode"
)

type V3Cache struct {
	tokens    map[common.Address]token.ERC20
	pools     map[common.Address]pool.Pool[uniswap.Slot0, uniswap.V3FeeType]
	factories map[common.Address]factory.Factory[uniswap.V3FeeType]
	lastSync  uint64
	m         sync.RWMutex
}

func NewV3Cache() *V3Cache {
	return &V3Cache{
		tokens:    make(map[common.Address]token.ERC20),
		pools:     make(map[common.Address]pool.Pool[uniswap.Slot0, uniswap.V3FeeType]),
		factories: make(map[common.Address]factory.Factory[uniswap.V3FeeType]),
		m:         sync.RWMutex{},
		lastSync:  0,
	}
}

///
/// Token Cache
///

func (c *V3Cache) ImportTokens(ctx context.Context, m generic.Multicall, tokens []common.Address) error {
	c.m.Lock()
	defer c.m.Unlock()

	// import tokens
	if err := c.importTokens(ctx, m, tokens); err != nil {
		return err
	}

	return nil
}

func (c *V3Cache) AddToken(t token.ERC20) error {
	c.m.Lock()
	defer c.m.Unlock()

	// validate token
	if ok := t.IsValid(); !ok {
		return InvalidToken
	}

	return c.addToken(t)
}

func (c *V3Cache) RemoveToken(address common.Address) error {
	c.m.Lock()
	defer c.m.Unlock()

	// check if token exists in cache
	if _, ok := c.tokens[address]; !ok {
		return TokenNotFound
	}

	return c.removeToken(address)
}

func (c *V3Cache) Token(address common.Address) (token.ERC20, error) {
	c.m.RLock()
	defer c.m.RUnlock()

	// get token from cache
	if t, ok := c.tokens[address]; ok {
		return t, nil
	}

	return token.ERC20{}, TokenNotFound
}

func (c *V3Cache) Tokens() ([]token.ERC20, error) {
	c.m.RLock()
	defer c.m.RUnlock()

	// get tokens from cache
	tokens := make([]token.ERC20, 0, len(c.tokens))
	for _, t := range c.tokens {
		tokens = append(tokens, t)
	}

	return tokens, nil
}

///
/// Pool Cache
///

func (c *V3Cache) InitializePools(factory factory.Factory[uniswap.V3FeeType]) error {
	// validate factory
	if ok := factory.IsValid(); !ok {
		return InvalidFactory
	}

	c.m.Lock()
	defer c.m.Unlock()

	// iterate through tokens
	for _, t0 := range c.tokens {
		// iterate through tokens again
		for _, t1 := range c.tokens {
			// skip if tokens are the same
			if bytes.EqualFold(t0.Address.Bytes(), t1.Address.Bytes()) {
				continue
			}

			// iterate through fee types
			for _, feeType := range factory.FeeTypes {
				// create pair & try to add pool to cache
				_p := pair.NewPair[uniswap.V3FeeType](t0, t1, feeType)
				if err := c.addPool(factory, _p, false); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *V3Cache) RemovePool(address common.Address) error {
	c.m.Lock()
	defer c.m.Unlock()

	// check if pool exists in cache
	if _, ok := c.pools[address]; !ok {
		return PoolNotFound
	}

	// remove pool from cache
	if err := c.removePool(address); err != nil {
		return err
	}

	return nil
}

func (c *V3Cache) Pool(address common.Address) (pool.Pool[uniswap.Slot0, uniswap.V3FeeType], error) {
	c.m.RLock()
	defer c.m.RUnlock()

	// get pool from cache
	if p, ok := c.pools[address]; ok {
		return p, nil
	}

	return nil, PoolNotFound
}

func (c *V3Cache) Pools() []pool.Pool[uniswap.Slot0, uniswap.V3FeeType] {
	c.m.RLock()
	defer c.m.RUnlock()

	// get pools from cache
	pools := make([]pool.Pool[uniswap.Slot0, uniswap.V3FeeType], 0, len(c.pools))
	for _, p := range c.pools {
		pools = append(pools, p)
	}

	return pools
}

///
/// Reserve Cache
///

func (c *V3Cache) SyncAll(ctx context.Context, m generic.Multicall, block uint64) error {
	c.m.Lock()
	defer c.m.Unlock()

	// check if block has already been synced
	if c.lastSync >= block {
		return BlockAlreadySynced
	}

	// get all pools
	allPools := make([]common.Address, 0, len(c.pools))
	for _, p := range c.pools {
		allPools = append(allPools, p.Address())
	}

	// sync reserves
	if err := c.sync(ctx, m, allPools, block); err != nil {
		return err
	}

	c.lastSync = block
	return nil
}

func (c *V3Cache) Sync(ctx context.Context, m generic.Multicall, pools []common.Address, block uint64) error {
	c.m.Lock()
	defer c.m.Unlock()

	// check if block has already been synced
	if c.lastSync >= block {
		return BlockAlreadySynced
	}

	// check if pools are valid
	for _, p := range pools {
		if _, ok := c.pools[p]; !ok {
			return PoolNotFound
		}
	}

	// sync reserves
	if err := c.sync(ctx, m, pools, block); err != nil {
		return err
	}

	c.lastSync = block
	return nil
}

func (c *V3Cache) LastSynced() uint64 {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.lastSync
}

///
/// Internal
/// (does not lock mutex)

// addToken adds a token to the cache
// overwrites existing token if it already exists in cache
// overwrites new pools to cache with existing tokens and factories
func (c *V3Cache) addToken(t token.ERC20) error {
	// check if token already exists in cache
	if _, ok := c.tokens[t.Address]; ok {
		return TokenAlreadyExists
	}

	// iterate through factories
	for _, f := range c.factories {
		// iterate through tokens
		for _, pairToken := range c.tokens {
			// iterate through fee types
			for _, feeType := range f.FeeTypes {
				newPair := pair.NewPair[uniswap.V3FeeType](pairToken, t, feeType)
				if err := c.addPool(f, newPair, true); err != nil {
					return err
				}
			}
		}
	}

	// add token to cache
	c.tokens[t.Address] = t
	return nil
}

// removeToken removes a token from the cache
// removes pools from cache that contain the token
func (c *V3Cache) removeToken(address common.Address) error {
	// remove token from cache
	delete(c.tokens, address)

	// remove pools from cache
	for _, p := range c.pools {
		if poolPair := p.Pair(); poolPair.Contains(address) {
			if err := c.removePool(p.Address()); err != nil {
				return err
			}
		}
	}

	return nil
}

// addPool adds a pool to the cache
func (c *V3Cache) addPool(f factory.Factory[uniswap.V3FeeType], pair pair.Pair[uniswap.V3FeeType], overwrite bool) error {
	// create pool & try to add to cache if it doesn't exist
	p := uniswap.NewV3Pool(f.Address, f.InitHash, pair)
	if _, ok := c.pools[p.Address()]; !ok || overwrite {
		c.pools[p.Address()] = p

		// add factory to cache if it doesn't exist
		if _, _ok := c.factories[f.Address]; !_ok || overwrite {
			c.factories[f.Address] = f
		}
	}

	return nil
}

// removePool removes a pool from the cache
// removes factory from cache if it doesn't have any pools
func (c *V3Cache) removePool(address common.Address) error {
	poolFactory := common.HexToAddress(c.pools[address].Factory().Hex())
	delete(c.pools, address)

	// remove factory from cache if it doesn't have any pools
	if _, ok := c.factories[poolFactory]; ok {
		for _, p := range c.pools {
			if bytes.EqualFold(p.Factory().Bytes(), poolFactory.Bytes()) {
				return nil
			}
		}
		delete(c.factories, poolFactory)
	}
	return nil
}

// sync syncs reserves for a list of pools
func (c *V3Cache) sync(ctx context.Context, m generic.Multicall, pools []common.Address, block uint64) error {
	// prepare calls
	calls := make([]generic.Call3, len(pools))
	for i, target := range pools {
		calls[i] = generic.Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("slot0()"))[:4],
			AllowFailure: true,
		}
	}

	// call the contract
	results, err := m.Aggregate(ctx, calls, block)
	if err != nil {
		return err
	}

	// check if results are valid
	if len(results) != len(pools) {
		return errors.New(fmt.Sprintf("wrong number of results: %v", len(results)))
	}

	// decode results
	for i, result := range results {
		poolAddr := pools[i]

		// check if pool initialized
		if len(result.ReturnData) == 0 {
			c.pools[poolAddr].Update(uniswap.Slot0{
				SqrtPriceX96:               big.NewInt(0),
				Tick:                       big.NewInt(0),
				ObservationIndex:           big.NewInt(0),
				ObservationCardinality:     big.NewInt(0),
				ObservationCardinalityNext: big.NewInt(0),
				FeeProtocol:                big.NewInt(0),
				Unlocked:                   false,
			}, block)
			continue
		}

		if len(result.ReturnData) != 224 {
			return errors.New(fmt.Sprintf("wrong return data length: %v", len(result.ReturnData)))
		}

		// update pool
		c.pools[poolAddr].Update(uniswap.Slot0{
			SqrtPriceX96:               new(big.Int).SetBytes(result.ReturnData[0:32]),
			Tick:                       new(big.Int).SetBytes(result.ReturnData[32:64]),
			ObservationIndex:           new(big.Int).SetBytes(result.ReturnData[64:96]),
			ObservationCardinality:     new(big.Int).SetBytes(result.ReturnData[96:128]),
			ObservationCardinalityNext: new(big.Int).SetBytes(result.ReturnData[128:160]),
			FeeProtocol:                new(big.Int).SetBytes(result.ReturnData[160:192]),
			Unlocked:                   result.ReturnData[223] != 0,
		}, block)
	}

	return nil
}

// importTokens imports tokens into the cache
func (c *V3Cache) importTokens(ctx context.Context, m generic.Multicall, tokens []common.Address) error {
	// prepare calls
	calls := make([]generic.Call3, 0)
	for _, target := range tokens {
		calls = append(calls, generic.Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("decimals()"))[:4],
			AllowFailure: false,
		})
		calls = append(calls, generic.Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("symbol()"))[:4],
			AllowFailure: false,
		})
		calls = append(calls, generic.Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("name()"))[:4],
			AllowFailure: false,
		})
	}

	// call the contract
	results, err := m.Aggregate(ctx, calls, 0)
	if err != nil {
		return err
	}

	// check if results are valid
	if len(results) != len(tokens)*3 {
		return errors.New(fmt.Sprintf("wrong number of results: %v", len(results)))
	}

	// decode results
	for i := 0; i < len(results); i += 3 {
		tokenAddr := tokens[i/3]

		// validate data
		if len(results[i].ReturnData) != 32 {
			return errors.New(fmt.Sprintf("invalid return data length for decimals: %v", len(results[i].ReturnData)))
		}
		if len(results[i+1].ReturnData) < 32 {
			return errors.New(fmt.Sprintf("invalid return data length for symbol: %v", len(results[i+1].ReturnData)))
		}
		if len(results[i+2].ReturnData) < 32 {
			return errors.New(fmt.Sprintf("invalid return data length for name: %v", len(results[i+2].ReturnData)))
		}

		// decode data
		decimals := new(big.Int).SetBytes(results[i].ReturnData)
		symbol := fmt.Sprintf("%s", results[i+1].ReturnData)
		name := fmt.Sprintf("%s", results[i+2].ReturnData)

		// trim invalid characters
		symbol = strings.TrimFunc(symbol, func(r rune) bool {
			return !unicode.IsDigit(r) && !unicode.IsLetter(r)
		})
		name = strings.TrimFunc(name, func(r rune) bool {
			return !unicode.IsDigit(r) && !unicode.IsLetter(r)
		})

		// create token info
		err = c.addToken(token.ERC20{
			Address:  tokenAddr,
			Decimals: decimals,
			Symbol:   symbol,
			Name:     name,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
