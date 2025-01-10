package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

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

	resp, err := client.Get(targetURL)
	if err != nil {
		log.Println(fmt.Sprintf("Error making request - %s:", targetURL), err)
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

// func scrapeATP(draw DrawRecord) (slotSlice, map[string]string) {
// 	slots := slotSlice{}
// 	seeds := make(map[string]string)

// 	html := scrapeWithProxy(draw.Url)
// 	reader := strings.NewReader(html)

// 	doc, err := goquery.NewDocumentFromReader(reader)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	roundContainers := doc.Find(".draw-content").FilterFunction(func(_ int, selection *goquery.Selection) bool {
// 		return !selection.Parents().Is("template")
// 	})

// 	round := 0
// 	roundContainers.Each(func(_ int, rc *goquery.Selection) {
// 		round++
// 		position := 1

// 		players := rc.Find(".name")
// 		players.Each(func(_ int, player *goquery.Selection) {
// 			name := trim(player.Find("a").Text())
// 			seed := trim(player.Find("span").Text())

// 			slots.add(Slot{DrawID: draw.ID, Round: round, Position: position, Name: name, Seed: seed})
// 			seeds[name] = seed

// 			position++
// 		})
// 	})

// 	round++
// 	winner := doc.Find(".draw-content").Last().Find(".winner").SiblingsFiltered(".name")
// 	winnerName := trim(winner.Find("a").Text())
// 	winnerSeed := trim(winner.Find("span").Text())
// 	slots.add(Slot{DrawID: draw.ID, Round: round, Position: 1, Name: winnerName, Seed: winnerSeed})

// 	return slots, seeds
// }

func scrapeWTA(draw DrawRecord) (slotSlice, map[string]string) {
	slots := slotSlice{}
	seeds := make(map[string]string)

	html := scrapeWithProxy(draw.Url)
	reader := strings.NewReader(html)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Println(err)
	}

	round := 0

	roundContainers := doc.Find(`.tournament-draw__tab[data-ui-tab="Singles"]`).Find(".tournament-draw__round-container")
	roundContainers.Each(func(_ int, rc *goquery.Selection) {
		round++
		position := 1

		matches := rc.Find(".tournament-draw__match-table")
		matches.Each(func(_ int, match *goquery.Selection) {
			rows := match.ChildrenMatcher(goquery.Single("table")).ChildrenMatcher(goquery.Single("tbody")).Children()
			rows.Each(func(_ int, row *goquery.Selection) {
				name, seed := wtaExtractName(row)

				slots.add(Slot{DrawID: draw.ID, Round: round, Position: position, Name: name, Seed: seed})
				seeds[name] = seed

				position++
			})
		})
	})

	winnerName := ""
	winnerSeed := ""
	if len(slots) > 0 && slots[len(slots)-1].Name != "" {
		winnerName = scrapeWTAFinal(draw)

		for _, slot := range slots {
			if slot.Round == round && getLastName(slot.Name) == getLastName(winnerName) {
				winnerName = slot.Name
				winnerSeed = slot.Seed
				break
			}
		}
	}

	round++
	slots.add(Slot{DrawID: draw.ID, Round: round, Position: 1, Name: winnerName, Seed: winnerSeed})

	return slots, seeds
}

func scrapeWTAFinal(draw DrawRecord) string {
	name := ""

	wtaDrawId := strings.Split(draw.Url, "/")[4]
	wtaDrawSlug := strings.Split(draw.Url, "/")[5]
	url := fmt.Sprintf(`https://www.wtatennis.com/tournament/%s/%s/%d/scores`, wtaDrawId, wtaDrawSlug, draw.Year)

	html := scrapeWithProxy(url)
	uncommented := regexp.MustCompile(`<!--|-->`).ReplaceAllString(html, "")
	reader := strings.NewReader(uncommented)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Println(err)
	}

	completed := doc.Find(`.tournament-scores__tab[data-ui-tab="Singles"]`).Find(".tennis-match--completed")
	completed.Each(func(_ int, match *goquery.Selection) {
		roundLabel := trim(match.Find(".tennis-match__round").Text())
		if roundLabel == "Final" {
			name, _ = wtaExtractName(match.Find(".match-table__team--winner"))
		}
	})

	return name
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
