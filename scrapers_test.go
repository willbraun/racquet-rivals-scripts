package main

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

type MockScraper struct{}

func (m *MockScraper) scrape(targetURL string) string {
	if strings.Contains(targetURL, "atptour.com") {
		html, err := readHTMLFromFile("scraped_pages/atp.html")
		if err != nil {
			log.Println("Error reading HTML from ATP file:", err)
			return ""
		}
		return html
	} else if strings.Contains(targetURL, "wtatennis.com") {
		html, err := readHTMLFromFile("scraped_pages/wta.html")
		if err != nil {
			log.Println("Error reading HTML from WTA file:", err)
			return ""
		}
		return html
	}
	log.Println("Unknown URL:", targetURL)
	return ""
}

type RealScraperSaveFile struct{}

func (s *RealScraperSaveFile) scrape(targetURL string) string {
	realScraper := &RealScraper{}
	html := realScraper.scrape(targetURL)

	if strings.Contains(targetURL, "atptour.com") {
		err := saveHTMLToFile(html, "scraped_pages/atp.html")
		if err != nil {
			log.Println("Error saving ATP HTML to file:", err)
		}
	} else if strings.Contains(targetURL, "wtatennis.com") {
		err := saveHTMLToFile(html, "scraped_pages/wta.html")
		if err != nil {
			log.Println("Error saving WTA HTML to file:", err)
		}
	}

	return html
}

func getScraper(draw DrawRecord) Scraper {
	if os.Getenv("SAVE_HTML_TO_FILE") == "atp" && strings.Contains(draw.Url, "atptour.com") {
		return &RealScraperSaveFile{}
	} else if os.Getenv("SAVE_HTML_TO_FILE") == "wta" && strings.Contains(draw.Url, "wtatennis.com") {
		return &RealScraperSaveFile{}
	}
	return &MockScraper{}
}

func TestScrapeATP(t *testing.T) {
	t.Parallel()

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file,", err)
	}

	// Use a completed draw for testing
	draw := DrawRecord{
		ID:               "test_mens_draw_id",
		Name:             "Australian Open",
		Event:            "Men's Singles",
		Year:             2025,
		Url:              "https://www.atptour.com/en/scores/current/roland-garros/520/draws",
		Start_Date:       "2025-01-12 12:00:00.000",
		End_Date:         "2025-01-26 12:00:00.000",
		Prediction_Close: "2025-01-19 12:00:00.000",
		Size:             128,
	}

	t.Run("Scrape ATP", func(t *testing.T) {
		t.Parallel()

		scrapedSlots, seeds := scrapeATP(getScraper(draw), draw)
		assert := assert.New(t)

		assert.Equal(255, len(scrapedSlots))
		assert.Equal(128, len(seeds))

		for _, slot := range scrapedSlots {
			assert.NotEmpty(slot.Name, "Slot name should not be empty")
		}

		assert.Equal(SetSlice{
			Set{
				Number:   1,
				Games:    6,
				Tiebreak: 0,
			},
			Set{
				Number:   2,
				Games:    7,
				Tiebreak: 0,
			},
			Set{
				Number:   3,
				Games:    4,
				Tiebreak: 0,
			},
			Set{
				Number:   4,
				Games:    6,
				Tiebreak: 3,
			},
			Set{
				Number:   5,
				Games:    6,
				Tiebreak: 2,
			},
		}, scrapedSlots[len(scrapedSlots)-3].Sets, "Slot R7P1 should have correct sets")
	})
}

func TestScrapeWTA(t *testing.T) {
	t.Parallel()

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file,", err)
	}

	// Use a completed draw for testing
	draw := DrawRecord{
		ID:               "test_womens_draw_id",
		Name:             "Australian Open",
		Event:            "Women's Singles",
		Year:             2025,
		Url:              "https://www.wtatennis.com/tournaments/wimbledon/draws",
		Start_Date:       "2025-01-12 12:00:00.000",
		End_Date:         "2025-01-26 12:00:00.000",
		Prediction_Close: "2025-01-19 12:00:00.000",
		Size:             128,
	}

	t.Run("Scrape WTA", func(t *testing.T) {
		scrapedSlots, seeds := scrapeWTA(getScraper(draw), draw)
		assert := assert.New(t)

		assert.Equal(255, len(scrapedSlots))
		assert.Equal(128, len(seeds))
		for _, slot := range scrapedSlots {
			assert.NotEmpty(slot.Name, "Slot name should not be empty")
		}

		assert.Equal(SetSlice{
			Set{
				Number:   1,
				Games:    6,
				Tiebreak: 3,
			},
			Set{
				Number:   2,
				Games:    6,
				Tiebreak: 2,
			},
		}, scrapedSlots[len(scrapedSlots)-11].Sets, "Slot R6P3 should have correct sets")
	})
}
