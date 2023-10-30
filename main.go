package main

import (
	"fmt"
	"go-colly/util"
	"log"
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
	cmdToken := util.ParseTokenFromParam()
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
