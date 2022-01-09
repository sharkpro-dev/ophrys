package market

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"ophrys/pkg/engine"
	"strconv"
	"strings"
	"time"
)

const (
	API_KEY_HEADER = "X-MBX-APIKEY"

	URL_ORDER               = "api/v3/order/test"
	URL_ALL_ORDERS          = "api/v3/openOrders"
	URL_ACCOUNT_INFORMATION = "api/v3/account"
)

type BinanceClient struct {
	baseUrl    string
	httpClient *http.Client
	apiKey     string
	signature  string
}

func NewBinanceClient(baseUrl string, apiKey string, signature string) *BinanceClient {
	return &BinanceClient{httpClient: &http.Client{}, baseUrl: baseUrl, apiKey: apiKey, signature: signature}
}

func (bc *BinanceClient) NewOrder(symbol string, side string, type_ string, timeInForce string, quantity float64, price float64) *engine.OrderResponse {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", bc.baseUrl, URL_ORDER), nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	req.Header.Add(API_KEY_HEADER, bc.apiKey)

	q := req.URL.Query()

	q.Add("symbol", symbol)
	q.Add("side", side)
	q.Add("type", type_)
	q.Add("timeInForce", timeInForce)
	q.Add("quantity", fmt.Sprintf("%f", quantity))
	q.Add("price", fmt.Sprintf("%f", price))
	q.Add("timestamp", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	e := bc.EncodeWithSignature(q)
	req.URL.RawQuery = e

	// Fetch Request
	response, err := bc.httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if err != nil {
		log.Println(err)
		return nil
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Println(err)
		return nil
	}

	responseMap := map[string]interface{}{}
	err = json.Unmarshal(body, &responseMap)

	if err != nil {
		log.Println(err)
		return nil
	}

	fmt.Println(responseMap)

	return &engine.OrderResponse{
		ClientOrderId: responseMap["clientOrderId"].(string),
		CumQty:        responseMap["cumQty"].(string),
		CumQuote:      responseMap["cumQuote"].(string),
		ExecutedQty:   responseMap["executedQty"].(string),
		OrderId:       responseMap["orderId"].(int),
		AvgPrice:      responseMap["avgPrice"].(string),
		OrigQty:       responseMap["origQty"].(string),
		Price:         responseMap["price"].(string),
		Side:          responseMap["side"].(string),
		PositionSide:  responseMap["positionSide"].(string),
		Status:        responseMap["status"].(string),
		Symbol:        responseMap["symbol"].(string),
		Type_:         responseMap["type"].(string),
		OrigType:      responseMap["origType"].(string),
		UpdateTime:    responseMap["updateTime"].(int),
	}

}

func (bc *BinanceClient) CancelOrder(symbol string, orderId int64) *engine.CancelOrderResponse {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", bc.baseUrl, URL_ORDER), nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	req.Header.Add(API_KEY_HEADER, bc.apiKey)

	q := req.URL.Query()

	q.Add("symbol", symbol)
	q.Add("orderId", fmt.Sprintf("%d", orderId))
	q.Add("timestamp", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	e := bc.EncodeWithSignature(q)
	req.URL.RawQuery = e

	// Fetch Request
	response, err := bc.httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Println(err)
		return nil
	}

	responseMap := map[string]interface{}{}
	err = json.Unmarshal(body, &responseMap)

	if err != nil {
		log.Println(err)
		return nil
	}

	return &engine.CancelOrderResponse{
		ClientOrderId: responseMap["clientOrderId"].(string),
		CumQty:        responseMap["cumQty"].(string),
		CumQuote:      responseMap["cumQuote"].(string),
		ExecutedQty:   responseMap["executedQty"].(string),
		OrderId:       responseMap["orderId"].(int),
		OrigQty:       responseMap["origQty"].(string),
		Price:         responseMap["price"].(string),
		Side:          responseMap["side"].(string),
		PositionSide:  responseMap["positionSide"].(string),
		Status:        responseMap["status"].(string),
		Symbol:        responseMap["symbol"].(string),
		Type_:         responseMap["type"].(string),
		OrigType:      responseMap["origType"].(string),
		ReduceOnly:    responseMap["reduceOnly"].(string),
		ClosePosition: responseMap["closePosition"].(string),
		TimeInForce:   responseMap["timeInForce"].(string),
	}

}

func (bc *BinanceClient) CancelAllOrders(symbol string) bool {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", bc.baseUrl, URL_ALL_ORDERS), nil)
	if err != nil {
		fmt.Println(err)
		return false
	}

	req.Header.Add(API_KEY_HEADER, bc.apiKey)

	q := req.URL.Query()

	q.Add("symbol", symbol)
	q.Add("timestamp", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	e := bc.EncodeWithSignature(q)
	req.URL.RawQuery = e

	// Fetch Request
	response, err := bc.httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Println(err)
		return false
	}

	responseMap := map[string]interface{}{}
	err = json.Unmarshal(body, &responseMap)

	if err != nil {
		log.Println(err)
		return false
	}

	if response.StatusCode != 200 {
		fmt.Println(responseMap["msg"])
		return false
	}

	return true
}

func (bc *BinanceClient) AllOrders(symbol string, startTime int64, endTime int64, limit int) (orders []*engine.OrderResponse) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", bc.baseUrl, URL_ALL_ORDERS), nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	q := req.URL.Query()

	q.Add("symbol", symbol)
	q.Add("startTime", strconv.FormatInt(startTime, 10))
	q.Add("endTime", strconv.FormatInt(endTime, 10))
	q.Add("limit", strconv.Itoa(limit))
	q.Add("timestamp", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	e := bc.EncodeWithSignature(q)
	req.URL.RawQuery = e

	// Fetch Request
	response, err := bc.httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return nil
	}

	responseMap := []map[string]interface{}{}
	err = json.Unmarshal(body, &responseMap)

	if err != nil {
		log.Println(err)
		return nil
	}

	for _, order := range responseMap {
		orderResponse := &engine.OrderResponse{
			ClientOrderId: order["clientOrderId"].(string),
			CumQty:        order["cumQty"].(string),
			CumQuote:      order["cumQuote"].(string),
			ExecutedQty:   order["executedQty"].(string),
			OrderId:       order["orderId"].(int),
			AvgPrice:      order["avgPrice"].(string),
			OrigQty:       order["origQty"].(string),
			Price:         order["price"].(string),
			Side:          order["side"].(string),
			PositionSide:  order["positionSide"].(string),
			Status:        order["status"].(string),
			Symbol:        order["symbol"].(string),
			Type_:         order["type"].(string),
			OrigType:      order["origType"].(string),
			UpdateTime:    order["updateTime"].(int),
		}
		orders = append(orders, orderResponse)
	}

	return orders
}

func (bc *BinanceClient) AccountInformation() *engine.AccountInformationResponse {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", bc.baseUrl, URL_ACCOUNT_INFORMATION), nil)
	if err != nil {
		fmt.Println("req")
		fmt.Println(err)
		return nil
	}

	req.Header.Add(API_KEY_HEADER, bc.apiKey)

	q := req.URL.Query()
	q.Add("timestamp", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	e := bc.EncodeWithSignature(q)
	req.URL.RawQuery = e

	// Fetch Request
	response, err := bc.httpClient.Do(req)
	if err != nil {
		fmt.Println("resp")
		fmt.Println(err)
		return nil
	}

	log.Println(response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("read")
		log.Println(err)
		return nil
	}

	responseMap := map[string]interface{}{}
	err = json.Unmarshal(body, &responseMap)
	log.Println(string(body))

	if err != nil {
		fmt.Println("unmarshall")
		log.Println(err)
		return nil
	}

	assetHoldings := make([]*engine.AssetHolding, 0)
	for _, asset := range responseMap["balances"].([]interface{}) {
		assetHoldings = append(assetHoldings, &engine.AssetHolding{
			Asset:  asset.(map[string]interface{})["asset"].(string),
			Free:   asset.(map[string]interface{})["free"].(string),
			Locked: asset.(map[string]interface{})["locked"].(string),
		})
	}

	return &engine.AccountInformationResponse{
		MakerCommission:  responseMap["makerCommission"].(float64),
		TakerCommission:  responseMap["takerCommission"].(float64),
		BuyerCommission:  responseMap["buyerCommission"].(float64),
		SellerCommission: responseMap["sellerCommission"].(float64),
		CanTrade:         responseMap["canTrade"].(bool),
		CanWithdraw:      responseMap["canWithdraw"].(bool),
		CanDeposit:       responseMap["canDeposit"].(bool),
		AccountType:      responseMap["accountType"].(string),
		Assets:           assetHoldings,
	}

}

func (bc *BinanceClient) EncodeWithSignature(query url.Values) string {
	var buf strings.Builder
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}

	for _, k := range keys {
		vs := query[k]
		keyEscaped := url.QueryEscape(k)
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(v))
		}
	}

	mac := hmac.New(sha256.New, []byte(bc.signature))
	mac.Write([]byte(buf.String()))
	hashedQueryString := mac.Sum(nil)

	if buf.Len() > 0 {
		buf.WriteByte('&')
	}
	buf.WriteString("signature=")

	buf.WriteString(hex.EncodeToString(hashedQueryString))

	return buf.String()
}
