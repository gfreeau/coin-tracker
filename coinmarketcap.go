package cointracker

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"fmt"
)

type Coin struct {
	Name             string
	Symbol           string
	PriceUSD         float64 `json:"price_usd,string"`
	PriceCAD         float64 `json:"price_cad,string"`
	PercentChange24h float64 `json:"percent_change_24h,string"`
}

type CoinList []Coin

func GetCoinData() (CoinList, error) {
	return GetTopCoinsData("CAD", 100)
}

func GetTopCoinsData(currency string, limit int) (CoinList, error) {
	var coins CoinList

	res, err := http.Get(fmt.Sprintf("https://api.coinmarketcap.com/v1/ticker/?convert=%s&limit=%s", currency, limit))
	if err != nil {
		return coins, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return coins, err
	}

	err = json.Unmarshal(body, &coins)
	if err != nil {
		return coins, err
	}

	return coins, nil
}

func FilterCoins(coins CoinList, test func(Coin) bool) CoinList {
	filteredCoins := make(CoinList, 0)

	for _, coin := range coins {
		if test(coin) {
			filteredCoins = append(filteredCoins, coin)
		}
	}

	return filteredCoins
}

func GetCoinMap(coins CoinList) map[string]Coin {
	coinMap := make(map[string]Coin, 0)

	for _, coin := range coins {
		coinMap[coin.Symbol] = coin
	}

	return coinMap
}