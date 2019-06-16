package main

import (
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"github.com/gfreeau/coin-tracker/coingecko"
	"gopkg.in/gomail.v2"
	"math"
	"os"
)

type Config struct {
	Coins     []Coin
	SendEmail bool
	Email     string
	Smtp      struct {
		Host     string
		Port     int
		Username string
		Password string
	}
	AlertPercent float64
}

type Coin struct {
	Name       string
	ExchangeId string
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

	exchangeIds := make([]string, len(conf.Coins))

	for i, c := range conf.Coins {
		exchangeIds[i] = c.ExchangeId
	}

	coinMap, err := coingecko.GetCoinMap(exchangeIds)
	if err != nil {
		cointracker.LogFatal(err.Error())
	}

	if len(coinMap) == 0 {
		cointracker.LogFatal("Coin data is unavailable")
	}

	alert := false
	output := ""

	for _, coin := range conf.Coins {
		coinData, ok := coinMap[coin.ExchangeId]

		if !ok {
			continue
		}

		if math.Abs(coinData.PercentChange24hCAD) >= conf.AlertPercent {
			output += fmt.Sprintf("%s (%.2f%%) is now CAD %.4f, USD %.4f\n", coin.Name, coinData.PercentChange24hCAD, coinData.PriceCAD, coinData.PriceUSD)
			alert = true
		}
	}

	if alert {
		fmt.Print(output)

		if conf.SendEmail {
			m := gomail.NewMessage()
			m.SetHeader("To", conf.Email)
			m.SetHeader("From", conf.Email)
			m.SetHeader("Subject", "Coin Percent Change Alert")
			m.SetBody("text/plain", output)

			d := gomail.NewDialer(conf.Smtp.Host, conf.Smtp.Port, conf.Smtp.Username, conf.Smtp.Password)

			if err := d.DialAndSend(m); err != nil {
				fmt.Println(os.Stderr, err.Error())
			}
		}
	}
}
