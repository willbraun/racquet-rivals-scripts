package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func scrapeWithProxy(targetURL string) string {
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
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return ""
	}

	// Bright Data header to wait for is-winner class to appear
	// Used for WTA draws to indicate that scores and winners have been rendered
	if strings.Contains(targetURL, "wtatennis.com") {
		req.Header.Set("x-unblock-expect", "{\"element\": \".is-winner\"}")
	}

	// Exponential backoff retry mechanism
	maxRetries := 5
	backoff := time.Second

	for i := 0; i < maxRetries; i++ {
		printWithTimestamp("Attempt:", i+1)
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			log.Println(fmt.Sprintf("Error making request - %s:", targetURL), err)
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

func scrapeATP(draw DrawRecord) (slotSlice, map[string]string) {
	slots := slotSlice{}
	seeds := make(map[string]string)

	// html := scrapeWithProxy(draw.Url)
	html, err := readHTMLFromFile("scraped_pages/atp.html")
	if err != nil {
		log.Println("Error reading HTML from ATP file:", err)
		return slots, seeds
	}

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

			setScores := []SetScore{}
			sets := rawSlot.Find(".score-item")
			sets.EachWithBreak(func(i int, set *goquery.Selection) bool {
				scores := set.Find("span")
				gamesStr := scores.Eq(0).Text()
				tiebreakStr := scores.Eq(1).Text()

				if gamesStr == "" {
					return false
				}

				games, err := strconv.Atoi(gamesStr)
				if err != nil {
					log.Println("Error converting games to int:", err)
				}

				tiebreak := 0
				if tiebreakStr != "" {
					tiebreak, err = strconv.Atoi(tiebreakStr)
					if err != nil {
						log.Println("Error converting tiebreak to int:", err)
					}
				}

				setScores = append(setScores, SetScore{Number: i + 1, Games: games, Tiebreak: tiebreak})

				return true
			})

			slots.add(Slot{DrawID: draw.ID, Round: round, Position: position, Name: name, Seed: seed, SetScores: setScores})
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

func scrapeWTA(draw DrawRecord) (slotSlice, map[string]string) {
	slots := slotSlice{}
	seeds := make(map[string]string)

	// html := scrapeWithProxy(draw.Url)
	html, err := readHTMLFromFile("scraped_pages/wtaRendered.html")
	if err != nil {
		log.Println("Error reading HTML from WTA file:", err)
		return slots, seeds
	}

	reader := strings.NewReader(html)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Println(err)
	}

	roundContainers := doc.Find(`.tournament-draw__tab[data-ui-tab="Singles"]`).Find(".tournament-draw__round-container")
	roundContainers.Each(func(i int, rc *goquery.Selection) {
		round := i + 1
		position := 1

		matches := rc.Find(".tournament-draw__match-table")
		matches.Each(func(_ int, match *goquery.Selection) {
			rawSlots := match.ChildrenMatcher(goquery.Single("table")).ChildrenMatcher(goquery.Single("tbody")).Children()
			rawSlots.Each(func(_ int, rawSlot *goquery.Selection) {
				name, seed := wtaExtractName(rawSlot)

				setScores := []SetScore{}
				sets := rawSlot.Find(".match-table__score-cell")
				sets.EachWithBreak(func(i int, set *goquery.Selection) bool {
					fields := strings.Fields(set.Text())

					if fields[0] == "-" {
						return false
					}

					games, err := strconv.Atoi(fields[0])
					if err != nil {
						log.Println("Error converting games to int:", err)
					}

					tiebreak := 0
					if len(fields) > 1 {
						tiebreak, err = strconv.Atoi(fields[1])
						if err != nil {
							log.Println("Error converting tiebreak to int:", err)
						}
					}

					setScores = append(setScores, SetScore{Number: i + 1, Games: games, Tiebreak: tiebreak})

					return true
				})

				slots.add(Slot{DrawID: draw.ID, Round: round, Position: position, Name: name, Seed: seed, SetScores: setScores})
				seeds[name] = seed

				position++
			})
		})

		if round == roundContainers.Length() {
			winner := rc.Find(".match-table__team.is-winner")
			winnerName, winnerSeed := wtaExtractName(winner)

			round++
			slots.add(Slot{DrawID: draw.ID, Round: round, Position: 1, Name: winnerName, Seed: winnerSeed})
		}
	})

	return slots, seeds
}

func wtaExtractName(x *goquery.Selection) (string, string) {
	firstNameRaw := trim(x.Find(".match-table__player-fname").Text())
	firstName := strings.ReplaceAll(firstNameRaw, ".", "")
	lastName := trim(x.Find(".match-table__player-lname").Text())
	name := trim(firstName + " " + lastName)

	if !hasAlphabet(name) {
		return "", ""
	}

	seed := trim(x.Find(".match-table__player-seed").Text())

	return name, seed
}
