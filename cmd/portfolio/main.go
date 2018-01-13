package main

import (
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"os"
)

type Config struct {
	InvestmentAmount float64
	Holdings         map[string]float64
	UseBittrexPrice  []string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s configfilepath\n", os.Args[0])
		os.Exit(1)
	}

	var conf Config
	err := cointracker.ParseJsonFile(os.Args[1], &conf)
	if err != nil {
		cointracker.LogFatal(err.Error())
	}

	portfolio := conf.Holdings

	coins, err := cointracker.GetCoinData()
	if err != nil {
		cointracker.LogFatal(err.Error())
	}

	if len(coins) == 0 {
		cointracker.LogFatal("Coin data is unavailable")
	}

	coins = cointracker.FilterCoins(coins, func(c cointracker.Coin) bool {
		_, ok := portfolio[c.Symbol]
		return ok
	})

	for _, symbol := range conf.UseBittrexPrice {
		// Sometimes price is skewed by foreign exchanges, use bittrex price to get a more accurate value
		marketData, err := cointracker.GetTickerData("BTC-" + symbol)

		if err != nil {
			continue
		}

		for i := range coins {
			if coins[i].Symbol == symbol {
				// use reference to overwrite value
				coin := &coins[i]

				BTCUSDRate := coin.PriceUSD / coin.PriceBTC
				USDCADRate := coin.PriceCAD / coin.PriceUSD

				coin.PriceBTC = marketData.Result.Ask
				coin.PriceUSD = coin.PriceBTC * BTCUSDRate
				coin.PriceCAD = coin.PriceUSD * USDCADRate

				break
			}
		}
	}

	var totalCAD float64 = 0
	var totalUSD float64 = 0
	var totalBTC float64 = 0
	var totalETH float64 = 0
	var ETHCADPrice float64 = 0
	var output string

	for _, coin := range coins {
		numberOfCoins := float64(portfolio[coin.Symbol])

		totalCAD += numberOfCoins * coin.PriceCAD
		totalUSD += numberOfCoins * coin.PriceUSD
		totalBTC += numberOfCoins * coin.PriceBTC

		if coin.Symbol == "ETH" {
			ETHCADPrice = coin.PriceCAD
		}
	}

	if ETHCADPrice > 0 {
		totalETH = totalCAD / ETHCADPrice
	}

	for _, coin := range coins {
		numberOfCoins := float64(portfolio[coin.Symbol])

		priceCAD := numberOfCoins * coin.PriceCAD
		priceUSD := numberOfCoins * coin.PriceUSD

		var percentage float64 = 0

		if totalUSD > 0 {
			percentage = priceUSD / totalUSD * 100
		}

		output += fmt.Sprintf("%s: %.2f%% CAD %.4f (%.4f), USD %.4f (%.4f)\n", coin.Symbol, percentage, priceCAD, coin.PriceCAD, priceUSD, coin.PriceUSD)
	}

	if output == "" {
		output = "None\n"
	}

	fmt.Print("Totals:\n")
	fmt.Printf("CAD: %.4f (%.2f%%)\n", totalCAD, cointracker.PercentDiff(conf.InvestmentAmount, totalCAD))
	fmt.Printf("USD: %.4f\n", totalUSD)
	fmt.Printf("BTC: %.4f\n", totalBTC)
	if ETHCADPrice > 0 {
		fmt.Printf("ETH: %.4f\n", totalETH)
	}
	fmt.Print("\nCoins:\n")
	fmt.Print(output)
}
