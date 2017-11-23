package main

import (
	"encoding/json"
	"fmt"
	"github.com/gfreeau/coin-tracker"
	"io/ioutil"
	"net/http"
	"sync"
)

// testing out goroutines

func getCoinData(coinId string) (cointracker.Coin, error) {
	var coin cointracker.Coin
	var coins cointracker.CoinList

	res, err := http.Get(fmt.Sprintf("https://api.coinmarketcap.com/v1/ticker/%s/?convert=%s", coinId, "CAD"))
	if err != nil {
		return coin, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return coin, err
	}

	err = json.Unmarshal(body, &coins)
	if err != nil {
		return coin, err
	}

	return coins[0], nil
}

func main() {
	coins := []string{
		"ripple",
		"basic-attention-token",
		"ethereum",
		"omisego",
	}

	var wg sync.WaitGroup

	for _, coinId := range coins {
		wg.Add(1)
		go func(coinId string) {
			defer wg.Done()
			coin, _ := getCoinData(coinId)
			fmt.Println(coinId, coin)
		}(coinId)
	}

	wg.Wait()
}
