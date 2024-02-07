package uniswap

import "errors"

var (
	TokenAlreadyExists = errors.New("token already exists in cache")
	TokenNotFound      = errors.New("token not found")
	InvalidToken       = errors.New("invalid token")
	InvalidFactory     = errors.New("invalid factory")
	PoolNotFound       = errors.New("pool not found")
	BlockAlreadySynced = errors.New("block already synced")
)
