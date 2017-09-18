package menus

import (
	"fmt"
	// "fmt"

	"net/http"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	menuURL    = "https://www.chu.cam.ac.uk/student-hub/catering/menus/"
	maxRetries = 5
)

func getRootNode(url string) (*html.Node, error) {
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	// TODO: Check StatusCode and use maxRetries
	defer res.Body.Close()

	// TODO: Custom error for improper input
	return html.Parse(res.Body)
}

func makeTableError(num int) error {
	return fmt.Errorf("Scraper: Incorrect number of tables scraped (Expected: 1, Actual: %v)", num)
}

func getTableNode(root *html.Node) (*html.Node, error) {
	tables := scrape.FindAll(root, scrape.ByTag(atom.Table))
	// Consistency Check
	if len(tables) == 0 {
		return nil, makeTableError(0)
	} else if len(tables) != 1 {
		return tables[0], makeTableError(len(tables))
	}
	return tables[0], nil
}

// This is a wrapper function for the above
func getTable() (*html.Node, error) {
	root, err := getRootNode(menuURL)
	if err != nil {
		return nil, err
	}

	return getTableNode(root)
}

func parseMeal(node *html.Node) (Meal, error) {
	if node == nil {
		// Custom Error
		return Meal{}, fmt.Errorf("Scraper: nil node passed into parseMeal")
	}

	// Parses a list of meal items into a Meal
	items := scrape.FindAll(node, scrape.ByTag(atom.Li))
	if len(items) == 0 || scrape.Text(items[0]) == "\xa0" || scrape.Text(items[0]) == "" { // Empty menu conditions
		return Meal{}, nil
	}

	var meal Meal = make([]string, len(items))
	for i, item := range items {
		meal[i] = postProcessing(scrape.Text(item))
	}

	return meal, nil
}

func parseDay(node *html.Node) (Menu, error) {
	if node == nil {
		// Custom Error
		return Menu{}, fmt.Errorf("Scraper: nil node passed into parseDay")
	}
	// Parses a day's menu into a Menu
	unparsedMeal := scrape.FindAll(node, scrape.ByTag(atom.Td))
	if len(unparsedMeal) != 3 {
		return Menu{}, fmt.Errorf("Scraper: Incorrect number of columns in table (Expected: 3, Actual: %v)", len(unparsedMeal))
	}

	menu := Menu{}
	meal, err := parseMeal(unparsedMeal[1])
	if err != nil {
		return menu, err
	}
	menu.Lunch = meal

	meal, err = parseMeal(unparsedMeal[2])
	if err != nil {
		return menu, err
	}
	menu.Dinner = meal

	return menu, nil
}

func parseWeek(node *html.Node) ([]Menu, error) {
	if node == nil {
		// Custom Error
		return []Menu{}, fmt.Errorf("Scraper: nil node passed into parseWeek")
	}
	// Parses the whole weeks menu and returns a slice of menu structs
	unparsedMenus := scrape.FindAll(node, scrape.ByTag(atom.Tr))
	if len(unparsedMenus) != 8 {
		return []Menu{}, fmt.Errorf("Scraper: Incorrect number of rows in table (Expected: 8, Actual: %v)", len(unparsedMenus))
	}

	week := make([]Menu, 7)
	for i, menu := range unparsedMenus[1:] {
		day, err := parseDay(menu)
		if err != nil {
			return week, err
		}
		week[i] = day
	}

	return week, nil
}

func benchmark(root *html.Node) ([]Menu, error) {
	table, err := getTableNode(root)
	if err != nil {
		return nil, err
	}

	return parseWeek(table)
}
