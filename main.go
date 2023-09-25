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
		fmt.Println("Response code:", r.StatusCode)
	})

	// get HTML from URL
	// format HTML into struct
	
	c.OnHTML(".scores-draw-table > tbody > tr > td[rowspan='1']", func(e *colly.HTMLElement) {
		fmt.Println(e.Text)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping", r.Request.URL)
	})

	c.Visit("https://www.atptour.com/en/scores/current/chengdu/7581/draws")
	
	
	// get current slots from pb, script_user
	// filter struct to remove positions already in pb
	// loop over remaining struct to upload to pb

	fmt.Println("Finished with script")
}