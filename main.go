package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Slot struct {
	ID       string
	DrawID   string
	Round    int
	Position int
	Name     string
	Seed     string
}

type slotSlice []Slot

func main() {
	location, err := time.LoadLocation("UTC")
	if err != nil {
		log.Fatal(err)
	}
	time.Local = location

	log.SetOutput(os.Stderr)

	currentDir, _ := os.Getwd()
	envErr := godotenv.Load(currentDir + "/.env")
	if envErr != nil {
		log.Println("Error loading .env file,", envErr)
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

		if draw.Event == "Mens Singles" {
			scrapedSlots, seeds = scrapeATP(draw)
		} else if draw.Event == "Womens Singles" {
			scrapedSlots, seeds = scrapeWTA(draw)
		} else {
			log.Println("Invalid event:", draw.Event)
		}

		newSlots := getNewSlots(scrapedSlots, currentSlots)
		postSlots(newSlots, token)

		updatedSlots := prepareUpdates(scrapedSlots, currentSlots, seeds)
		updateSlots(updatedSlots, token)
	}
}
