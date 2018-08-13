package main

import (
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"os"

	"github.com/olekukonko/tablewriter"
)

type Config struct {
	Purchases []Purchase
}

type Purchase struct {
	Symbol         string
	CurrencyAmount float64
	UnitAmount     float64
	Currency       string
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

	coins, err := cointracker.GetCoinData()
	if err != nil {
		cointracker.LogFatal(err.Error())
	}

	if len(coins) == 0 {
		cointracker.LogFatal("Coin data is unavailable")
	}

	{
		set := make(map[string]bool, 0)

		for _, p := range conf.Purchases {
			set[p.Symbol] = true
		}

		coins = cointracker.FilterCoins(coins, func(c cointracker.Coin) bool {
			_, ok := set[c.Symbol]
			return ok
		})
	}

	coinMap := cointracker.GetCoinMap(coins)

	tableRows := make([][]string, len(conf.Purchases))

	for i, purchase := range conf.Purchases {
		coin, ok := coinMap[purchase.Symbol]

		if !ok {
			continue
		}

		units := purchase.UnitAmount
		currencyAmount := purchase.CurrencyAmount
		coinPrice := coin.PriceUSD
		if purchase.Currency == "CAD" {
			coinPrice = coin.PriceCAD
		}

		if units > 0 {
			currencyAmount = purchase.UnitAmount * coinPrice
		} else {
			units = purchase.CurrencyAmount / coinPrice
		}

		tableRows[i] = []string{
			purchase.Currency,
			fmt.Sprintf("$%.2f", currencyAmount),
			coin.Symbol,
			fmt.Sprintf("%.4f", units),
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Currency", "CurrencyAmount", "Symbol", "Units"})

	table.AppendBulk(tableRows)
	table.Render()

}
