package main

import (
	"fmt"

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
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	token := login()
	draws := getDraws(token)

	for _, draw := range draws {
		currentSlots := toSlotSlice(getSlots(draw.ID, token))
		var scrapedSlots slotSlice
		var seeds map[string]string

		if draw.Event == "Mens Singles" {
			scrapedSlots, seeds = scrapeATP(draw)
		} else if draw.Event == "Womens Singles" {
			scrapedSlots, seeds = scrapeWTA(draw)
		} else {
			fmt.Println("Invalid event:", draw.Event)
		}

		newSlots := getNewSlots(scrapedSlots, currentSlots)
		postSlots(newSlots, token)

		updatedSlots := prepareUpdates(scrapedSlots, currentSlots, seeds)
		updateSlots(updatedSlots, token)
	}
}
