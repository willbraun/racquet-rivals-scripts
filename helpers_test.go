package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var dummySetScores = []SetScore{
	{Number: 1, Games: 6, Tiebreak: 0},
	{Number: 2, Games: 6, Tiebreak: 0},
}

var allFilled = slotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", SetScores: dummySetScores},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
}

var allFilledPartialSets = slotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: []SetScore{
		{Number: 1, Games: 4, Tiebreak: 0},
	}},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", SetScores: []SetScore{}},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: []SetScore{}},
}

var twoFilled = slotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", SetScores: dummySetScores},
}

var twoFilledOneWithName = slotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", SetScores: dummySetScores},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: []SetScore{}},
}

var twoFilledOneBlank = slotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", SetScores: dummySetScores},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "", Seed: "", SetScores: []SetScore{}},
}

var allBlank = slotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "", Seed: "", SetScores: []SetScore{}},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "", Seed: "", SetScores: []SetScore{}},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "", Seed: "", SetScores: []SetScore{}},
}

var seeds = map[string]string{
	"Roger Federer": "(1)",
	"Rafael Nadal":  "(2)",
}

func TestGetUpdates(t *testing.T) {
	t.Parallel()

	t.Run("Add a slot", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, twoFilled, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
		})
		assert.Equal(updatedSlots, slotSlice{})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{
			{DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{})
	})

	t.Run("Add all slots", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, slotSlice{}, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", SetScores: dummySetScores},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
		})
		assert.Equal(updatedSlots, slotSlice{})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{
			{DrawSlotID: "aaa", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "aaa", Number: 2, Games: 6, Tiebreak: 0},
			{DrawSlotID: "bbb", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "bbb", Number: 2, Games: 6, Tiebreak: 0},
			{DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{})
	})


	t.Run("Update slot name", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(twoFilledOneWithName, twoFilledOneBlank, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{})
		assert.Equal(updatedSlots, slotSlice{
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: []SetScore{}},
		})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{})
	})

	t.Run("Add slot score", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, twoFilledOneWithName, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{})
		assert.Equal(updatedSlots, slotSlice{})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{
			{DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{})
	})

	t.Run("Update slot name and add score", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, twoFilledOneBlank, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{})
		assert.Equal(updatedSlots, slotSlice{
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
		})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{
			{DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{})
	})

	t.Run("Update and add score", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, allFilledPartialSets, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{})
		assert.Equal(updatedSlots, slotSlice{})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{
			{DrawSlotID: "aaa", Number: 2, Games: 6, Tiebreak: 0},
			{DrawSlotID: "bbb", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "bbb", Number: 2, Games: 6, Tiebreak: 0},
			{DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{
			{DrawSlotID: "aaa", Number: 1, Games: 6, Tiebreak: 0},
		})
	})

	t.Run("Update all slots", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, allBlank, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{})
		assert.Equal(updatedSlots, slotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", SetScores: dummySetScores},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
		})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{
			{DrawSlotID: "aaa", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "aaa", Number: 2, Games: 6, Tiebreak: 0},
			{DrawSlotID: "bbb", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "bbb", Number: 2, Games: 6, Tiebreak: 0},
			{DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{})
	})

	t.Run("Scraped all blanks, do not clear", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allBlank, allFilled, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{})
		assert.Equal(updatedSlots, slotSlice{})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{})
	})

	t.Run("Scraped one blank, do not clear", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(twoFilledOneBlank, allFilled, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{})
		assert.Equal(updatedSlots, slotSlice{})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{})
	})

	t.Run("One updated, manual entry ignored", func(t *testing.T) {
		// round 1 slot 2 is scraped successfully and updates a blank slot
		// round 2 slot 1 is scraped as a blank due to an error with the website
		scraped := slotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", SetScores: dummySetScores},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "", Seed: "", SetScores: []SetScore{}},
		}

		// round 1 slot 2 hasn't been filled yet
		// round 2 slot 1 is entered manually
		current := slotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "", Seed: "", SetScores: []SetScore{}},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
		}

		// only round 1 slot 2 should be updated
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(scraped, current, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{})
		assert.Equal(updatedSlots, slotSlice{
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", SetScores: dummySetScores},
		})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{
			{DrawSlotID: "bbb", Number: 1, Games: 6, Tiebreak: 0},
			{DrawSlotID: "bbb", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{})
	})

	t.Run("No changes", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, allFilled, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{})
		assert.Equal(updatedSlots, slotSlice{})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{})
	})

	t.Run("Empty scrape", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(slotSlice{}, allFilled, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, slotSlice{})
		assert.Equal(updatedSlots, slotSlice{})
		assert.Equal(newSets, []CreateUpdateSetScoreReq{})
		assert.Equal(updatedSets, []CreateUpdateSetScoreReq{})
	})
}

func TestToSlotSlice(t *testing.T) {
	t.Parallel()

	six := 6
	sixPtr := &six
	four := 4
	fourPtr := &four
	zero := 0
	zeroPtr := &zero

	slotRecords := []SlotRecord{
		{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Set1Games: sixPtr, Set1Tiebreak: fourPtr, Set2Games: fourPtr, Set2Tiebreak: zeroPtr, Set3Games: sixPtr, Set3Tiebreak: zeroPtr, Set4Games: sixPtr, Set4Tiebreak: zeroPtr, Set5Games: sixPtr, Set5Tiebreak: zeroPtr},
		{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Set1Games: sixPtr, Set1Tiebreak: zeroPtr, Set2Games: sixPtr, Set2Tiebreak: zeroPtr},
		{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Set1Games: sixPtr, Set1Tiebreak: zeroPtr, Set2Games: sixPtr, Set2Tiebreak: zeroPtr},
	}

	t.Run("toSlotSlice all filled", func(t *testing.T) {
		testSlice := toSlotSlice(slotRecords)
		assert := assert.New(t)
		assert.Equal(testSlice, slotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: []SetScore{
				{Number: 1, Games: 6, Tiebreak: 4},
				{Number: 2, Games: 4, Tiebreak: 0},
				{Number: 3, Games: 6, Tiebreak: 0},
				{Number: 4, Games: 6, Tiebreak: 0},
				{Number: 5, Games: 6, Tiebreak: 0},
			}},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", SetScores: dummySetScores},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", SetScores: dummySetScores},
		})
	})

	t.Run("toSlotSlice empty", func(t *testing.T) {
		testSlice := toSlotSlice([]SlotRecord{})
		assert := assert.New(t)
		assert.Equal(testSlice, slotSlice{})
	})
}

func TestNoAlphabet(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	cases := []struct {
		s        string
		expected bool
	}{
		{"", false},
		{" ", false},
		{"-", false},
		{" - ", false},
		{"04-=+./?!@%{}[]()", false},
		{"a", true},
		{"A", true},
		{"name", true},
		{"Roger Federer", true},
		{"N. Osaka", true},
		{"-23489.<;@T972347", true},
	}

	for _, item := range cases {
		assert.Equal(hasAlphabet(item.s), item.expected)
	}
}
