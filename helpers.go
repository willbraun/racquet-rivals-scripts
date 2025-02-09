package main

import (
	"fmt"
	"os"
	"reflect"
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

func toSlotSlice(sr []SlotRecord) slotSlice {
	result := slotSlice{}
	for _, r := range sr {
		setScores := []SetScore{}

		for i := 1; i <= 5; i++ {
			gamesField := fmt.Sprintf("Set%dGames", i)
			tiebreakField := fmt.Sprintf("Set%dTiebreak", i)

			gamesValue := reflect.ValueOf(r).FieldByName(gamesField)
			tiebreakValue := reflect.ValueOf(r).FieldByName(tiebreakField)

			if gamesValue.IsValid() && !gamesValue.IsNil() {
				setScores = append(setScores, SetScore{
					Number:   i,
					Games:    *gamesValue.Interface().(*int),
					Tiebreak: *tiebreakValue.Interface().(*int),
				})
			} else {
				break
			}
		}

		result.add(Slot{
			ID:        r.ID,
			DrawID:    r.DrawID,
			Round:     r.Round,
			Position:  r.Position,
			Name:      r.Name,
			Seed:      r.Seed,
			SetScores: setScores,
		})
	}
	return result
}

func getSlotKey(s Slot) string {
	return fmt.Sprintf("%d.%d", s.Round, s.Position)
}

func getUpdates(scraped slotSlice, current slotSlice, seeds map[string]string) (slotSlice, slotSlice, []CreateUpdateSetScoreReq, []CreateUpdateSetScoreReq) {
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

	newSlots := slotSlice{}
	updatedSlots := slotSlice{}
	newSets := []CreateUpdateSetScoreReq{}
	updatedSets := []CreateUpdateSetScoreReq{}

	for key := range allKeys {
		scrapedSlot, scrapedExists := scrapedMap[key]
		currentSlot, currentExists := currentMap[key]

		if !currentExists {
			newSlots.add(scrapedSlot)
			continue
		}

		if !scrapedExists {
			continue
		}

		for j, scrapedSetScore := range scrapedSlot.SetScores {
			if j < len(currentSlot.SetScores) {
				currentSetScore := currentSlot.SetScores[j]
				if !reflect.DeepEqual(scrapedSetScore, currentSetScore) {
					updatedSets = append(updatedSets, CreateUpdateSetScoreReq{
						DrawSlotID: currentSlot.ID,
						Number: scrapedSetScore.Number,
						Games: scrapedSetScore.Games,
						Tiebreak: scrapedSetScore.Tiebreak,
					})
				}
			} else {
				newSets = append(newSets, CreateUpdateSetScoreReq{
					DrawSlotID: currentSlot.ID,
					Number: scrapedSetScore.Number,
					Games: scrapedSetScore.Games,
					Tiebreak: scrapedSetScore.Tiebreak,
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

		updatedSlot := currentSlot
		updatedSlot.Name = newName
		updatedSlot.Seed = newSeed
		updatedSlots.add(updatedSlot)
	}

	return newSlots, updatedSlots, newSets, updatedSets
}

// Example usage:
// err:= saveHTMLToFile(html, "scraped_pages/wtaRendered.html")
// if err != nil {
//   log.Println("Error saving HTML to file:", err)
// }
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
