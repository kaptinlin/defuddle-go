// Package main demonstrates custom extractor usage.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/kaptinlin/defuddle-go"
	"github.com/kaptinlin/defuddle-go/extractors"
)

// CustomBlogExtractor implements a custom extractor for blog sites
type CustomBlogExtractor struct {
	*extractors.ExtractorBase
}

// NewCustomBlogExtractor creates a new custom blog extractor
func NewCustomBlogExtractor(doc *goquery.Document, url string, schemaOrgData any) extractors.BaseExtractor {
	return &CustomBlogExtractor{
		ExtractorBase: extractors.NewExtractorBase(doc, url, schemaOrgData),
	}
}

// CanExtract determines if this extractor can handle the content
func (e *CustomBlogExtractor) CanExtract() bool {
	return e.GetDocument().Find(".blog-post, .post-content").Length() > 0
}

// GetName returns the name of this extractor
func (e *CustomBlogExtractor) Name() string {
	return "CustomBlogExtractor"
}

// Extract performs the custom extraction logic
func (e *CustomBlogExtractor) Extract() *extractors.ExtractorResult {
	doc := e.GetDocument()

	// Extract title
	title := ""
	if titleElement := doc.Find(".post-title, h1").First(); titleElement.Length() > 0 {
		title = titleElement.Text()
	}

	// Extract main content
	contentHTML := ""
	if contentElement := doc.Find(".post-content").First(); contentElement.Length() > 0 {
		if html, err := contentElement.Html(); err == nil {
			contentHTML = html
		}
	}

	variables := map[string]string{
		"title": title,
		"site":  "Custom Blog",
	}

	return &extractors.ExtractorResult{
		ContentHTML: contentHTML,
		Variables:   variables,
	}
}

func main() {
	// Register custom extractor for blog.example.com
	extractors.Register(extractors.ExtractorMapping{
		Patterns:  []any{"blog.example.com"},
		Extractor: NewCustomBlogExtractor,
	})

	// HTML content with blog structure
	html := `
	<html>
	<head>
		<title>My Blog Post</title>
	</head>
	<body>
		<h1 class="post-title">Custom Extractor Demo</h1>
		<div class="post-content">
			<p>This content will be extracted by our custom blog extractor.</p>
			<p>The extractor looks for specific CSS classes like .post-content.</p>
		</div>
	</body>
	</html>
	`

	// URL matches our registered pattern
	options := &defuddle.Options{
		URL:   "https://blog.example.com/post/123",
		Debug: true,
	}

	defuddleInstance, err := defuddle.NewDefuddle(html, options)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	result, err := defuddleInstance.Parse(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("=== Custom Extractor Demo ===")
	fmt.Printf("URL: %s\n", options.URL)
	fmt.Printf("Title: %s\n", result.Title)
	fmt.Printf("Site: %s\n", result.Site)
	fmt.Printf("Word Count: %d\n", result.WordCount)

	fmt.Println("\n=== Extracted Content ===")
	fmt.Println(result.Content)

	fmt.Println("\n✅ Custom extractor successfully used!")
}
