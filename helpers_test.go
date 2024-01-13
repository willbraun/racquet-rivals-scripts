package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var allFilled = slotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)"},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)"},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)"},
}

var twoFilled = slotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)"},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)"},
}

var twoFilledOneBlank = slotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)"},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)"},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "", Seed: ""},
}

var allBlank = slotSlice{
	Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "", Seed: ""},
	Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "", Seed: ""},
	Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "", Seed: ""},
}

var seeds = map[string]string{
	"Roger Federer": "(1)",
	"Rafael Nadal":  "(2)",
}

func TestGetNewSlots(t *testing.T) {
	t.Parallel()

	t.Run("Add a slot", func(t *testing.T) {
		newSlots := getNewSlots(allFilled, twoFilled)
		assert := assert.New(t)

		assert.Equal(newSlots, slotSlice{
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)"},
		})
	})

	t.Run("Add all slots", func(t *testing.T) {
		newSlots := getNewSlots(allFilled, slotSlice{})
		assert := assert.New(t)

		assert.Equal(newSlots, slotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)"},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)"},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)"},
		})
	})

	t.Run("Add no slots, name update", func(t *testing.T) {
		newSlots := getNewSlots(allFilled, twoFilledOneBlank)
		assert := assert.New(t)

		assert.Equal(newSlots, slotSlice{})
	})

	t.Run("No changes", func(t *testing.T) {
		newSlots := getNewSlots(allFilled, allFilled)
		assert := assert.New(t)

		assert.Equal(newSlots, slotSlice{})
	})

	t.Run("Empty scrape", func(t *testing.T) {
		newSlots := getNewSlots(slotSlice{}, allFilled)
		assert := assert.New(t)

		assert.Equal(newSlots, slotSlice{})
	})
}

func TestPrepareUpdates(t *testing.T) {
	t.Parallel()

	t.Run("Update a slot", func(t *testing.T) {
		updatedSlots := prepareUpdates(allFilled, twoFilledOneBlank, seeds)
		assert := assert.New(t)

		assert.Equal(updatedSlots, slotSlice{
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)"},
		})
	})

	t.Run("Update all slots", func(t *testing.T) {
		updatedSlots := prepareUpdates(allFilled, allBlank, seeds)
		assert := assert.New(t)

		assert.Equal(updatedSlots, slotSlice{
			Slot{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)"},
			Slot{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)"},
			Slot{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)"},
		})
	})

	t.Run("No update, just add", func(t *testing.T) {
		updatedSlots := prepareUpdates(allFilled, twoFilled, seeds)
		assert := assert.New(t)

		assert.Equal(updatedSlots, slotSlice{})
	})

	t.Run("No changes", func(t *testing.T) {
		updatedSlots := prepareUpdates(allFilled, allFilled, seeds)
		assert := assert.New(t)

		assert.Equal(updatedSlots, slotSlice{})
	})

	t.Run("No current slots, just add", func(t *testing.T) {
		updatedSlots := prepareUpdates(allFilled, slotSlice{}, seeds)
		assert := assert.New(t)

		assert.Equal(updatedSlots, slotSlice{})
	})

	t.Run("Empty scrape", func(t *testing.T) {
		updatedSlots := prepareUpdates(slotSlice{}, allFilled, seeds)
		assert := assert.New(t)

		assert.Equal(updatedSlots, allBlank)
	})
}

func TestToSlotSlice(t *testing.T) {
	t.Parallel()

	slotRecords := []SlotRecord{
		{ID: "aaa", DrawID: "draw1", Round: 1, Position: 1, Name: "Roger Federer", Seed: "(1)"},
		{ID: "bbb", DrawID: "draw1", Round: 1, Position: 2, Name: "Rafael Nadal", Seed: "(2)"},
		{ID: "ccc", DrawID: "draw1", Round: 2, Position: 1, Name: "Roger Federer", Seed: "(1)"},
	}

	t.Run("toSlotSlice all filled", func(t *testing.T) {
		slotSlice := toSlotSlice(slotRecords)
		assert := assert.New(t)

		assert.Equal(slotSlice, allFilled)
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
