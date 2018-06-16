package menus

import (
	"fmt"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html/atom"
)

// GetData returns a Datablock which contains the menus for the provided weekday and the following day
func GetData(weekday time.Weekday) (Datablock, error) {
	table, err := getTable()
	if err != nil {
		return Datablock{}, err
	}

	menus := scrape.FindAll(table, scrape.ByTag(atom.Tr))
	if len(menus) != 8 {
		return Datablock{}, fmt.Errorf("Scraper: Menus array is not the correct length (Expected: 8, Actual: %v)", len(menus))
	}

	// Convert from Weekday to slice index
	// Only need to change Sunday, int value of everything else is correct
	index := int(weekday)
	if weekday == time.Sunday {
		index = 7
	}

	ret := Datablock{}
	menu, err := parseDay(menus[index])
	if err != nil {
		return ret, err
	}
	ret.Current = menu

	menu, err = parseDay(menus[(index%7)+1])
	if err != nil {
		return ret, err
	}
	ret.Next = menu

	return ret, nil
}

// GetMenus returns a slice containing the menus for the entire week
func GetMenus() ([]Menu, error) {
	table, err := getTable()
	if err != nil {
		return nil, err
	}

	return parseWeek(table)
}
