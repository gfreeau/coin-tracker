Some simple golang CLI tools for keeping tracking of cryptocurrencies.

This code predates modules and other ways of organizing dependencies so you'll need to run the following to build this code:

```
go env -w GO111MODULE=off
go get github.com/gfreeau/coin-tracker
go get gopkg.in/gomail.v2 
go get github.com/olekukonko/tablewriter 
```

Then change into the directory set for `$GOPATH`, run `go env` to determine what it is by default. Then navigate to `src/github.com/gfreeau/coin-tracker`.

You will need the `ExchangeId` for each coin for the config files, you can get it from this API: https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=100&page=1&sparkline=false&price_change_percentage=true. The field is called `id`

Portfolio Balance
-----------------

```
go build -o ./bin/portfolio cmd/portfolio/main.go
./bin/portfolio ./config/portfolio.json
```

Create a json config file based on `portfolio.json.dist`.
In the config file list the coin symbols and the amount of coins in a json object.

This command will show the total balance in USD and CAD and the value of each coin.


Price Alert
-----------

```
go build -o ./bin/pricealert cmd/pricealert/main.go
./bin/pricealert ./config/pricealert.json
```

Create a json config file based on `pricealert.json.dist`.

This command will monitor the coin symbols and e-mail you when the price increases by a percentage compared to the previous value.

It should be setup as a cron job i.e hourly.

Price Percent Alert
-----------

```
go build -o ./bin/pricepercentalert cmd/pricepercentalert/main.go
./bin/pricepercentalert ./config/pricepercentalert.json
```

Create a json config file based on `pricepercentalert.json.dist`.

This command will monitor the coin symbols and e-mail you when the price changes by a percentage in the last day.

It should be setup as a cron job i.e daily.

Optimal Trade Alert
-----------

```
go build -o ./bin/optimaltrade cmd/optimaltrade/main.go
./bin/optimaltrade ./config/optimaltrade.json
```

Create a json config file based on `optimaltrade.json.dist`.

This command will monitor the tickers and e-mail you when the trade conditions are met.

It should be setup as a cron job i.e hourly.

Optimal Purchase Alert
-----------

```
go build -o ./bin/optimalpurchase cmd/optimalpurchase/main.go
./bin/optimalpurchase ./config/optimalpurchase.json
```

Create a json config file based on `optimalpurchase.json.dist`.

This command will monitor the tickers and e-mail you when the trade conditions are met.

It should be setup as a cron job i.e hourly.

