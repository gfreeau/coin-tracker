package cointracker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func ParseJsonFile(filename string, v interface{}) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, &v)
	if err != nil {
		return err
	}

	return nil
}

func PercentDiff(from float64, to float64) float64 {
	return ((to - from) / from) * 100
}

func LogFatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
