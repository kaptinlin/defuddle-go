package defuddle

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSelectsReadableTableCellWhenSemanticContainersAreAbsent(t *testing.T) {
	t.Parallel()

	body := strings.Repeat("Table based article paragraph with enough original reporting and analysis. ", 12)
	html := `<html><head><title>Table Article</title></head><body>
		<table><tr><td><nav><a href="/a">Home</a><a href="/b">Archive</a></nav></td><td><h1>Table Article</h1><p>` + body + `</p></td></tr></table>
	</body></html>`

	result, err := ParseFromString(context.Background(), html, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Contains(t, result.Content, "Table based article paragraph")
	assert.Contains(t, result.Content, "Archive")
	assert.Greater(t, result.WordCount, 50)
}

func TestParseSelectsHighestScoredContentWhenNoSemanticContainerExists(t *testing.T) {
	t.Parallel()

	body := strings.Repeat("Scored article text with meaningful sentences for readers. ", 14)
	html := `<html><head><title>Scored Article</title></head><body>
		<div class="site-nav"><a href="/one">One</a><a href="/two">Two</a><a href="/three">Three</a></div>
		<section class="layout"><h1>Scored Article</h1><p>` + body + `</p></section>
	</body></html>`

	result, err := ParseFromString(context.Background(), html, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Contains(t, result.Content, "Scored article text")
	assert.NotContains(t, result.Content, "site-nav")
	assert.Greater(t, result.WordCount, 50)
}

func TestParseFallsBackToBodyWhenNoContentCandidateQualifies(t *testing.T) {
	t.Parallel()

	html := `<html><head><title>Tiny Page</title></head><body><span>Short body</span></body></html>`

	result, err := ParseFromString(context.Background(), html, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Contains(t, result.Content, "Short body")
	assert.Equal(t, 2, result.WordCount)
}
