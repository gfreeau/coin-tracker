package main

import (
	"fmt"
	"os"

	cointracker "github.com/gfreeau/coin-tracker"

	"github.com/gfreeau/coin-tracker/coingecko"
	"github.com/olekukonko/tablewriter"
)

type Config struct {
	Purchases []Purchase
}

type Purchase struct {
	Name           string
	ExchangeId     string
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
		cointracker.LogFatal("config file error: " + err.Error())
	}

	exchangeIds := make([]string, len(conf.Purchases))

	for _, p := range conf.Purchases {
		exchangeIds = append(exchangeIds, p.ExchangeId)
	}

	coinMap, err := coingecko.GetCoinMap(exchangeIds)
	if err != nil {
		cointracker.LogFatal(err.Error())
	}

	if len(coinMap) == 0 {
		cointracker.LogFatal("Coin data is unavailable")
	}

	tableRows := make([][]string, len(conf.Purchases))

	for i, purchase := range conf.Purchases {
		coinData, ok := coinMap[purchase.ExchangeId]

		if !ok {
			continue
		}

		units := purchase.UnitAmount
		currencyAmount := purchase.CurrencyAmount
		currencySymbol := "$"
		coinPrice := coinData.PriceUSD

		switch purchase.Currency {
		case "AUD":
			coinPrice = coinData.PriceAUD
		case "CAD":
			coinPrice = coinData.PriceCAD
		case "EUR":
			currencySymbol = "€"
			coinPrice = coinData.PriceEUR
		}

		if units > 0 {
			currencyAmount = purchase.UnitAmount * coinPrice
		} else {
			units = purchase.CurrencyAmount / coinPrice
		}

		tableRows[i] = []string{
			purchase.Currency,
			fmt.Sprintf("%s%.2f", currencySymbol, currencyAmount),
			purchase.Name,
			fmt.Sprintf("%.4f", units),
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Currency", "Currency Amount", "Symbol", "Units"})

	table.AppendBulk(tableRows)
	table.Render()

}
