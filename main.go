package main

import (
	"fmt"

	"github.com/gocolly/colly"
)

func main() {
	fmt.Println("Started")
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
    fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)
	})

	c.OnHTML("tr td:nth-of-type(1)", func(e *colly.HTMLElement) {
		fmt.Println("First column of a table row:", e.Text)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping", r.Request.URL)
	})

	c.Visit("https://www.atptour.com/en/scores/current/chengdu/7581/draws")

	fmt.Println("Finished with script")
}