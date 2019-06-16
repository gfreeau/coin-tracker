package cointracker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Coin struct {
	Name             string
	Symbol           string
	PriceUSD         float64 `json:"price_usd,string"`
	PriceCAD         float64 `json:"price_cad,string"`
	PriceBTC         float64 `json:"price_btc,string"`
	Rank             int     `json:"rank,string"`
	PercentChange24h float64 `json:"percent_change_24h,string"`
}

type CoinList []Coin

func GetTopCoinsData(currency string, limit int) (CoinList, error) {
	var coins CoinList

	res, err := http.Get(fmt.Sprintf("https://api.coinmarketcap.com/v1/ticker/?convert=%s&limit=%d", currency, limit))
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

func FindCoin(symbol string, coins CoinList) *Coin {
	for _, coin := range coins {
		if symbol == coin.Symbol {
			return &coin
		}
	}

	return nil
}

func GetCoinMap(coins CoinList) map[string]Coin {
	coinMap := make(map[string]Coin, 0)

	for _, coin := range coins {
		coinMap[coin.Symbol] = coin
	}

	return coinMap
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

func PercentDiff(from float64, to float64) float64 {
	if from == 0 {
		return 0
	}

	return ((to - from) / from) * 100
}

func LogFatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
