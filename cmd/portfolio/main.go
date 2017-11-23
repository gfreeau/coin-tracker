package main

import (
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s configfilepath\n", os.Args[0])
		os.Exit(1)
	}

	var portfolio map[string]float64
	err := cointracker.ParseJsonFile(os.Args[1], &portfolio)
	if err != nil {
		cointracker.LogFatal(err.Error())
	}

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

	var totalCAD float64 = 0
	var totalUSD float64 = 0
	var output string

	for _, coin := range coins {
		numberOfCoins := float64(portfolio[coin.Symbol])

		priceCAD := numberOfCoins * coin.PriceCAD
		priceUSD := numberOfCoins * coin.PriceUSD

		output += fmt.Sprintf("%s: CAD %.4f, USD %.4f\n", coin.Symbol, priceCAD, priceUSD)

		totalUSD += priceUSD
		totalCAD += priceCAD
	}

	if output == "" {
		output = "None\n"
	}

	fmt.Print("Totals:\n")
	fmt.Printf("CAD: %.4f\n", totalCAD)
	fmt.Printf("USD: %.4f\n", totalUSD)
	fmt.Print("\nCoins:\n")
	fmt.Print(output)
}