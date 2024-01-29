package pair

import (
	"PoolHelper/src/structs/token"
	"bytes"
	"github.com/ethereum/go-ethereum/common"
)

type Pair[PairOption any] struct {
	TokenA      token.ERC20
	TokenB      token.ERC20
	PairOptions PairOption
}

func NewPair[PairOption any](tokenA token.ERC20, tokenB token.ERC20, options PairOption) Pair[PairOption] {
	return Pair[PairOption]{
		TokenA:      tokenA,
		TokenB:      tokenB,
		PairOptions: options,
	}
}

func (p Pair[any]) Equals(other Pair[any]) bool {
	p0t0, p0t1 := p.SortAddresses()
	p1t0, p1t1 := other.SortAddresses()
	return p0t0 == p1t0 && p0t1 == p1t1
}

func (p Pair[any]) SortAddresses() (common.Address, common.Address) {
	if p.TokenA.Address.Hex() < p.TokenB.Address.Hex() {
		return p.TokenA.Address, p.TokenB.Address
	}
	return p.TokenB.Address, p.TokenA.Address
}

func (p Pair[any]) SortTokens() (token.ERC20, token.ERC20) {
	if p.TokenA.Address.Hex() < p.TokenB.Address.Hex() {
		return p.TokenA, p.TokenB
	}
	return p.TokenB, p.TokenA
}

func (p Pair[any]) Reverse() Pair[any] {
	return Pair[any]{
		TokenA: p.TokenB,
		TokenB: p.TokenA,
	}
}

func (p Pair[any]) String() string {
	// sort & return
	t0, t1 := p.SortTokens()
	return t0.Symbol + "/" + t1.Symbol
}

func (p Pair[any]) Options() any {
	return p.PairOptions
}

func (p Pair[any]) Contains(addr common.Address) bool {
	return bytes.EqualFold(p.TokenA.Address.Bytes(), addr.Bytes()) ||
		bytes.EqualFold(p.TokenB.Address.Bytes(), addr.Bytes())
}

func (p Pair[any]) IsValid() bool {
	return p.TokenA.IsValid() && p.TokenB.IsValid()
}
