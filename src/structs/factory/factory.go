package factory

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
)

type Factory[FeeType any] struct {
	Name     string
	Address  common.Address
	InitHash common.Hash
	FeeTypes []FeeType
}

func (f Factory[any]) IsValid() bool {
	return f.Name != "" &&
		!bytes.EqualFold(f.Address.Bytes(), common.Address{}.Bytes()) &&
		!bytes.EqualFold(f.InitHash.Bytes(), common.Hash{}.Bytes())
}
