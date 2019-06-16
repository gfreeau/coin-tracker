package coingecko

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Coin struct {
	PriceUSD            float64 `json:"usd"`
	PercentChange24hUSD float64 `json:"usd_24h_change"`
	PriceCAD            float64 `json:"cad"`
	PercentChange24hCAD float64 `json:"cad_24h_change"`
	PriceEUR            float64 `json:"eur"`
	PercentChange24hEUR float64 `json:"eur_24h_change"`
	PriceBTC            float64 `json:"btc"`
	PercentChange24hBTC float64 `json:"btc_24h_change"`
	PriceETH            float64 `json:"eth"`
	PercentChange24hETH float64 `json:"eth_24h_change"`
}

type CoinMap map[string]Coin

var currencies = []string{"usd", "cad", "eur", "btc", "eth"}

func GetCoinMap(exchangeIds []string) (CoinMap, error) {
	var coins CoinMap

	exchangeIds = uniqueStrings(exchangeIds)

	currenciesParam := strings.Join(currencies, ",")
	exchangeIdsParam := strings.Join(exchangeIds, ",")

	apiUrl := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s&include_24hr_change=true", url.QueryEscape(exchangeIdsParam), url.QueryEscape(currenciesParam))

	res, err := http.Get(apiUrl)
	if err != nil {
		return coins, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return coins, err
	}

	err = json.Unmarshal(body, &coins)
	if err != nil {
		return coins, err
	}

	return coins, nil
}

func uniqueStrings(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}
