package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func scrapeATP(draw DrawRecord) (slotSlice, map[string]string) {
	slots := slotSlice{}
	seeds := make(map[string]string)

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		printWithTimestamp("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong - ATP:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		printWithTimestamp("Response code:", r.StatusCode)
		if r.StatusCode != 200 {
			log.Println("Response code:", r.StatusCode)
		}
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		// first .draw-content element is part of the template, not the document we want to scrape
		e.DOM.Find(".draw-content").First().Remove()
		roundContainers := e.DOM.Find(".draw-content")

		round := 0
		roundContainers.Each(func(_ int, rc *goquery.Selection) {
			round++
			position := 1

			players := rc.Find(".name")
			players.Each(func(_ int, player *goquery.Selection) {
				name := trim(player.Find("a").Text())
				seed := trim(player.Find("span").Text())

				slots.add(Slot{DrawID: draw.ID, Round: round, Position: position, Name: name, Seed: seed})
				seeds[name] = seed

				position++
			})
		})

		round++
		winner := e.DOM.Find(".draw-content").Last().Find(".winner").SiblingsFiltered(".name")
		winnerName := trim(winner.Find("a").Text())
		winnerSeed := trim(winner.Find("span").Text())
		slots.add(Slot{DrawID: draw.ID, Round: round, Position: 1, Name: winnerName, Seed: winnerSeed})
	})

	c.OnScraped(func(r *colly.Response) {
		printWithTimestamp("Finished scraping ATP")
	})

	printWithTimestamp("Start scraping ATP")
	c.Visit(draw.Url)

	return slots, seeds
}

func scrapeWTA(draw DrawRecord) (slotSlice, map[string]string) {
	slots := slotSlice{}
	seeds := make(map[string]string)

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		printWithTimestamp("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong - WTA:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		printWithTimestamp("Response code:", r.StatusCode)
		if r.StatusCode != 200 {
			log.Println("Response code:", r.StatusCode)
		}
	})

	c.OnHTML(`.tournament-draw__tab[data-ui-tab="Singles"]`, func(e *colly.HTMLElement) {
		round := 0

		roundContainers := e.DOM.Find(".tournament-draw__round-container")
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

		round++
		winnerName, winnerSeed := scrapeWTAFinal(draw)
		slots.add(Slot{DrawID: draw.ID, Round: round, Position: 1, Name: winnerName, Seed: winnerSeed})
	})

	c.OnScraped(func(r *colly.Response) {
		printWithTimestamp("Finished scraping WTA")
	})

	printWithTimestamp("Start scraping WTA")
	c.Visit(draw.Url)

	return slots, seeds
}

func wtaExtractName(x *goquery.Selection) (string, string) {
	firstInitial := trim(x.Find(".match-table__player-fname").Text())
	lastName := trim(x.Find(".match-table__player-lname").Text())
	name := trim(fmt.Sprintf(`%s %s`, firstInitial, lastName))
	seed := trim(x.Find(".match-table__player-seed").Text())

	return name, seed
}

func scrapeWTAFinal(draw DrawRecord) (string, string) {
	name := ""
	seed := ""

	wtaDrawId := strings.Split(draw.Url, "/")[4]
	wtaDrawSlug := strings.Split(draw.Url, "/")[5]
	url := fmt.Sprintf(`https://www.wtatennis.com/tournament/%s/%s/%d/scores`, wtaDrawId, wtaDrawSlug, draw.Year)

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		printWithTimestamp("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		printWithTimestamp("Response code:", r.StatusCode)
		if r.StatusCode != 200 {
			log.Println("Response code:", r.StatusCode)
		}
	})

	c.OnScraped(func(r *colly.Response) {
		text := string(r.Body)
		uncommented := strings.ReplaceAll(strings.ReplaceAll(text, "<!--", ""), "-->", "")
		reader := strings.NewReader(uncommented)

		doc, err := goquery.NewDocumentFromReader(reader)
		if err != nil {
			log.Println(err)
		}

		completed := doc.Find(`.tournament-scores__tab[data-ui-tab="Singles"]`).Find(".tennis-match--completed")
		completed.Each(func(_ int, match *goquery.Selection) {
			roundLabel := trim(match.Find(".tennis-match__round").Text())
			if roundLabel == "Final" {
				name, seed = wtaExtractName(match.Find(".match-table__team--winner"))
			}
		})
	})

	c.OnScraped(func(r *colly.Response) {
		printWithTimestamp("Finished scraping WTA final")
	})

	printWithTimestamp("Start scraping WTA final")
	c.Visit(url)

	return name, seed
}
