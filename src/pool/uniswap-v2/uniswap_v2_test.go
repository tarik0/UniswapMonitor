package uniswap_v2_test

import (
	uniswap_v2 "PoolHelper/src/pool/uniswap-v2"
	"PoolHelper/src/token"
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

const initHash = "0x96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f"

// TestAddressCalculation tests the address calculation for a UniswapV2 pool.
func TestAddressCalculation(t *testing.T) {
	factory := common.HexToAddress("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f")
	tokenA := common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2")
	tokenB := common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7")

	p := uniswap_v2.NewUniswapV2Pool(factory, common.HexToHash(initHash), &token.Pair{
		TokenA: token.ERC20{
			Address: tokenA,
		},
		TokenB: token.ERC20{
			Address: tokenB,
		},
	})

	expected := common.HexToAddress("0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852")
	if p.Address() != expected {
		t.Errorf("wrong address")
		t.Fatalf("expected %v, got %v", expected, p.Address())
	}
}
