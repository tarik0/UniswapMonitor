package uniswap_v2

import (
	"PoolHelper/src/pool"
	"PoolHelper/src/token"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"sync"
)

type UniswapV2Pool struct {
	factory  common.Address
	pair     *token.Pair
	reserve0 *big.Int
	reserve1 *big.Int
	initHash common.Hash

	lastUpdateBlock uint64
	m               *sync.RWMutex
}

func NewUniswapV2Pool(factory common.Address, initCode common.Hash, pair *token.Pair) *UniswapV2Pool {
	return &UniswapV2Pool{
		pair:            pair,
		factory:         factory,
		reserve0:        big.NewInt(0),
		reserve1:        big.NewInt(0),
		m:               &sync.RWMutex{},
		initHash:        initCode,
		lastUpdateBlock: 0,
	}
}

///
/// Reserves
///

func (p *UniswapV2Pool) UpdateSafe(reserve0 *big.Int, reserve1 *big.Int, block uint64) {
	p.m.Lock()
	defer p.m.Unlock()

	p.reserve0.Set(reserve0)
	p.reserve1.Set(reserve1)
	p.lastUpdateBlock = block
}

func (p *UniswapV2Pool) Update(reserve0 *big.Int, reserve1 *big.Int, block uint64) {
	p.reserve0.Set(reserve0)
	p.reserve1.Set(reserve1)
	p.lastUpdateBlock = block
}

///
/// Pool Implementation
///

func (p *UniswapV2Pool) Type() pool.Type {
	return pool.UniswapV2
}

func (p *UniswapV2Pool) Pair() token.Pair {
	return *p.pair
}

func (p *UniswapV2Pool) Address() common.Address {
	token0, token1 := p.pair.Sort()

	data := append([]byte{0xff}, p.factory.Bytes()...)
	data = append(data, crypto.Keccak256(token0.Bytes(), token1.Bytes())...)
	data = append(data, p.initHash.Bytes()[:]...)

	hash := crypto.Keccak256(data)
	addressBytes := hash[:]

	return common.BytesToAddress(addressBytes)
}

func (p *UniswapV2Pool) Reserves() (*big.Int, *big.Int, uint64) {
	p.m.RLock()
	defer p.m.RUnlock()

	// return a copy of the reserves
	return new(big.Int).Set(p.reserve0), new(big.Int).Set(p.reserve1), p.lastUpdateBlock
}
