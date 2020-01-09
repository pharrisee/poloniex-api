package poloniex

import (
	"github.com/k0kubun/pp"
)

func ExampleWS() {
	p := NewWithCredentials("Key goes here", "secret goes here")
	p.Subscribe("ticker")
	p.Subscribe("USDT_BTC")

	p.On("ticker", func(m WSTicker) {
		pp.Println(m)
	}).On("USDT_BTC-trade", func(m WSOrderbook) {
		pp.Println(m)
	})

	select {}
}
