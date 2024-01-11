package token

import "github.com/ethereum/go-ethereum/common"

///
/// Token
/// Represents an ERC20 token.

type Token struct {
	Address  common.Address
	Decimals uint8
	Name     string
	Symbol   string
}

///
/// Pair
/// Represents a pair of tokens.

type Pair struct {
	TokenA Token
	TokenB Token
}

func (p Pair) Equals(other Pair) bool {
	return p.TokenA.Address == other.TokenA.Address && p.TokenB.Address == other.TokenB.Address
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
