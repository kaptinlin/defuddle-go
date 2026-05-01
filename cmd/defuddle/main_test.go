package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/defuddle-go"
)

func TestMarkdownContentUsesExistingMarkdown(t *testing.T) {
	available := "# Existing"
	result := &defuddle.Result{
		Content:         "<h1>Fallback</h1>",
		ContentMarkdown: &available,
	}

	content := markdownContent(result, &ParseOptions{Source: "test.html"})

	assert.Equal(t, available, content)
}

func TestMarkdownContentConvertsHTMLContent(t *testing.T) {
	result := &defuddle.Result{Content: "<article><h1>Generated</h1><p>Readable content.</p></article>"}

	content := markdownContent(result, &ParseOptions{Source: "test.html"})

	require.NotEqual(t, result.Content, content)
	assert.Contains(t, content, "Generated")
	assert.Contains(t, content, "Readable content.")
}

func TestValidateFilePathRejectsParentSegments(t *testing.T) {
	assert.ErrorIs(t, validateFilePath("../article.html"), ErrDirectoryTraversal)
}
