package uniswap

import (
	"PoolHelper/src3/cache/newcache"
	"PoolHelper/src3/multicaller/generic"
	"PoolHelper/src3/pool"
	unipools "PoolHelper/src3/pool/uniswap"
	"PoolHelper/src3/token"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"sync"
	"time"
)

type Factory[ReserveType any] struct {
	Details newcache.FactoryDetails
	Pools   map[common.Address]pool.Pool[ReserveType]
}

type Cache[ReserveType any] struct {
	tokens    map[common.Address]token.ERC20
	factories map[common.Address]Factory[ReserveType]
	lastSync  uint64
	version   int
	m         *sync.RWMutex
}

func NewV2Cache() *Cache[unipools.Reserves] {
	return &Cache[unipools.Reserves]{
		tokens:    make(map[common.Address]token.ERC20),
		factories: make(map[common.Address]Factory[unipools.Reserves]),
		version:   2,
		m:         &sync.RWMutex{},
	}
}

func NewV3Cache() *Cache[unipools.Slot0] {
	return &Cache[unipools.Slot0]{
		tokens:    make(map[common.Address]token.ERC20),
		factories: make(map[common.Address]Factory[unipools.Slot0]),
		version:   3,
		m:         &sync.RWMutex{},
	}
}

///
/// Token Cache
///

func (m *Cache[any]) AddToken(t token.ERC20) error {
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

func (m *Cache[any]) RemoveToken(address common.Address) error {
	m.m.Lock()
	defer m.m.Unlock()

	// remove token
	delete(m.tokens, address)

	// remove v2 pools
	isDeleted := false
	for _, dex := range m.factories {
		// iterate and remove if pair
		for _, _pool := range dex.Pools {
			if _pool.Pair().TokenA.Address == address || _pool.Pair().TokenB.Address == address {
				delete(dex.Pools, _pool.Address())
				isDeleted = true
			}
		}
	}
	if isDeleted {
		return nil
	}
	return fmt.Errorf("token not found")
}

func (m *Cache[any]) Token(address common.Address) (token.ERC20, bool) {
	m.m.RLock()
	defer m.m.RUnlock()
	val, ok := m.tokens[address]
	return val, ok
}

func (m *Cache[any]) Tokens() []token.ERC20 {
	m.m.RLock()
	defer m.m.RUnlock()

	tokens := make([]token.ERC20, 0)
	for _, _token := range m.tokens {
		tokens = append(tokens, _token)
	}

	return tokens
}

///
/// Pool Cache
///

func (m *Cache[any]) RemovePool(address common.Address) error {
	m.m.Lock()
	defer m.m.Unlock()

	// remove pool
	isDeleted := false
	for _, dex := range m.factories {
		delete(dex.Pools, address)
		if len(dex.Pools) == 0 {
			delete(m.factories, dex.Details.Address)
			isDeleted = true
		}
	}

	if isDeleted {
		return nil
	}
	return fmt.Errorf("pool not found")
}

func (m *Cache[any]) Pools() []pool.Pool[any] {
	m.m.RLock()
	defer m.m.RUnlock()

	pools := make([]pool.Pool[any], 0)
	for _, dex := range m.factories {
		for _, _pool := range dex.Pools {
			pools = append(pools, _pool)
		}
	}

	return pools
}

func (m *Cache[any]) Pool(address common.Address) (pool.Pool[any], bool) {
	m.m.RLock()
	defer m.m.RUnlock()

	for _, dex := range m.factories {
		if _pool, ok := dex.Pools[address]; ok {
			return _pool, true
		}
	}

	return nil, false
}

///
/// Initializers
///

func (m *Cache[any]) ImportTokens(ctx context.Context, tokens []common.Address, multicaller generic.Multicaller, block uint64) ([]token.ERC20, error) {
	// fetch token infos
	tokensWithDetails, err := fetchTokens(ctx, multicaller, tokens, block)
	if err != nil {
		return nil, err
	}

	// update tokens
	m.m.Lock()
	defer m.m.Unlock()
	for _, _token := range tokensWithDetails {
		m.tokens[_token.Address] = _token
	}

	// import pools
	for _, factory := range m.factories {
		_, err = m.ImportPools(factory.Details)
		if err != nil {
			return nil, err
		}
	}

	return tokensWithDetails, nil
}

func (m *Cache[ReserveType]) ImportPools(factory newcache.FactoryDetails) ([]common.Address, error) {
	m.m.Lock()
	defer m.m.Unlock()

	// create dex if not exists
	isFound := false
	if _, isFound = m.factories[factory.Address]; !isFound {
		m.factories[factory.Address] = Factory[ReserveType]{
			Details: factory,
			Pools:   make(map[common.Address]pool.Pool[ReserveType]),
		}
	}

	// iterate over tokens
	newPoolAddrs := make([]common.Address, 0)
	for _, tokenA := range m.tokens {
		for _, tokenB := range m.tokens {
			// skip same token
			if tokenA.Address == tokenB.Address {
				continue
			}

			// create pair
			pair := token.Pair{
				TokenA: tokenA,
				TokenB: tokenB,
			}

			// create pools.
			var poolAddr common.Address
			switch m.version {
			case 2:
				poolAddr = m.createV2Pool(factory, &pair)
			case 3:
				for _, fee := range factory.Options {
					poolAddr = m.createV3Pool(factory, &pair, fee)
				}
			default:
				return nil, fmt.Errorf("invalid pool version")
			}

		}
	}

	// create pools
	newPools := make([]common.Address, 0)
	for _, fee := range factory.Options {
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

				// switch on pool version
				var v2Pool pool.Pool[ReserveType]
				switch m.version {
				case 2:
					_pool := unipools.NewV2Pool(factory.Address, factory.InitCode, &pair)
					poolAddress = _pool.Address()
				case 3:
					// get fee type
					val, ok := fee.(unipools.FeeType)
					if !ok {
						return nil, fmt.Errorf("invalid fee type: %v", fee)
					}
					_pool := unipools.NewUniswapV3Pool(factory.Address, factory.InitCode, &pair, val)
					poolAddress = _pool.Address()
				default:
					return nil, fmt.Errorf("invalid reserve type")
				}

				// add pool to tokens
				if _, ok := m.factories[factory.Address].Pools[poolAddress]; !ok {
					newPools = append(newPools, _pool.Address())
				}

				_pool := unipools.NewUniswapV3Pool(factory, initCode, &pair, fee)

				// add pool to newcache
				if _, ok := m.dexV3[factory].Pools[_pool.Address()]; !ok {
					newPools = append(newPools, _pool.Address())
				}
				m.dexV3[factory].Pools[_pool.Address()] = _pool
			}
		}
	}

	if isFound {
		// combine fees
		var combined []unipools.FeeType
		copy(m.dexV3[factory].PoolOptions, combined)
		for _, fee := range fees {
			if !slices.Contains(m.dexV3[factory].PoolOptions, fee) {
				combined = append(combined, fee)
			}
		}

		// overwrite dex
		m.dexV3[factory] = FactoryDetails[unipools.Slot0, unipools.FeeType]{
			Factory:     factory,
			InitCode:    initCode,
			PoolOptions: combined,
			Pools:       make(map[common.Address]pool.Pool[unipools.Slot0]),
		}
	}

	return newPools
}

func (m *Cache) ImportV2Pools(factory common.Address, initCode common.Hash) []common.Address {
	m.m.Lock()
	defer m.m.Unlock()

	if _, ok := m.factories[factory]; !ok {
		m.factories[factory] = FactoryDetails[unipools.Reserves, any]{
			Factory:  factory,
			InitCode: initCode,
			Pools:    make(map[common.Address]pool.Pool[unipools.Reserves]),
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
			_pool := unipools.NewV2Pool(factory, initCode, &pair)

			// add pool to newcache
			if _, ok := m.factories[factory].Pools[_pool.Address()]; !ok {
				m.factories[factory].Pools[_pool.Address()] = _pool
				newPools = append(newPools, _pool.Address())
			}
		}
	}

	return newPools
}

// ImportV3Pools finds all Uniswap V3 pools for the given factories and init code.
// It does not overwrite the existing pool if it already exists.
// It appends the given fees to the existing fees.
func (m *Cache) ImportV3Pools(factory common.Address, initCode common.Hash, fees []unipools.FeeType) []common.Address {
	m.m.Lock()
	defer m.m.Unlock()

	// create dex if not exists
	isFound := false
	if _, isFound = m.dexV3[factory]; !isFound {
		m.dexV3[factory] = FactoryDetails[unipools.Slot0, unipools.FeeType]{
			Factory:     factory,
			InitCode:    initCode,
			PoolOptions: fees,
			Pools:       make(map[common.Address]pool.Pool[unipools.Slot0]),
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
				_pool := unipools.NewUniswapV3Pool(factory, initCode, &pair, fee)

				// add pool to newcache
				if _, ok := m.dexV3[factory].Pools[_pool.Address()]; !ok {
					newPools = append(newPools, _pool.Address())
				}
				m.dexV3[factory].Pools[_pool.Address()] = _pool
			}
		}
	}

	if isFound {
		// combine fees
		var combined []unipools.FeeType
		copy(m.dexV3[factory].PoolOptions, combined)
		for _, fee := range fees {
			if !slices.Contains(m.dexV3[factory].PoolOptions, fee) {
				combined = append(combined, fee)
			}
		}

		// overwrite dex
		m.dexV3[factory] = FactoryDetails[unipools.Slot0, unipools.FeeType]{
			Factory:     factory,
			InitCode:    initCode,
			PoolOptions: combined,
			Pools:       make(map[common.Address]pool.Pool[unipools.Slot0]),
		}
	}

	return newPools
}

///
/// States
///

// SyncAll updates the pool states for all pools in the newcache.
func (m *Cache) SyncAll(ctx context.Context, multicaller generic.Multicaller, block uint64) (error, time.Duration) {
	// skip if already synced
	if m.lastSync >= block {
		return nil, 0
	}

	var err error
	res := make([]unipools.Reserves, 0)
	slots := make([]unipools.Slot0, 0)

	start := time.Now()
	m.m.RLock()

	// fetch uni_pools v2 reserves
	v2Pools := make([]pool.Pool[unipools.Reserves], 0)
	for _, dex := range m.factories {
		for _, _pool := range dex.Pools {
			v2Pools = append(v2Pools, _pool)
		}
	}
	if len(v2Pools) > 0 {
		// pools to addr
		poolAddrs := make([]common.Address, 0)
		for _, _pool := range v2Pools {
			poolAddrs = append(poolAddrs, _pool.Address())
		}

		res, err = fetchReserves(ctx, multicaller, poolAddrs, block)
		if err != nil {
			return err, 0
		}
	}

	// fetch uni_pools v3 slots
	v3Pools := m.PoolsV3()
	if len(v3Pools) > 0 {
		// pools to addr
		poolAddrs := make([]common.Address, 0)
		for _, _pool := range v3Pools {
			poolAddrs = append(poolAddrs, _pool.Address())
		}

		slots, err = fetchSlots(ctx, multicaller, poolAddrs, block)
		if err != nil {
			return err, 0
		}
	}

	m.m.RUnlock()
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
	return nil, time.Now().Sub(start)
}

// Sync updates the pool states for the given pools.
func (m *Cache) Sync(ctx context.Context, multicaller generic.Multicaller, pools []common.Address, block uint64) (error, time.Duration) {
	// skip if already synced
	if m.lastSync >= block {
		return nil, 0
	}

	var err error
	res := make([]unipools.Reserves, 0)
	slots := make([]unipools.Slot0, 0)

	start := time.Now()
	m.m.RLock()

	// fetch uni_pools v2 reserves
	v2Pools := make([]pool.Pool[unipools.Reserves], 0)
	for _, _pool := range pools {
		if _poolV2, ok := m.factories[_pool]; ok {
			v2Pools = append(v2Pools, _poolV2.Pools[_pool])
		}
	}
	if len(v2Pools) > 0 {
		// pools to addr
		poolAddrs := make([]common.Address, 0)
		for _, _pool := range v2Pools {
			poolAddrs = append(poolAddrs, _pool.Address())
		}

		res, err = fetchReserves(ctx, multicaller, poolAddrs, block)
		if err != nil {
			return err, 0
		}
	}

	// fetch uni_pools v3 slots
	v3Pools := make([]pool.Pool[unipools.Slot0], 0)
	for _, _pool := range pools {
		if _poolV3, ok := m.dexV3[_pool]; ok {
			v3Pools = append(v3Pools, _poolV3.Pools[_pool])
		}
	}
	if len(v3Pools) > 0 {
		// pools to addr
		poolAddrs := make([]common.Address, 0)
		for _, _pool := range v3Pools {
			poolAddrs = append(poolAddrs, _pool.Address())
		}

		slots, err = fetchSlots(ctx, multicaller, poolAddrs, block)
		if err != nil {
			return err, 0
		}
	}
	m.m.RUnlock()

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
	return nil, time.Now().Sub(start)
}

// LastSyncBlock returns the last synced block number.
func (m *Cache) LastSyncBlock() uint64 {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.lastSync
}

///
/// Utils
///

func fetchSlots(ctx context.Context, m generic.Multicaller, targets []common.Address, blockNumber uint64) ([]unipools.Slot0, error) {
	// prepare calls
	calls := make([]generic.Call3, len(targets))
	for i, target := range targets {
		calls[i] = generic.Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("slot0()"))[:4],
			AllowFailure: true,
		}
	}

	// call
	result, err := m.Aggregate(ctx, calls, blockNumber)
	if err != nil {
		return nil, err
	}

	// decode results
	slots := make([]unipools.Slot0, len(targets))
	for i, data := range result {
		// pair doesn't exist.
		if len(data.ReturnData) == 0 {
			slots[i] = unipools.Slot0{
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

		if len(data.ReturnData) != 224 {
			return nil, errors.New(fmt.Sprintf("wrong return data length: %v", len(data.ReturnData)))
		}

		slot := unipools.Slot0{
			SqrtPriceX96:               new(big.Int).SetBytes(data.ReturnData[0:32]),
			Tick:                       new(big.Int).SetBytes(data.ReturnData[32:64]),
			ObservationIndex:           new(big.Int).SetBytes(data.ReturnData[64:96]),
			ObservationCardinality:     new(big.Int).SetBytes(data.ReturnData[96:128]),
			ObservationCardinalityNext: new(big.Int).SetBytes(data.ReturnData[128:160]),
			FeeProtocol:                new(big.Int).SetBytes(data.ReturnData[160:192]),
			Unlocked:                   data.ReturnData[223] != 0,
		}

		// store in the final result
		slots[i] = slot
	}

	return slots, nil
}

func fetchReserves(ctx context.Context, m generic.Multicaller, targets []common.Address, blockNumber uint64) ([]unipools.Reserves, error) {
	// prepare calls
	calls := make([]generic.Call3, len(targets))
	for i, target := range targets {
		calls[i] = generic.Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("getReserves()"))[:4],
			AllowFailure: true,
		}
	}

	// call
	results, err := m.Aggregate(ctx, calls, blockNumber)
	if err != nil {
		return nil, err
	}

	// decode results
	reserves := make([]unipools.Reserves, len(targets))
	for i, result := range results {
		if len(result.ReturnData) == 0 {
			reserves[i] = unipools.Reserves{
				Reserve0: big.NewInt(0),
				Reserve1: big.NewInt(0),
			}
			continue
		}

		if len(result.ReturnData) != 32*3 {
			return nil, errors.New(fmt.Sprintf("wrong return data length: %v", len(result.ReturnData)))
		}

		reserve0 := new(big.Int).SetBytes(result.ReturnData[0:32])
		reserve1 := new(big.Int).SetBytes(result.ReturnData[32:64])

		// store in the final result
		reserves[i] = unipools.Reserves{
			Reserve0: reserve0,
			Reserve1: reserve1,
		}
	}

	return reserves, nil
}

func fetchTokens(ctx context.Context, m generic.Multicaller, targets []common.Address, blockNumber uint64) ([]token.ERC20, error) {
	// prepare calls
	calls := make([]generic.Call3, 0)
	for _, target := range targets {
		// decimals
		calls = append(calls, generic.Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("decimals()"))[:4],
			AllowFailure: false,
		})

		// symbol
		calls = append(calls, generic.Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("symbol()"))[:4],
			AllowFailure: false,
		})

		// name
		calls = append(calls, generic.Call3{
			Target:       target,
			CallData:     crypto.Keccak256([]byte("name()"))[:4],
			AllowFailure: false,
		})
	}

	// call
	results, err := m.Aggregate(ctx, calls, blockNumber)
	if err != nil {
		return nil, err
	}

	// decode results
	tokens := make([]token.ERC20, 0)
	stringType, _ := abi.NewType("string", "", nil)
	for i := 0; i < len(results); i += 3 {
		// validate data
		if len(results[i].ReturnData) != 32 {
			return nil, errors.New(fmt.Sprintf("invalid return data length for decimals: %v", len(results[i].ReturnData)))
		}
		if len(results[i+1].ReturnData) < 32 {
			return nil, errors.New(fmt.Sprintf("invalid return data length for symbol: %v", len(results[i+1].ReturnData)))
		}
		if len(results[i+2].ReturnData) < 32 {
			return nil, errors.New(fmt.Sprintf("invalid return data length for name: %v", len(results[i+2].ReturnData)))
		}

		// decode data
		args := abi.Arguments{
			{
				Name: "str",
				Type: stringType,
			},
		}
		decimals := new(big.Int).SetBytes(results[i].ReturnData)
		symbol, err := args.Unpack(results[i+1].ReturnData)
		name, err := args.Unpack(results[i+2].ReturnData)
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
