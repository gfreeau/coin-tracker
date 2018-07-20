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
	Symbol   string
	Amount   float64
	Currency string
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

		units := purchase.Amount / coin.PriceUSD

		if purchase.Currency == "CAD" {
			units = purchase.Amount / coin.PriceCAD
		}

		tableRows[i] = []string{
			purchase.Currency,
			fmt.Sprintf("$%.2f", purchase.Amount),
			coin.Symbol,
			fmt.Sprintf("%.4f", units),
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Currency", "Amount", "Symbol", "Units"})

	table.AppendBulk(tableRows)
	table.Render()

}
