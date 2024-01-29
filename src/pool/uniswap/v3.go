package uniswap

import (
	"PoolHelper/src/structs/pair"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"time"
)

type V3FeeType uint64

const (
	MAX    V3FeeType = 10000
	NORMAL V3FeeType = 3000
	LOW    V3FeeType = 500
	MIN    V3FeeType = 100
)

type V3Pool struct {
	pair     pair.Pair[V3FeeType]
	factory  common.Address
	initHash common.Hash

	// slot
	slot                Slot0
	lastUpdateBlock     uint64
	lastUpdateTimestamp uint64
}

func NewV3Pool(factory common.Address, initHash common.Hash, pair pair.Pair[V3FeeType]) *V3Pool {
	return &V3Pool{
		pair:     pair,
		factory:  factory,
		initHash: initHash,
	}
}

///
/// Slot
///

type Slot0 struct {
	SqrtPriceX96               *big.Int
	Tick                       *big.Int
	ObservationIndex           *big.Int
	ObservationCardinality     *big.Int
	ObservationCardinalityNext *big.Int
	FeeProtocol                *big.Int
	Unlocked                   bool
}

func (p *V3Pool) Pair() pair.Pair[V3FeeType] {
	return p.pair
}

func (p *V3Pool) Address() common.Address {
	token0, token1 := p.pair.SortAddresses()

	// abi.encode(token0, token1, fee)
	addrType, _ := abi.NewType("address", "", nil)
	uint24Type, _ := abi.NewType("uint24", "", nil)
	encodedData, err := abi.Arguments{
		{Type: addrType},
		{Type: addrType},
		{Type: uint24Type},
	}.Pack(token0, token1, new(big.Int).SetUint64(uint64(p.pair.PairOptions)))
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

func (p *V3Pool) Update(slot Slot0, block uint64) {
	p.slot = slot
	p.lastUpdateBlock = block
	p.lastUpdateTimestamp = uint64(time.Now().Unix())
}

func (p *V3Pool) State() (Slot0, uint64, uint64) {
	return p.slot, p.lastUpdateBlock, p.lastUpdateTimestamp
}

func (p *V3Pool) Factory() common.Address {
	return p.factory
}
