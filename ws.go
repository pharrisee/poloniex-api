package poloniex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
)

const (
	apiURL = "wss://api2.poloniex.com/"
)

type (
	// WSTicker describes a ticker item
	WSTicker struct {
		Pair          string
		Last          float64
		Ask           float64
		Bid           float64
		PercentChange float64
		BaseVolume    float64
		QuoteVolume   float64
		IsFrozen      bool
		DailyHigh     float64
		DailyLow      float64
		PairID        int64
	}

	// WSOrderbook ::::
	WSOrderbook struct {
		Pair    string
		Event   string
		TradeID int64
		Type    string
		Rate    float64
		Amount  float64
		Total   float64
		TS      time.Time
	}

	// WSReportFunc is used whilst idling
	WSReportFunc = func(time.Time)
)

// StartWS opens the websocket connection, and waits for message events
func (p *Poloniex) StartWS() {
	ctx := context.Background()
	for {
		select {
		case <-ctx.Done():
			go p.ws.Close()
			log.Printf("Websocket closed %s", p.ws.GetURL())
			return
		default:
			// if !p.ws.IsConnected() {
			// 	log.Printf("Websocket disconnected %s", p.ws.GetURL())
			// 	continue
			// }
			message := []interface{}{}
			if err := p.ws.ReadJSON(&message); err != nil {
				log.Println(err)
				continue
			}
			chid := int64(message[0].(float64)) // first element is the channel id
			chids := toString(chid)
			// we only handle informational and pair based channels, assuming the informational channels are orderbooks
			if chid > 100.0 && chid < 1000.0 { //
				if err := p.handleOrderBook(message); err != nil {
					continue
				}
			} else if chids == p.ByName["ticker"] {
				if err := p.handleTicker(message); err != nil {
					continue
				}
			}
		}
	}
}

// takes a message and emits relevant events
func (p *Poloniex) handleOrderBook(message []interface{}) error {
	// it's an orderbook
	orderbook, err := p.parseOrderbook(message)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, v := range orderbook {
		p.Emit(v.Event, v).Emit(v.Pair, v).Emit(v.Pair+"-"+v.Event, v)
	}
	return nil
}

// takes a message and emits relevant events
func (p *Poloniex) handleTicker(message []interface{}) error {
	// it's a ticker
	ticker, err := p.parseTicker(message)
	if err != nil {
		log.Printf("%s: (%s)\n", err, message)
		return err
	}
	p.Emit("ticker", ticker)
	return nil
}

// Subscribe to a particular channel specified by channel id:
// 1000	Private	Account Notifications (Beta)
// 1002	Public	Ticker Data
// 1003	Public	24 Hour Exchange Volume
// 1010	Public	Heartbeat
//<currency pair>	Public	Price Aggregated Book
func (p *Poloniex) Subscribe(chid string) error {
	if c, ok := p.ByName[chid]; ok {
		chid = c
	} else if c, ok := p.ByID[chid]; ok {
		chid = c
	} else {
		return errors.New("unrecognised channelid in subscribe")
	}

	p.subscriptions[chid] = true
	message := subscription{Command: "subscribe", Channel: chid}
	return p.sendWSMessage(message)
}

// Unsubscribe from the specified channel:
// 1000	Private	Account Notifications (Beta)
// 1002	Public	Ticker Data
// 1003	Public	24 Hour Exchange Volume
// 1010	Public	Heartbeat
//<currency pair>	Public	Price Aggregated Book
func (p *Poloniex) Unsubscribe(chid string) error {
	if c, ok := p.ByName[chid]; ok {
		chid = c
	} else if c, ok := p.ByID[chid]; ok {
		chid = c
	} else {
		return errors.New("unrecognised channelid in subscribe")
	}
	message := subscription{Command: "subscribe", Channel: chid}
	delete(p.subscriptions, chid)
	return p.sendWSMessage(message)
}

// parse the ticker supplied
func (p *Poloniex) parseTicker(raw []interface{}) (WSTicker, error) {
	wt := WSTicker{}
	var rawInner []interface{}
	if len(raw) <= 2 {
		return wt, errors.New("cannot parse to ticker")
	}
	rawInner = raw[2].([]interface{})
	marketID := int64(toFloat(rawInner[0]))
	pair, ok := p.ByID[fmt.Sprintf("%d", marketID)]
	if !ok {
		return wt, errors.New("cannot parse to ticker - invalid marketID")
	}

	wt.Pair = pair
	wt.PairID = marketID
	wt.Last = toFloat(rawInner[1])
	wt.Ask = toFloat(rawInner[2])
	wt.Bid = toFloat(rawInner[3])
	wt.PercentChange = toFloat(rawInner[4])
	wt.BaseVolume = toFloat(rawInner[5])
	wt.QuoteVolume = toFloat(rawInner[6])
	wt.IsFrozen = toFloat(rawInner[7]) != 0.0
	wt.DailyHigh = toFloat(rawInner[8])
	wt.DailyLow = toFloat(rawInner[9])

	return wt, nil
}

// parse the supplied orderbook
func (p *Poloniex) parseOrderbook(raw []interface{}) ([]WSOrderbook, error) {
	trades := []WSOrderbook{}
	marketID := int64(toFloat(raw[0]))
	pair, ok := p.ByID[fmt.Sprintf("%d", marketID)]
	if !ok {
		return trades, errors.New("cannot parse to orderbook - invalid marketID")
	}
	for _, _v := range raw[2].([]interface{}) {
		v := _v.([]interface{})
		trade := WSOrderbook{}
		trade.Pair = pair
		switch v[0].(string) {
		case "i":
		case "o":
			trade.Event = "modify"
			if t := toFloat(v[3]); t == 0.0 {
				trade.Event = "remove"
			}
			trade.Type = "ask"
			if t := toFloat(v[1]); t == 1.0 {
				trade.Type = "bid"
			}
			trade.Rate = toFloat(v[2])
			trade.Amount = toFloat(v[3])
			trade.TS = time.Now()
		case "t":
			trade.Event = "trade"
			trade.TradeID = int64(toFloat(raw[1]))
			trade.Type = "sell"
			if t := toFloat(v[2]); t == 1.0 {
				trade.Type = "buy"
			}
			trade.Rate = toFloat(v[3])
			trade.Amount = toFloat(v[4])
			trade.Total = trade.Rate * trade.Amount
			t := time.Unix(int64(toFloat(v[5])), 0)
			trade.TS = t
		default:
		}
		trades = append(trades, trade)
	}
	return trades, nil
}

// WSIdle idles whilst waiting for callbacks
func (p *Poloniex) WSIdle(dur time.Duration, callbacks ...WSReportFunc) {
	for t := range time.Tick(dur) {
		for _, cb := range callbacks {
			cb(t)
		}
	}
}
