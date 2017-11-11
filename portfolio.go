package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"os"
)

type Portfolio map[string]float64

type Coin struct {
	Name string
	Symbol string
	PriceUSD float64 `json:"price_usd,string"`
	PriceCAD float64 `json:"price_cad,string"`
}

type CoinList []Coin

func main() {
	rawConfig, err := ioutil.ReadFile("./config/portfolio.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var portfolio Portfolio
	json.Unmarshal(rawConfig, &portfolio)

	coins, err := getCoinData()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	coins = filterCoins(coins, portfolio)

	var totalCAD float64 = 0
	var totalUSD float64 = 0
	var output string

	for _, coin := range coins {
		numberOfCoins := float64(portfolio[coin.Symbol])

		priceCAD := numberOfCoins * coin.PriceCAD
		priceUSD := numberOfCoins * coin.PriceUSD

		output += fmt.Sprintf("%s: CAD %6.2f, USD %6.2f\n", coin.Symbol, priceCAD, priceUSD)

		totalUSD += priceUSD
		totalCAD += priceCAD
	}

	fmt.Print("Totals:\n")
	fmt.Printf("CAD: %6.2f\n", totalCAD)
	fmt.Printf("USD: %6.2f\n", totalUSD)
	fmt.Print("\nCoins:\n")
	fmt.Print(output)
}

func getCoinData() (CoinList, error) {
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

func filterCoins(coins CoinList, portfolio Portfolio) CoinList {
	filteredCoins := make(CoinList, 0)

	for _, coin := range coins {
		if _, ok := portfolio[coin.Symbol]; ok {
			filteredCoins = append(filteredCoins, coin)
		}
	}

	return filteredCoins
}