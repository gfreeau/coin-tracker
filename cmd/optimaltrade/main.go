package main

import (
	"bytes"
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"github.com/gfreeau/coin-tracker/binance"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/gomail.v2"
	"os"
)

type Config struct {
	AlertMode bool
	Trades    []Trade
	SendEmail bool
	Email     string
	Smtp      struct {
		Host     string
		Port     int
		Username string
		Password string
	}
}

type Trade struct {
	IntermediarySymbol string
	MajorSymbol        string
	MinorSymbol        string
	SellSymbol         string
	SellUnits          float64
	BuySymbol          string
	BuyUnits           float64
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

	priceMap, err := binance.GetMarketPricesMap()

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

		for _, t := range conf.Trades {
			set[t.SellSymbol] = true
			set[t.BuySymbol] = true
		}

		coins = cointracker.FilterCoins(coins, func(c cointracker.Coin) bool {
			_, ok := set[c.Symbol]
			return ok
		})
	}

	coinMap := cointracker.GetCoinMap(coins)

	alert := false
	tableRows := make([][]string, 0)

	for _, trade := range conf.Trades {
		var price float64 = 0

		sellCoin, exists := coinMap[trade.SellSymbol]

		if !exists {
			continue
		}

		buyCoin, exists := coinMap[trade.BuySymbol]

		if !exists {
			continue
		}

		if len(trade.IntermediarySymbol) > 0 {
			sellMarket := trade.BuySymbol + trade.IntermediarySymbol
			sellMarketPrice := priceMap[sellMarket]

			buyMarket := trade.SellSymbol + trade.IntermediarySymbol
			buyMarketPrice := priceMap[buyMarket]

			if buyMarketPrice <= 0 {
				continue
			}

			price = sellMarketPrice / buyMarketPrice
		} else {
			market := trade.MinorSymbol + trade.MajorSymbol
			price = priceMap[market]
		}

		if price <= 0 {
			continue
		}

		if trade.SellUnits <= 0 {
			continue
		}

		var currentBuy, targetSellPrice, targetBuyPrice, diff, targetSellPriceUSD, targetBuyPriceUSD, currentBuyPriceUSD float64

		if trade.SellSymbol == trade.MinorSymbol {
			currentBuy = trade.SellUnits * price
			targetSellPrice = trade.SellUnits / trade.BuyUnits
			targetBuyPrice = trade.BuyUnits / trade.SellUnits
			diff = cointracker.PercentDiff(targetBuyPrice, price)
			targetSellPriceUSD = buyCoin.PriceUSD / targetSellPrice
			targetBuyPriceUSD = sellCoin.PriceUSD / targetBuyPrice
			currentBuyPriceUSD = sellCoin.PriceUSD / price
		} else {
			currentBuy = trade.SellUnits / price
			targetSellPrice = trade.BuyUnits / trade.SellUnits
			targetBuyPrice = trade.SellUnits / trade.BuyUnits
			diff = cointracker.PercentDiff(price, targetBuyPrice)
			targetSellPriceUSD = targetSellPrice * buyCoin.PriceUSD
			targetBuyPriceUSD = targetBuyPrice * sellCoin.PriceUSD
			currentBuyPriceUSD = price * sellCoin.PriceUSD
		}

		if !conf.AlertMode || currentBuy >= trade.BuyUnits {
			tableRows = append(tableRows, []string{
				fmt.Sprintf("%.2f %s", trade.SellUnits, trade.SellSymbol),
				fmt.Sprintf("%.2f %s", trade.BuyUnits, trade.BuySymbol),
				fmt.Sprintf("%.8f %s", currentBuy, trade.BuySymbol),
				fmt.Sprintf("%.2f%%", diff),
				fmt.Sprintf("%s: %.4f USD", trade.SellSymbol, sellCoin.PriceUSD),
				fmt.Sprintf("%s: %.4f USD", trade.SellSymbol, targetSellPriceUSD),
				fmt.Sprintf("%s: %.4f USD", trade.BuySymbol, currentBuyPriceUSD),
				fmt.Sprintf("%s: %.4f USD", trade.BuySymbol, targetBuyPriceUSD),
			})

			if conf.AlertMode {
				alert = true
			}
		}
	}

	if len(tableRows) == 0 {
		os.Exit(0)
	}

	buf := new(bytes.Buffer)

	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"Sell", "Target Buy", "Current Buy", "Diff", "Current Sell Price", "Target Sell Price", "Current Buy Price", "Target Buy Price"})

	table.AppendBulk(tableRows)
	table.Render()

	fmt.Print(buf)

	if alert {
		if conf.SendEmail {
			m := gomail.NewMessage()
			m.SetHeader("To", conf.Email)
			m.SetHeader("From", conf.Email)
			m.SetHeader("Subject", "Optimal Trade Alert")
			m.SetBody("text/plain", buf.String())

			d := gomail.NewDialer(conf.Smtp.Host, conf.Smtp.Port, conf.Smtp.Username, conf.Smtp.Password)

			if err := d.DialAndSend(m); err != nil {
				fmt.Println(os.Stderr, err.Error())
			}
		}
	}
}
