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

	// If this fails, change the draw URL to a current draw (with current in the URL)
	draw := DrawRecord{
		ID:               "abc123",
		Name:             "Australian Open",
		Event:            "Men's Singles",
		Year:             2025,
		Url:              "https://www.atptour.com/en/scores/current/australian-open/580/draws",
		Start_Date:       "2025-01-12 12:00:00.000",
		End_Date:         "2024-01-26 12:00:00.000",
		Prediction_Close: "2024-01-19 12:00:00.000",
		Size:             128,
	}

	t.Run("Scrape ATP", func(t *testing.T) {
		scrapedSlots, seeds := scrapeATP(draw)
		assert := assert.New(t)

		assert.Equal(len(scrapedSlots), 255)

		// If the draw is not complete, there will be an extra empty seed representing empty slots
		delete(seeds, "")
		assert.Equal(len(seeds), 128)
	})
}

func TestScrapeWTA(t *testing.T) {
	t.Parallel()

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file,", err)
	}

	draw := DrawRecord{
		ID:               "abc123",
		Name:             "Australian Open",
		Event:            "Women's Singles",
		Year:             2025,
		Url:              "https://www.wtatennis.com/tournament/901/australian-open/2025/draws",
		Start_Date:       "2025-01-12 12:00:00.000",
		End_Date:         "2025-01-26 12:00:00.000",
		Prediction_Close: "2025-01-19 12:00:00.000",
		Size:             128,
	}

	t.Run("Scrape WTA", func(t *testing.T) {
		scrapedSlots, seeds := scrapeWTA(draw)
		log.Println(scrapedSlots)
		assert := assert.New(t)

		assert.Equal(len(scrapedSlots), 255)
		for _, slot := range scrapedSlots {
			assert.NotEmpty(slot.Name)
		}

		delete(seeds, "")
		assert.Equal(len(seeds), 128)
	})
}
