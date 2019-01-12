package poloniex

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/franela/goreq"
)

type (
	//Base is the common fields returned by a call tothepoloniex API
	Base struct {
		Error    string
		Success  int64
		Response string
	}

	//Balances are the complete balance map returned by the poloniex API
	Balances map[string]Balance
	//Balance is a single balance entry used in the Balances map
	Balance struct {
		Available float64 `json:",string"`
		OnOrders  float64 `json:"onOrders,string"`
		BTCValue  float64 `json:"btcValue,string"`
	}

	accountBalancesTemp struct {
		Exchange map[string]string
		Margin   map[string]string
		Lending  map[string]string
	}

	//Account holds the balances in the various wallet accounts
	AccountBalances struct {
		Exchange map[string]float64
		Margin   map[string]float64
		Lending  map[string]float64
	}

	//Addresses holds the various deposit addresses foreach coin
	Addresses map[string]string

	//DepositsWithdrawals holds the history of deposit and withdrawal
	DepositsWithdrawals struct {
		Deposits    []deposit
		Withdrawals []withdrawal
	}
	deposit struct {
		Currency      string
		Address       string
		Amount        float64 `json:",string"`
		Confirmations int64
		TXID          string `json:"txid"`
		Timestamp     int64
		Status        string
	}
	withdrawal struct {
		WithdrawalNumber int64 `json:"withdrawalNumber"`
		Currency         string
		Address          string
		Amount           float64 `json:",string"`
		Timestamp        int64
		Status           string
	}

	//OpenOrders is the list of open orders for the pair specified
	OpenOrders []OpenOrder
	//OpenOrder is a singular entry used in the OpenOrders type
	OpenOrder struct {
		OrderNumber int64 `json:",string"`
		Type        string
		Rate        float64 `json:",string"`
		Amount      float64 `json:",string"`
		Total       float64 `json:",string"`
	}
	//OpenOrdersAll is used for all pairs
	OpenOrdersAll map[string]OpenOrders

	PrivateTradeHistory      []PrivateTradeHistoryEntry
	PrivateTradeHistoryEntry struct {
		Date          string
		Rate          float64 `json:",string"`
		Amount        float64 `json:",string"`
		Total         float64 `json:",string"`
		OrderNumber   int64   `json:",string"`
		Type          string
		GlobalTradeID int64 `json:"globalTradeID"`
	}
	PrivateTradeHistoryAll map[string]PrivateTradeHistory

	OrderTrades []OrderTrade
	OrderTrade  struct {
		GlobalTradeID int64   `json:"globalTradeID"`
		TradeID       int64   `json:"tradeID"`
		CurrencyPair  string  `json:"currencyPair"`
		Type          string  `json:"type"`
		Rate          float64 `json:"rate,string"`
		Amount        float64 `json:"amount,string"`
		Total         float64 `json:"total,string"`
		Fee           float64 `json:"fee,string"`
		Date          string  `json:"date"`
	}

	Buy struct {
		OrderNumber int64 `json:",string"`
		// ResultingTrades []ResultingTrade
	}
	ResultingTrade struct {
		Amount  float64 `json:",string"`
		Rate    float64 `json:",string"`
		Date    string
		Total   float64 `json:",string"`
		TradeID string  `json:"tradeID"`
		Type    string
	}
	Sell struct {
		Buy
	}

	MoveOrder struct {
		Base
		OrderNumber int64 `json:",string"`
		// ResultingTrades []ResultingTrade
	}

	Withdraw struct {
		Base
	}

	FeeInfo struct {
		MakerFee        float64 `json:"makerFee,string"`
		TakerFee        float64 `json:"takerFee,string"`
		ThirtyDayVolume float64 `json:"thirtyDayVolume,string"`
		NextTier        float64 `json:"nextTier,string"`
	}

	AvailableAccountBalances struct {
		Exchange map[string]float64
		Margin   map[string]float64
		Lending  map[string]float64
	}
	AvailableAccountBalancesTemp struct {
		Exchange map[string]json.Number
		Margin   map[string]json.Number
		Lending  map[string]json.Number
	}

	TradableBalances map[string]TradableBalance
	TradableBalance  map[string]float64

	TradableBalancesTemp map[string]TradableBalanceTemp
	TradableBalanceTemp  map[string]json.Number

	TransferBalance struct {
		Base
		Message string `json:"message"`
	}

	MarginAccountSummary struct {
		TotalValue         float64 `json:"totalValue,string"`
		ProfitLoss         float64 `json:"pl,string"`
		LendingFees        float64 `json:"lendingFees,string"`
		NetValue           float64 `json:"netValue,string"`
		TotalBorrowedValue float64 `json:"totalBorrowedValue,string"`
		CurrentMargin      float64 `json:"currentMargin,string"`
	}

	LoanOffer struct {
		Base
		OrderID int64 `json:"orderID"`
	}

	OpenLoanOffers map[string][]OpenLoanOffer
	OpenLoanOffer  struct {
		ID        int64   `json:"id"`
		Rate      float64 `json:",string"`
		Amount    float64 `json:",string"`
		Duration  int64
		Renewable bool
		AutoRenew int64 `json:"autoRenew"`
		Date      string
		DateTaken time.Time
	}

	ActiveLoans struct {
		Provided []ActiveLoan
	}
	ActiveLoan struct {
		ID        int64 `json:"id"`
		Currency  string
		Rate      float64 `json:",string"`
		Amount    float64 `json:",string"`
		Range     int64
		Renewable bool
		AutoRenew int64 `json:"autoRenew"`
		Date      string
		DateTaken time.Time
		Fees      float64 `json:",string"`
	}
)

func (p *Poloniex) Balances() (balances Balances, err error) {
	err = p.private("returnCompleteBalances", nil, &balances)
	return balances, err
}

func (p *Poloniex) AccountBalances() (balances AccountBalances, err error) {
	b := accountBalancesTemp{}
	p.private("returnAvailableAccountBalances", nil, &b)
	balances = AccountBalances{Exchange: map[string]float64{}, Margin: map[string]float64{}, Lending: map[string]float64{}}
	for k, v := range b.Exchange {
		balances.Exchange[k] = toFloat(v)
	}
	for k, v := range b.Margin {
		balances.Margin[k] = toFloat(v)
	}
	for k, v := range b.Lending {
		balances.Lending[k] = toFloat(v)
	}
	return
}

func (p *Poloniex) Addresses() (addresses Addresses, err error) {
	p.private("returnDepositAddresses", nil, &addresses)
	return
}

func (p *Poloniex) GenerateNewAddress(currency string) (address string, err error) {
	params := url.Values{}
	params.Add("currency", currency)
	b := Base{}
	err = p.private("generateNewAddress", params, &b)
	address = b.Response
	return
}

func (p *Poloniex) DepositsWithdrawals() (depositsWithdrawals DepositsWithdrawals, err error) {
	params := url.Values{}
	params.Add("start", fmt.Sprintf("%d", time.Now().Add(-5208*time.Hour).Unix()))
	params.Add("end", "9999999999")
	err = p.private("returnDepositsWithdrawals", params, &depositsWithdrawals)
	return
}

func (p *Poloniex) OpenOrders(pair string) (openOrders OpenOrders, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	err = p.private("returnOpenOrders", params, &openOrders)
	return
}

func (p *Poloniex) OpenOrdersAll() (openOrders OpenOrdersAll, err error) {
	params := url.Values{}
	params.Add("currencyPair", "all")
	err = p.private("returnOpenOrders", params, &openOrders)
	return
}

//PrivateTradeHistory takes a string pair and 2 unix timestamps as the start and end date period for the request.
func (p *Poloniex) PrivateTradeHistory(pair string, dates ...int64) (history PrivateTradeHistory, err error) {
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
	err = p.private("returnTradeHistory", params, &history)
	return
}

//PrivateTradeHistoryAll takes 2 unix timestamps as the start and end date period for the request.
func (p *Poloniex) PrivateTradeHistoryAll(dates ...int64) (history PrivateTradeHistoryAll, err error) {
	params := url.Values{}
	if len(dates) > 0 {
		// we have a start date
		params.Add("start", fmt.Sprintf("%d", dates[0]))
	}
	if len(dates) > 1 {
		// we have an end date
		params.Add("end", fmt.Sprintf("%d", dates[1]))
	}
	params.Add("currencyPair", "all")
	err = p.private("returnTradeHistory", params, &history)
	return
}

func (p *Poloniex) OrderTrades(orderNumber int64) (ot OrderTrades, err error) {
	params := url.Values{}
	params.Add("orderNumber", fmt.Sprintf("%d", orderNumber))
	err = p.private("returnOrderTrades", params, &ot)
	return
}

func (p *Poloniex) CancelOrder(orderNumber int64) (success bool, err error) {
	params := url.Values{}
	params.Add("orderNumber", fmt.Sprintf("%d", orderNumber))
	b := Base{}
	err = p.private("cancelOrder", params, &b)
	success = b.Success == 1
	return
}

func (p *Poloniex) Buy(pair string, rate, amount float64) (buy Buy, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("rate", fmt.Sprintf("%.8f", rate))
	params.Add("amount", fmt.Sprintf("%.8f", amount))
	err = p.private("buy", params, &buy)
	return
}

func (p *Poloniex) BuyPostOnly(pair string, rate, amount float64) (buy Buy, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("rate", fmt.Sprintf("%.8f", rate))
	params.Add("amount", fmt.Sprintf("%.8f", amount))
	params.Add("postOnly", "1")
	err = p.private("buy", params, &buy)
	return
}

func (p *Poloniex) BuyFillKill(pair string, rate, amount float64) (buy Buy, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("rate", fmt.Sprintf("%.8f", rate))
	params.Add("amount", fmt.Sprintf("%.8f", amount))
	params.Add("fillOrKill", "1")
	err = p.private("buy", params, &buy)
	return
}

func (p *Poloniex) Sell(pair string, rate, amount float64) (sell Sell, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("rate", fmt.Sprintf("%.8f", rate))
	params.Add("amount", fmt.Sprintf("%.8f", amount))
	err = p.private("sell", params, &sell)
	return
}

func (p *Poloniex) SellPostOnly(pair string, rate, amount float64) (sell Sell, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("rate", fmt.Sprintf("%.8f", rate))
	params.Add("amount", fmt.Sprintf("%.8f", amount))
	params.Add("postOnly", "1")
	err = p.private("sell", params, &sell)
	return
}

func (p *Poloniex) SellFillKill(pair string, rate, amount float64) (sell Sell, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("rate", fmt.Sprintf("%.8f", rate))
	params.Add("amount", fmt.Sprintf("%.8f", amount))
	params.Add("fillOrKill", "1")
	err = p.private("sell", params, &sell)
	return
}

func (p *Poloniex) Move(orderNumber int64, rate float64) (moveOrder MoveOrder, err error) {
	params := url.Values{}
	params.Add("orderNumber", fmt.Sprintf("%d", orderNumber))
	params.Add("rate", fmt.Sprintf("%.8f", rate))
	err = p.private("moveOrder", params, &moveOrder)
	return
}

func (p *Poloniex) MovePostOnly(orderNumber int64, rate float64) (moveOrder MoveOrder, err error) {
	return p.Move(orderNumber, rate)
}

func (p *Poloniex) Withdraw(currency string, amount float64, address string) (w Withdraw, err error) {
	params := url.Values{}
	params.Add("currency", currency)
	params.Add("amount", fmt.Sprintf("%f", amount))
	params.Add("address", address)
	err = p.private("withdraw", params, w)
	return
}

func (p *Poloniex) FeeInfo() (fi FeeInfo, err error) {
	err = p.private("returnFeeInfo", nil, &fi)
	return
}

func (p *Poloniex) AvailableAccountBalances() (aab AvailableAccountBalances, err error) {
	aabt := AvailableAccountBalancesTemp{}
	err = p.private("returnAvailableAccountBalances", nil, &aabt)
	if err != nil {
		return
	}
	aab.Exchange = map[string]float64{}
	aab.Margin = map[string]float64{}
	aab.Lending = map[string]float64{}
	for k, v := range aabt.Exchange {
		aab.Exchange[k] = toFloat(v)
	}
	for k, v := range aabt.Margin {
		aab.Margin[k] = toFloat(v)
	}
	for k, v := range aabt.Lending {
		aab.Lending[k] = toFloat(v)
	}
	return
}

func (p *Poloniex) TradableBalances() (tb TradableBalances, err error) {
	tbt := TradableBalancesTemp{}
	err = p.private("returnTradableBalances", nil, &tbt)
	if err != nil {
		return
	}
	tb = TradableBalances{}
	for k, v := range tbt {
		tb[k] = TradableBalance{}
		for kk, vv := range v {
			tb[k][kk] = toFloat(vv)
		}
	}
	return
}

func (p *Poloniex) TransferBalance(currency string, amount float64, from string, to string) (tb TransferBalance, err error) {
	params := url.Values{}
	params.Add("currency", currency)
	params.Add("amount", toString(amount))
	params.Add("fromAccount", from)
	params.Add("toAccount", to)
	fmt.Printf("%+v", params)
	err = p.private("transferBalance", params, &tb)
	return
}

func (p *Poloniex) MarginAccountSummary() (mas MarginAccountSummary, err error) {
	err = p.private("returnMarginAccountSummary", nil, &mas)
	return
}

func (p *Poloniex) LoanOffer(currency string, amount float64, duration int, renew bool, lendingRate float64) (loanOffer LoanOffer, err error) {
	params := url.Values{}
	params.Add("currency", currency)
	params.Add("amount", toString(amount))
	params.Add("lendingRate", toString(lendingRate/100.0))
	params.Add("duration", fmt.Sprintf("%d", duration))
	r := 0
	if renew {
		r = 1
	}
	params.Add("autoRenew", fmt.Sprintf("%d", r))
	err = p.private("createLoanOffer", params, &loanOffer)
	return
}

func (p *Poloniex) CancelLoanOffer(orderNumber int64) (success bool, err error) {
	params := url.Values{}
	params.Add("orderNumber", fmt.Sprintf("%d", orderNumber))
	b := Base{}
	err = p.private("cancelLoanOffer", params, &b)
	success = b.Success == 1
	return
}

func (p *Poloniex) OpenLoanOffers() (openLoanOffers OpenLoanOffers, err error) {
	err = p.private("returnOpenLoanOffers", nil, &openLoanOffers)
	return
}

func (p *Poloniex) ActiveLoans() (activeLoans ActiveLoans, err error) {
	err = p.private("returnActiveLoans", nil, &activeLoans)
	provided := activeLoans.Provided
	n := []ActiveLoan{}
	for k := range provided {
		v := provided[k]
		v.Renewable = v.AutoRenew == 1
		t, err := time.Parse("2006-01-02 15:04:05", v.Date)
		if err == nil {
			v.DateTaken = t
		}
		n = append(n, v)
	}
	activeLoans.Provided = n
	return
}

func (p *Poloniex) ToggleAutoRenew(orderNumber int64) (success bool, err error) {
	params := url.Values{}
	params.Add("orderNumber", fmt.Sprintf("%d", orderNumber))
	b := Base{}
	err = p.private("toggleAutoRenew", params, &b)
	success = b.Success == 1
	return
}

// make a call to the jsonrpc api, marshal into v
func (p *Poloniex) private(method string, params url.Values, retval interface{}) error {
	if p.debug {
		defer un(trace("private: " + method))
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()
	if params == nil {
		params = url.Values{}
	}
	params.Set("nonce", p.getNonce())
	params.Set("command", method)
	postData := params.Encode()

	req := goreq.Request{
		Method:      "POST",
		Uri:         PRIVATEURI,
		Body:        postData,
		ContentType: "application/x-www-form-urlencoded",
		Accept:      "application/json",
		Timeout:     130 * time.Second,
	}

	req.AddHeader("Sign", p.sign(postData))
	req.AddHeader("Key", p.Key)
	req.AddHeader("Content-Length", strconv.Itoa(len(postData)))

	res, err := req.Do()
	if err != nil {
		return err
	}

	defer res.Body.Close()

	s, err := res.Body.ToString()
	if err != nil {
		return err
	}

	if p.debug {
		fmt.Println(s)
	}

	if strings.HasPrefix(s, "[") {
		// poloniex only ever returns an array type when there is no real data
		// e.g. no data in a time range
		// if this ever changes then this breaks badly
		return nil
	}

	// do we have an error message from the server?
	perr := PoloniexError{}
	err = json.Unmarshal([]byte(s), &perr)
	if err == nil && perr.Error != "" {
		// looks like we have an error from poloniex
		return fmt.Errorf(perr.Error)
	}

	err = json.Unmarshal([]byte(s), retval)
	if err != nil && retval == nil {
		log.Println(err)
		return err
	}
	return err
}

// generate hmac-sha512 hash, hex encoded
func (p *Poloniex) sign(payload string) string {
	mac := hmac.New(sha512.New, []byte(p.Secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
