package main

import (
	"encoding/json"
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"os"
	"path/filepath"
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

	priceHistoryFile := filepath.Dir(os.Args[1]) + "/coin-prices.json"

	var priceHistory PriceHistory
	err = cointracker.ParseJsonFile(priceHistoryFile, &priceHistory)

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

	for _, v := range conf.Coins {
		if _, ok := priceHistory[v]; !ok {
			priceHistory[v] = 0
		}
	}

	coins = cointracker.FilterCoins(coins, func(c cointracker.Coin) bool {
		_, ok := priceHistory[c.Symbol]
		return ok
	})

	alert := false
	output := ""

	for _, coin := range coins {
		alertPrice, ok := priceHistory[coin.Symbol]

		if !ok {
			continue
		}

		if coin.PriceCAD > alertPrice {
			output += fmt.Sprintf("%s is now CAD %.4f, USD %.4f\n", coin.Name, coin.PriceCAD, coin.PriceUSD)
			alert = true
			priceHistory[coin.Symbol] = coin.PriceCAD + (coin.PriceCAD * (conf.IncreasePercent / 100))
		}
	}

	if alert {
		fmt.Print(output)

		jsonData, _ := json.Marshal(priceHistory)
		err := ioutil.WriteFile(priceHistoryFile, jsonData, 0644)
		if err != nil {
			fmt.Println(os.Stderr, err.Error())
		}

		if conf.SendEmail {
			m := gomail.NewMessage()
			m.SetHeader("To", conf.Email)
			m.SetHeader("From", conf.Email)
			m.SetHeader("Subject", "Coin Price Increase Alert")
			m.SetBody("text/plain", output)

			d := gomail.NewDialer(conf.Smtp.Host, conf.Smtp.Port, conf.Smtp.Username, conf.Smtp.Password)

			if err := d.DialAndSend(m); err != nil {
				fmt.Println(os.Stderr, err.Error())
			}
		}
	}
}
