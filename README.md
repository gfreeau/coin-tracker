A WIP, working on some simple golang CLI tools for keeping tracking of cryptocurrencies.

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