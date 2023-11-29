package main

import (
	"fmt"
	"strconv"

	// "time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

// type WTAMatchData struct {
// 	BinPacketBase64   string    `json:"BinPacketBase64"`
// 	CourtID           int       `json:"CourtID"`
// 	DateSeq           int       `json:"DateSeq"`
// 	DrawLevelType     string    `json:"DrawLevelType"`
// 	DrawMatchType     string    `json:"DrawMatchType"`
// 	EntryTypeA        string    `json:"EntryTypeA"`
// 	EntryTypeB        string    `json:"EntryTypeB"`
// 	EventID           string    `json:"EventID"`
// 	EventYear         int       `json:"EventYear"`
// 	LastUpdated       time.Time `json:"LastUpdated"`
// 	MatchID           string    `json:"MatchID"`
// 	MatchState        string    `json:"MatchState"`
// 	MatchTimeStamp    time.Time `json:"MatchTimeStamp"`
// 	MatchTimeTotal    string    `json:"MatchTimeTotal"`
// 	Message           string    `json:"Message"`
// 	NumSets           int       `json:"NumSets"`
// 	PlayerCountryA    string    `json:"PlayerCountryA"`
// 	PlayerCountryA2   string    `json:"PlayerCountryA2"`
// 	PlayerCountryB    string    `json:"PlayerCountryB"`
// 	PlayerCountryB2   string    `json:"PlayerCountryB2"`
// 	PlayerIDA         string    `json:"PlayerIDA"`
// 	PlayerIDA2        string    `json:"PlayerIDA2"`
// 	PlayerIDB         string    `json:"PlayerIDB"`
// 	PlayerIDB2        string    `json:"PlayerIDB2"`
// 	PlayerNameFirstA  string    `json:"PlayerNameFirstA"`
// 	PlayerNameFirstA2 string    `json:"PlayerNameFirstA2"`
// 	PlayerNameFirstB  string    `json:"PlayerNameFirstB"`
// 	PlayerNameFirstB2 string    `json:"PlayerNameFirstB2"`
// 	PlayerNameLastA   string    `json:"PlayerNameLastA"`
// 	PlayerNameLastA2  string    `json:"PlayerNameLastA2"`
// 	PlayerNameLastB   string    `json:"PlayerNameLastB"`
// 	PlayerNameLastB2  string    `json:"PlayerNameLastB2"`
// 	PointA            string    `json:"PointA"`
// 	PointB            string    `json:"PointB"`
// 	ResultString      string    `json:"ResultString"`
// 	RoundID           string    `json:"RoundID"`
// 	ScoreSet1A        string    `json:"ScoreSet1A"`
// 	ScoreSet1B        string    `json:"ScoreSet1B"`
// 	ScoreSet2A        string    `json:"ScoreSet2A"`
// 	ScoreSet2B        string    `json:"ScoreSet2B"`
// 	ScoreSet3A        string    `json:"ScoreSet3A"`
// 	ScoreSet3B        string    `json:"ScoreSet3B"`
// 	ScoreSet4A        string    `json:"ScoreSet4A"`
// 	ScoreSet4B        string    `json:"ScoreSet4B"`
// 	ScoreSet5A        string    `json:"ScoreSet5A"`
// 	ScoreSet5B        string    `json:"ScoreSet5B"`
// 	ScoreString       string    `json:"ScoreString"`
// 	ScoreSys          string    `json:"ScoreSys"`
// 	ScoreTbSet1       string    `json:"ScoreTbSet1"`
// 	ScoreTbSet2       string    `json:"ScoreTbSet2"`
// 	ScoreTbSet3       string    `json:"ScoreTbSet3"`
// 	ScoreTbSet4       string    `json:"ScoreTbSet4"`
// 	SeedA             string    `json:"SeedA"`
// 	SeedB             string    `json:"SeedB"`
// 	Serve             string    `json:"Serve"`
// 	Winner            string    `json:"Winner"`
// }

// type WTATournamentData struct {
// 	TournamentGroup   WTATournamentGroupData `json:"tournamentGroup"`
// 	Year              int             `json:"year"`
// 	Title             string          `json:"title"`
// 	StartDate         string          `json:"startDate"`
// 	EndDate           string          `json:"endDate"`
// 	Surface           string          `json:"surface"`
// 	InOutdoor         string          `json:"inOutdoor"`
// 	City              string          `json:"city"`
// 	Country           string          `json:"country"`
// 	SinglesDrawSize   int             `json:"singlesDrawSize"`
// 	DoublesDrawSize   int             `json:"doublesDrawSize"`
// 	PrizeMoney        int             `json:"prizeMoney"`
// 	PrizeMoneyCurrency string          `json:"prizeMoneyCurrency"`
// 	LiveScoringId     string          `json:"liveScoringId"`
// }

// type WTATournamentGroupData struct {
// 	ID       int    `json:"id"`
// 	Name     string `json:"name"`
// 	Level    string `json:"level"`
// 	Metadata interface{} `json:"metadata"`
// }

// type WTADataResponse struct {
// 	Matches    []WTAMatchData `json:"matches"`
// 	Tournament WTATournamentData     `json:"tournament"`
// }

func scrapeATP(draw DrawRecord) (slotSlice, map[string]string) {
	slots := slotSlice{}
	seeds := make(map[string]string)

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

	// first round
	c.OnHTML(".scores-draw-entry-box-table", func(e *colly.HTMLElement) {
		rows := e.DOM.ChildrenMatcher(goquery.Single("tbody")).Children()
		rows.Each(func(_ int, row *goquery.Selection) {
			values := row.Children().Map(func(_ int, s *goquery.Selection) string {
				return trim(s.Text())
			})
			positionStr, seed, name := values[0], values[1], values[2]

			position, err := strconv.Atoi(positionStr)
			if err != nil {
				fmt.Println(err)
			}

			slots.add(Slot{DrawID: draw.ID, Round: 1, Position: position, Name: name, Seed: seed})

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

		slots.add(Slot{DrawID: draw.ID, Round: currentRound, Position: position, Name: name, Seed: seed})

		currentRound++
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping")
	})

	fmt.Println("Start scraping")
	c.Visit(draw.Url)

	return slots, seeds
}

func scrapeWTA(draw DrawRecord) (slotSlice, map[string]string) {
	slots := slotSlice{}
	seeds := make(map[string]string)

	// wtaDrawId := strings.Split(draw.Url, "/")[4]
	// fmt.Println(wtaDrawId)

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
					name, seed := wtaExtractRow(row)

					slots.add(Slot{DrawID: draw.ID, Round: round, Position: position, Name: name, Seed: seed})

					seeds[name] = seed
					position++
				})
			})
		})

		// url := fmt.Sprintf(`https://api.wtatennis.com/tennis/tournaments/%s/2023/matches/?states=%s`, wtaDrawId, "L%2C%20C")
		// fmt.Println(url)
		// res, err := makeHTTPRequest("GET", url, "", nil)
		// defer res.Body.Close()

		// wtaDataResponse := &WTADataResponse{}
		// fmt.Println("Auth request status:", res.Status)
		// derr := json.NewDecoder(res.Body).Decode(wtaDataResponse)
		// if derr != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		// matches := wtaDataResponse.Matches
		// winner := strings.Split(matches[len(matches)-1].ResultString, " d ")[0]
		// fmt.Println(winner)

		champion := roundContainers.Last().Find(".is-winner").Find(".match-table__player-name")
		name, seed := wtaExtractRow(champion)
		round++

		slots.add(Slot{DrawID: draw.ID, Round: round, Position: 1, Name: name, Seed: seed})
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping")
	})

	fmt.Println("Start scraping")
	c.Visit(draw.Url)

	return slots, seeds
}

func wtaExtractRow(r *goquery.Selection) (string, string) {
	firstInitial := trim(r.Find(".match-table__player-fname").Text())
	lastName := trim(r.Find(".match-table__player-lname").Text())
	name := trim(fmt.Sprintf(`%s %s`, firstInitial, lastName))
	seed := trim(r.Find(".match-table__player-seed").Text())

	return name, seed
}
