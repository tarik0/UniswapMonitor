package uniswap_v3

import (
	"PoolHelper/src/pool"
	"PoolHelper/src/token"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

type UniswapV3Pool struct {
	pair     *token.Pair
	factory  common.Address
	fee      uint64
	initHash common.Hash

	// slot
	sqrtPriceX96               *big.Int
	tick                       int64
	observationIndex           uint16
	observationCardinality     uint16
	observationCardinalityNext uint16
	feeProtocol                uint8
	unlocked                   bool
}

func NewUniswapV3Pool(factory common.Address, initHash common.Hash, pair *token.Pair, fee uint64) *UniswapV3Pool {
	return &UniswapV3Pool{
		pair:     pair,
		factory:  factory,
		fee:      fee,
		initHash: initHash,
	}
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
	}.Pack(token0, token1, new(big.Int).SetUint64(p.fee))
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

func (p *UniswapV3Pool) Reserves() (*big.Int, *big.Int, uint64) {
	return nil, nil, 0
}
