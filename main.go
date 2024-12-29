package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type SlotSet struct {
	Number   int
	Games    int
	Tiebreak int
}

type Slot struct {
	ID       string
	DrawID   string
	Round    int
	Position int
	Name     string
	Seed     string
	Winner   bool
	Sets     []SlotSet
}

type slotSlice []Slot

type Set struct {
	ID         string
	MatchID    string
	Number     int
	P1Games    int
	P2Games    int
	P1Tiebreak int
	P2Tiebreak int
}

type Match struct {
	ID           string
	WinnerSlotID string
	Status       string
	Sets         []Set
}

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

	for _, draw := range draws {
		currentSlots := toSlotSlice(getSlots(draw.ID, token))
		var scrapedSlots slotSlice
		var seeds map[string]string

		if draw.Event == "Men's Singles" {
			scrapedSlots, seeds = scrapeATP(draw)
		} else if draw.Event == "Women's Singles" {
			scrapedSlots, seeds = scrapeWTA(draw)
		} else {
			log.Println("Invalid event:", draw.Event)
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

		newSlots := getNewSlots(scrapedSlots, currentSlots)
		postSlots(newSlots, token)

		updatedSlots := prepareUpdates(scrapedSlots, currentSlots, seeds)
		updateSlots(updatedSlots, token)
	}
}
