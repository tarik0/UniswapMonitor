package pool

import (
	"PoolHelper/src/structs/pair"
	"github.com/ethereum/go-ethereum/common"
)

type Pool[ReserveType any, PairOption any] interface {
	Pair() pair.Pair[PairOption]
	Factory() common.Address
	Address() common.Address
	Update(ReserveType, uint64)
	State() (ReserveType, uint64, uint64)
}
