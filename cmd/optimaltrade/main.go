package main

import (
	"bytes"
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"github.com/gfreeau/coin-tracker/coingecko"
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
	SellName  string
	SellId    string
	SellUnits float64
	BuyName   string
	BuyId     string
	BuyUnits  float64
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

	exchangeIds := make([]string, 0)

	{
		set := make(map[string]bool, 0)

		for _, p := range conf.Trades {
			set[p.SellId] = true
			set[p.BuyId] = true
		}

		for k := range set {
			exchangeIds = append(exchangeIds, k)
		}
	}

	coinMap, err := coingecko.GetCoinMap(exchangeIds)
	if err != nil {
		cointracker.LogFatal(err.Error())
	}

	if len(coinMap) == 0 {
		cointracker.LogFatal("Coin data is unavailable")
	}

	alert := false
	tableRows := make([][]string, 0)

	for _, trade := range conf.Trades {
		sellCoin, exists := coinMap[trade.SellId]

		if !exists {
			continue
		}

		buyCoin, exists := coinMap[trade.BuyId]

		if !exists {
			continue
		}

		price := sellCoin.PriceBTC / buyCoin.PriceBTC

		if price <= 0 {
			continue
		}

		if trade.SellUnits <= 0 {
			continue
		}

		var currentBuy, targetSellPrice, targetBuyPrice, diff, targetSellPriceUSD, targetBuyPriceUSD, currentBuyPriceUSD float64

		currentBuy = trade.SellUnits * price
		targetSellPrice = trade.SellUnits / trade.BuyUnits
		targetBuyPrice = trade.BuyUnits / trade.SellUnits
		targetSellPriceUSD = buyCoin.PriceUSD / targetSellPrice
		targetBuyPriceUSD = sellCoin.PriceUSD / targetBuyPrice
		currentBuyPriceUSD = sellCoin.PriceUSD / price

		diff = cointracker.PercentDiff(targetBuyPrice, price)

		if !conf.AlertMode || currentBuy >= trade.BuyUnits {
			tableRows = append(tableRows, []string{
				fmt.Sprintf("%.2f %s", trade.SellUnits, trade.SellName),
				fmt.Sprintf("%.2f %s", trade.BuyUnits, trade.BuyName),
				fmt.Sprintf("%.8f %s", currentBuy, trade.BuyName),
				fmt.Sprintf("%.2f%%", diff),
				fmt.Sprintf("%s: %.4f USD", trade.SellName, sellCoin.PriceUSD),
				fmt.Sprintf("%s: %.4f USD", trade.SellName, targetSellPriceUSD),
				fmt.Sprintf("%s: %.4f USD", trade.BuyName, currentBuyPriceUSD),
				fmt.Sprintf("%s: %.4f USD", trade.BuyName, targetBuyPriceUSD),
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
