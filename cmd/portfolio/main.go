package main

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"os"
	"github.com/gfreeau/coin-tracker"
	"log"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s configfilepath\n", os.Args[0])
		os.Exit(1)
	}

	rawConfig, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err.Error())
	}

	var portfolio map[string]float64
	err = json.Unmarshal(rawConfig, &portfolio)
	if err != nil {
		log.Fatal(err.Error())
	}

	coins, err := cointracker.GetCoinData()
	if err != nil {
		log.Fatal(err.Error())
	}

	if len(coins) == 0 {
		log.Fatal("Coin data is unavailable")
	}

	coins = cointracker.FilterCoins(coins, func(symbol string) bool {
		_, ok := portfolio[symbol]
		return ok
	})

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

	if output == "" {
		output = "None\n"
	}

	fmt.Print("Totals:\n")
	fmt.Printf("CAD: %6.2f\n", totalCAD)
	fmt.Printf("USD: %6.2f\n", totalUSD)
	fmt.Print("\nCoins:\n")
	fmt.Print(output)
}