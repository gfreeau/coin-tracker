package binance

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type MarketMap map[string]float64

type Market struct {
	Symbol string
	Price  float64 `json:"price,string"`
}

type Markets []Market

func getHttpRequest(url string) ([]byte, error) {
	res, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetMarketPrices() (Markets, error) {
	var data Markets

	body, err := getHttpRequest("https://api.binance.com/api/v3/ticker/price")

	if err != nil {
		return data, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func GetMarketPricesMap() (MarketMap, error) {
	prices, err := GetMarketPrices()

	if err != nil {
		return nil, err
	}

	priceMap := make(MarketMap)

	for _, p := range prices {
		priceMap[p.Symbol] = p.Price
	}

	return priceMap, nil
}
