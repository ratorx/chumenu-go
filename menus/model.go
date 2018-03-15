package menus

import (
	"fmt"
	"strings"
)

type Meal []string

func (m Meal) String() string {
	// fmt.Println()
	if len(m) == 0 {
		return " - TBC"
	}

	return fmt.Sprintf(" - %s", strings.Join(m, "\n - "))
}

type Menu struct {
	Lunch  Meal
	Dinner Meal
}

func (m Menu) String() string {
	return fmt.Sprintf("\nLunch:\n%s\nDinner:\n%s\n", m.Lunch, m.Dinner)
}

type Datablock struct {
	Current Menu
	Next    Menu
}
