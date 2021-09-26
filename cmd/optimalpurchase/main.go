package main

import (
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"github.com/gfreeau/coin-tracker/coingecko"
	"gopkg.in/gomail.v2"
	"os"
	"bytes"
	"github.com/olekukonko/tablewriter"
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
	tableRows := make([][]string, 0)

	for _, purchase := range conf.Purchases {
		coinData, ok := coinMap[purchase.ExchangeId]

		if !ok {
			continue
		}

		currentUnitPrice := coinData.PriceUSD
		currentPurchasePrice := purchase.BuyUnits * coinData.PriceUSD

		switch purchase.Currency {
		case "AUD":
			currentUnitPrice = coinData.PriceAUD
			currentPurchasePrice = purchase.BuyUnits * coinData.PriceAUD
		case "CAD":
			currentUnitPrice = coinData.PriceCAD
			currentPurchasePrice = purchase.BuyUnits * coinData.PriceCAD
		case "EUR":
			currentUnitPrice = coinData.PriceEUR
			currentPurchasePrice = purchase.BuyUnits * coinData.PriceEUR
		}

		if !conf.AlertMode || currentPurchasePrice <= purchase.Price {
			targetPrice := purchase.Price / purchase.BuyUnits
			targetDiff := cointracker.PercentDiff(currentPurchasePrice, purchase.Price)

			tableRows = append(tableRows, []string{
				fmt.Sprintf("%.2f %s", purchase.BuyUnits, purchase.Name),
				fmt.Sprintf("%.2f %s", currentPurchasePrice, purchase.Currency),
				fmt.Sprintf("%.2f %s", purchase.Price, purchase.Currency),
				fmt.Sprintf("%.4f %s", targetPrice, purchase.Currency),
				fmt.Sprintf("%.4f %s", currentUnitPrice, purchase.Currency),
				fmt.Sprintf("%.2f%%", targetDiff),
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
	table.SetHeader([]string{"Buy", "Current Buy Price", "Target Buy Price", "Target Unit Price", "Current Unit Price", "Target Diff"})

	table.AppendBulk(tableRows)
	table.Render()

	fmt.Print(buf)

	if alert {
		if conf.SendEmail {
			m := gomail.NewMessage()
			m.SetHeader("To", conf.Email)
			m.SetHeader("From", conf.Email)
			m.SetHeader("Subject", "Optimal Purchase Alert")
			m.SetBody("text/plain", buf.String())

			d := gomail.NewDialer(conf.Smtp.Host, conf.Smtp.Port, conf.Smtp.Username, conf.Smtp.Password)

			if err := d.DialAndSend(m); err != nil {
				fmt.Println(os.Stderr, err.Error())
			}
		}
	}
}
