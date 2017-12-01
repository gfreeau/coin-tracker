package main

import (
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"gopkg.in/gomail.v2"
	"os"
	"strings"
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
	Market    string
	SellUnits float64
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
		cointracker.LogFatal(err.Error())
	}

	alert := false
	output := ""

	for _, trade := range conf.Trades {
		marketSymbols := strings.Split(trade.Market, "-")

		if len(marketSymbols) != 2 {
			continue
		}

		data, err := cointracker.GetTickerData(trade.Market)

		if err != nil {
			continue
		}

		if data.Success == false {
			continue
		}

		if data.Result.Ask <= 0 {
			continue
		}

		sellSymbol := marketSymbols[0]
		buySymbol := marketSymbols[1]

		targetAskPrice := trade.SellUnits / trade.BuyUnits
		targetDiff := cointracker.PercentDiff(data.Result.Ask, targetAskPrice)
		currentBuy := trade.SellUnits / data.Result.Ask

		if !conf.AlertMode || currentBuy >= trade.BuyUnits {
			output += fmt.Sprintf("%s: %.2f %s = %.2f %s (%.2f%%)\n", trade.Market, trade.SellUnits, sellSymbol, currentBuy, buySymbol, targetDiff)
			output += fmt.Sprintf("Target: %.4f %s (%.8f)\n", trade.BuyUnits, buySymbol, targetAskPrice)
			output += fmt.Sprintf("Current: %.4f %s (%.8f)\n\n", data.Result.Ask * trade.BuyUnits, sellSymbol, data.Result.Ask)

			if conf.AlertMode {
				alert = true
			}
		}
	}

	fmt.Print(output)

	if alert {
		if conf.SendEmail {
			m := gomail.NewMessage()
			m.SetHeader("To", conf.Email)
			m.SetHeader("From", conf.Email)
			m.SetHeader("Subject", "Optimal Trade Alert")
			m.SetBody("text/plain", output)

			d := gomail.NewDialer(conf.Smtp.Host, conf.Smtp.Port, conf.Smtp.Username, conf.Smtp.Password)

			if err := d.DialAndSend(m); err != nil {
				fmt.Println(os.Stderr, err.Error())
			}
		}
	}
}
