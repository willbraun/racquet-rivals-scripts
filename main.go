package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

type Slot struct {
	ID       string
	DrawID   string
	Round    int
	Position int
	Name     string
	Seed     string
}

type slotSlice []Slot

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	token := login()
	draws := getDraws(token)
	testDraw := draws[0]
	currentSlots := toSlotSlice(getSlots(testDraw.ID, token))

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

	scrapedSlots := slotSlice{}
	seeds := make(map[string]string)

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

			scrapedSlots.add(Slot{DrawID: testDraw.ID, Round: 1, Position: position, Name: name, Seed: seed})

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

		scrapedSlots.add(Slot{DrawID: testDraw.ID, Round: currentRound, Position: position, Name: name, Seed: seed})

		currentRound++
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping", r.Request.URL)
	})

	fmt.Println("Start scraping")
	c.Visit(testDraw.Url)

	newSlots := getNewSlots(scrapedSlots, currentSlots)
	postSlots(newSlots, token)

	updatedSlots := prepareUpdates(scrapedSlots, currentSlots, seeds)
	updateSlots(updatedSlots, token)
}

func trim(s string) string {
	return strings.Trim(s, " \n\r")
}

func (ss *slotSlice) add(s Slot) {
	*ss = append(*ss, s)
}

func toSlotSlice(s []SlotRecord) slotSlice {
	result := slotSlice{}
	for _, v := range s {
		result.add(Slot{
			ID:       v.ID,
			DrawID:   v.DrawID,
			Position: v.Position,
			Round:    v.Round,
			Name:     v.Name,
			Seed:     v.Seed,
		})
	}
	return result
}

func getNewSlots(scraped slotSlice, current slotSlice) slotSlice {
	currentMap := make(map[string]bool)
	for _, v := range current {
		key := getSlotKey(v)
		currentMap[key] = true
	}

	result := slotSlice{}
	for _, v := range scraped {
		key := getSlotKey(v)
		if !currentMap[key] {
			result.add(v)
		}
	}

	return result
}

func prepareUpdates(scraped slotSlice, current slotSlice, seeds map[string]string) slotSlice {
	scrapedMap := make(map[string]string)
	for _, v := range scraped {
		key := getSlotKey(v)
		scrapedMap[key] = v.Name
	}

	result := slotSlice{}
	for _, v := range current {
		key := getSlotKey(v)
		newName := scrapedMap[key]
		newSeed := seeds[newName]
		if newName != v.Name || newSeed != v.Seed {
			v.Name = newName
			v.Seed = newSeed
			result.add(v)
		}
	}

	return result
}

func getSlotKey(s Slot) string {
	return fmt.Sprintf("%d.%d", s.Round, s.Position)
}
