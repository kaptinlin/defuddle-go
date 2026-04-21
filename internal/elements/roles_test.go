package elements

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleProcessorProcessRolesConvertsSemanticElements(t *testing.T) {
	html := `
	<div role="paragraph" id="intro">Intro</div>
	<div role="list" id="steps">
		<div role="listitem">
			<span class="label">1)</span>
			<div class="content"><div role="paragraph">First item</div></div>
		</div>
		<div role="listitem">
			<span class="label">2)</span>
			<div class="content"><div role="paragraph">Second item</div></div>
		</div>
	</div>
	<div role="button" id="cta">Click</div>
	<div role="link" id="docs-link" href="https://example.com/docs">Docs</div>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	processor := NewRoleProcessor(doc)
	processor.ProcessRoles(DefaultRoleProcessingOptions())

	assert.Equal(t, 1, doc.Find("p#intro").Length())
	assert.Equal(t, 1, doc.Find("ol#steps").Length())
	assert.Equal(t, 2, doc.Find("ol#steps > li").Length())
	assert.Zero(t, doc.Find(".label").Length())
	assert.Equal(t, 1, doc.Find("button#cta").Length())
	assert.Equal(t, 1, doc.Find(`a#docs-link[href="https://example.com/docs"]`).Length())
	assert.Zero(t, doc.Find(`[role]`).Length())
}

func TestRoleProcessorProcessRolesUsesDefaultOptionsWhenNil(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<div role="paragraph">Default paragraph</div>`))
	require.NoError(t, err)

	processor := NewRoleProcessor(doc)
	processor.ProcessRoles(nil)

	assert.Equal(t, 1, doc.Find("p").Length())
	assert.Equal(t, "Default paragraph", strings.TrimSpace(doc.Find("p").Text()))
}
