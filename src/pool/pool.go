package pool

import (
	"PoolHelper/src/structs/pair"
	"github.com/ethereum/go-ethereum/common"
)

type Pool[State any] interface {
	Pair() pair.Pair
	Address() common.Address
	Update(State, uint64)
	State() (State, uint64, uint64)
}
