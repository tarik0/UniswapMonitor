package token

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ERC20 struct {
	Address  common.Address
	Decimals *big.Int
	Name     string
	Symbol   string
}
