package token

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ERC20 struct {
	Address  common.Address
	Decimals *big.Int
	Name     string
	Symbol   string
}

func (t ERC20) IsValid() bool {
	return bytes.EqualFold(t.Address.Bytes(), common.Address{}.Bytes()) &&
		t.Decimals != nil &&
		t.Decimals.Int64() > 0 &&
		t.Name != "" &&
		t.Symbol != ""
}
