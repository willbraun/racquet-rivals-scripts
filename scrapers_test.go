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
	// Current draw will have some empty slots
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
		scrapedSlots, seeds := scrapeATP(draw)
		assert := assert.New(t)

		uniqueNames := make(map[string]bool)

		assert.Equal(255, len(scrapedSlots))
		for _, slot := range scrapedSlots {
			uniqueNames[slot.Name] = true
		}

		// If the draw is not complete, there will be empty slots and an extra empty seed representing those
		delete(seeds, "")
		delete(uniqueNames, "")
		assert.Equal(len(uniqueNames), len(seeds))
	})
}

func TestScrapeWTA(t *testing.T) {
	t.Parallel()

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file,", err)
	}

	draw := DrawRecord{
		ID:               "test_womens_draw_id",
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
		assert := assert.New(t)

		uniqueNames := make(map[string]bool)

		assert.Equal(255, len(scrapedSlots))
		for _, slot := range scrapedSlots {
			assert.NotEmpty(slot.Name)
			uniqueNames[slot.Name] = true
		}

		delete(seeds, "")
		assert.Equal(len(uniqueNames), len(seeds))
	})
}
