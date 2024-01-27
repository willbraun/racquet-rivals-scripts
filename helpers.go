package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func printWithTimestamp(a ...any) {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprintf("%s %v", timestamp, a)
	formatted := strings.ReplaceAll(strings.ReplaceAll(message, "[", ""), "]", "")
	fmt.Println(formatted)
}

func trim(s string) string {
	return strings.Trim(s, " \n\r")
}

func hasAlphabet(input string) bool {
	hasAlphabetPattern := regexp.MustCompile("[a-zA-Z]")
	return hasAlphabetPattern.MatchString(input)
}

func getLastName(name string) string {
	nameSlice := strings.Split(name, " ")
	return nameSlice[len(nameSlice)-1]
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
