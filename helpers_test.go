package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var setScoresA = SetSlice{
	{ID: "ss_a_1", DrawSlotID: "aaa", Number: 1, Games: 6, Tiebreak: 0},
	{ID: "ss_a_2", DrawSlotID: "aaa", Number: 2, Games: 6, Tiebreak: 0},
}

var setScoresB = SetSlice{
	{ID: "ss_b_1", DrawSlotID: "bbb", Number: 1, Games: 6, Tiebreak: 0},
	{ID: "ss_b_2", DrawSlotID: "bbb", Number: 2, Games: 6, Tiebreak: 0},
}

var setScoresC = SetSlice{
	{ID: "ss_c_1", DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
	{ID: "ss_c_2", DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
}

var allFilled = SlotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresA},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Sets: setScoresB},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresC},
}

var allFilledPartialSets = SlotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: SetSlice{
		{ID: "ss_a_1", DrawSlotID: "aaa", Number: 1, Games: 4, Tiebreak: 0},
	}},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Sets: SetSlice{
		{ID: "ss_b_1", DrawSlotID: "bbb", Number: 1, Games: 2, Tiebreak: 0},
	}},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: SetSlice{}},
}

var twoFilled = SlotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresA},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Sets: setScoresB},
}

var twoFilledOneWithName = SlotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresA},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Sets: setScoresB},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: SetSlice{}},
}

var twoFilledOneBlank = SlotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresA},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Sets: setScoresB},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "", Seed: "", Sets: SetSlice{}},
}

var allBlank = SlotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "", Seed: "", Sets: SetSlice{}},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "", Seed: "", Sets: SetSlice{}},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "", Seed: "", Sets: SetSlice{}},
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
		assert.Equal(newSlots, SlotSlice{
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresC},
		})
		assert.Equal(updatedSlots, SlotSlice{})
		assert.Equal(newSets, SetSlice{
			{ID: "", DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, SetSlice{})
	})

	t.Run("Add all slots", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, SlotSlice{}, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, SlotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresA},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Sets: setScoresB},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresC},
		})
		assert.Equal(updatedSlots, SlotSlice{})
		assert.Equal(newSets, SetSlice{
			{ID: "", DrawSlotID: "aaa", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "aaa", Number: 2, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "bbb", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "bbb", Number: 2, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, SetSlice{})
	})

	t.Run("Update slot name", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(twoFilledOneWithName, twoFilledOneBlank, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, SlotSlice{})
		assert.Equal(updatedSlots, SlotSlice{
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: SetSlice{}},
		})
		assert.Equal(newSets, SetSlice{})
		assert.Equal(updatedSets, SetSlice{})
	})

	t.Run("Add slot score", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, twoFilledOneWithName, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, SlotSlice{})
		assert.Equal(updatedSlots, SlotSlice{})
		assert.Equal(newSets, SetSlice{
			{ID: "", DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, SetSlice{})
	})

	t.Run("Update slot name and add score", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, twoFilledOneBlank, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, SlotSlice{})
		assert.Equal(updatedSlots, SlotSlice{
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresC},
		})
		assert.Equal(newSets, SetSlice{
			{ID: "", DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, SetSlice{})
	})

	t.Run("Update and add score", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, allFilledPartialSets, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, SlotSlice{})
		assert.Equal(updatedSlots, SlotSlice{})
		assert.Equal(newSets, SetSlice{
			{ID: "", DrawSlotID: "aaa", Number: 2, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "bbb", Number: 2, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, SetSlice{
			{ID: "ss_a_1", DrawSlotID: "aaa", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "ss_b_1", DrawSlotID: "bbb", Number: 1, Games: 6, Tiebreak: 0},
		})
	})

	t.Run("Update all slots", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, allBlank, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, SlotSlice{})
		assert.Equal(updatedSlots, SlotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresA},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Sets: setScoresB},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresC},
		})
		assert.Equal(newSets, SetSlice{
			{ID: "", DrawSlotID: "aaa", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "aaa", Number: 2, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "bbb", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "bbb", Number: 2, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "ccc", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "ccc", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, SetSlice{})
	})

	t.Run("Scraped all blanks, do not clear", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allBlank, allFilled, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, SlotSlice{})
		assert.Equal(updatedSlots, SlotSlice{})
		assert.Equal(newSets, SetSlice{})
		assert.Equal(updatedSets, SetSlice{})
	})

	t.Run("Scraped one blank, do not clear", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(twoFilledOneBlank, allFilled, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, SlotSlice{})
		assert.Equal(updatedSlots, SlotSlice{})
		assert.Equal(newSets, SetSlice{})
		assert.Equal(updatedSets, SetSlice{})
	})

	t.Run("One updated, manual entry ignored", func(t *testing.T) {
		// round 1 slot 2 is scraped successfully and updates a blank slot
		// round 2 slot 1 is scraped as a blank due to an error with the website
		scraped := SlotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresA},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Sets: setScoresB},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "", Seed: "", Sets: SetSlice{}},
		}

		// round 1 slot 2 hasn't been filled yet
		// round 2 slot 1 is entered manually
		current := SlotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresA},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "", Seed: "", Sets: SetSlice{}},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresC},
		}

		// only round 1 slot 2 should be updated
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(scraped, current, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, SlotSlice{})
		assert.Equal(updatedSlots, SlotSlice{
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Sets: setScoresB},
		})
		assert.Equal(newSets, SetSlice{
			{ID: "", DrawSlotID: "bbb", Number: 1, Games: 6, Tiebreak: 0},
			{ID: "", DrawSlotID: "bbb", Number: 2, Games: 6, Tiebreak: 0},
		})
		assert.Equal(updatedSets, SetSlice{})
	})

	t.Run("No changes", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(allFilled, allFilled, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, SlotSlice{})
		assert.Equal(updatedSlots, SlotSlice{})
		assert.Equal(newSets, SetSlice{})
		assert.Equal(updatedSets, SetSlice{})
	})

	t.Run("Empty scrape", func(t *testing.T) {
		newSlots, updatedSlots, newSets, updatedSets := getUpdates(SlotSlice{}, allFilled, seeds)
		assert := assert.New(t)
		assert.Equal(newSlots, SlotSlice{})
		assert.Equal(updatedSlots, SlotSlice{})
		assert.Equal(newSets, SetSlice{})
		assert.Equal(updatedSets, SetSlice{})
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
		{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Set1ID: "ss_a_1", Set1Games: sixPtr, Set1Tiebreak: fourPtr, Set2ID: "ss_a_2", Set2Games: fourPtr, Set2Tiebreak: zeroPtr, Set3ID: "ss_a_3", Set3Games: sixPtr, Set3Tiebreak: zeroPtr, Set4ID: "ss_a_4", Set4Games: sixPtr, Set4Tiebreak: zeroPtr, Set5ID: "ss_a_5", Set5Games: sixPtr, Set5Tiebreak: zeroPtr},
		{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Set1ID: "ss_b_1", Set1Games: sixPtr, Set1Tiebreak: zeroPtr, Set2ID: "ss_b_2", Set2Games: sixPtr, Set2Tiebreak: zeroPtr},
		{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Set1ID: "ss_c_1", Set1Games: sixPtr, Set1Tiebreak: zeroPtr, Set2ID: "ss_c_2", Set2Games: sixPtr, Set2Tiebreak: zeroPtr},
	}

	t.Run("toSlotSlice all filled", func(t *testing.T) {
		testSlice := toSlotSlice(slotRecords)
		assert := assert.New(t)
		assert.Equal(testSlice, SlotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: SetSlice{
				{ID: "ss_a_1", DrawSlotID: "aaa", Number: 1, Games: 6, Tiebreak: 4},
				{ID: "ss_a_2", DrawSlotID: "aaa", Number: 2, Games: 4, Tiebreak: 0},
				{ID: "ss_a_3", DrawSlotID: "aaa", Number: 3, Games: 6, Tiebreak: 0},
				{ID: "ss_a_4", DrawSlotID: "aaa", Number: 4, Games: 6, Tiebreak: 0},
				{ID: "ss_a_5", DrawSlotID: "aaa", Number: 5, Games: 6, Tiebreak: 0},
			}},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)", Sets: setScoresB},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)", Sets: setScoresC},
		})
	})

	t.Run("toSlotSlice empty", func(t *testing.T) {
		testSlice := toSlotSlice([]SlotRecord{})
		assert := assert.New(t)
		assert.Equal(testSlice, SlotSlice{})
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
