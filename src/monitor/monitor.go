package monitor

import (
	"PoolHelper/src/multicaller"
	"PoolHelper/src/pool"
	uniswapv2 "PoolHelper/src/pool/uniswap-v2"
	uniswapv3 "PoolHelper/src/pool/uniswap-v3"
	"PoolHelper/src/token"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"sync"
)

///
/// Monitor
///

type Monitor struct {
	tokens     map[common.Address]token.Token
	pools      map[common.Address]pool.Pool
	poolsMutex *sync.RWMutex
}

func NewMonitor() *Monitor {
	return &Monitor{
		tokens:     make(map[common.Address]token.Token),
		pools:      make(map[common.Address]pool.Pool),
		poolsMutex: &sync.RWMutex{},
	}
}

///
/// Token Management
///

// Setters

// AddTokenSafe adds a token to the cache.
// It is thread-safe.
func (m *Monitor) AddTokenSafe(address common.Address) {
	m.poolsMutex.Lock()
	defer m.poolsMutex.Unlock()

	// add token
	m.tokens[address] = token.Token{
		Address: address,
	}
}

// RemoveTokenSafe removes a token from the cache.
// It is thread-safe.
func (m *Monitor) RemoveTokenSafe(address common.Address) {
	m.poolsMutex.Lock()
	defer m.poolsMutex.Unlock()

	// remove token
	delete(m.tokens, address)

	// remove pools
	for _, _pool := range m.pools {
		if _pool.Pair().TokenA.Address == address || _pool.Pair().TokenB.Address == address {
			delete(m.pools, _pool.Address())
		}
	}
}

// InitializeTokens updates the token states for all tokens in the cache.
// It uses the multicaller to fetch the data.
func (m *Monitor) InitializeTokens(ctx context.Context, multicaller multicaller.Multicaller, block uint64) ([]token.Token, error) {
	// fetch addresses
	m.poolsMutex.RLock()
	addresses := make([]common.Address, 0)
	for _, _token := range m.tokens {
		addresses = append(addresses, _token.Address)
	}
	m.poolsMutex.RUnlock()

	// fetch token infos
	tokens, err := multicaller.FetchTokens(ctx, addresses, block)
	if err != nil {
		return nil, err
	}

	// update tokens
	for _, _token := range tokens {
		m.tokens[_token.Address] = _token
	}

	return tokens, nil
}

// Getters

func (m *Monitor) Token(address common.Address) (token.Token, bool) {
	m.poolsMutex.RLock()
	defer m.poolsMutex.RUnlock()
	val, ok := m.tokens[address]
	return val, ok
}

func (m *Monitor) Tokens() []token.Token {
	m.poolsMutex.RLock()
	defer m.poolsMutex.RUnlock()

	tokens := make([]token.Token, 0)
	for _, _token := range m.tokens {
		tokens = append(tokens, _token)
	}

	return tokens
}

func (m *Monitor) TokenCount() int {
	m.poolsMutex.RLock()
	defer m.poolsMutex.RUnlock()
	return len(m.tokens)
}

///
/// Pool Management
///

// Setters

// FindPoolsV2 finds all pools for the given factory and init code.
// It overwrites the existing pool cache.
func (m *Monitor) FindPoolsV2(factory common.Address, initCode common.Hash) []common.Address {
	m.poolsMutex.Lock()
	defer m.poolsMutex.Unlock()

	// create pools
	newPools := make([]common.Address, 0)
	for _, tokenA := range m.tokens {
		for _, tokenB := range m.tokens {
			// skip same token
			if tokenA.Address == tokenB.Address {
				continue
			}

			// create pair & pool
			pair := token.Pair{
				TokenA: tokenA,
				TokenB: tokenB,
			}
			_pool := uniswapv2.NewUniswapV2Pool(factory, initCode, &pair)

			// add pool to cache
			if _, ok := m.pools[_pool.Address()]; !ok {
				newPools = append(newPools, _pool.Address())
			}
			m.pools[_pool.Address()] = _pool
		}
	}

	return newPools
}

// FindPoolsV3 finds all pools for the given factory and init code.
// It overwrites the existing pool cache.
func (m *Monitor) FindPoolsV3(factory common.Address, initHash common.Hash, fees []uniswapv3.FeeType) []common.Address {
	m.poolsMutex.Lock()
	defer m.poolsMutex.Unlock()

	// create pools
	newPools := make([]common.Address, 0)
	for _, fee := range fees {
		for _, tokenA := range m.tokens {
			for _, tokenB := range m.tokens {
				// skip same token
				if tokenA.Address == tokenB.Address {
					continue
				}

				// create pair & pool
				pair := token.Pair{
					TokenA: tokenA,
					TokenB: tokenB,
				}
				_pool := uniswapv3.NewUniswapV3Pool(factory, initHash, &pair, fee)

				// add pool to cache
				if _, ok := m.pools[_pool.Address()]; !ok {
					newPools = append(newPools, _pool.Address())
				}
				m.pools[_pool.Address()] = _pool
			}
		}
	}

	return newPools
}

// InitializePools updates the pool states for all pools in the cache.
// It uses the multicaller to fetch the data.
func (m *Monitor) InitializePools(ctx context.Context, multicaller multicaller.Multicaller, block uint64) error {
	v2Pools := make([]common.Address, 0)
	v3Pools := make([]common.Address, 0)

	// get pool addresses
	for _, _pool := range m.pools {
		switch _pool.Type() {
		case pool.UniswapV2:
			v2Pools = append(v2Pools, _pool.Address())
		case pool.UniswapV3:
			v3Pools = append(v3Pools, _pool.Address())
		default:
			return errors.New(fmt.Sprintf("unknown pool type: %v", _pool.Type()))
		}
	}

	var err error
	res := make([]uniswapv2.Reserves, 0)
	slots := make([]uniswapv3.Slot0, 0)

	// fetch uniswap v2 reserves
	if len(v2Pools) > 0 {
		res, err = multicaller.FetchReserves(ctx, v2Pools, block)
		if err != nil {
			return err
		}
	}

	// fetch uniswap v3 slots
	if len(v3Pools) > 0 {
		slots, err = multicaller.FetchSlots(ctx, v3Pools, block)
		if err != nil {
			return err
		}
	}

	m.poolsMutex.Lock()
	defer m.poolsMutex.Unlock()

	// update v2 pools
	for i, _pool := range v2Pools {
		_pool := m.pools[_pool].(*uniswapv2.UniswapV2Pool)
		_pool.UpdateSafe(res[i].Reserve0, res[i].Reserve1, block)
	}

	// update v3 pools
	for i, _pool := range v3Pools {
		_pool := m.pools[_pool].(*uniswapv3.UniswapV3Pool)
		_pool.UpdateSafe(slots[i], block)
	}

	return nil
}

// Getters

func (m *Monitor) PoolsByPair(pair token.Pair) ([]pool.Pool, bool) {
	m.poolsMutex.RLock()
	defer m.poolsMutex.RUnlock()

	pools := make([]pool.Pool, 0)
	for _, _pool := range m.pools {
		if _pool.Pair().Equals(pair) {
			pools = append(pools, _pool)
		}
	}

	return nil, false
}

func (m *Monitor) Pool(address common.Address) (pool.Pool, bool) {
	m.poolsMutex.RLock()
	defer m.poolsMutex.RUnlock()
	val, ok := m.pools[address]
	return val, ok
}

func (m *Monitor) Pools() []pool.Pool {
	m.poolsMutex.RLock()
	defer m.poolsMutex.RUnlock()

	pools := make([]pool.Pool, 0)
	for _, _pool := range m.pools {
		pools = append(pools, _pool)
	}

	return pools
}

func (m *Monitor) PoolCount() int {
	m.poolsMutex.RLock()
	defer m.poolsMutex.RUnlock()
	return len(m.pools)
}
