// Package markdown provides HTML to Markdown conversion functionality.
// It uses the html-to-markdown library to convert HTML content to clean Markdown format.
package markdown

import (
	"fmt"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

// ConvertHTML converts HTML content to Markdown with default settings
func ConvertHTML(htmlContent string) (string, error) {
	markdownContent, err := htmltomarkdown.ConvertString(htmlContent)
	if err != nil {
		return "", fmt.Errorf("failed to convert HTML to Markdown: %w", err)
	}

	// Clean up the markdown content
	markdownContent = strings.TrimSpace(markdownContent)

	// Remove excessive newlines
	markdownContent = strings.ReplaceAll(markdownContent, "\n\n\n", "\n\n")

	return markdownContent, nil
}
