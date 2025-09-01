package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func scrapeATP(scraper Scraper, draw DrawRecord) (SlotSlice, map[string]string) {
	slots := SlotSlice{}
	seeds := make(map[string]string)

	html := scraper.scrape(draw.Url)
	reader := strings.NewReader(html)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Println(err)
	}

	roundContainers := doc.Find(".draw-content").FilterFunction(func(_ int, selection *goquery.Selection) bool {
		return !selection.Parents().Is("template")
	})

	round := 0
	roundContainers.Each(func(_ int, rc *goquery.Selection) {
		round++
		position := 1

		rawSlots := rc.Find(".stats-item")
		rawSlots.Each(func(_ int, rawSlot *goquery.Selection) {
			player := rawSlot.Find(".name")
			name := trim(player.Find("a").Text())
			seed := trim(player.Find("span").Text())

			sets := SetSlice{}
			rawSets := rawSlot.Find(".score-item")
			rawSets.EachWithBreak(func(i int, set *goquery.Selection) bool {
				scores := set.Find("span").Map(func(_ int, span *goquery.Selection) string {
					return trim(span.Text())
				})

				if len(scores) == 0 {
					return false
				}

				gamesStr := scores[0]
				if gamesStr == "" || gamesStr == "-" {
					return false
				}

				games, err := strconv.Atoi(gamesStr)
				if err != nil {
					log.Println("ATP - Error converting games to int:", err)
				}

				tiebreakStr := ""
				if len(scores) > 1 {
					tiebreakStr = scores[1]
				}

				tiebreak := 0
				if tiebreakStr != "" {
					tiebreak, err = strconv.Atoi(tiebreakStr)
					if err != nil {
						log.Println("ATP - Error converting tiebreak to int:", err)
					}
				}

				sets.add(Set{Number: i + 1, Games: games, Tiebreak: tiebreak})

				return true
			})

			slots.add(Slot{DrawID: draw.ID, Round: round, Position: position, Name: name, Seed: seed, Sets: sets})
			seeds[name] = seed

			position++
		})
	})

	round++
	winner := doc.Find(".draw-content").Last().Find(".winner").SiblingsFiltered(".name")
	winnerName := trim(winner.Find("a").Text())
	winnerSeed := trim(winner.Find("span").Text())
	slots.add(Slot{DrawID: draw.ID, Round: round, Position: 1, Name: winnerName, Seed: winnerSeed})

	return slots, seeds
}
