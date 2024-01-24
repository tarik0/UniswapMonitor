package cache

import (
	"PoolHelper/src/multicaller"
	"PoolHelper/src/pool"
	uniswapv2 "PoolHelper/src/pool/uniswap-v2"
	uniswapv3 "PoolHelper/src/pool/uniswap-v3"
	"PoolHelper/src/token"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"slices"
	"sync"
)

type PoolType int

type DEXCache[R any, F any] struct {
	Factory  common.Address
	InitCode common.Hash
	Fees     []F
	Pools    map[common.Address]pool.Pool[R]
}

type Cache struct {
	tokens map[common.Address]token.ERC20

	dexV2 map[common.Address]DEXCache[uniswapv2.Reserves, any]
	dexV3 map[common.Address]DEXCache[uniswapv3.Slot0, uniswapv3.FeeType]

	lastSync uint64
	m        *sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		tokens: make(map[common.Address]token.ERC20),
		dexV2:  make(map[common.Address]DEXCache[uniswapv2.Reserves, any]),
		dexV3:  make(map[common.Address]DEXCache[uniswapv3.Slot0, uniswapv3.FeeType]),
		m:      &sync.RWMutex{},
	}
}

///
/// Tokens
///

// AddToken adds a token to the cache.
// It overwrites the existing token if it already exists.
func (m *Cache) AddToken(t token.ERC20) error {
	// validate token
	if t.Address.Cmp(common.Address{}) == 0 {
		return fmt.Errorf("invalid token address")
	}
	if t.Decimals == nil || t.Decimals.Cmp(common.Big0) == 0 {
		return fmt.Errorf("invalid token decimals")
	}

	m.m.Lock()
	defer m.m.Unlock()

	// add token
	m.tokens[t.Address] = t
	return nil
}

// RemoveToken removes a token from the cache.
// It does remove any pools that use this token.
func (m *Cache) RemoveToken(address common.Address) {
	m.m.Lock()
	defer m.m.Unlock()

	// remove token
	delete(m.tokens, address)

	// remove v2 pools
	for _, dex := range m.dexV2 {
		// iterate and remove if pair
		for _, _pool := range dex.Pools {
			if _pool.Pair().TokenA.Address == address || _pool.Pair().TokenB.Address == address {
				delete(dex.Pools, _pool.Address())
			}
		}
	}

	// remove v2 pools
	for _, dex := range m.dexV3 {
		// iterate and remove if pair
		for _, _pool := range dex.Pools {
			if _pool.Pair().TokenA.Address == address || _pool.Pair().TokenB.Address == address {
				delete(dex.Pools, _pool.Address())
			}
		}
	}
}

func (m *Cache) Token(address common.Address) (token.ERC20, bool) {
	m.m.RLock()
	defer m.m.RUnlock()
	val, ok := m.tokens[address]
	return val, ok
}

func (m *Cache) Tokens() []token.ERC20 {
	m.m.RLock()
	defer m.m.RUnlock()

	tokens := make([]token.ERC20, 0)
	for _, _token := range m.tokens {
		tokens = append(tokens, _token)
	}

	return tokens
}

///
/// Pools
///

// RemovePool removes a pool from the cache.
func (m *Cache) RemovePool(address common.Address) {
	m.m.Lock()
	defer m.m.Unlock()

	// remove pool
	for _, dex := range m.dexV2 {
		delete(dex.Pools, address)
		if len(dex.Pools) == 0 {
			delete(m.dexV2, dex.Factory)
		}
	}
	for _, dex := range m.dexV3 {
		delete(dex.Pools, address)
		if len(dex.Pools) == 0 {
			delete(m.dexV3, dex.Factory)
		}
	}
}

func (m *Cache) PoolsV2() []pool.Pool[uniswapv2.Reserves] {
	m.m.RLock()
	defer m.m.RUnlock()

	pools := make([]pool.Pool[uniswapv2.Reserves], 0)
	for _, dex := range m.dexV2 {
		for _, _pool := range dex.Pools {
			pools = append(pools, _pool)
		}
	}

	return pools
}

func (m *Cache) PoolsV3() []pool.Pool[uniswapv3.Slot0] {
	m.m.RLock()
	defer m.m.RUnlock()

	pools := make([]pool.Pool[uniswapv3.Slot0], 0)
	for _, dex := range m.dexV3 {
		for _, _pool := range dex.Pools {
			pools = append(pools, _pool)
		}
	}

	return pools
}

func (m *Cache) Pools() []interface{} {
	m.m.RLock()
	defer m.m.RUnlock()

	pools := make([]interface{}, 0)
	for _, dex := range m.dexV2 {
		for _, _pool := range dex.Pools {
			pools = append(pools, _pool)
		}
	}
	for _, dex := range m.dexV3 {
		for _, _pool := range dex.Pools {
			pools = append(pools, _pool)
		}
	}

	return pools
}

///
/// Initializers
///

// ImportV2Pools finds all Uniswap V2 pools for the given factory and init code.
// It does not overwrite the existing pool if it already exists.
func (m *Cache) ImportV2Pools(factory common.Address, initCode common.Hash) []common.Address {
	m.m.Lock()
	defer m.m.Unlock()

	if _, ok := m.dexV2[factory]; !ok {
		m.dexV2[factory] = DEXCache[uniswapv2.Reserves, any]{
			Factory:  factory,
			InitCode: initCode,
			Pools:    make(map[common.Address]pool.Pool[uniswapv2.Reserves]),
		}
	}

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
			if _, ok := m.dexV2[factory].Pools[_pool.Address()]; !ok {
				m.dexV2[factory].Pools[_pool.Address()] = _pool
				newPools = append(newPools, _pool.Address())
			}
		}
	}

	return newPools
}

// ImportV3Pools finds all Uniswap V3 pools for the given factory and init code.
// It does not overwrite the existing pool if it already exists.
// It appends the given fees to the existing fees.
func (m *Cache) ImportV3Pools(factory common.Address, initCode common.Hash, fees []uniswapv3.FeeType) []common.Address {
	m.m.Lock()
	defer m.m.Unlock()

	// create dex if not exists
	isFound := false
	if _, isFound = m.dexV3[factory]; !isFound {
		m.dexV3[factory] = DEXCache[uniswapv3.Slot0, uniswapv3.FeeType]{
			Factory:  factory,
			InitCode: initCode,
			Fees:     fees,
			Pools:    make(map[common.Address]pool.Pool[uniswapv3.Slot0]),
		}
	}

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
				_pool := uniswapv3.NewUniswapV3Pool(factory, initCode, &pair, fee)

				// add pool to cache
				if _, ok := m.dexV3[factory].Pools[_pool.Address()]; !ok {
					newPools = append(newPools, _pool.Address())
				}
				m.dexV3[factory].Pools[_pool.Address()] = _pool
			}
		}
	}

	if isFound {
		// combine fees
		var combined []uniswapv3.FeeType
		copy(m.dexV3[factory].Fees, combined)
		for _, fee := range fees {
			if !slices.Contains(m.dexV3[factory].Fees, fee) {
				combined = append(combined, fee)
			}
		}

		// overwrite dex
		m.dexV3[factory] = DEXCache[uniswapv3.Slot0, uniswapv3.FeeType]{
			Factory:  factory,
			InitCode: initCode,
			Fees:     combined,
			Pools:    make(map[common.Address]pool.Pool[uniswapv3.Slot0]),
		}
	}

	return newPools
}

// ImportTokens finds all tokens for the given addresses.
// It overwrites the existing token if it already exists.
// It also imports all pools for the given tokens.
func (m *Cache) ImportTokens(ctx context.Context, tokens []common.Address, multicaller multicaller.Multicaller, block uint64) ([]token.ERC20, error) {
	// fetch token infos
	tokensWithDetails, err := multicaller.FetchTokens(ctx, tokens, block)
	if err != nil {
		return nil, err
	}

	// update tokens
	m.m.Lock()
	defer m.m.Unlock()
	for _, _token := range tokensWithDetails {
		m.tokens[_token.Address] = _token
	}

	// import v2 pools
	for _, dex := range m.dexV2 {
		m.ImportV2Pools(dex.Factory, dex.InitCode)
	}

	// import v3 pools
	for _, dex := range m.dexV3 {
		m.ImportV3Pools(dex.Factory, dex.InitCode, dex.Fees)
	}

	return tokensWithDetails, nil
}

///
/// States
///

// Sync updates the pool states for all pools in the cache.
func (m *Cache) Sync(ctx context.Context, multicaller multicaller.Multicaller, block uint64) error {
	// skip if already synced
	if m.lastSync >= block {
		return nil
	}

	var err error
	res := make([]uniswapv2.Reserves, 0)
	slots := make([]uniswapv3.Slot0, 0)

	// fetch uniswap v2 reserves
	v2Pools := m.PoolsV2()
	if len(v2Pools) > 0 {
		// pools to addr
		poolAddrs := make([]common.Address, 0)
		for _, _pool := range v2Pools {
			poolAddrs = append(poolAddrs, _pool.Address())
		}

		res, err = multicaller.FetchReserves(ctx, poolAddrs, block)
		if err != nil {
			return err
		}
	}

	// fetch uniswap v3 slots
	v3Pools := m.PoolsV3()
	if len(v3Pools) > 0 {
		// pools to addr
		poolAddrs := make([]common.Address, 0)
		for _, _pool := range v3Pools {
			poolAddrs = append(poolAddrs, _pool.Address())
		}

		slots, err = multicaller.FetchSlots(ctx, poolAddrs, block)
		if err != nil {
			return err
		}
	}

	m.m.Lock()
	defer m.m.Unlock()

	// update v2 pairs
	for i, _pool := range v2Pools {
		// skip empty pairs
		if res[i].Reserve0 == nil || res[i].Reserve1 == nil {
			continue
		}
		_pool.Update(res[i], block)
	}

	// update v3 pools
	for i, _pool := range v3Pools {
		// skip empty pools
		if slots[i].SqrtPriceX96 == nil {
			continue
		}
		_pool.Update(slots[i], block)
	}

	m.lastSync = block
	return nil
}

// Block returns the last synced block number.
func (m *Cache) Block() uint64 {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.lastSync
}
