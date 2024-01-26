package pair

import (
	"PoolHelper/src/structs/token"
	"github.com/ethereum/go-ethereum/common"
)

type Pair struct {
	TokenA token.ERC20
	TokenB token.ERC20
}

func (p Pair) Equals(other Pair) bool {
	p0t0, p0t1 := p.SortAddresses()
	p1t0, p1t1 := other.SortAddresses()
	return p0t0 == p1t0 && p0t1 == p1t1
}

func (p Pair) SortAddresses() (common.Address, common.Address) {
	if p.TokenA.Address.Hex() < p.TokenB.Address.Hex() {
		return p.TokenA.Address, p.TokenB.Address
	}
	return p.TokenB.Address, p.TokenA.Address
}

func (p Pair) SortTokens() (token.ERC20, token.ERC20) {
	if p.TokenA.Address.Hex() < p.TokenB.Address.Hex() {
		return p.TokenA, p.TokenB
	}
	return p.TokenB, p.TokenA
}

func (p Pair) Reverse() Pair {
	return Pair{
		TokenA: p.TokenB,
		TokenB: p.TokenA,
	}
}

func (p Pair) String() string {
	// sort & return
	t0, t1 := p.SortTokens()
	return t0.Symbol + "/" + t1.Symbol
}
