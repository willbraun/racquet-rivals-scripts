package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Scraper interface {
	scrape(targetURL string) string
}

type RealScraper struct{}

func (r *RealScraper) scrape(targetURL string) string {
	printWithTimestamp("Visiting:", targetURL)

	proxyURL, err := url.Parse(os.Getenv("PROXY_URL"))
	if err != nil {
		log.Println(fmt.Sprintf("Error parsing proxy URL - %s:", targetURL), err)
		return ""
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		// Waiting for rendering can take a while so we set a longer timeout
		Timeout: 600 * time.Second,
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return ""
	}

	// Bright Data header to wait for is-winner class to appear
	// Used for WTA draws to indicate that scores and winners have been rendered
	if strings.Contains(targetURL, "wtatennis.com") {
		req.Header.Set("x-unblock-expect", "{\"element\": \".match-table__tie-break\"}")
	}

	// Exponential backoff retry mechanism
	maxRetries := 5
	backoff := time.Second

	for i := range maxRetries {
		printWithTimestamp("Attempt:", i+1)
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			log.Println(fmt.Sprintf("Error making request - %s:", targetURL), err, "Status Code:", resp.Status, "Response Body:", resp.Body)
			if i < maxRetries-1 {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return ""
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(fmt.Sprintf("Error reading response body - %s:", targetURL), err)
			return ""
		}

		printWithTimestamp("Finished scraping:", targetURL)
		return string(body)
	}

	return ""
}

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
