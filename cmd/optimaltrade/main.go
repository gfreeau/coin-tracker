package main

import (
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"gopkg.in/gomail.v2"
	"os"
	"strings"
)

type Config struct {
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

		SellSymbol := marketSymbols[0]
		buySymbol := marketSymbols[1]

		currentBuy := trade.SellUnits / data.Result.Bid

		if currentBuy >= trade.BuyUnits {
			alert = true
			output += fmt.Sprintf("%s: %.2f %s = %.2f %s\n", trade.Market, trade.SellUnits, SellSymbol, currentBuy, buySymbol)
		}
	}

	if alert {
		fmt.Print(output)

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
