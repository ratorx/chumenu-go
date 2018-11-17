package menus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type postProcessingTest struct {
	Case     string
	Expected string
}

func postProcessingCases() []postProcessingTest {
	return []postProcessingTest{
		// Invalid item tests
		{"\xa0\n", ""},
		{"  asdf\xa0\n", "asdf"},
		// TBC tests
		{"TBC", ""},
		{"To be Confirmed", ""},
		{"To be confirmed \n", ""},
		{"to Be cONFIRmED", ""},
		// Fish of the Day tests
		{"Fish of the Day", "Fish of the Day"},
		{"FoD", "Fish of the Day"},
		{"fIsh OF ThE DAy", "Fish of the Day"},
		// No post processing tests
		{"Beef Steak", "Beef Steak"},
		{"SoME WEIrd fOOd", "SoME WEIrd fOOd"},
	}
}

func TestPostProcessing(t *testing.T) {
	for _, ppt := range postProcessingCases() {
		assert.Equal(t, ppt.Expected, postProcessing(ppt.Case), "Incorrect post processing")
	}
}
