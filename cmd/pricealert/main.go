package main

import (
	"encoding/json"
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"gopkg.in/gomail.v2"
	"io/ioutil"
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
	IncreasePercent  float64
	PriceHistoryFile string
}

type PriceHistory map[string]float64

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

	var priceHistory PriceHistory
	err = cointracker.ParseJsonFile(conf.PriceHistoryFile, &priceHistory)

	if err != nil {
		priceHistory = make(PriceHistory, 0)
	}

	coins, err := cointracker.GetCoinData()
	if err != nil {
		cointracker.LogFatal(err.Error())
	}

	if len(coins) == 0 {
		cointracker.LogFatal("Coin data is unavailable")
	}

	{
		set := make(map[string]bool)

		for _, v := range conf.Coins {
			set[v] = true

			// default initial price for coin if it does not exist
			if _, ok := priceHistory[v]; !ok {
				priceHistory[v] = 0
			}
		}

		coins = cointracker.FilterCoins(coins, func(c cointracker.Coin) bool {
			_, ok := set[c.Symbol]
			return ok
		})
	}

	alert := false
	output := ""

	for _, coin := range coins {
		alertPrice, ok := priceHistory[coin.Symbol]

		if !ok {
			continue
		}

		if coin.PriceCAD > alertPrice {
			output += fmt.Sprintf("%s is now CAD %6.2f\n", coin.Name, coin.PriceCAD)
			alert = true
			priceHistory[coin.Symbol] = coin.PriceCAD + (coin.PriceCAD * conf.IncreasePercent)
		}
	}

	if alert {
		fmt.Print(output)

		jsonData, _ := json.Marshal(priceHistory)
		err := ioutil.WriteFile(conf.PriceHistoryFile, jsonData, 0644)
		if err != nil {
			fmt.Println(os.Stderr, err.Error())
		}

		if conf.SendEmail {
			m := gomail.NewMessage()
			m.SetHeader("To", conf.Email)
			m.SetHeader("From", conf.Email)
			m.SetHeader("Subject", "New Coin Price Alert")
			m.SetBody("text/plain", output)

			d := gomail.NewDialer(conf.Smtp.Host, conf.Smtp.Port, conf.Smtp.Username, conf.Smtp.Password)

			if err := d.DialAndSend(m); err != nil {
				fmt.Println(os.Stderr, err.Error())
			}
		}
	}
}
