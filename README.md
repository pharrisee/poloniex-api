<a href="https://godoc.org/github.com/pharrisee/poloniex-api" target="_blank"><img src="https://godoc.org/github.com/pharrisee/poloniex-api?status.svg"></a>

# Go Poloniex API wrapper
This API should be a complete wrapper for the [Poloniex api](https://poloniex.com/support/api/), including the public, private and websocket APIs.

## Install

```
go get -u github.com/pharrisee/poloniex-api
```

## Usage
To use create a copy of config-example.json and fill in your API key and secret.

```json
{
    "key":"put your key here",
    "secret":"put your secret here"
}
```

You can also pass your key/secret pair in code rather than creating a config.json.

## Examples

### Public API

```go
package main

import (
    "log"

    "github.com/k0kubun/pp"
    "github.com/pharrisee/poloniex-api"
)

func main() {
    p := poloniex.NewPublicOnly()
    ob, err := p.OrderBook("BTC_ETH")
    if err != nil {
        log.Fatalln(err)
    }
    pp.Println(ob.Asks[0], ob.Bids[0])
}
```

### Private API

```go
package main

import (
    "fmt"
    "log"

    "github.com/pharrisee/poloniex-api"
)

func main() {
    p := poloniex.New("config.json")
    balances, err := p.Balances()
    if err != nil {
        log.Fatalln(err)
    }
    fmt.Printf("%+v\n", balances)
}
```

### Websocket API

```go
package main

import (
    "log"

    poloniex "github.com/pharrisee/poloniex-api"

    "github.com/k0kubun/pp"
)

func main() {
	p := poloniex.NewWithCredentials("Key goes here", "secret goes here")
	p.Subscribe("ticker")
	p.Subscribe("USDT_BTC")
	p.StartWS()

	p.On("ticker", func(m poloniex.WSTicker) {
		pp.Println(m)
	}).On("USDT_BTC-trade", func(m poloniex.WSOrderbook) {
		pp.Println(m)
	})

	for _ = range time.Tick(1 * time.Second) {

	}
}

```
### Websocket Events
When subscribing to an event stream there are a few input types, and strangely more output types.

### Ticker 
event name: _ticker_

Sends ticker updates when any of currencyPair, last, lowestAsk, highestBid, percentChange, baseVolume, quoteVolume, isFrozen, 24hrHigh or 24hrLow changes for any market.

You are required to filter which markets you are interested in.

### Market_Name

Subscribing to an orderbook change stream can be confusing (both to think about and describe), since a single subscription can lead to multiple event streams being created.

**_using USDT_BTC as an example market below, any valid market name could be used (e.g. BTC_NXT or ETH_ETC)_**

Subscribing to USDT_BTC will lead to these events being emitted.

| Event           | Purpose                                                  |
| :-------------- | -------------------------------------------------------- |
| USDT_BTC        | all events, trade, modify and remove for a single market |
| trade           | trade events for all markets                             |
| modify          | modify events for all markets                            |
| remove          | remove events for all markets                            |
| USDT_BTC-trade  | trade events for single market                           |
| USDT_BTC-modify | modify events for single market                          |
| USDT_BTC-remove | remove event for single market                           |

This gives flexibility when writing the event handlers, meaning that you could for example have one routing which sends all trades for all markets to a local database for later processing.

see https://poloniex.com/support/api/ for a fuller description of the event types.

