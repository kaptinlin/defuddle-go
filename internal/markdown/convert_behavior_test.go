package markdown

import (
	"strings"
	"testing"
)

func TestConvertHTMLReturnsConversionErrors(t *testing.T) {
	t.Parallel()

	_, err := ConvertHTML(strings.Repeat("<div>", 20000))

	if err == nil {
		t.Fatal("ConvertHTML() error = nil, want conversion error")
	}
}
