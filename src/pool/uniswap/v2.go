package uniswap

import (
	"PoolHelper/src/structs/pair"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"time"
)

type V2Pool struct {
	factory  common.Address
	pair     pair.Pair[any]
	reserve0 *big.Int
	reserve1 *big.Int
	initHash common.Hash

	lastUpdateBlock     uint64
	lastUpdateTimestamp uint64
}

func NewV2Pool(factory common.Address, initCode common.Hash, pair pair.Pair[any]) *V2Pool {
	return &V2Pool{
		pair:                pair,
		factory:             factory,
		reserve0:            big.NewInt(0),
		reserve1:            big.NewInt(0),
		initHash:            initCode,
		lastUpdateBlock:     0,
		lastUpdateTimestamp: 0,
	}
}

///
/// State
///

type Reserves struct {
	Reserve0 *big.Int
	Reserve1 *big.Int
}

func (p *V2Pool) Pair() pair.Pair[any] {
	return p.pair
}

func (p *V2Pool) Address() common.Address {
	token0, token1 := p.pair.SortAddresses()

	data := append([]byte{0xff}, p.factory.Bytes()...)
	data = append(data, crypto.Keccak256(token0.Bytes(), token1.Bytes())...)
	data = append(data, p.initHash.Bytes()[:]...)

	hash := crypto.Keccak256(data)
	addressBytes := hash[:]

	return common.BytesToAddress(addressBytes)
}

func (p *V2Pool) Update(res Reserves, block uint64) {
	p.reserve0.Set(res.Reserve0)
	p.reserve1.Set(res.Reserve1)
	p.lastUpdateTimestamp = uint64(time.Now().Unix())
	p.lastUpdateBlock = block
}

func (p *V2Pool) State() (Reserves, uint64, uint64) {
	return Reserves{
		Reserve0: new(big.Int).Set(p.reserve0),
		Reserve1: new(big.Int).Set(p.reserve1),
	}, p.lastUpdateBlock, p.lastUpdateTimestamp
}

func (p *V2Pool) Factory() common.Address {
	return p.factory
}
