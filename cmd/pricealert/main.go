package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	cointracker "github.com/gfreeau/coin-tracker"
	"github.com/gfreeau/coin-tracker/coingecko"
	"gopkg.in/gomail.v2"
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
	IncreasePercent float64
}

type Coin struct {
	Name       string
	ExchangeId string
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
		cointracker.LogFatal("config file error: " + err.Error())
	}

	priceHistoryFile := filepath.Dir(os.Args[1]) + "/coin-prices.json"

	var priceHistory PriceHistory
	err = cointracker.ParseJsonFile(priceHistoryFile, &priceHistory)

	if err != nil {
		priceHistory = make(PriceHistory, 0)
	}

	exchangeIds := make([]string, len(conf.Coins))

	for _, c := range conf.Coins {
		exchangeIds = append(exchangeIds, c.ExchangeId)
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

		if _, ok := priceHistory[coin.ExchangeId]; !ok {
			priceHistory[coin.ExchangeId] = 0
		}

		alertPrice := priceHistory[coin.ExchangeId]

		if coinData.PriceAUD > alertPrice {
			output += fmt.Sprintf("%s is now AUD %.4f, USD %.4f\n", coin.Name, coinData.PriceAUD, coinData.PriceUSD)
			alert = true
			priceHistory[coin.ExchangeId] = coinData.PriceAUD + (coinData.PriceAUD * (conf.IncreasePercent / 100))
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
