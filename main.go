package main

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Token struct {
	Address  common.Address
	Decimals uint8
	Name     string
	Symbol   string
}

type Type int

type Pool interface {
	Address() common.Address
	Tokens() (Token, Token)
	Reserves() (*big.Int, *big.Int)
	Type() Type
}

func main() {

}
