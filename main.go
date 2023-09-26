package main

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func main() {
	// get HTML from URL
	// format HTML into struct

	type slot struct {
		draw_id     string
		player_name string
		seed        string
		round       int
		position    int
	}

	slots := []slot{}

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

	c.OnHTML(".scores-draw-entry-box", func(e *colly.HTMLElement) {
		table := e.DOM.ChildrenMatcher(goquery.Single(".scores-draw-entry-box-table"))
		if table.Length() > 0 {
			// add each player from round 1

			entry := slot{draw_id: "1", player_name: "test", seed: "1", round: 1, position: 1}
			slots = append(slots, entry)
		} else {
			// add the single player in the other rounds
		}

	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping", r.Request.URL)
	})

	c.Visit("https://www.atptour.com/en/scores/current/chengdu/7581/draws")

	// get current slots from pb, script_user
	// filter struct to remove positions already in pb
	// loop over remaining struct to upload to pb
}
