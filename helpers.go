package main

import (
	"fmt"
	"log"
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

func toSlotSlice(sr []SlotRecord) SlotSlice {
	result := SlotSlice{}
	for _, record := range sr {
		sets := SetSlice{}

		for i := 1; i <= 5; i++ {
			idField := fmt.Sprintf("Set%dID", i)
			gamesField := fmt.Sprintf("Set%dGames", i)
			tiebreakField := fmt.Sprintf("Set%dTiebreak", i)

			idValue := reflect.ValueOf(record).FieldByName(idField)
			gamesValue := reflect.ValueOf(record).FieldByName(gamesField)
			tiebreakValue := reflect.ValueOf(record).FieldByName(tiebreakField)

			if gamesValue.IsValid() && !gamesValue.IsNil() {
				sets.add(Set{
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
			ID:       record.ID,
			DrawID:   record.DrawID,
			Round:    record.Round,
			Position: record.Position,
			Name:     record.Name,
			Seed:     record.Seed,
			Sets:     sets,
		})
	}
	return result
}

func getUpdates(scraped SlotSlice, current SlotSlice, seeds map[string]string) (SlotSlice, SlotSlice, SetSlice, SetSlice) {
	scrapedMap := make(map[SlotKey]Slot)
	currentMap := make(map[SlotKey]Slot)
	allKeys := make(map[SlotKey]bool)

	for _, slot := range scraped {
		key := SlotKey{
			Round:    slot.Round,
			Position: slot.Position,
		}
		scrapedMap[key] = slot
		allKeys[key] = true
	}

	for _, slot := range current {
		key := SlotKey{
			Round:    slot.Round,
			Position: slot.Position,
		}
		currentMap[key] = slot
		allKeys[key] = true
	}

	keys := []SlotKey{}
	for k := range allKeys {
		keys = append(keys, k)
	}

	// Sort keys to ensure consistent order for testing/debugging
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Round == keys[j].Round {
			return keys[i].Position < keys[j].Position
		}
		return keys[i].Round < keys[j].Round
	})

	newSlots := SlotSlice{}
	updatedSlots := SlotSlice{}
	newSets := SetSlice{}
	updatedSets := SetSlice{}

	for _, key := range keys {
		scrapedSlot, scrapedExists := scrapedMap[key]
		currentSlot, currentExists := currentMap[key]

		// New slot
		if !currentExists {
			newSlots.add(scrapedSlot)
			for _, set := range scrapedSlot.Sets {
				newSets.add(Set{
					DrawSlotID: scrapedSlot.ID,
					Number:     set.Number,
					Games:      set.Games,
					Tiebreak:   set.Tiebreak,
				})
			}
			continue
		}

		// Existing slot isn't scraped
		if !scrapedExists {
			continue
		}

		// Update set scores
		// Scraped slots and sets don't have IDs so we can't compare them directly
		for j, scrapedSet := range scrapedSlot.Sets {
			if j < len(currentSlot.Sets) {
				currentSet := currentSlot.Sets[j]

				// Current and scraped sets are in order on each slot so set numbers should match
				if currentSet.Number != scrapedSet.Number {
					log.Println("Set numbers don't match for Set with ID:", currentSet.ID, "Current:", currentSet.Number, "Scraped:", scrapedSet.Number)
					continue
				}

				if currentSet.Games != scrapedSet.Games || currentSet.Tiebreak != scrapedSet.Tiebreak {
					updatedSets.add(Set{
						ID:         currentSet.ID,
						DrawSlotID: currentSlot.ID,
						Number:     scrapedSet.Number,
						Games:      scrapedSet.Games,
						Tiebreak:   scrapedSet.Tiebreak,
					})
				}
			} else {
				// Add new set score
				newSets.add(Set{
					DrawSlotID: currentSlot.ID,
					Number:     scrapedSet.Number,
					Games:      scrapedSet.Games,
					Tiebreak:   scrapedSet.Tiebreak,
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

		updatedSlot := Slot{
			ID:       currentSlot.ID,
			DrawID:   currentSlot.DrawID,
			Round:    currentSlot.Round,
			Position: currentSlot.Position,
			Name:     newName,
			Seed:     newSeed,
			Sets:     scrapedSlot.Sets,
		}

		updatedSlots.add(updatedSlot)
	}

	return newSlots, updatedSlots, newSets, updatedSets
}

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

type MockScraper struct{}

func (m *MockScraper) scrape(targetURL string) string {
	if strings.Contains(targetURL, "atptour.com") {
		html, err := readHTMLFromFile("scraped_pages/atp.html")
		if err != nil {
			log.Println("Error reading HTML from ATP file:", err)
			return ""
		}
		return html
	} else if strings.Contains(targetURL, "wtatennis.com") {
		html, err := readHTMLFromFile("scraped_pages/wta.html")
		if err != nil {
			log.Println("Error reading HTML from WTA file:", err)
			return ""
		}
		return html
	}
	log.Println("Unknown URL:", targetURL)
	return ""
}

type RealScraperSaveFile struct{}

func (s *RealScraperSaveFile) scrape(targetURL string) string {
	realScraper := &RealScraper{}
	html := realScraper.scrape(targetURL)

	if strings.Contains(targetURL, "atptour.com") {
		err := saveHTMLToFile(html, "scraped_pages/atp.html")
		if err != nil {
			log.Println("Error saving ATP HTML to file:", err)
		}
	} else if strings.Contains(targetURL, "wtatennis.com") {
		err := saveHTMLToFile(html, "scraped_pages/wta.html")
		if err != nil {
			log.Println("Error saving WTA HTML to file:", err)
		}
	}

	return html
}

func getScraper(draw DrawRecord) Scraper {
	if os.Getenv("SAVE_HTML_TO_FILE") == "atp" && strings.Contains(draw.Url, "atptour.com") {
		return &RealScraperSaveFile{}
	} else if os.Getenv("SAVE_HTML_TO_FILE") == "wta" && strings.Contains(draw.Url, "wtatennis.com") {
		return &RealScraperSaveFile{}
	}
	return &MockScraper{}
}
