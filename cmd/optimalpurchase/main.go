package main

import (
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"github.com/gfreeau/coin-tracker/coingecko"
	"gopkg.in/gomail.v2"
	"os"
)

type Config struct {
	AlertMode bool
	Purchases []Purchase
	SendEmail bool
	Email     string
	Smtp      struct {
		Host     string
		Port     int
		Username string
		Password string
	}
}

type Purchase struct {
	Name       string
	ExchangeId string
	BuyUnits   float64
	Price      float64
	Currency   string
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

	exchangeIds := make([]string, len(conf.Purchases))

	for _, p := range conf.Purchases {
		exchangeIds = append(exchangeIds, p.ExchangeId)
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

	for _, purchase := range conf.Purchases {
		coinData, ok := coinMap[purchase.ExchangeId]

		if !ok {
			continue
		}

		currentUnitPrice := coinData.PriceUSD
		currentPurchasePrice := purchase.BuyUnits * coinData.PriceUSD

		if purchase.Currency == "CAD" {
			currentUnitPrice = coinData.PriceCAD
			currentPurchasePrice = purchase.BuyUnits * coinData.PriceCAD
		}

		if !conf.AlertMode || currentPurchasePrice <= purchase.Price {
			targetPrice := purchase.Price / purchase.BuyUnits
			targetDiff := cointracker.PercentDiff(currentPurchasePrice, purchase.Price)

			output += fmt.Sprintf("%s: %.2f = %.2f %s\n", purchase.Name, purchase.BuyUnits, currentPurchasePrice, purchase.Currency)
			output += fmt.Sprintf("Target: %.2f %s (%.2f%%)\n", purchase.Price, purchase.Currency, targetDiff)
			output += fmt.Sprintf("Current Unit Price: %.4f %s\n", currentUnitPrice, purchase.Currency)
			output += fmt.Sprintf("Target Unit Price: %.4f %s\n\n", targetPrice, purchase.Currency)

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
			m.SetHeader("Subject", "Optimal Purchase Alert")
			m.SetBody("text/plain", output)

			d := gomail.NewDialer(conf.Smtp.Host, conf.Smtp.Port, conf.Smtp.Username, conf.Smtp.Password)

			if err := d.DialAndSend(m); err != nil {
				fmt.Println(os.Stderr, err.Error())
			}
		}
	}
}
