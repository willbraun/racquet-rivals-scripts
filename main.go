package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type slot struct {
	draw_id     string
	round       int
	position    int
	player_name string
	seed        string
}

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

	slots := []slot{}

	c.OnHTML(".scores-draw-entry-box", func(e *colly.HTMLElement) {
		table := e.DOM.ChildrenMatcher(goquery.Single(".scores-draw-entry-box-table"))
		if table.Length() > 0 {
			// round 1
			rows := table.ChildrenMatcher(goquery.Single("tbody")).Children()
			rows.Each(func(i int, row *goquery.Selection) {
				values := row.Children().Map(func(i int, s *goquery.Selection) string {
					return trim(s.Text())
				})
				position, seed, name := values[0], values[1], values[2]
				positionInt, err := strconv.Atoi(position)

				if err != nil {
					fmt.Println(err)
				}

				entry := slot{draw_id: "1", round: 1, position: positionInt, player_name: name, seed: seed}
				slots = append(slots, entry)
			})
		} else {
			// other rounds
		}

	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping", r.Request.URL)
	})

	c.Visit("https://www.atptour.com/en/scores/current/chengdu/7581/draws")

	fmt.Println(slots)
	// get current slots from pb, script_user
	// filter struct to remove positions already in pb
	// loop over remaining struct to upload to pb
}

func trim(s string) string {
	return strings.Trim(s, " \n\r")
}
