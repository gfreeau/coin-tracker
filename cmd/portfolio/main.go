package main

import (
	"fmt"
	"os"

	cointracker "github.com/gfreeau/coin-tracker"
	"github.com/gfreeau/coin-tracker/coingecko"
	"github.com/olekukonko/tablewriter"
)

type Config struct {
	InvestmentAmount float64
	Holdings         []Holding
}

type Holding struct {
	Name       string
	ExchangeId string
	Units      float64
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s configfilepath\n", os.Args[0])
		os.Exit(1)
	}

	var conf Config
	err := cointracker.ParseJsonFile(os.Args[1], &conf)
	if err != nil {
		cointracker.LogFatal("config file error: " + err.Error())
	}

	holdings := conf.Holdings

	exchangeIds := make([]string, len(holdings))

	for _, holding := range holdings {
		exchangeIds = append(exchangeIds, holding.ExchangeId)
	}

	coinMap, err := coingecko.GetCoinMap(exchangeIds)
	if err != nil {
		cointracker.LogFatal(err.Error())
	}

	if len(coinMap) == 0 {
		cointracker.LogFatal("Coin data is unavailable")
	}

	var totalAUD float64 = 0
	var totalCAD float64 = 0
	var totalUSD float64 = 0
	var totalEUR float64 = 0
	var totalBTC float64 = 0
	var totalETH float64 = 0
	var ChangeAUD24hAgo float64 = 0

	for _, holding := range holdings {
		coin, ok := coinMap[holding.ExchangeId]

		if !ok {
			continue
		}

		totalAUD += holding.Units * coin.PriceAUD
		totalCAD += holding.Units * coin.PriceCAD
		totalUSD += holding.Units * coin.PriceUSD
		totalEUR += holding.Units * coin.PriceEUR
		totalBTC += holding.Units * coin.PriceBTC
		totalETH += holding.Units * coin.PriceETH

		ChangeAUD24hAgo += holding.Units * coin.PriceAUD * (coin.PercentChange24hAUD / 100)
	}

	tableRows := make([][]string, len(holdings))

	for i, holding := range holdings {
		coin, ok := coinMap[holding.ExchangeId]

		if !ok {
			continue
		}

		priceAUD := holding.Units * coin.PriceAUD
		priceCAD := holding.Units * coin.PriceCAD
		priceUSD := holding.Units * coin.PriceUSD
		priceBTC := holding.Units * coin.PriceBTC
		priceETH := holding.Units * coin.PriceETH

		var percentage float64 = 0

		if totalAUD > 0 {
			percentage = priceAUD / totalAUD * 100
		}

		tableRows[i] = []string{
			holding.Name,
			fmt.Sprintf("%.2f%%", percentage),
			fmt.Sprintf("$%.4f", priceAUD),
			fmt.Sprintf("$%.4f", coin.PriceAUD),
			fmt.Sprintf("%.2f%%", coin.PercentChange24hAUD),
			fmt.Sprintf("$%.4f", priceCAD),
			fmt.Sprintf("$%.4f", coin.PriceCAD),
			fmt.Sprintf("$%.4f", priceUSD),
			fmt.Sprintf("$%.4f", coin.PriceUSD),
			fmt.Sprintf("%.4f", priceETH),
			fmt.Sprintf("%.8f", coin.PriceETH),
			fmt.Sprintf("%.4f", priceBTC),
			fmt.Sprintf("%.8f", coin.PriceBTC),
		}
	}

	summaryTable := tablewriter.NewWriter(os.Stdout)
	summaryTable.SetHeader([]string{"Return %", "AUD", "CAD", "USD", "ETH", "BTC", "Return (AUD)", "24H (AUD)", "24H %"})
	summaryTable.Append([]string{
		fmt.Sprintf("%.2f%%", cointracker.PercentDiff(conf.InvestmentAmount, totalAUD)),
		fmt.Sprintf("$%.2f", totalAUD),
		fmt.Sprintf("$%.2f", totalCAD),
		fmt.Sprintf("$%.2f", totalUSD),
		fmt.Sprintf("%.4f", totalETH),
		fmt.Sprintf("%.4f", totalBTC),
		fmt.Sprintf("$%.2f", totalAUD-conf.InvestmentAmount),
		fmt.Sprintf("$%.2f", ChangeAUD24hAgo),
		fmt.Sprintf("%.2f%%", ChangeAUD24hAgo/(totalAUD-ChangeAUD24hAgo)*100),
	})
	summaryTable.Render()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Alloc", "AUD", "Price (AUD)", "24H % (AUD)", "CAD", "Price (CAD)", "USD", "Price (USD)", "ETH", "Price (ETH)", "BTC", "Price (BTC)"})

	table.AppendBulk(tableRows)
	table.Render()
}
