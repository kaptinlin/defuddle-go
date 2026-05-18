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

func TestFindFootnoteDefinitionMatchesTextPrefixes(t *testing.T) {
	tests := []struct {
		name string
		key  string
		text string
	}{
		{name: "dot", key: "1", text: "1. Dot note"},
		{name: "brackets", key: "2", text: "[2] Bracket note"},
		{name: "paren", key: "3", text: "3) Paren note"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<article><section class="footnotes"><ol><li>` + tt.text + `</li></ol></section></article>`

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			require.NoError(t, err)

			processor := NewFootnoteProcessor(doc)
			definition := processor.findFootnoteDefinition(tt.key)
			require.NotNil(t, definition)
			assert.Equal(t, tt.text, strings.TrimSpace(definition.Text()))
		})
	}
}
