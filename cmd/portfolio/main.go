package main

import (
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"os"
	"github.com/olekukonko/tablewriter"
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
	var ETHUSDPrice float64 = 0

	for _, coin := range coins {
		numberOfCoins := float64(portfolio[coin.Symbol])

		totalCAD += numberOfCoins * coin.PriceCAD
		totalUSD += numberOfCoins * coin.PriceUSD
		totalBTC += numberOfCoins * coin.PriceBTC

		if coin.Symbol == "ETH" {
			ETHUSDPrice = coin.PriceUSD
		}
	}

	if ETHUSDPrice > 0 {
		totalETH = totalUSD / ETHUSDPrice
	}

	tableRows := make([][]string, len(coins))

	for i, coin := range coins {
		numberOfCoins := float64(portfolio[coin.Symbol])

		priceCAD := numberOfCoins * coin.PriceCAD
		priceUSD := numberOfCoins * coin.PriceUSD

		var percentage float64 = 0

		if totalUSD > 0 {
			percentage = priceUSD / totalUSD * 100
		}

		var priceETH float64 = 0
		var coinPriceETH float64 = 0

		if ETHUSDPrice > 0 {
			priceETH = priceCAD / ETHUSDPrice
			coinPriceETH = coin.PriceUSD / ETHUSDPrice
		}

		tableRows[i] = []string{
			coin.Symbol,
			fmt.Sprintf("%.2f%%", percentage),
			fmt.Sprintf("$%.4f", priceCAD),
			fmt.Sprintf("$%.4f", coin.PriceCAD),
			fmt.Sprintf("$%.4f", priceUSD),
			fmt.Sprintf("$%.4f", coin.PriceUSD),
			fmt.Sprintf("%.4f", priceETH),
			fmt.Sprintf("%.8f", coinPriceETH),
		}
	}

	summaryTable := tablewriter.NewWriter(os.Stdout)
	summaryTable.SetHeader([]string{"Return %", "Return", "CAD", "USD", "BTC", "ETH"})
	summaryTable.Append([]string{
		fmt.Sprintf("%.2f%%", cointracker.PercentDiff(conf.InvestmentAmount, totalCAD)),
		fmt.Sprintf("$%.2f", totalCAD - conf.InvestmentAmount),
		fmt.Sprintf("$%.2f", totalCAD),
		fmt.Sprintf("$%.2f", totalUSD),
		fmt.Sprintf("%.4f", totalBTC),
		fmt.Sprintf("%.4f", totalETH),
	})
	summaryTable.Render()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Alloc", "CAD", "Price (CAD)", "USD", "Price (USD)", "ETH", "Price (ETH)"})

	table.AppendBulk(tableRows)
	table.Render()
}
