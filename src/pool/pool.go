package pool

import (
	"PoolHelper/src/token"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

///
/// Pool
/// Represents a generic trading pool.

type Type uint8

const (
	UniswapV2 Type = iota + 1
	UniswapV3
)

type Pool interface {
	Type() Type
	Pair() token.Pair
	Address() common.Address
	Reserves() (*big.Int, *big.Int, uint64)
}
