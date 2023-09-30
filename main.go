package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

type Slot struct {
	Draw_id     string
	Round       int
	Position    int
	Player_name string
	Seed        string
}

type slotSlice []Slot

type UserRecord struct {
	ID              string `json:"id"`
	CollectionID    string `json:"collectionId"`
	CollectionName  string `json:"collectionName"`
	Created         string `json:"created"`
	Updated         string `json:"updated"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	Verified        bool   `json:"verified"`
	EmailVisibility bool   `json:"emailVisibility"`
	SomeCustomField string `json:"someCustomField"`
}

type UserAuthRes struct {
	Token  string     `json:"token"`
	Record UserRecord `json:"record"`
}

type DrawRecord struct {
	ID               string `json:"id"`
	CollectionID     string `json:"collectionId"`
	CollectionName   string `json:"collectionName"`
	Name             string `json:"name"`
	Event            string `json:"event"`
	Year             int    `json:"year"`
	Url              string `json:"url"`
	Start_Date       string `json:"start_date"`
	End_Date         string `json:"end_date"`
	Prediction_Close string `json:"prediction_close"`
	Updated          string `json:"updated"`
	Created          string `json:"created"`
}

type DrawRes struct {
	Page       int          `json:"page"`
	PerPage    int          `json:"perPage"`
	TotalItems int          `json:"totalItems"`
	TotalPages int          `json:"totalPages"`
	Items      []DrawRecord `json:"items"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// fmt.Println(os.Getenv("SCRIPT_USER_USERNAME"))

	fmt.Println("Started")
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

				slots.add(Slot{Draw_id: "1", Round: 1, Position: position, Player_name: name, Seed: seed})

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

			slots.add(Slot{Draw_id: "1", Round: round, Position: position, Player_name: name, Seed: seed})

			currentRound++
		}
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished scraping", r.Request.URL)
	})

	// c.Visit("https://www.atptour.com/en/scores/current/beijing/747/draws")

	// fmt.Println(slots)

	// get current slots from pb, script_user
	authUrl := "https://tennisbracket.willbraun.dev/api/collections/user/auth-with-password"
	username := os.Getenv("SCRIPT_USER_USERNAME")
	password := os.Getenv("SCRIPT_USER_PASSWORD")

	authRequestData := struct {
		Identity string `json:"identity"`
		Password string `json:"password"`
	}{
		Identity: username,
		Password: password,
	}

	authJSON, err := json.Marshal(authRequestData)
	perror(err)

	body := []byte(authJSON)

	userRequest, err := http.NewRequest("POST", authUrl, bytes.NewBuffer(body))
	perror(err)

	userRequest.Header.Add("Content-Type", "application/json")

	authClient := &http.Client{Timeout: 10 * time.Second}
	ures, err := authClient.Do(userRequest)
	perror(err)

	defer ures.Body.Close()

	userAuthRes := &UserAuthRes{}
	fmt.Println("auth request status:", ures.Status)
	uderr := json.NewDecoder(ures.Body).Decode(userAuthRes)
	perror(uderr)

	token := userAuthRes.Token

	getDrawsUrl := "https://tennisbracket.willbraun.dev/api/collections/draw/records"
	drawRequest, err := http.NewRequest("GET", getDrawsUrl, nil)
	drawRequest.Header.Add("Authorization", token)
	drawClient := &http.Client{Timeout: 10 * time.Second}
	dres, err := drawClient.Do(drawRequest)
	perror(err)

	defer dres.Body.Close()

	drawRes := &DrawRes{}
	dderr := json.NewDecoder(dres.Body).Decode(drawRes)
	perror(dderr)

	fmt.Println(drawRes.Items)

	// filter struct to remove positions already in pb
	// loop over remaining struct to upload to pb
}

func trim(s string) string {
	return strings.Trim(s, " \n\r")
}

func (ss *slotSlice) add(s Slot) {
	*ss = append(*ss, s)
}

func perror(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
