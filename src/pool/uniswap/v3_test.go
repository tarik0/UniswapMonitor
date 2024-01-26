package uniswap_test

import (
	"PoolHelper/src/pool/uniswap"
	"PoolHelper/src/structs/pair"
	"PoolHelper/src/structs/token"
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

const initHashV3 = "0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54"

func TestVV3AddressCalculation(t *testing.T) {
	factory := common.HexToAddress("0x1f98431c8ad98523631ae4a59f267346ea31f984")
	token0 := common.HexToAddress("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599")
	token1 := common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	p := uniswap.NewV3Pool(factory, common.HexToHash(initHashV3), pair.Pair{
		TokenA: token.ERC20{
			Address: token0,
		},
		TokenB: token.ERC20{
			Address: token1,
		},
	}, 3000)

	expected := common.HexToAddress("0xcbcdf9626bc03e24f779434178a73a0b4bad62ed")
	if p.Address() != expected {
		t.Errorf("wrong address")
		t.Fatalf("expected %v, got %v", expected, p.Address())
	}
}
