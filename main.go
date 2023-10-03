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

				scrapedSlots.add(Slot{DrawID: testDraw.ID, Round: 1, Position: position, Name: name, Seed: seed})

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

			scrapedSlots.add(Slot{DrawID: testDraw.ID, Round: round, Position: position, Name: name, Seed: seed})

			currentRound++
		}
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping", r.Request.URL)
	})

	fmt.Println("Start scraping")
	c.Visit(testDraw.Url)

	// load initial draw, should only happen once at beginning
	newSlots := getNewSlots(scrapedSlots, currentSlots)
	postSlots(newSlots, token)

	// filter scrapedSlots to remove the currentSlots WITH NAME
	// take the remaining slots (without names), update name and seed, send update back to pb
	updatedSlots := prepareUpdates(scrapedSlots, currentSlots)
	updateSlots(updatedSlots, token)
	fmt.Println(updatedSlots)
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
		key := getSlotKey(v.Round, v.Position)
		currentMap[key] = true
	}

	result := slotSlice{}
	for _, v := range scraped {
		key := getSlotKey(v.Round, v.Position)
		if !currentMap[key] {
			result.add(v)
		}
	}

	return result
}

func prepareUpdates(scraped slotSlice, current slotSlice) slotSlice {
	currentMap := make(map[string]string)
	for _, v := range current {
		key := getSlotKey(v.Round, v.Position)
		currentMap[key] = v.Name
	}

	result := slotSlice{}
	for _, v := range scraped {
		key := getSlotKey(v.Round, v.Position)
		if currentMap[key] != v.Name {
			result.add(v)
		}
	}

	return result
}

func getSlotKey(r int, p int) string {
	return fmt.Sprintf("%d.%d", r, p)
}
