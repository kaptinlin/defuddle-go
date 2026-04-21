package elements

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFootnoteProcessorPublicHelpers(t *testing.T) {
	html := `<article><p>Text<sup><a href="#fn1">1</a></sup></p><div id="fn1">Note</div></article>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	processor := NewFootnoteProcessor(doc)
	footnotes := processor.GetFootnotes()
	assert.NotEmpty(t, footnotes)
	assert.True(t, processor.HasFootnotes())

	cleaned := processor.CleanupFootnotes([]*Footnote{
		{ID: "fn1", Content: "Note"},
		{ID: "fn1", Content: "Duplicate"},
		{ID: "", Content: "Invalid"},
	})
	assert.Len(t, cleaned, 1)
	assert.Equal(t, "fn1", cleaned[0].ID)
}
