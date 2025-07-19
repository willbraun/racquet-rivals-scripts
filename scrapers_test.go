package main

import (
	"log"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

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
		}, scrapedSlots[len(scrapedSlots)-3].Sets, "Sinner in final should have correct sets")
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
		}, scrapedSlots[len(scrapedSlots)-11].Sets, "Andreeva in quarterfinal should have correct sets")
	})
}
