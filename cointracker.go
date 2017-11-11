package cointracker

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Coin struct {
	Name     string
	Symbol   string
	PriceUSD float64 `json:"price_usd,string"`
	PriceCAD float64 `json:"price_cad,string"`
}

type CoinList []Coin

func GetCoinData() (CoinList, error) {
	var coins CoinList

	res, err := http.Get("https://api.coinmarketcap.com/v1/ticker/?convert=CAD&limit=100")
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

func ParseJsonFile(filename string, v interface{}) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, &v)
	if err != nil {
		return err
	}

	return nil
}
