package poloniex

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/franela/goreq"
	"github.com/k0kubun/pp"
)

type (
	Ticker      map[string]TickerEntry
	TickerEntry struct {
		Last        float64 `json:",string"`
		Ask         float64 `json:"lowestAsk,string"`
		Bid         float64 `json:"highestBid,string"`
		Change      float64 `json:"percentChange,string"`
		BaseVolume  float64 `json:"baseVolume,string"`
		QuoteVolume float64 `json:"quoteVolume,string"`
		IsFrozen    int64   `json:"isFrozen,string"`
		ID          int64   `json:"id"`
	}

	DailyVolume          map[string]DailyVolumeEntry
	DailyVolumeEntry     map[string]float64
	DailyVolumeTemp      map[string]interface{}
	DailyVolumeEntryTemp map[string]interface{}

	OrderBook struct {
		Asks     []Order
		Bids     []Order
		IsFrozen bool
	}
	Order struct {
		Rate   float64
		Amount float64
	}

	OrderBookTemp struct {
		Asks     []OrderTemp
		Bids     []OrderTemp
		IsFrozen interface{}
	}
	OrderTemp        []interface{}
	OrderBookAll     map[string]OrderBook
	OrderBookAllTemp map[string]OrderBookTemp

	TradeHistory      []TradeHistoryEntry
	TradeHistoryEntry struct {
		ID     int64 `json:"globalTradeID"`
		Date   string
		Type   string
		Rate   float64 `json:",string"`
		Amount float64 `json:",string"`
		Total  float64 `json:",string"`
	}

	ChartData      []ChartDataEntry
	ChartDataEntry struct {
		Date            int64
		High            float64
		Low             float64
		Open            float64
		Close           float64
		Volume          float64
		QuoteVolume     float64
		WeightedAverage float64
	}

	Currencies map[string]Currency
	Currency   struct {
		Name           string
		TxFee          float64 `json:",string"`
		MinConf        float64
		DepositAddress string
		Disabled       int64
		Delisted       int64
		Frozen         int64
	}

	LoanOrders struct {
		Offers  []LoanOrder
		Demands []LoanOrder
	}
	LoanOrder struct {
		Rate     float64 `json:",string"`
		Amount   float64 `json:",string"`
		RangeMin float64
		RangeMax float64
	}
)

func (p *Poloniex) Ticker() (ticker Ticker, err error) {
	err = p.public("returnTicker", nil, &ticker)
	return
}

func (p *Poloniex) DailyVolume() (dailyVolume DailyVolume, err error) {
	dvt := DailyVolumeTemp{}
	err = p.public("return24hVolume", nil, &dvt)
	if err != nil {
		return
	}
	dailyVolume = DailyVolume{}
	for k := range dvt {
		v := dvt[k]
		dve := DailyVolumeEntry{}
		switch i := v.(type) {
		default:
			v := i.(map[string]interface{})
			for kk, vv := range v {
				dve[kk] = toFloat(vv)
			}
			dailyVolume[k] = dve
		case string:
			//ignore anything that isn't a map
		}
	}
	return
}

func (p *Poloniex) OrderBook(pair string) (orderBook OrderBook, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("depth", "40")
	obt := OrderBookTemp{}
	err = p.public("returnOrderBook", params, &obt)
	if err != nil {
		return
	}
	orderBook = tempToOrderBook(obt)
	return
}

func (p *Poloniex) OrderBookAll() (orderBook OrderBookAll, err error) {
	params := url.Values{}
	params.Add("depth", "5")
	params.Add("currencyPair", "all")
	obt := OrderBookAllTemp{}
	err = p.public("returnOrderBook", params, &obt)
	if err != nil {
		return
	}
	orderBook = OrderBookAll{}
	for k, v := range obt {
		orderBook[k] = tempToOrderBook(v)
	}
	return
}

func (p *Poloniex) TradeHistory(pair string, dates ...int64) (tradeHistory TradeHistory, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	if len(dates) > 0 {
		// we have a start date
		params.Add("start", fmt.Sprintf("%d", dates[0]))
	}
	if len(dates) > 1 {
		// we have an end date
		params.Add("end", fmt.Sprintf("%d", dates[1]))
	}
	err = p.public("returnTradeHistory", params, &tradeHistory)
	return
}

func (p *Poloniex) ChartData(pair string) (chartData ChartData, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("start", fmt.Sprintf("%d", time.Now().Add(-24*time.Hour).Unix()))
	params.Add("end", "9999999999")
	params.Add("period", "300")
	err = p.public("returnChartData", params, &chartData)
	return
}

func (p *Poloniex) ChartDataPeriod(pair string, start, end time.Time, period ...int) (chartData ChartData, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("start", fmt.Sprintf("%d", start.Unix()))
	params.Add("end", fmt.Sprintf("%d", end.Unix()))
	pi := 300
	if len(period) > 0 {
		pi = period[0]
	}
	ps := fmt.Sprintf("%d", pi)
	params.Add("period", ps)
	err = p.public("returnChartData", params, &chartData)
	return
}

func (p *Poloniex) ChartDataCurrent(pair string) (chartData ChartData, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("start", fmt.Sprintf("%d", time.Now().Add(-5*time.Minute).Unix()))
	params.Add("end", "9999999999")
	params.Add("period", "300")
	err = p.public("returnChartData", params, &chartData)
	return
}

func (p *Poloniex) Currencies() (currencies Currencies, err error) {
	err = p.public("returnCurrencies", nil, &currencies)
	return
}

func (p *Poloniex) LoanOrders(currency string) (loanOrders LoanOrders, err error) {
	params := url.Values{}
	params.Add("currency", currency)
	err = p.public("returnLoanOrders", params, &loanOrders)
	return
}

func tempToOrderBook(obt OrderBookTemp) (ob OrderBook) {
	asks := obt.Asks
	bids := obt.Bids
	ob.IsFrozen = obt.IsFrozen.(string) != "0"
	ob.Asks = []Order{}
	ob.Bids = []Order{}
	for k := range asks {
		v := asks[k]
		price := toFloat(v[0])
		amount := toFloat(v[1])
		o := Order{Rate: price, Amount: amount}
		ob.Asks = append(ob.Asks, o)
	}
	for k := range bids {
		v := bids[k]
		price := toFloat(v[0])
		amount := toFloat(v[1])
		o := Order{Rate: price, Amount: amount}
		ob.Bids = append(ob.Bids, o)
	}
	return
}

func (p *Poloniex) public(command string, params url.Values, retval interface{}) (err error) {
	if p.debug {
		defer un(trace("public: " + command))
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if params == nil {
		params = url.Values{}
	}
	params.Add("command", command)
	req := goreq.Request{Uri: PUBLICURI, QueryString: params, Timeout: 130 * time.Second}
	res, err := req.Do()
	if err != nil {
		return
	}
	if p.debug {
		pp.Println(res.Request.URL.String())
	}

	defer res.Body.Close()

	s, err := res.Body.ToString()
	if err != nil {
		return
	}
	if p.debug {
		pp.Println(s)
	}
	err = json.Unmarshal([]byte(s), retval)
	return
}
