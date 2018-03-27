package cointracker

import (
	"testing"
)

func getTestCoinList() CoinList {
	return CoinList{
		Coin{Symbol: "ETH"},
		Coin{Symbol: "XRP"},
		Coin{Symbol: "BTC"},
		Coin{Symbol: "BCH"},
		Coin{Symbol: "LTC"},
	}
}

func TestFindCoin(t *testing.T) {
	coin := FindCoin("ETH", getTestCoinList())

	if coin == nil {
		t.Fatal("coin was not found")
	}

	if coin.Symbol != "ETH" {
		t.Fatalf("coin symbol is not correct, got: %s", coin.Symbol)
	}
}

func TestGetCoinMap(t *testing.T) {
	coinMap := GetCoinMap(getTestCoinList())

	coin, ok := coinMap["BTC"]

	if !ok {
		t.Fatal("coin was not found")
	}

	if coin.Symbol != "BTC" {
		t.Fatalf("coin symbol is not correct, got: %s", coin.Symbol)
	}
}
