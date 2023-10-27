package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var (
	// Clients and Transports are safe for concurrent use by multiple goroutines and for efficiency should only be created once and re-used.
	customizedClient = http.Client{Timeout: time.Second * 10}
)

// HttpRequest
// GET:  HttpRequest("http..", "GET", nil, "")
// POST: HttpRequest("http..", "POST", [content-type=application/x-www-form-urlencoded], "a=1&b=2")
// POST: HttpRequest("http..", "POST", [content-type=application/json], "{a:1, b:2}")
func HttpRequest(link string, method string, headers map[string]string, body string) (response string, err error) {
	return HttpClientRequest(&customizedClient, link, method, headers, body)
}

func HttpClientRequest(client *http.Client, link string, method string, headers map[string]string, body string) (response string, err error) {
	method = strings.ToUpper(method)
	req, err := http.NewRequest(method, link, strings.NewReader(body))
	if err != nil {
		return "", errors.New("NewRequest error")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("client.Do error")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		return string(data), errors.New(resp.Status)
	}
	return string(data), nil
}

func CurlGetWithParam(link string, param map[string]interface{}) (response string, err error) {
	u, _ := url.Parse(link)
	q := u.Query()
	for k, v := range param {
		var val string
		switch ins := v.(type) {
		case string:
			val = ins
		case int:
			val = strconv.Itoa(ins)
		case bool:
			val = strconv.FormatBool(ins)
		case float64:
			val = strconv.FormatFloat(ins, 'f', -1, 64)
		}
		q.Set(k, val)
	}
	u.RawQuery = q.Encode()
	return HttpRequest(u.String(), "GET", nil, "")
}

func CurlWithParam(link string, method string, param map[string]string) (response string, err error) {
	u, _ := url.Parse(link)
	q := u.Query()
	for k, val := range param {
		q.Set(k, val)
	}
	u.RawQuery = q.Encode()
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	method = strings.ToUpper(method)
	if method == "GET" {
		link = u.String()
		headers = nil
	}
	return HttpRequest(link, method, headers, q.Encode())
}

var folder = "files/"

func Save(filename string, content string) {
	full := folder + filename

	file, err := os.OpenFile(full, os.O_CREATE|os.O_RDWR, 755)
	if err != nil {
		log.Fatal(err)
	}

	file.WriteString(content)
}

func PrettyPrint(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Println(v)
		return
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "  ")
	if err != nil {
		fmt.Println(v)
		return
	}

	fmt.Println(out.String())
}

func GetStockData() ([]KlineData, error) {
	conf, err := ParseConfigFile()
	if err != nil {
		return nil, err
	}

	var result []KlineData

	for _, v := range conf.Stock {
		code := strings.Split(v, "_")
		var data KlineData
		var err error
		if code[3] == "1" {
			data, err = GetStockDataFromJFZT(code[0], code[1], conf.Token)
			if err != nil {
				log.Println(err)
			}
		} else if code[3] == "2" {
			data, err = GetEtfDataFromJFZT(code[0], code[1])
		}

		data.StockCode = code[1]
		data.StockName = code[2]

		result = append(result, data)
	}

	return result, nil
}

/*
	{
	    "Code": "0000",
	    "Message": "ok",
	    "ReqID": 1,
	    "QuoteData": {
	        "KlineData": [
	            {
	                "TradingDay": 1698076800,
	                "Time": 1698076800,
	                "High": 9.95,
	                "Open": 9.9,
	                "Low": 9.34,
	                "Close": 9.76,
	                "Volume": 35561516,
	                "Amount": 341673611.45,
	                "TickCount": 0,
	                "AfterTradeVolume": 0,
	                "AfterTradeAmount": 0,
	                "PreClose": 10.07,
	                "SettlementPrice": 0
	            }
	        ]
	    }
	}
*/
type ResponseData struct {
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	ReqID     int    `json:"ReqID"`
	QuoteData map[string][]KlineData
}
type KlineData struct {
	TradingDay       int64   `json:"TradingDay"`
	Time             int64   `json:"Time"`
	High             float64 `json:"High"`  // 今日最高价
	Open             float64 `json:"Open"`  //今日开盘价
	Low              float64 `json:"Low"`   //今日最低价
	Close            float64 `json:"Close"` //当前报价
	Volume           int64   `json:"Volume"`
	Amount           float64 `json:"Amount"`
	TickCount        int64   `json:"TickCount"`
	AfterTradeVolume int64   `json:"AfterTradeVolume"`
	AfterTradeAmount float64 `json:"AfterTradeAmount"`
	PreClose         float64 `json:"PreClose"` // 上一天收盘价
	SettlementPrice  float64 `json:"SettlementPrice"`
	StockCode        string
	StockName        string
}

/*
{"Market":"SZ","Inst":"002139","Period":"DAY","ReqID":1,"servicetype":"KLINE","StartID":0,"EndID":-1}
*/
type RequestData struct {
	Market      string `json:"Market"`
	Inst        string `json:"Inst"`
	Period      string `json:"Period"`
	ReqID       int    `json:"ReqID"`
	Servicetype string `json:"servicetype"`
	StartID     int    `json:"StartID"`
	EndID       int    `json:"EndID"`
}

type Config struct {
	Stock []string
	Token string
}

func ParseConfigFile() (Config, error) {
	var conf Config
	file, err := os.Open("code.txt")
	if err != nil {
		return Config{}, err
	}

	ids := make([]string, 0)
	r := bufio.NewReader(file)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			break
		}
		bs := string(b)
		if strings.HasPrefix(bs, "//") {
			continue
		}
		if strings.HasPrefix(bs, "token_") {
			conf.Token = strings.ReplaceAll(bs, "token_", "")
			continue
		}
		ids = append(ids, bs)
	}
	conf.Stock = ids

	return conf, nil
}

// 九方智投
func GetStockDataFromJFZT(market string, inst string, token string) (KlineData, error) {
	req := RequestData{
		Market:      market,
		Inst:        inst,
		Period:      "DAY",
		ReqID:       1,
		Servicetype: "KLINE",
		StartID:     0,
		EndID:       -1,
	}
	byt, _ := json.Marshal(req)

	header := map[string]string{
		"token":        token,
		"content-type": "application/x-www-form-urlencoded",
	}
	resp, err := HttpRequest("https://qas.sylapp.cn/api/v30/busi", "POST", header, string(byt))
	if err != nil {
		return KlineData{}, nil
	}

	var respData ResponseData
	err = json.Unmarshal([]byte(resp), &respData)
	if err != nil {
		log.Println(err)
	}
	if respData.Code == "0000" {
		return respData.QuoteData["KlineData"][0], nil
	} else if respData.Code == "6403" {
		// token invalid
		// 访问网站 https://stock.9fzt.com/index/sz_002139.html，从接口 https://qas.sylapp.cn/api/v30/busi 中找到 token ，目前是24小时过期。
		return KlineData{}, errors.New(respData.Message)
	} else {
		return KlineData{}, errors.New(respData.Message)
	}
}

// 九方智投
// https://hq.chongnengjihua.com/rjhy-gmg-quote/api/1/stock/getastockfundamentals?symbol=shetf510300
// https://hq.chongnengjihua.com/rjhy-gmg-quote/api/1/stock/getastockfundamentals?symbol=szetf159673
/*
{
    "code": 0,
    "data": {
        "amplitude": 0,
        "bidGrp": null,
        "bps": 0,
        "businessAmount": 0,
        "businessAmountAm": 0,
        "businessAmountIn": 0,
        "businessAmountOut": 0,
        "businessBalance": 0,
        "businessBalanceAm": 0,
        "businessCount": 0,
        "circulationAmount": 0,
        "circulationValue": 0,
        "currentAmount": 0,
        "dataTimestamp": 90018370,
        "day5Vol": 0,
        "downPx": 3194,
        "dynPbRate": 0,
        "entrustDiff": 0,
        "entrustRate": 0,
        "eps": 0,
        "epsTtm": 0,
        "epsYear": 0,
        "finEndDate": 0,
        "finQuarter": 0,
        "highPx": 0,
        "hqTypeCode": "XSHG.EM.ETF",
        "ipoPrice": 0,
        "issueDate": 0,
        "lastPx": 0,
        "lowPx": 0,
        "marketDate": 20231025,
        "marketValue": 0,
        "min5Chgpct": 0,
        "neeqMakerCount": 0,
        "offerGrp": null,
        "openPrice": 0,
        "peRate": 0,
        "preClosePx": 3549,
        "prodCode": "510300",
        "prodName": "300ETF  ",
        "pxChange": 0,
        "pxChangeRate": 0,
        "sharesPerHand": 100,
        "staticPeRate": 0,
        "totalBidTurnover": 0,
        "totalBuyAmount": 0,
        "totalOfferTurnover": 0,
        "totalSellAmount": 0,
        "totalShares": 0,
        "tradeMins": 0,
        "tradeStatus": "START",
        "ttmPeRate": 0,
        "turnoverRatio": 0,
        "upPx": 3904,
        "volRatio": 0,
        "w52HighPx": 4267,
        "w52LowPx": 3488,
        "wAvgPx": 0,
        "withdrawBuyAmount": 0,
        "withdrawBuyNumber": 0,
        "withdrawSellAmount": 0,
        "withdrawSellNumber": 0
    },
    "errorMessage": null,
    "timestamp": 1698195619837
}
*/

type ResponseDataEtf struct {
	Code         int                    `json:"Code"`
	ErrorMessage string                 `json:"errorMessage"`
	Data         map[string]interface{} `json:"data"`
}

func GetEtfDataFromJFZT(market string, inst string) (KlineData, error) {
	link := fmt.Sprintf("https://hq.chongnengjihua.com/rjhy-gmg-quote/api/1/stock/getastockfundamentals?symbol=%setf%s", strings.ToLower(market), inst)
	resp, err := HttpRequest(link, "GET", nil, "")
	if err != nil {
		return KlineData{}, nil
	}

	var respData ResponseDataEtf
	err = json.Unmarshal([]byte(resp), &respData)
	if err != nil {
		log.Println(err)
	}
	if respData.Code == 0 {
		preClosePx := respData.Data["preClosePx"].(float64)
		highPx := respData.Data["highPx"].(float64)
		lowPx := respData.Data["lowPx"].(float64)
		closePx := respData.Data["lastPx"].(float64)
		openPx := respData.Data["openPrice"].(float64)
		var scale float64 = 0.001
		da := KlineData{
			High:     highPx * scale,
			Low:      lowPx * scale,
			PreClose: preClosePx * scale,
			Close:    closePx * scale,
			Open:     openPx * scale,
		}
		return da, nil
	} else {
		return KlineData{}, errors.New(respData.ErrorMessage)
	}
}

// https://blog.csdn.net/Meepoljd/article/details/129422612
func BuildTable(result []KlineData) string {
	t := table.NewWriter()
	header := table.Row{"Code", "Name", "Yesterday", "Current", "Open", "High", "Low"}
	t.AppendHeader(header)
	t.SetAutoIndex(true)

	for _, v := range result {
		var curr, high, low, open float64
		if v.Close > 0 {
			curr = (v.Close - v.PreClose) / v.PreClose
		}
		if v.High > 0 {
			high = (v.High - v.PreClose) / v.PreClose
		}
		if v.Low > 0 {
			low = (v.Low - v.PreClose) / v.PreClose
		}
		if v.Open > 0 {
			open = (v.Open - v.PreClose) / v.PreClose
		}

		preclode := fmt.Sprintf("%.3f", v.PreClose)
		currperc := fmt.Sprintf("%.3f [%.2f%%]", v.Close, (math.Round(10000*curr))/100)
		openperc := fmt.Sprintf("%.3f [%.2f%%]", v.Open, (math.Round(10000*open))/100)
		highperc := fmt.Sprintf("%.3f [%.2f%%]", v.High, (math.Round(10000*high))/100)
		lowperc := fmt.Sprintf("%.3f [%.2f%%]", v.Low, (math.Round(10000*low))/100)

		// 样式
		SetColumnStyle(t, []string{"Current", "Open", "High", "Low"}, nil)

		row := table.Row{v.StockCode, v.StockName, preclode, currperc, openperc, highperc, lowperc}
		t.AppendRow(row)
	}

	return t.Render()
}

/*
\033[0m 关闭所有属性
\033[1m 设置高亮度
\033[4m 下划线
\033[5m 闪烁
\033[7m 反显
\033[8m 消隐
\033[30m — \033[37m 设置前景色
\033[40m — \033[47m 设置背景色
\033[nA 光标上移n行
\033[nB 光标下移n行
\033[nC 光标右移n行
\033[nD 光标左移n行
\033[y;xH设置光标位置
\033[2J 清屏
\033[K 清除从光标到行尾的内容
\033[s 保存光标位置
\033[u 恢复光标位置
\033[?25l 隐藏光标
\033[?25h 显示光标
*/
func RefreshTable(data string) string {
	lines := strings.Split(data, "\n")
	num := strconv.Itoa(len(lines))
	for k, line := range lines {
		lines[k] = "\033[K" + line
	}

	lines[0] = "\033[" + num + "A" + lines[0]
	dst := strings.Join(lines, "\n")
	return dst
}

// 设置文字为红色和绿色
func GetColumnTransformer() text.Transformer {
	warnTransformer := text.Transformer(func(val interface{}) string {
		WarnColor := text.Colors{text.FgRed}
		GreenColor := text.Colors{text.FgGreen}
		
		if strings.Contains(val.(string), "-") {
			return GreenColor.Sprint(val)
		} else if strings.Contains(val.(string), "0.00") {
			return fmt.Sprint(val)
		} else {
			return WarnColor.Sprint(val)
		}
	})

	return warnTransformer
}

func SetColumnStyle(t table.Writer, columns []string, warnTransformer text.Transformer) {
	if len(columns) == 0 {
		return
	}

	tableColumnConfig := make([]table.ColumnConfig, 0)
	for _, column := range columns {
		tableColumnConfig = append(tableColumnConfig, table.ColumnConfig{
			Name:        column,
			AutoMerge:   true,
			Align:       text.AlignRight,
			AlignHeader: text.AlignCenter,
			AlignFooter: text.AlignCenter,
			Transformer: warnTransformer,
		})
	}

	t.SetColumnConfigs(tableColumnConfig)
}
