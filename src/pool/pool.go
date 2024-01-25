package pool

import (
	"PoolHelper/src/token"
	"github.com/ethereum/go-ethereum/common"
)

///
/// Pool
/// Represents a generic trading pool.

type Pool[State any] interface {
	Pair() token.Pair
	Address() common.Address
	Update(State, uint64)
	State() (State, uint64, uint64)

	// todo: PriceOf(token.ERC20) (*big.Int, uint64)
}
