package monitor_test

import (
	"PoolHelper/src/monitor"
	uniswapv3 "PoolHelper/src/pool/uniswap-v3"
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

func TestMonitor_FindPoolsV2_Count(t *testing.T) {
	initHash := "0x96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f"
	factory := common.HexToAddress("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f")
	tokenA := common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2")
	tokenB := common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7")
	tokenC := common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f")
	tokenD := common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599")
	tokenE := common.HexToAddress("0x514910771af9ca656af840dff83e8264ecf986ca")

	m := monitor.NewMonitor()
	m.AddTokenSafe(tokenA)
	m.AddTokenSafe(tokenB)
	m.AddTokenSafe(tokenC)
	m.AddTokenSafe(tokenD)
	m.AddTokenSafe(tokenE)

	p := m.FindPoolsV2(factory, common.HexToHash(initHash))
	if len(p) != 10 {
		t.Errorf("wrong number of pools")
		t.Fatalf("expected %v, got %v", 10, len(p))
	}
}

func TestMonitor_FindPoolsV3_Count(t *testing.T) {
	initHash := "0x96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f"
	factory := common.HexToAddress("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f")
	tokenA := common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2")
	tokenB := common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7")
	tokenC := common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f")
	tokenD := common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599")
	tokenE := common.HexToAddress("0x514910771af9ca656af840dff83e8264ecf986ca")

	m := monitor.NewMonitor()
	m.AddTokenSafe(tokenA)
	m.AddTokenSafe(tokenB)
	m.AddTokenSafe(tokenC)
	m.AddTokenSafe(tokenD)
	m.AddTokenSafe(tokenE)

	feeTypes := []uniswapv3.FeeType{uniswapv3.MAX, uniswapv3.NORMAL, uniswapv3.LOW}

	p := m.FindPoolsV3(factory, common.HexToHash(initHash), feeTypes)
	if len(p) != 10*len(feeTypes) {
		t.Errorf("wrong number of pools")
		t.Fatalf("expected %v, got %v", 10*len(feeTypes), len(p))
	}
}
