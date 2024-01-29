package multicall

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"math/big"
)

type ClientDispatcher interface {
	CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error)
}

type Multicaller[CallType any, ResultType any] interface {
	Aggregate(context.Context, []CallType, uint64) ([]ResultType, error)
}
