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

	draw := DrawRecord{
		ID:               "abc123",
		Name:             "Brisbane International",
		Event:            "Men's Singles",
		Year:             2024,
		Url:              "https://www.atptour.com/en/scores/archive/brisbane/339/2024/draws",
		Start_Date:       "2023-12-31 12:00:00.000",
		End_Date:         "2024-01-07 12:00:00.000",
		Prediction_Close: "2024-01-03 12:00:00.000",
		Size:             32,
	}

	t.Run("Scrape ATP", func(t *testing.T) {
		scrapedSlots, seeds := scrapeATP(draw)
		assert := assert.New(t)

		assert.Equal(len(scrapedSlots), 63)
		for _, slot := range scrapedSlots {
			assert.NotEmpty(slot.Name)
		}
		assert.Equal(len(seeds), 32)
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
		Name:             "Brisbane International",
		Event:            "Women's Singles",
		Year:             2024,
		Url:              "https://www.wtatennis.com/tournament/800/brisbane/2024/draws",
		Start_Date:       "2023-12-31 12:00:00.000",
		End_Date:         "2024-01-07 12:00:00.000",
		Prediction_Close: "2024-01-03 12:00:00.000",
		Size:             64,
	}

	t.Run("Scrape WTA", func(t *testing.T) {
		scrapedSlots, seeds := scrapeWTA(draw)
		assert := assert.New(t)

		assert.Equal(len(scrapedSlots), 127)
		for _, slot := range scrapedSlots {
			assert.NotEmpty(slot.Name)
		}
		assert.Equal(len(seeds), 49) // only 48 players, plus the BYE in this draw
	})
}
