package token

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

///
/// ERC20
/// Represents an ERC20 token.

type ERC20 struct {
	Address  common.Address
	Decimals *big.Int
	Name     string
	Symbol   string
}

///
/// Pair
/// Represents a pair of tokens.

type Pair struct {
	TokenA ERC20
	TokenB ERC20
}

func (p Pair) Equals(other Pair) bool {
	p0t0, p0t1 := p.Sort()
	p1t0, p1t1 := other.Sort()
	return p0t0 == p1t0 && p0t1 == p1t1
}

func (p Pair) Sort() (common.Address, common.Address) {
	if p.TokenA.Address.Hex() < p.TokenB.Address.Hex() {
		return p.TokenA.Address, p.TokenB.Address
	}
	return p.TokenB.Address, p.TokenA.Address
}

func (p Pair) Reverse() Pair {
	return Pair{
		TokenA: p.TokenB,
		TokenB: p.TokenA,
	}
}
