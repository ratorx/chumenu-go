package menus

import (
	"fmt"
	"strings"
)

const emptyMeal = " - To Be Confirmed"

// Meal represents a list of food items
type Meal []string

func (m Meal) String() string {
	// fmt.Println()
	if len(m) == 0 {
		return emptyMeal
	}

	return fmt.Sprintf(" - %s", strings.Join(m, "\n - "))
}

// Menu is a struct which contains the meals provided for lunch and dinner
type Menu struct {
	Lunch  Meal
	Dinner Meal
}

func (m Menu) String() string {
	return fmt.Sprintf("\nLunch:\n%s\nDinner:\n%s\n", m.Lunch, m.Dinner)
}

// Datablock is a struct which contains menus of 2 consecutive days
type Datablock struct {
	Current Menu
	Next    Menu
}
