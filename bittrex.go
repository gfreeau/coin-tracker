package cointracker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type TickerResponse struct {
	Success bool
	Message string
	Result  struct {
		Bid  float64
		Ask  float64
		Last float64
	}
}

func GetTickerData(market string) (TickerResponse, error) {
	var response TickerResponse

	res, err := http.Get(fmt.Sprintf("https://bittrex.com/api/v1.1/public/getticker?market=%s", market))
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}
