package metadata

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func mustMetadataDocument(t *testing.T, html string) *goquery.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("goquery.NewDocumentFromReader() error = %v", err)
	}

	return doc
}

func TestCleanTitleRemovesSiteName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		title    string
		siteName string
		want     string
	}{
		{
			name:     "site name at end",
			title:    "Advanced Test Article - Test Site",
			siteName: "Test Site",
			want:     "Advanced Test Article",
		},
		{
			name:     "site name at start",
			title:    "Test Site | Advanced Test Article",
			siteName: "Test Site",
			want:     "Advanced Test Article",
		},
		{
			name:     "regex meta characters in site name",
			title:    "Advanced Test Article - Test (Site)+",
			siteName: "Test (Site)+",
			want:     "Advanced Test Article",
		},
		{
			name:     "no match keeps title",
			title:    "Advanced Test Article",
			siteName: "Different Site",
			want:     "Advanced Test Article",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := cleanTitle(tt.title, tt.siteName)
			if got != tt.want {
				t.Fatalf("cleanTitle(%q, %q) = %q, want %q", tt.title, tt.siteName, got, tt.want)
			}
		})
	}
}

func TestGetSchemaPropertyHandlesArrayIndex(t *testing.T) {
	t.Parallel()

	schema := map[string]any{
		"author": []any{
			map[string]any{"name": "First Author"},
			map[string]any{"name": "Second Author"},
		},
	}

	got := getSchemaProperty(schema, "author.[1].name")
	if got != "Second Author" {
		t.Fatalf("getSchemaProperty() = %q, want %q", got, "Second Author")
	}
}

func TestExtractPrefersBaseURLAndMetaData(t *testing.T) {
	t.Parallel()

	authorName := "author"
	authorContent := "Meta Author"
	descriptionName := "description"
	descriptionContent := "Meta description"
	imageProperty := "og:image"
	imageContent := "https://cdn.example.com/image.jpg"
	siteProperty := "og:site_name"
	siteContent := "Example Site"
	publishedProperty := "article:published_time"
	publishedContent := "2026-04-21"

	doc := mustMetadataDocument(t, `<html><head>
		<title>Example Article - Example Site</title>
		<link rel="icon" href="/favicon.ico">
	</head><body>
		<time datetime="2025-01-01"></time>
	</body></html>`)

	metaTags := []MetaTag{
		{Name: &authorName, Content: &authorContent},
		{Name: &descriptionName, Content: &descriptionContent},
		{Property: &imageProperty, Content: &imageContent},
		{Property: &siteProperty, Content: &siteContent},
		{Property: &publishedProperty, Content: &publishedContent},
	}

	metadata := Extract(doc, nil, metaTags, "https://www.example.com/articles/test")
	if metadata == nil {
		t.Fatal("Extract() returned nil")
	}
	if metadata.Domain != "example.com" {
		t.Fatalf("Domain = %q, want %q", metadata.Domain, "example.com")
	}
	if metadata.Favicon != "https://www.example.com/favicon.ico" {
		t.Fatalf("Favicon = %q, want resolved favicon URL", metadata.Favicon)
	}
	if metadata.Title != "Example Article" {
		t.Fatalf("Title = %q, want %q", metadata.Title, "Example Article")
	}
	if metadata.Author != "Meta Author" {
		t.Fatalf("Author = %q, want %q", metadata.Author, "Meta Author")
	}
	if metadata.Description != "Meta description" {
		t.Fatalf("Description = %q, want %q", metadata.Description, "Meta description")
	}
	if metadata.Image != "https://cdn.example.com/image.jpg" {
		t.Fatalf("Image = %q, want %q", metadata.Image, "https://cdn.example.com/image.jpg")
	}
	if metadata.Site != "Example Site" {
		t.Fatalf("Site = %q, want %q", metadata.Site, "Example Site")
	}
	if metadata.Published != "2026-04-21" {
		t.Fatalf("Published = %q, want %q", metadata.Published, "2026-04-21")
	}
}

func TestExtractFallsBackToSchemaAndDOM(t *testing.T) {
	t.Parallel()

	doc := mustMetadataDocument(t, `<html><head>
		<title>Schema Headline | Publisher Name</title>
		<base href="https://blog.example.org/posts/123">
	</head><body>
		<div class="author">DOM Author</div>
	</body></html>`)

	schema := map[string]any{
		"headline":      "Schema Headline",
		"description":   "Schema description",
		"datePublished": "2026-04-20",
		"image": map[string]any{
			"url": "https://blog.example.org/schema-image.jpg",
		},
		"author": []any{
			map[string]any{"name": "Schema Author"},
			map[string]any{"name": "Schema Author"},
			map[string]any{"name": "Another Author"},
		},
		"publisher": map[string]any{"name": "Publisher Name"},
	}

	metadata := Extract(doc, schema, nil, "")
	if metadata == nil {
		t.Fatal("Extract() returned nil")
	}
	if metadata.Domain != "blog.example.org" {
		t.Fatalf("Domain = %q, want %q", metadata.Domain, "blog.example.org")
	}
	if metadata.Title != "Schema Headline" {
		t.Fatalf("Title = %q, want %q", metadata.Title, "Schema Headline")
	}
	if metadata.Author != "Schema Author, Another Author" {
		t.Fatalf("Author = %q, want deduplicated schema authors", metadata.Author)
	}
	if metadata.Description != "Schema description" {
		t.Fatalf("Description = %q, want %q", metadata.Description, "Schema description")
	}
	if metadata.Image != "https://blog.example.org/schema-image.jpg" {
		t.Fatalf("Image = %q, want schema image URL", metadata.Image)
	}
	if metadata.Site != "Publisher Name" {
		t.Fatalf("Site = %q, want %q", metadata.Site, "Publisher Name")
	}
	if metadata.Published != "2026-04-20" {
		t.Fatalf("Published = %q, want %q", metadata.Published, "2026-04-20")
	}
	if metadata.Favicon != "https://blog.example.org/favicon.ico" {
		t.Fatalf("Favicon = %q, want default favicon resolved from base tag", metadata.Favicon)
	}
}
