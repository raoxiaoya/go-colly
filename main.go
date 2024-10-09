package main

import (
	"database/sql"
	"fmt"
	"go-colly/util"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/gocolly/colly"
)

func main() {
	fun3()
}

func fun1() {
	c := colly.NewCollector()

	c.OnHTML("html", func(e *colly.HTMLElement) {

		util.Save("sz_002139.html", e.Text)
		// e.Request.Visit(e.Attr("href"))
	})

	c.OnHTML("tr td:nth-of-type(1)", func(e *colly.HTMLElement) {
		fmt.Println("First column of a table row:", e.Text)
	})

	c.Visit("https://stock.9fzt.com/index/sz_002139.html")

}

func fun3() {
	// cmdToken := util.ParseTokenFromParam()

	cmdToken := util.GetTokenFromWebsite()
	if cmdToken == "" {
		log.Fatalln("token is empty")
	}

	var FormatBool bool
	for {
		result, err := util.GetStockData(cmdToken)
		if err != nil {
			log.Fatal(err)
		}

		if FormatBool {
			fmt.Println(util.RefreshTable(util.BuildTable(result)))
		} else {
			fmt.Println(util.BuildTable(result))
			FormatBool = true
		}

		time.Sleep(5 * time.Second)
	}
}

var period = []string{
	"2024Q1",
	"2024Q2",
	"2024Q3",
	"2024Q4",
	"2024",
	"2023Q1",
	"2023Q2",
	"2023Q3",
	"2023Q4",
	"2023",
	"2022Q1",
	"2022Q2",
	"2022Q3",
	"2022Q4",
	"2022",
	"2021Q1",
	"2021Q2",
	"2021Q3",
	"2021Q4",
	"2021",
	"2020Q1",
	"2020Q2",
	"2020Q3",
	"2020Q4",
	"2020",
	"2019Q1",
	"2019Q2",
	"2019Q3",
	"2019Q4",
	"2019",
	"2018Q1",
	"2018Q2",
	"2018Q3",
	"2018Q4",
	"2018",
}

type Item struct {
	Id         int
	StockName  string
	TargetName string
	P2018Q1    float64
	P2018Q2    float64
	P2018Q3    float64
	P2018Q4    float64
	P2018      float64
	P2019Q1    float64
	P2019Q2    float64
	P2019Q3    float64
	P2019Q4    float64
	P2019      float64
	P2020Q1    float64
	P2020Q2    float64
	P2020Q3    float64
	P2020Q4    float64
	P2020      float64
	P2021Q1    float64
	P2021Q2    float64
	P2021Q3    float64
	P2021Q4    float64
	P2021      float64
	P2022Q1    float64
	P2022Q2    float64
	P2022Q3    float64
	P2022Q4    float64
	P2022      float64
	P2023Q1    float64
	P2023Q2    float64
	P2023Q3    float64
	P2023Q4    float64
	P2023      float64
	P2024Q1    float64
	P2024Q2    float64
	P2024Q3    float64
	P2024Q4    float64
	P2024      float64
}

func fun4() {
	sqldb, err := util.CreateSqlite3()
	if err != nil {
		log.Fatal(err)
	}
	defer sqldb.Close()

	target_names := getTargetNames(sqldb)

	sql := `
	SELECT
		id
	FROM
		stock
	WHERE status = 1
	LIMIT 2
	`

	results, err := sqldb.Query(sql)
	if err != nil {
		log.Fatal(err)
	}

	stockList := make([]Item, 0)
	defer results.Close()

	for results.Next() {
		var id int
		err := results.Scan(&id)
		if err != nil {
			fmt.Println("scan error: ", err)
			break
		}
		lt := one(sqldb, id, target_names)
		stockList = append(stockList, lt...)
	}

	// util.PrettyPrint(stockList)

	// 字段标题，和 item 字段一一对应
	headers := []string{"ID", "名称", "指标"}
	headers = append(headers, period...)
	title := "财报数据"
	filepath := "file"
	filename := fmt.Sprintf("stock-%d.xlsx", time.Now().Unix())

	fmt.Println(util.OutPutDataWithXLSX(stockList, headers, title, filepath, filename, len(target_names)))
}

func one(sqldb *sql.DB, stockId int, target_names []string) []Item {
	sql := `
	SELECT
		b.id,
		b.name AS stock_name,
		a.period,
		c.name AS target_name,
		a.data
	FROM
		stock_data a
	INNER JOIN stock b ON a.stock_id = b.id
	INNER JOIN stock_target c ON a.target_id = c.id
	WHERE a.stock_id = ?
	ORDER BY
		a.target_id ASC
	`

	results, err := sqldb.Query(sql, stockId)
	if err != nil {
		log.Fatal(err)
	}

	list := make(map[string]Item)
	res := make([]Item, 0)
	defer results.Close()

	for results.Next() {
		var id int
		var stock_name, period, target_name string
		var data float64
		err := results.Scan(&id, &stock_name, &period, &target_name, &data)
		if err != nil {
			fmt.Println("scan error: ", err)
			break
		}

		if _, ok := list[target_name]; !ok {
			list[target_name] = Item{
				Id:         id,
				StockName:  stock_name,
				TargetName: target_name,
			}
		}

		line := list[target_name]
		reflect.ValueOf(&line).Elem().FieldByName("P" + period).SetFloat(data)

		list[target_name] = line
	}

	for _, target_name := range target_names {
		if line, ok := list[target_name]; ok {
			res = append(res, line)
		}
	}

	return res
}

func getTargetNames(sqldb *sql.DB) []string {
	sql := "select name from stock_target ORDER BY id asc"
	targets, err := sqldb.Query(sql)
	if err != nil {
		log.Fatal(err)
	}

	target_names := make([]string, 0)
	defer targets.Close()

	for targets.Next() {
		var name string
		err := targets.Scan(&name)
		if err != nil {
			fmt.Println("scan error: ", err)
			break
		}
		target_names = append(target_names, name)
	}

	return target_names
}

func createdata() {
	sqldb, _ := util.CreateSqlite3()
	defer sqldb.Close()

	target := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	stock := make([]int, 3)
	format := "insert into stock_data(stock_id, period, target_id, data) values(?,?,?,?)"
	for k := range stock {
		stockId := k + 1
		for _, p := range period {
			for _, t := range target {
				d := int64(100 * rand.Float64())
				_, err := sqldb.Exec(format, stockId, p, t, d)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}
