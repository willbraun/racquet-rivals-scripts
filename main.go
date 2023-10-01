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
	Draw_id  string
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

	token := Login()
	draws := getDraws(token)
	testDraw := draws[0]

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

				slots.add(Slot{Draw_id: testDraw.ID, Round: 1, Position: position, Name: name, Seed: seed})

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

			slots.add(Slot{Draw_id: testDraw.ID, Round: round, Position: position, Name: name, Seed: seed})

			currentRound++
		}
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping", r.Request.URL)
	})

	fmt.Println("Start scraping")
	c.Visit(testDraw.Url)

	// filter struct to remove positions already in pb
	// loop over remaining struct to upload to pb

	postSlots(slots, token)
}

func trim(s string) string {
	return strings.Trim(s, " \n\r")
}

func (ss *slotSlice) add(s Slot) {
	*ss = append(*ss, s)
}
