package uniswap_v3

import (
	"PoolHelper/src/pool"
	"PoolHelper/src/token"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"sync"
	"time"
)

type FeeType uint64

const (
	MAX    FeeType = 10000
	NORMAL FeeType = 3000
	LOW    FeeType = 500
	MIN    FeeType = 100
)

type UniswapV3Pool struct {
	pair     *token.Pair
	factory  common.Address
	fee      FeeType
	initHash common.Hash

	// slot
	slot                Slot0
	lastUpdateBlock     uint64
	lastUpdateTimestamp uint64
	m                   *sync.RWMutex
}

func NewUniswapV3Pool(factory common.Address, initHash common.Hash, pair *token.Pair, fee FeeType) *UniswapV3Pool {
	return &UniswapV3Pool{
		pair:     pair,
		factory:  factory,
		fee:      fee,
		initHash: initHash,
	}
}

///
/// Slot
///

type Slot0 struct {
	SqrtPriceX96               *big.Int
	Tick                       *big.Int
	ObservationIndex           uint16
	ObservationCardinality     uint16
	ObservationCardinalityNext uint16
	FeeProtocol                uint8
	Unlocked                   bool
}

func (p *UniswapV3Pool) UpdateSafe(slot Slot0, block uint64) {
	p.m.Lock()
	defer p.m.Unlock()

	p.slot = slot
	p.lastUpdateBlock = block
	p.lastUpdateTimestamp = uint64(time.Now().Unix())
}

func (p *UniswapV3Pool) Update(slot Slot0, block uint64) {
	p.slot = slot
	p.lastUpdateBlock = block
	p.lastUpdateTimestamp = uint64(time.Now().Unix())
}

func (p *UniswapV3Pool) Slot0() (Slot0, uint64, uint64) {
	p.m.RLock()
	defer p.m.RUnlock()

	return p.slot, p.lastUpdateBlock, p.lastUpdateTimestamp
}

///
/// Pool Implementation
///

func (p *UniswapV3Pool) Type() pool.Type {
	return pool.UniswapV3
}

func (p *UniswapV3Pool) Pair() token.Pair {
	return *p.pair
}

func (p *UniswapV3Pool) Address() common.Address {
	token0, token1 := p.pair.Sort()

	// abi.encode(token0, token1, fee)
	addrType, _ := abi.NewType("address", "", nil)
	uint24Type, _ := abi.NewType("uint24", "", nil)
	encodedData, err := abi.Arguments{
		{Type: addrType},
		{Type: addrType},
		{Type: uint24Type},
	}.Pack(token0, token1, new(big.Int).SetUint64(uint64(p.fee)))
	if err != nil {
		panic(err)
	}

	data := append([]byte{0xff}, p.factory.Bytes()...)
	data = append(data, crypto.Keccak256(encodedData)...)
	data = append(data, p.initHash.Bytes()[:]...)

	hash := crypto.Keccak256(data)
	addressBytes := hash[:]

	return common.BytesToAddress(addressBytes)
}
