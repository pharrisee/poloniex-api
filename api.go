package poloniex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/recws-org/recws"

	"github.com/chuckpreslar/emission"

	"github.com/pkg/errors"
)

type (
	// Poloniex describes the API
	Poloniex struct {
		Key           string
		Secret        string
		ws            recws.RecConn
		debug         bool
		nonce         int64
		mutex         sync.Mutex
		emitter       *emission.Emitter
		subscriptions map[string]bool
		ByID          map[string]string
		ByName        map[string]string
	}

	// Error is a domain specific error
	Error struct {
		Error string `json:"error"`
	}
)

const (
	// PUBLICURI is the address of the public API on Poloniex
	PUBLICURI = "https://poloniex.com/public"
	// PRIVATEURI is the address of the public API on Poloniex
	PRIVATEURI = "https://poloniex.com/tradingApi"
)

// Debug turns on debugmode, which basically dumps all responses from the poloniex API REST server
func (p *Poloniex) Debug() {
	p.debug = true
}

func (p *Poloniex) getNonce() string {
	p.nonce++
	return fmt.Sprintf("%d", p.nonce)
}

// NewWithCredentials allows to pass in the key and secret directly
func NewWithCredentials(key, secret string) *Poloniex {
	p := &Poloniex{}
	p.Key = key
	p.Secret = secret
	p.nonce = time.Now().UnixNano()
	p.mutex = sync.Mutex{}
	p.emitter = emission.NewEmitter()
	p.subscriptions = map[string]bool{}
	p.ws = recws.RecConn{}
	p.ws.Dial(apiURL, http.Header{})
	p.getMarkets()

	return p
}

// NewWithConfig is the replacement function for New, pass in a configfile to use
func NewWithConfig(configfile string) *Poloniex {
	p := map[string]string{}
	// we have a configfile
	b, err := ioutil.ReadFile(configfile)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "reading "+configfile+" failed."))
	}
	err = json.Unmarshal(b, &p)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "unmarshal of config failed."))
	}
	return NewWithCredentials(p["key"], p["secret"])
}

// NewPublicOnly allows the use of the public and websocket api only
func NewPublicOnly() *Poloniex {
	p := &Poloniex{}
	p.nonce = time.Now().UnixNano()
	p.mutex = sync.Mutex{}
	p.emitter = emission.NewEmitter()
	p.subscriptions = map[string]bool{}
	p.ws = recws.RecConn{}
	p.ws.Dial(apiURL, http.Header{})
	p.getMarkets()
	return p
}

// New is the legacy way to create a new client, here just to maintain api
func New(configfile string) *Poloniex {
	return NewWithConfig(configfile)
}

func (p *Poloniex) getMarkets() {
	markets, err := p.Ticker()
	if err != nil {
		log.Fatalln("error getting markets for lookups", err)
	}
	ByName := map[string]string{}
	ByID := map[string]string{}
	for k, v := range markets {
		id := fmt.Sprintf("%d", v.ID)
		ByName[k] = id
		ByID[id] = k
	}

	ByID["1001"] = "trollbox"
	ByID["1002"] = "ticker"
	ByID["1003"] = "footer"
	ByID["1010"] = "heartbeat"

	ByName["trollbox"] = "1001"
	ByName["ticker"] = "1002"
	ByName["footer"] = "1003"
	ByName["heartbeat"] = "1010"

	p.ByID = ByID
	p.ByName = ByName
}

func trace(s string) (string, time.Time) {
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	elapsed := time.Since(startTime)
	log.Printf("trace end: %s, elapsed %f secs\n", s, elapsed.Seconds())
}

func toFloat(i interface{}) float64 {
	maxFloat := float64(math.MaxFloat64)
	switch i := i.(type) {
	case string:
		a, err := strconv.ParseFloat(i, 64)
		if err != nil {
			return maxFloat
		}
		return a
	case float64:
		return i
	case int64:
		return float64(i)
	case json.Number:
		a, err := i.Float64()
		if err != nil {
			return maxFloat
		}
		return a
	}
	return maxFloat
}

func toString(i interface{}) string {
	switch i := i.(type) {
	case string:
		return i
	case float64:
		return fmt.Sprintf("%.8f", i)
	case int64:
		return fmt.Sprintf("%d", i)
	case json.Number:
		return i.String()
	}
	return ""
}
