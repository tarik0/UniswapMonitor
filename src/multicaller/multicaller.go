package multicaller

import (
	uniswapv2 "PoolHelper/src/pool/uniswap-v2"
	uniswapv3 "PoolHelper/src/pool/uniswap-v3"
	"PoolHelper/src/token"
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ClientDispatcher interface {
	CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error)
}

type Multicaller interface {
	FetchReserves(context.Context, []common.Address, uint64) ([]uniswapv2.Reserves, error)
	FetchSlots(context.Context, []common.Address, uint64) ([]uniswapv3.Slot0, error)
	FetchTokens(context.Context, []common.Address, uint64) ([]token.ERC20, error)
}
