package menus

import (
	"fmt"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html/atom"
)

func postProcessing(item string) string {
	item = strings.TrimSpace(item)
	item = strings.TrimSuffix(item, "\xa0")
	if item == "FOD" || item == "Fish of the day" {
		item = "Fish of the Day"
	}

	return item
}

func GetData(day uint8) (Datablock, error) {
	table, err := getTable()
	if err != nil {
		return Datablock{}, err
	}

	menus := scrape.FindAll(table, scrape.ByTag(atom.Tr))
	if len(menus) != 8 {
		return Datablock{}, fmt.Errorf("Scraper: Menus array is not the correct length (Expected: 8, Actual: %v)", len(menus))
	}

	ret := Datablock{}
	menu, err := parseDay(menus[day])
	if err != nil {
		return ret, err
	}
	ret.Current = menu

	menu, err = parseDay(menus[(day%7)+1])
	if err != nil {
		return ret, err
	}
	ret.Next = menu

	return ret, nil
}

func GetMenus() ([]Menu, error) {
	table, err := getTable()
	if err != nil {
		return nil, err
	}

	return parseWeek(table)
}
