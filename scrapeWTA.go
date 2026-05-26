package main

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

func scrapeWTA(scraper Scraper, draw DrawRecord) (SlotSlice, map[string]string, error) {
	if strings.Contains(draw.Url, "wtatennis.com") {
		return scrapeWtaOfficial(scraper, draw)
	} else if strings.Contains(draw.Url, "live-tennis.eu/en/wta-singles-draws") {
		return scrapeWtaLiveTennisEu(scraper, draw)
	}
	log.Println("Unsupported WTA site:", draw.Url)
	return SlotSlice{}, nil, fmt.Errorf("unsupported WTA site: %s", draw.Url)
}

func scrapeWtaOfficial(scraper Scraper, draw DrawRecord) (SlotSlice, map[string]string, error) {
	slots := SlotSlice{}
	seeds := make(map[string]string)
	var errs []error

	html, err := scraper.scrape(draw.Url)
	if err != nil {
		return SlotSlice{}, nil, fmt.Errorf("error scraping WTA: %w", err)
	}
	reader := strings.NewReader(html)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Println("error creating document:", err)
		errs = append(errs, fmt.Errorf("error creating document: %w", err))
	}

	slotMap := make(map[SlotKey]*Slot)

	roundContainers := doc.Find(`.tournament-draw__tab[data-event-type="LS"]`).Find(".tournament-draw__round-container")
	roundContainers.Each(func(i int, rc *goquery.Selection) {
		round := i + 1
		position := 1

		rawSlots := rc.Find(".match-table__row")
		rawSlots.Each(func(_ int, rawSlot *goquery.Selection) {
			name, seed := wtaOfficialExtractName(rawSlot)
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
					errs = append(errs, fmt.Errorf("WTA - error converting games to int: %w", err))
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
						errs = append(errs, fmt.Errorf("WTA - error converting tiebreak to int: %w", err))
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

	cleanedSlots, cleanedSeeds := cleanScrapedResults(slots, seeds)

	received := len(cleanedSlots)
	expected := (draw.Size * 2) - 1

	if received != expected {
		errs = append(errs, fmt.Errorf("WTA - Incorrect number of scraped slots: expected %d, got %d", expected, received))
	}

	if len(errs) > 0 {
		return cleanedSlots, cleanedSeeds, errors.Join(errs...)
	}
	return cleanedSlots, cleanedSeeds, nil
}

func wtaOfficialExtractName(x *goquery.Selection) (string, string) {
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

func scrapeWtaLiveTennisEu(scraper Scraper, draw DrawRecord) (SlotSlice, map[string]string, error) {
	slots := SlotSlice{}
	seeds := make(map[string]string)
	var errs []error

	html, err := scraper.scrape(draw.Url)
	if err != nil {
		return SlotSlice{}, nil, fmt.Errorf("error scraping WTA Live Tennis EU: %w", err)
	}
	reader := strings.NewReader(html)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Println("error creating document:", err)
		errs = append(errs, fmt.Errorf("error creating document: %w", err))
	}

	majorNames := []string{"Australian Open", "French Open", "Wimbledon", "US Open"}
	var htmlButtonId string
	var exists bool

	doc.Find("figcaption").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		text := selection.Text()
		if slices.Contains(majorNames, text) {
			htmlButtonId, exists = selection.Closest("button").Attr("id")
			return false // found the first major, break
		}
		return true // keep searching
	})
	if !exists {
		log.Println("WTA - Live Tennis EU draw not found")
		return SlotSlice{}, nil, fmt.Errorf("WTA Live Tennis EU draw not found")
	}

	htmlDrawId := "dr" + htmlButtonId[len(htmlButtonId)-1:]
	htmlDraw := doc.Find("#" + htmlDrawId)
	htmlDrawRows := htmlDraw.ChildrenFiltered("table").ChildrenFiltered("tbody").ChildrenFiltered("tr")

	receivedHtmlRows := htmlDrawRows.Length()

	// Expected states
	// 0 = draw was found but not started
	// 66 = draw is available
	expectedHtmlRows := []int{0, 66}

	if !slices.Contains(expectedHtmlRows, htmlDrawRows.Length()) {
		errs = append(errs, fmt.Errorf("WTA - Incorrect number of HTML rows: expected %d, got %d", expectedHtmlRows, receivedHtmlRows))
	}

	// This site shows a live view of upcoming tournaments. If the draw hasn't started,
	// there will be no rows to scrape. A failure is expected until the draw has been created.
	if receivedHtmlRows == 0 {
		log.Println("WTA - Live Tennis EU draw not started")
		return SlotSlice{}, nil, fmt.Errorf("WTA Live Tennis EU draw not started")
	}

	rowspanToRoundMap := map[string]int{
		"1":  1,
		"2":  2,
		"4":  3,
		"8":  4,
		"16": 5,
		"32": 6,
		"64": 7,
	}
	positionByRound := make(map[int]int)
	winnerName := ""
	winnerSeed := ""

	htmlDrawRows.Each(func(i int, row *goquery.Selection) {
		matches := row.ChildrenFiltered("td")
		matches.Each(func(j int, match *goquery.Selection) {
			round := rowspanToRoundMap[match.AttrOr("rowspan", "1")]

			rawSlots := match.Find("tr")
			rawSlots.Each(func(k int, slot *goquery.Selection) {
				text := slot.Children().Eq(1)
				name := strings.TrimSpace(text.Text())
				seed := ""
				span := text.ChildrenFiltered("span").First()
				if span.Length() > 0 {
					seed = strings.TrimSpace(span.Text())
					name = strings.TrimSpace(strings.ReplaceAll(name, seed, ""))
				}

				formattedName := name
				if name != "" && name != "-" {
					firstName := strings.Split(name, " ")[0]
					lastName := strings.Join(strings.Split(name, " ")[1:], " ")
					firstInitial := string(unicode.ToUpper(rune(firstName[0])))
					formattedName = removeAccents(fmt.Sprintf("%s. %s", firstInitial, lastName))

					seeds[formattedName] = seed
				}

				sets := SetSlice{}
				rawSets := slot.Children().Slice(2, goquery.ToEnd)
				rawSets.EachWithBreak(func(l int, set *goquery.Selection) bool {
					text := set.Text() // contains games and tiebreak as one string

					if text == "Ret." || text == "" {
						return false
					}

					runes := []rune(text)
					var gamesStr string
					var tiebreakStr string

					if len(runes) > 0 {
						gamesStr = string(runes[0])
						if len(runes) > 1 {
							tiebreakStr = string(runes[1:])
						}
					}

					// Walkover, skip adding sets
					if strings.Contains(gamesStr, "w") {
						return false
					}

					games, err := strconv.Atoi(gamesStr)
					if err != nil {
						log.Println("WTA Live Tennis EU - Error converting games to int:", err)
						errs = append(errs, fmt.Errorf("WTA Live Tennis EU - error converting games to int: %w", err))
					}

					tiebreak := 0
					if tiebreakStr != "" {
						tiebreak, err = strconv.Atoi(tiebreakStr)
						if err != nil {
							log.Println("WTA Live Tennis EU - Error converting tiebreak to int:", err)
							errs = append(errs, fmt.Errorf("WTA Live Tennis EU - error converting tiebreak to int: %w", err))
						}
					}

					sets = append(sets, Set{Number: l + 1, Games: games, Tiebreak: tiebreak})
					return true
				})

				positionByRound[round]++
				slots.add(Slot{DrawID: draw.ID, Round: round, Position: positionByRound[round], Name: formattedName, Seed: seed, Sets: sets})

				if round == 7 && text.ChildrenFiltered("b").Length() > 0 {
					winnerName = formattedName
					winnerSeed = seed
				}
			})
		})
	})

	slots.add(Slot{DrawID: draw.ID, Round: 8, Position: 1, Name: winnerName, Seed: winnerSeed})

	cleanedSlots, cleanedSeeds := cleanScrapedResults(slots, seeds)

	receivedSlots := len(cleanedSlots)
	expectedSlots := (draw.Size * 2) - 1

	if receivedSlots != expectedSlots {
		errs = append(errs, fmt.Errorf("WTA - Incorrect number of scraped slots: expected %d, got %d", expectedSlots, receivedSlots))
	}

	if len(errs) > 0 {
		return cleanedSlots, cleanedSeeds, errors.Join(errs...)
	}
	return cleanedSlots, cleanedSeeds, nil
}
