package main

import (
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func scrapeWTA(scraper Scraper, draw DrawRecord) (SlotSlice, map[string]string) {
	slots := SlotSlice{}
	seeds := make(map[string]string)

	html := scraper.scrape(draw.Url)
	reader := strings.NewReader(html)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Println(err)
	}

	slotMap := make(map[SlotKey]*Slot)

	roundContainers := doc.Find(`.tournament-draw__tab[data-event-type="LS"]`).Find(".tournament-draw__round-container")
	roundContainers.Each(func(i int, rc *goquery.Selection) {
		round := i + 1
		position := 1

		rawSlots := rc.Find(".match-table__row")
		rawSlots.Each(func(_ int, rawSlot *goquery.Selection) {
			name, seed := wtaExtractName(rawSlot)
			seeds[name] = seed

			sets := SetSlice{}
			rawSets := rawSlot.Find(".match-table__score-cell")
			rawSets.EachWithBreak(func(i int, set *goquery.Selection) bool {
				scores := strings.Fields(set.Text())

				if len(scores) == 0 {
					return false
				}

				gameStr := scores[0]
				if gameStr == "." || gameStr == "" {
					return false
				}

				games, err := strconv.Atoi(gameStr)
				if err != nil {
					log.Println("WTA - Error converting games to int:", err)
				}

				tiebreakStr := ""
				if len(scores) > 1 {
					tiebreakStr = scores[1]
				}

				tiebreak := 0
				if tiebreakStr != "" {
					tiebreak, err = strconv.Atoi(tiebreakStr)
					if err != nil {
						log.Println("WTA - Error converting tiebreak to int:", err)
					}
				}

				sets.add(Set{Number: i + 1, Games: games, Tiebreak: tiebreak})

				return true
			})

			// Add slot for round 1
			// For other rounds, update slot with sets, other fields should be the same
			key := SlotKey{Round: round, Position: position}
			if slot, ok := slotMap[key]; ok {
				slot.Sets = sets
			} else {
				slotMap[key] = &Slot{
					DrawID:   draw.ID,
					Round:    round,
					Position: position,
					Name:     name,
					Seed:     seed,
					Sets:     sets,
				}
			}

			// Placeholder final slot
			if round == roundContainers.Length() {
				nextRound := round + 1
				nextKey := SlotKey{Round: nextRound, Position: 1}
				slotMap[nextKey] = &Slot{
					DrawID:   draw.ID,
					Round:    nextRound,
					Position: 1,
					Name:     "",
					Seed:     "",
				}
			}

			// Check if the player is a winner
			// WTA site only fills slots when matches are complete, so we fill in the next round
			// Add slot for the next round
			if rawSlot.HasClass("is-winner") {
				nextRound := round + 1
				nextRoundPosition := (position + 1) / 2
				nextKey := SlotKey{Round: nextRound, Position: nextRoundPosition}
				slotMap[nextKey] = &Slot{
					DrawID:   draw.ID,
					Round:    nextRound,
					Position: nextRoundPosition,
					Name:     name,
					Seed:     seed,
				}
			}

			position++
		})
	})

	for _, slot := range slotMap {
		slots.add(*slot)
	}

	sort.Slice(slots, func(i, j int) bool {
		if slots[i].Round == slots[j].Round {
			return slots[i].Position < slots[j].Position
		}
		return slots[i].Round < slots[j].Round
	})

	return slots, seeds
}

func wtaExtractName(x *goquery.Selection) (string, string) {
	data := x.Find(".match-table__player-name")

	if data.Length() == 0 {
		return "", ""
	}

	name := trim(data.Find(".match-table__player-fullname").Text())

	if !hasAlphabet(name) {
		return "", ""
	}

	seed := trim(data.Find(".match-table__player-seed").Text())

	return name, seed
}
