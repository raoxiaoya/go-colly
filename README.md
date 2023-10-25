# go-colly
获取股票数据

colly仓库：https://github.com/gocolly/colly
colly文档：http://go-colly.org/docs/examples/basic/

```bash
go get -u github.com/gocolly/colly/...
```


1、同花顺问财

https://q.10jqka.com.cn 

http://search.10jqka.com.cn/unifiedwap/home/index

2、新浪财经：最低拉取5分钟间隔的实时数据
http://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=sz002139&scale=5&ma=5&datalen=10
（参数为：symbol=【股票编号】&scale=【分钟间隔（5、15、30、60）】&ma=【均值（5、10、15、20、25）】&datalen=【查询个数点（最大值242）】）
获取的数据为：日期、开盘价、最高价、最低价、收盘价、成交量。

3、腾讯股票接口：最低拉取1分钟间隔的实时数据
https://web.ifzq.gtimg.cn/appstock/app/minute/query?code=sz002139

上证：sh，深圳：sz

4、九方智投：获取实时报价
股票
    地址：https://qas.sylapp.cn/api/v30/busi
    类型：post
    header：token
    rawdata：{"Market":"SZ","Inst":"002139","Period":"DAY","ReqID":1,"servicetype":"KLINE","StartID":0,"EndID":-1} 

ETF
    类型：get
    https://hq.chongnengjihua.com/rjhy-gmg-quote/api/1/stock/getastockfundamentals?symbol=shetf510300
    https://hq.chongnengjihua.com/rjhy-gmg-quote/api/1/stock/getastockfundamentals?symbol=szetf159673