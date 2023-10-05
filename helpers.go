package main

import (
	"fmt"
	"strings"
)

func trim(s string) string {
	return strings.Trim(s, " \n\r")
}

func (ss *slotSlice) add(s Slot) {
	*ss = append(*ss, s)
}

func toSlotSlice(s []SlotRecord) slotSlice {
	result := slotSlice{}
	for _, v := range s {
		result.add(Slot{
			ID:       v.ID,
			DrawID:   v.DrawID,
			Position: v.Position,
			Round:    v.Round,
			Name:     v.Name,
			Seed:     v.Seed,
		})
	}
	return result
}

func getSlotKey(s Slot) string {
	return fmt.Sprintf("%d.%d", s.Round, s.Position)
}

func getNewSlots(scraped slotSlice, current slotSlice) slotSlice {
	currentMap := make(map[string]bool)
	for _, slot := range current {
		key := getSlotKey(slot)
		currentMap[key] = true
	}

	result := slotSlice{}
	for _, slot := range scraped {
		key := getSlotKey(slot)
		if !currentMap[key] {
			result.add(slot)
		}
	}

	return result
}

func prepareUpdates(scraped slotSlice, current slotSlice, seeds map[string]string) slotSlice {
	scrapedMap := make(map[string]string)
	for _, slot := range scraped {
		key := getSlotKey(slot)
		scrapedMap[key] = slot.Name
	}

	result := slotSlice{}
	for _, slot := range current {
		key := getSlotKey(slot)
		newName := scrapedMap[key]
		newSeed := seeds[newName]
		if newName != slot.Name || newSeed != slot.Seed {
			slot.Name = newName
			slot.Seed = newSeed
			result.add(slot)
		}
	}

	return result
}
