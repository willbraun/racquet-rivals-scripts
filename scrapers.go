package main

import (
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func scrapeATP(draw DrawRecord) (slotSlice, map[string]string) {
	slots := slotSlice{}
	seeds := make(map[string]string)

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

	// first round
	c.OnHTML(".scores-draw-entry-box-table", func(e *colly.HTMLElement) {
		rows := e.DOM.ChildrenMatcher(goquery.Single("tbody")).Children()
		rows.Each(func(i int, row *goquery.Selection) {
			values := row.Children().Map(func(i int, s *goquery.Selection) string {
				return trim(s.Text())
			})
			positionStr, seed, name := values[0], values[1], values[2]

			position, err := strconv.Atoi(positionStr)
			if err != nil {
				fmt.Println(err)
			}

			slots.add(Slot{DrawID: draw.ID, Round: 1, Position: position, Name: name, Seed: seed})

			seeds[name] = seed
		})
	})

	currentRound := 2
	positions := make(map[int]int)

	// other rounds
	c.OnHTML(".scores-draw-entry-box-wrapper", func(e *colly.HTMLElement) {
		rowspan, _ := e.DOM.Parent().Attr("rowspan")
		if rowspan == "1" {
			currentRound = 2
		}

		positions[currentRound]++
		position := positions[currentRound]
		name := trim(e.DOM.Children().ChildrenMatcher(goquery.Single(".scores-draw-entry-box-players-item")).Text())
		seed := seeds[name]

		slots.add(Slot{DrawID: draw.ID, Round: currentRound, Position: position, Name: name, Seed: seed})

		currentRound++
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping", r.Request.URL)
	})

	fmt.Println("Start scraping")
	c.Visit(draw.Url)

	return slots, seeds
}

func scrapeWTA(draw DrawRecord) (slotSlice, map[string]string) {
	slots := slotSlice{}
	seeds := make(map[string]string)

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

	c.OnHTML(".tournament-draw__round-container", func(e *colly.HTMLElement) {
		fmt.Println(e.DOM.Attr("data-round"))
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping", r.Request.URL)
	})

	fmt.Println("Start scraping")
	c.Visit(draw.Url)

	fmt.Println("Scraped WTA:", slots)
	return slots, seeds
}
