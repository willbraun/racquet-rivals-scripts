package main

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
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

func toSlotSlice(sr []SlotRecord) slotSlice {
	result := slotSlice{}
	for _, record := range sr {
		setScores := []SetScore{}

		for i := 1; i <= 5; i++ {
			idField := fmt.Sprintf("Set%dID", i)
			gamesField := fmt.Sprintf("Set%dGames", i)
			tiebreakField := fmt.Sprintf("Set%dTiebreak", i)

			idValue := reflect.ValueOf(record).FieldByName(idField)
			gamesValue := reflect.ValueOf(record).FieldByName(gamesField)
			tiebreakValue := reflect.ValueOf(record).FieldByName(tiebreakField)

			if gamesValue.IsValid() && !gamesValue.IsNil() {
				setScores = append(setScores, SetScore{
					ID:         idValue.String(),
					DrawSlotID: record.ID,
					Number:     i,
					Games:      *gamesValue.Interface().(*int),
					Tiebreak:   *tiebreakValue.Interface().(*int),
				})
			} else {
				break
			}
		}

		result.add(Slot{
			ID:        record.ID,
			DrawID:    record.DrawID,
			Round:     record.Round,
			Position:  record.Position,
			Name:      record.Name,
			Seed:      record.Seed,
			SetScores: setScores,
		})
	}
	return result
}

func getSlotKey(s Slot) string {
	return fmt.Sprintf("%d.%d", s.Round, s.Position)
}

func getUpdates(scraped slotSlice, current slotSlice, seeds map[string]string) (slotSlice, slotSlice, []SetScore, []SetScore) {
	scrapedMap := make(map[string]Slot)
	currentMap := make(map[string]Slot)
	allKeys := make(map[string]bool)

	for _, slot := range scraped {
		key := getSlotKey(slot)
		scrapedMap[key] = slot
		allKeys[key] = true
	}

	for _, slot := range current {
		key := getSlotKey(slot)
		currentMap[key] = slot
		allKeys[key] = true
	}

	keys := []string{}
	for k := range allKeys {
		keys = append(keys, k)
	}

	// Sort keys to ensure consistent order for unit tests
	// Sorting by string is fine as the order of API calls doesn't matter
	sort.Strings(keys)

	newSlots := slotSlice{}
	updatedSlots := slotSlice{}
	newSets := []SetScore{}
	updatedSets := []SetScore{}

	for _, key := range keys {
		scrapedSlot, scrapedExists := scrapedMap[key]
		currentSlot, currentExists := currentMap[key]

		// New slot
		if !currentExists {
			newSlots.add(scrapedSlot)
			for _, setScore := range scrapedSlot.SetScores {
				newSets = append(newSets, SetScore{
					DrawSlotID: scrapedSlot.ID,
					Number:     setScore.Number,
					Games:      setScore.Games,
					Tiebreak:   setScore.Tiebreak,
				})
			}
			continue
		}

		// Existing slot isn't scraped
		if !scrapedExists {
			continue
		}

		// Update set scores
		for j, scrapedSetScore := range scrapedSlot.SetScores {
			if j < len(currentSlot.SetScores) {
				currentSetScore := currentSlot.SetScores[j]
				if !reflect.DeepEqual(scrapedSetScore, currentSetScore) {
					updatedSets = append(updatedSets, SetScore{
						ID:         currentSetScore.ID,
						DrawSlotID: currentSlot.ID,
						Number:     scrapedSetScore.Number,
						Games:      scrapedSetScore.Games,
						Tiebreak:   scrapedSetScore.Tiebreak,
					})
				}
			} else {
				// Add new set score
				newSets = append(newSets, SetScore{
					DrawSlotID: currentSlot.ID,
					Number:     scrapedSetScore.Number,
					Games:      scrapedSetScore.Games,
					Tiebreak:   scrapedSetScore.Tiebreak,
				})
			}
		}

		newName := scrapedSlot.Name
		newSeed := seeds[newName]

		// Don't clear slots with existing name
		if newName == "" {
			continue
		}

		// No update needed
		if newName == currentSlot.Name && newSeed == currentSlot.Seed {
			continue
		}

		updatedSlots.add(scrapedSlot)
	}

	return newSlots, updatedSlots, newSets, updatedSets
}

// Example usage:
// err:= saveHTMLToFile(html, "scraped_pages/wtaRendered.html")
//
//	if err != nil {
//	  log.Println("Error saving HTML to file:", err)
//	}
func saveHTMLToFile(html, filename string) error {
	return os.WriteFile(filename, []byte(html), 0644)
}

func readHTMLFromFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
