package main

import (
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"gopkg.in/gomail.v2"
	"math"
	"os"
)

type Config struct {
	Coins     []string
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

		for _, v := range conf.Coins {
			set[v] = true
		}

		coins = cointracker.FilterCoins(coins, func(c cointracker.Coin) bool {
			_, ok := set[c.Symbol]
			return ok
		})
	}

	alert := false
	output := ""

	for _, coin := range coins {
		if math.Abs(coin.PercentChange24h) >= conf.AlertPercent {
			output += fmt.Sprintf("%s (%.2f%%) is now CAD %.4f, USD %.4f\n", coin.Name, coin.PercentChange24h, coin.PriceCAD, coin.PriceUSD)
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
