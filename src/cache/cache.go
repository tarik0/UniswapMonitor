package cache

import (
	"PoolHelper/src/multicall/generic"
	"PoolHelper/src/pool"
	"PoolHelper/src/structs/factory"
	"PoolHelper/src/structs/token"
	"context"
	"github.com/ethereum/go-ethereum/common"
)

// TokenCache is an interface for adding and removing ERC20 tokens
type TokenCache interface {
	AddToken(token.ERC20) error
	ImportTokens(context.Context, generic.Multicall, []common.Address) error
	RemoveToken(common.Address) error
	Token(common.Address) (token.ERC20, error)
	Tokens() ([]token.ERC20, error)
}

// PoolCache is an interface for adding and removing pools
type PoolCache[ReserveType any, OptionType any] interface {
	InitializePools(factory.Factory[OptionType]) error
	RemovePool(common.Address) error
	Pool(common.Address) (pool.Pool[ReserveType, OptionType], error)
	Pools() []pool.Pool[ReserveType, OptionType]
}

// ReserveCache is an interface for updating pool reserves
type ReserveCache[ReserveType any] interface {
	SyncAll(context.Context, generic.Multicall, uint64) error
	Sync(context.Context, generic.Multicall, []common.Address, uint64) error
	LastSynced() uint64
}

type DEXCache[ReserveType any, OptionType any] interface {
	TokenCache
	PoolCache[ReserveType, OptionType]
	ReserveCache[ReserveType]
}
