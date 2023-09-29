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

type slotSlice []slot

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

	slots := slotSlice{}
	seeds := make(map[string]string)
	currentRound := 2
	positions := make(map[int]int)

	c.OnHTML(".scores-draw-entry-box", func(e *colly.HTMLElement) {
		table := e.DOM.ChildrenMatcher(goquery.Single(".scores-draw-entry-box-table"))
		if table.Length() > 0 {
			// round 1
			rows := table.ChildrenMatcher(goquery.Single("tbody")).Children()
			rows.Each(func(i int, row *goquery.Selection) {
				values := row.Children().Map(func(i int, s *goquery.Selection) string {
					return trim(s.Text())
				})
				positionStr, seed, name := values[0], values[1], values[2]
				
				position, err := strconv.Atoi(positionStr)
				if err != nil {
					fmt.Println(err)
				}

				slots.add(slot{draw_id: "1", round: 1, position: position, player_name: name, seed: seed})
				
				seeds[name] = seed
				currentRound = 2
			})
		} else {
			// other rounds
			round := currentRound
			positions[round]++
			position := positions[round]
			name := trim(e.DOM.ChildrenMatcher(goquery.Single(".scores-draw-entry-box-players-item")).Text())
			seed := seeds[name]

			slots.add(slot{draw_id: "1", round: round, position: position, player_name: name, seed: seed})

			currentRound++
		}
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping", r.Request.URL)
	})

	c.Visit("https://www.atptour.com/en/scores/current/beijing/747/draws")

	fmt.Println(slots)

	// get current slots from pb, script_user
	// filter struct to remove positions already in pb
	// loop over remaining struct to upload to pb
}

func trim(s string) string {
	return strings.Trim(s, " \n\r")
}

func (ss *slotSlice) add(s slot) {
	*ss = append(*ss, s)
}