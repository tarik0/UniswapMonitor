package cache

import (
	"PoolHelper/src/token"
	"github.com/ethereum/go-ethereum/common"
)

type Cache interface {
	AddToken(token.ERC20) error
	RemoveToken(address common.Address) error
}
