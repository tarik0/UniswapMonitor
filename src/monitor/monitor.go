package monitor

import (
	"PoolHelper/src/pool"
	uniswap_v2 "PoolHelper/src/pool/uniswap-v2"
	"PoolHelper/src/token"
	"github.com/ethereum/go-ethereum/common"
	"sync"
)

///
/// Monitor
///

type Monitor struct {
	tokens []token.Token

	pools      map[common.Address]pool.Pool // pool cache
	poolsMutex *sync.RWMutex
}

func NewMonitor() *Monitor {
	return &Monitor{
		tokens:     make([]token.Token, 0),
		pools:      make(map[common.Address]pool.Pool),
		poolsMutex: &sync.RWMutex{},
	}
}

///
/// Token Management
///

func (m *Monitor) AddToken(token token.Token) {
	m.poolsMutex.Lock()
	defer m.poolsMutex.Unlock()

	// add token
	m.tokens = append(m.tokens, token)
}

func (m *Monitor) RemoveToken(address common.Address) {
	m.poolsMutex.Lock()
	defer m.poolsMutex.Unlock()

	// remove token
	for i, _token := range m.tokens {
		if _token.Address == address {
			m.tokens = append(m.tokens[:i], m.tokens[i+1:]...)
			break
		}
	}

	// remove pools
	for _, _pool := range m.pools {
		if _pool.Pair().TokenA.Address == address || _pool.Pair().TokenB.Address == address {
			delete(m.pools, _pool.Address())
		}
	}
}

///
/// Pool Management
///

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
			_pool := uniswap_v2.NewUniswapV2Pool(factory, initCode, &pair)

			// add pool to cache
			m.pools[_pool.Address()] = _pool
			newPools = append(newPools, _pool.Address())
		}
	}

	return newPools
}

///
/// Getters
///

func (m *Monitor) Pool(address common.Address) pool.Pool {
	m.poolsMutex.RLock()
	defer m.poolsMutex.RUnlock()
	return m.pools[address]
}

func (m *Monitor) Tokens() []token.Token {
	m.poolsMutex.RLock()
	defer m.poolsMutex.RUnlock()
	return m.tokens
}
