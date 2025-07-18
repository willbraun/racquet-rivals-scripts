package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	location, err := time.LoadLocation("UTC")
	if err != nil {
		log.Fatal(err)
	}
	time.Local = location

	log.SetOutput(os.Stderr)

	// set environment variables for local testing
	// on remote (https), environment variables are set before running script
	if !strings.Contains(os.Getenv("BASE_URL"), "https://") {
		envErr := godotenv.Load()
		if envErr != nil {
			log.Println("Error loading .env file,", envErr)
		}
	}

	token := login()
	draws := getDraws(token)

	if len(draws) == 0 {
		printWithTimestamp("No active draws")
		return
	}

	scraper := &RealScraper{}

	for _, draw := range draws {
		currentSlots := getSlots(draw.ID, token)
		var scrapedSlots SlotSlice
		var seeds map[string]string

		switch draw.Event {
		case "Men's Singles":
			scrapedSlots, seeds = scrapeATP(scraper, draw)
		case "Women's Singles":
			scrapedSlots, seeds = scrapeWTA(scraper, draw)
		default:
			log.Println("Invalid event:", draw.Event)
			continue
		}

		received := len(scrapedSlots)
		expected := (draw.Size * 2) - 1

		if received != expected {
			log.Printf("Incorrect number of scraped slots for %s %s %d. Expected: %d, received: %d.",
				draw.Name,
				draw.Event,
				draw.Year,
				expected,
				received)
			continue
		}

		newSlots, updatedSlots, newSets, updatedSets := getUpdates(scrapedSlots, currentSlots, seeds)

		postSlots(newSlots, token)
		updateSlots(updatedSlots, token)
		postSets(newSets, token)
		updateSets(updatedSets, token)
	}
}
