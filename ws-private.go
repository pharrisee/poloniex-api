package poloniex

import (
	turnpike "gopkg.in/beatgammit/turnpike.v2"
)

type (
	subscription struct {
		Command string `json:"command"`
		Channel string `json:"channel"`
	}

	notificationSubscription struct {
		subscription
		Key string `json:"key"`
	}
)

func (p *Poloniex) sendWSMessage(msg interface{}) error {
	p.ws.WriteJSON(msg)
	return nil
}

// messageHandler takes a WS Order or Trade and send it over the channel specified by the user
func (p *Poloniex) messageHandler(ch chan WSTicker) turnpike.EventHandler {
	return func(p []interface{}, n map[string]interface{}) {
		t := WSTicker{
			Pair:          p[0].(string),
			Last:          toFloat(p[1]),
			Ask:           toFloat(p[2]),
			Bid:           toFloat(p[3]),
			PercentChange: toFloat(p[4]) * 100.0,
			BaseVolume:    toFloat(p[5]),
			QuoteVolume:   toFloat(p[6]),
			IsFrozen:      toFloat(p[7]) != 0.0,
			DailyHigh:     toFloat(p[8]),
			DailyLow:      toFloat(p[9]),
		}
		ch <- t
	}
}
