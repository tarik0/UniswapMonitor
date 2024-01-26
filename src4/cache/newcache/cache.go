package newcache

import (
	"PoolHelper/src3/multicaller/generic"
	"PoolHelper/src3/pool"
	"PoolHelper/src3/token"
	"context"
	"github.com/ethereum/go-ethereum/common"
)

type FactoryDetails struct {
	Address  common.Address
	InitCode common.Hash
	Options  []any
}

// TokenCache is an interface for caching ERC20 tokens.
type TokenCache interface {
	AddToken(token.ERC20) error
	RemoveToken(address common.Address) error
	Token(common.Address) (token.ERC20, error)
	Tokens() ([]token.ERC20, error)
}

// PoolCache is an interface for caching pools.
type PoolCache[ReserveType any] interface {
	RemovePool(address common.Address) error
	Pool(common.Address) (pool.Pool[ReserveType], error)
	Pools() ([]pool.Pool[ReserveType], error)
}

// Initializer is an interface for initializing the newcache.
type Initializer[PoolOptions any] interface {
	ImportTokens(context.Context, []common.Address, generic.Multicaller, uint64) ([]token.ERC20, error)
	ImportPools(FactoryDetails[PoolOptions]) ([]common.Address, error)
}

// DEXCache is an interface for caching ERC20 tokens and pools.
type DEXCache[PoolOptions any, ReserveType any] interface {
	TokenCache
	PoolCache[ReserveType]
	Initializer[PoolOptions]
}
