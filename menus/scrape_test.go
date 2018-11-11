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
		postProcessingTest{"\xa0\n", ""},
		postProcessingTest{"  asdf\xa0\n", "asdf"},
		// TBC tests
		postProcessingTest{"TBC", ""},
		postProcessingTest{"To be Confirmed", ""},
		postProcessingTest{"To be confirmed \n", ""},
		postProcessingTest{"to Be cONFIRmED", ""},
		// Fish of the Day tests
		postProcessingTest{"Fish of the Day", "Fish of the Day"},
		postProcessingTest{"FoD", "Fish of the Day"},
		postProcessingTest{"fIsh OF ThE DAy", "Fish of the Day"},
	}
}

func TestPostProcessing(t *testing.T) {
	for _, ppt := range postProcessingCases() {
		assert.Equal(t, ppt.Expected, postProcessing(ppt.Case), "Incorrect post processing")
	}
}
