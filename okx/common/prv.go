package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nntaoli-project/goex/v2/httpcli"
	"github.com/nntaoli-project/goex/v2/logger"
	"github.com/nntaoli-project/goex/v2/model"
	"github.com/nntaoli-project/goex/v2/options"
	"github.com/nntaoli-project/goex/v2/util"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Prv struct {
	*OKxV5
	apiOpts options.ApiOptions
}

func (prv *Prv) GetAccount(coin string) (map[string]model.Account, []byte, error) {
	reqUrl := fmt.Sprintf("%s%s", prv.UriOpts.Endpoint, prv.UriOpts.GetAccountUri)
	params := url.Values{}
	params.Set("ccy", coin)
	data, responseBody, err := prv.DoAuthRequest(http.MethodGet, reqUrl, &params, nil)
	if err != nil {
		return nil, responseBody, err
	}
	acc, err := prv.UnmarshalOpts.GetAccountResponseUnmarshaler(data)
	return acc, responseBody, err
}

func (prv *Prv) CreateOrder(pair model.CurrencyPair, qty, price float64, side model.OrderSide, orderTy model.OrderType, opts ...model.OptionParameter) (*model.Order, []byte, error) {
	reqUrl := fmt.Sprintf("%s%s", prv.UriOpts.Endpoint, prv.UriOpts.NewOrderUri)
	params := url.Values{}

	params.Set("instId", pair.Symbol)
	//params.Set("tdMode", "cash")
	//params.Set("posSide", "")
	params.Set("ordType", adaptOrderTypeToSym(orderTy))
	params.Set("px", util.FloatToString(price, pair.PricePrecision))
	params.Set("sz", util.FloatToString(qty, pair.QtyPrecision))

	side2, posSide := adaptOrderSideToSym(side)
	params.Set("side", side2)
	if posSide != "" {
		params.Set("posSide", posSide)
	}
	
	util.MergeOptionParams(&params, opts...)
	AdaptOrderClientIDOptionParameter(&params)

	data, responseBody, err := prv.DoAuthRequest(http.MethodPost, reqUrl, &params, nil)
	if err != nil {
		logger.Errorf("[CreateOrder] response body =%s", string(responseBody))
		return nil, responseBody, err
	}

	ord, err := prv.UnmarshalOpts.CreateOrderResponseUnmarshaler(data)
	if err != nil {
		return nil, responseBody, err
	}

	ord.Pair = pair
	ord.Price = price
	ord.Qty = qty
	ord.Side = side
	ord.OrderTy = orderTy
	ord.Status = model.OrderStatus_Pending

	return ord, responseBody, err
}

func (prv *Prv) AmendOrder(pair model.CurrencyPair, orderId string, newQty, newPrice float64, opts ...model.OptionParameter) ([]byte, error) {
	reqUrl := fmt.Sprintf("%s%s", prv.UriOpts.Endpoint, prv.UriOpts.AmendOrderUri)
	params := url.Values{}
	params.Set("instId", pair.Symbol)
	if orderId != "" {
		params.Set("ordId", orderId)
	}
	if newQty > 0 {
		params.Set("newSz", util.FloatToString(newQty, pair.QtyPrecision))
	}
	if newPrice > 0 {
		params.Set("newPx", util.FloatToString(newPrice, pair.PricePrecision))
	}
	util.MergeOptionParams(&params, opts...)
	AdaptOrderClientIDOptionParameter(&params)
	data, responseBody, err := prv.DoAuthRequest(http.MethodPost, reqUrl, &params, nil)
	if err != nil {
		logger.Errorf("[AmendOrder] response body =%s", string(responseBody))
		return responseBody, err
	}
	err = prv.UnmarshalOpts.AmendOrderResponseUnmarshaler(data)
	return responseBody, err
}

func (prv *Prv) GetOrderInfo(pair model.CurrencyPair, id string, opt ...model.OptionParameter) (*model.Order, []byte, error) {
	reqUrl := fmt.Sprintf("%s%s", prv.UriOpts.Endpoint, prv.UriOpts.GetOrderUri)
	params := url.Values{}
	params.Set("instId", pair.Symbol)
	params.Set("ordId", id)

	util.MergeOptionParams(&params, opt...)

	data, responseBody, err := prv.DoAuthRequest(http.MethodGet, reqUrl, &params, nil)
	if err != nil {
		return nil, responseBody, err
	}

	ord, err := prv.UnmarshalOpts.GetOrderInfoResponseUnmarshaler(data[1 : len(data)-1])
	if err != nil {
		return nil, responseBody, err
	}

	ord.Pair = pair

	return ord, responseBody, nil
}

func (prv *Prv) GetPendingOrders(pair model.CurrencyPair, opt ...model.OptionParameter) ([]model.Order, []byte, error) {
	reqUrl := fmt.Sprintf("%s%s", prv.UriOpts.Endpoint, prv.UriOpts.GetPendingOrdersUri)
	params := url.Values{}
	params.Set("instId", pair.Symbol)

	util.MergeOptionParams(&params, opt...)

	data, responseBody, err := prv.DoAuthRequest(http.MethodGet, reqUrl, &params, nil)
	if err != nil {
		return nil, responseBody, err
	}

	orders, err := prv.UnmarshalOpts.GetPendingOrdersResponseUnmarshaler(data)
	return orders, responseBody, err
}

func (prv *Prv) GetHistoryOrders(pair model.CurrencyPair, opt ...model.OptionParameter) ([]model.Order, []byte, error) {
	reqUrl := fmt.Sprintf("%s%s", prv.UriOpts.Endpoint, prv.UriOpts.GetHistoryOrdersUri)
	params := url.Values{}
	params.Set("instId", pair.Symbol)
	params.Set("limit", "50")

	util.MergeOptionParams(&params, opt...)

	data, responseBody, err := prv.DoAuthRequest(http.MethodGet, reqUrl, &params, nil)
	if err != nil {
		return nil, responseBody, err
	}

	orders, err := prv.UnmarshalOpts.GetHistoryOrdersResponseUnmarshaler(data)
	return orders, responseBody, err
}

func (prv *Prv) CancelOrder(pair model.CurrencyPair, id string, opt ...model.OptionParameter) ([]byte, error) {
	reqUrl := fmt.Sprintf("%s%s", prv.UriOpts.Endpoint, prv.UriOpts.CancelOrderUri)
	params := url.Values{}
	params.Set("instId", pair.Symbol)
	params.Set("ordId", id)
	util.MergeOptionParams(&params, opt...)

	data, responseBody, err := prv.DoAuthRequest(http.MethodPost, reqUrl, &params, nil)
	if data != nil && len(data) > 0 {
		return responseBody, prv.UnmarshalOpts.CancelOrderResponseUnmarshaler(data)
	}

	return responseBody, err
}

func (prv *Prv) SetPositionMode(mode string, opt ...model.OptionParameter) ([]byte, error) {
	reqUrl := fmt.Sprintf("%s%s", prv.UriOpts.Endpoint, prv.UriOpts.SetPositionModeUri)
	params := url.Values{}
	params.Set("posMode", AdaptPositionMode(mode))
	util.MergeOptionParams(&params, opt...)

	data, responseBody, err := prv.DoAuthRequest(http.MethodPost, reqUrl, &params, nil)
	if err != nil {
		return responseBody, err
	}

	if data != nil && len(data) > 0 {
		_, err := prv.UnmarshalOpts.SetPositionModeResponseUnmarshaler(data)
		return responseBody, err
	}

	return responseBody, err
}

func (prv *Prv) SetLeverage(symbol, lever string, opts ...model.OptionParameter) ([]byte, error) {
	reqUrl := fmt.Sprintf("%s%s", prv.UriOpts.Endpoint, prv.UriOpts.SetLeverageUri)
	params := url.Values{}
	params.Set("instId", symbol)
	params.Set("lever", lever)
	util.MergeOptionParams(&params, opts...)
	data, responseBody, err := prv.DoAuthRequest(http.MethodPost, reqUrl, &params, nil)
	if err != nil {
		return responseBody, err
	}
	err = prv.UnmarshalOpts.SetLeverageResponseUnmarshaler(data)
	return responseBody, err
}

func (prv *Prv) GetLeverage(symbol string, opts ...model.OptionParameter) (string, []byte, error) {
	reqUrl := fmt.Sprintf("%s%s", prv.UriOpts.Endpoint, prv.UriOpts.GetLeverageUri)
	params := url.Values{}
	params.Set("instId", symbol)
	util.MergeOptionParams(&params, opts...)

	data, responseBody, err := prv.DoAuthRequest(http.MethodGet, reqUrl, &params, nil)
	if err != nil {
		return "", responseBody, err
	}

	lever, err := prv.UnmarshalOpts.GetLeverageResponseUnmarshaler(data)
	return lever, responseBody, err
}

func (prv *Prv) DoSignParam(httpMethod, apiUri, apiSecret, reqBody string) (signStr, timestamp string) {
	timestamp = time.Now().UTC().Format("2006-01-02T15:04:05.000Z") //iso time style
	payload := fmt.Sprintf("%s%s%s%s", timestamp, strings.ToUpper(httpMethod), apiUri, reqBody)
	signStr, _ = util.HmacSHA256Base64Sign(apiSecret, payload)
	return
}

func (prv *Prv) DoAuthRequest(httpMethod, reqUrl string, params *url.Values, headers map[string]string) ([]byte, []byte, error) {
	var (
		reqBodyStr string
		reqUri     string
	)

	if http.MethodGet == httpMethod {
		reqUrl += "?" + params.Encode()
	}

	if http.MethodPost == httpMethod {
		params.Set("tag", "86d4a3bf87bcBCDE")

		// 检查是否有 attachAlgoOrds
		attachVal := params.Get("attachAlgoOrds")
		if attachVal != "" {
			// 转成 JSON 数组
			var attachArr []attachAlgoOrd
			if err := json.Unmarshal([]byte(attachVal), &attachArr); err != nil {
				return nil, nil, fmt.Errorf("invalid attachAlgoOrds json: %w", err)
			}

			// 先把原有 params 复制到 map[string]interface{}
			jsonMap := make(map[string]interface{})
			for k, v := range *params {
				if k == "attachAlgoOrds" {
					continue
				}
				jsonMap[k] = v[0]
			}
			jsonMap["attachAlgoOrds"] = attachArr

			bodyBytes, _ := json.Marshal(jsonMap)
			reqBodyStr = string(bodyBytes)
		} else {
			// 原有逻辑
			reqBody, _ := util.ValuesToJson(*params)
			reqBodyStr = string(reqBody)
		}
	}

	_url, _ := url.Parse(reqUrl)
	reqUri = _url.RequestURI()
	signStr, timestamp := prv.DoSignParam(httpMethod, reqUri, prv.apiOpts.Secret, reqBodyStr)
	logger.Debugf("[DoAuthRequest] sign base64: %s, timestamp: %s", signStr, timestamp)

	headers = map[string]string{
		"Content-Type": "application/json; charset=UTF-8",
		//"Accept":               "application/json",
		"OK-ACCESS-KEY":        prv.apiOpts.Key,
		"OK-ACCESS-PASSPHRASE": prv.apiOpts.Passphrase,
		"OK-ACCESS-SIGN":       signStr,
		"OK-ACCESS-TIMESTAMP":  timestamp}

	respBody, err := httpcli.Cli.DoRequest(httpMethod, reqUrl, reqBodyStr, headers)
	if err != nil {
		return nil, respBody, err
	}
	logger.Debugf("[DoAuthRequest] response body: %s", string(respBody))

	var baseResp BaseResp
	err = prv.OKxV5.UnmarshalOpts.ResponseUnmarshaler(respBody, &baseResp)
	if err != nil {
		return nil, respBody, err
	}

	if baseResp.Code != 0 {
		var errData []ErrorResponseData
		err = prv.OKxV5.UnmarshalOpts.ResponseUnmarshaler(baseResp.Data, &errData)
		if err != nil {
			logger.Errorf("unmarshal error data error: %s", err.Error())
			return nil, respBody, errors.New(string(respBody))
		}
		if len(errData) > 0 {
			return nil, respBody, errors.New(errData[0].SMsg)
		}
		return nil, respBody, errors.New(baseResp.Msg)
	} // error response process

	return baseResp.Data, respBody, nil
}

func NewPrvApi(opts ...options.ApiOption) *Prv {
	var api = new(Prv)
	for _, opt := range opts {
		opt(&api.apiOpts)
	}
	return api
}

type attachAlgoOrd struct {
	TpTriggerPx string `json:"tpTriggerPx"`
	TpOrdPx     string `json:"tpOrdPx"`
	SlTriggerPx string `json:"slTriggerPx"`
	SlOrdPx     string `json:"slOrdPx"`
}
