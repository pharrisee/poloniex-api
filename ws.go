package poloniex

import (
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
)

const (
	apiURL = "wss://api2.poloniex.com/"
)

type (
	//WSTicker describes a ticker item
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

	WSReportFunc = func(time.Time)
)

func (p *Poloniex) StartWS() {
	go func() {
		for {
			message := []interface{}{}
			err := p.ws.ReadJSON(&message)
			if err != nil {
				log.Println("read:", err)
				continue
			}
			chid := int64(message[0].(float64))
			chids := toString(chid)
			if chid > 100.0 && chid < 1000.0 {
				if err := p.handleOrderBook(message); err != nil {
					continue
				}
			} else if chids == p.ByName["ticker"] {
				if err := p.handleTicker(message); err != nil {
					continue
				}
			}
		}
	}()
}

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

func (p *Poloniex) WSIdle(dur time.Duration, callbacks ...WSReportFunc) {
	for t := range time.Tick(dur) {
		for _, cb := range callbacks {
			cb(t)
		}
	}
}
