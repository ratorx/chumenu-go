package menus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mealTest struct {
	Case     Meal
	Expected string
}

func mealCases() []mealTest {
	return []mealTest{
		mealTest{Meal{}, " - TBC"},
		mealTest{Meal{"one"}, " - one"},
		mealTest{Meal{"one", "two"}, " - one\n - two"},
	}
}

func TestMeal_String(t *testing.T) {
	for _, mt := range mealCases() {
		assert.Equal(t, mt.Expected, mt.Case.String(), "Meal String conversion failed")
	}
}

type menuTest struct {
	Case     Menu
	Expected string
}

func menuCases() []menuTest {
	return []menuTest{
		menuTest{Menu{Meal{}, Meal{}}, "\nLunch:\n - TBC\nDinner:\n - TBC\n"},
	}
}

func TestMenu_String(t *testing.T) {
	for _, mt := range menuCases() {
		assert.Equal(t, mt.Expected, mt.Case.String(), "Menu String conversion failed")
	}
}
