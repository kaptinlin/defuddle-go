package extractors

import (
	"strings"
	"testing"
)

func TestYouTubeExtractorNameAndCanExtract(t *testing.T) {
	t.Parallel()

	extractor := NewYouTubeExtractor(newTestDocument(t, `<html><body></body></html>`), "https://youtube.com/watch?v=abc", nil)

	if !extractor.CanExtract() {
		t.Fatal("CanExtract() = false, want true")
	}
	if got := extractor.Name(); got != "YouTubeExtractor" {
		t.Fatalf("Name() = %q, want YouTubeExtractor", got)
	}
}

func TestYouTubeExtractorUsesVideoObjectFromMapSchema(t *testing.T) {
	t.Parallel()

	schema := map[string]any{
		"@type":        "VideoObject",
		"name":         "Map Schema Video",
		"description":  "Map schema description",
		"author":       "Channel Name",
		"uploadDate":   "2026-05-07",
		"thumbnailUrl": "https://cdn.example.com/map.jpg",
	}
	extractor := NewYouTubeExtractor(newTestDocument(t, `<html><body></body></html>`), "https://youtube.com/watch?v=map123", schema)

	result := extractor.Extract()
	if got := result.Variables["title"]; got != "Map Schema Video" {
		t.Fatalf("title = %q, want Map Schema Video", got)
	}
	if got := result.Variables["author"]; got != "Channel Name" {
		t.Fatalf("author = %q, want Channel Name", got)
	}
	if got := result.Variables["image"]; got != "https://cdn.example.com/map.jpg" {
		t.Fatalf("image = %q, want thumbnail URL", got)
	}
	if !strings.Contains(result.ContentHTML, "https://www.youtube.com/embed/map123") {
		t.Fatalf("ContentHTML = %q, want map123 iframe", result.ContentHTML)
	}
}

func TestYouTubeExtractorFallsBackWhenSchemaIsNotVideoObject(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><head><title>Fallback Video - YouTube</title></head><body><meta name="description" content="Meta description"><div class="content">DOM description</div></body></html>`)
	extractor := NewYouTubeExtractor(doc, "https://youtube.com/watch?v=fallback123", map[string]any{"@type": "Article"})

	result := extractor.Extract()
	if got := result.Variables["title"]; got != "Fallback Video" {
		t.Fatalf("title = %q, want DOM title fallback", got)
	}
	if got := result.ExtractedContent["videoId"]; got != "fallback123" {
		t.Fatalf("videoId = %#v, want fallback123", got)
	}
	if got := result.Variables["image"]; got != "https://img.youtube.com/vi/fallback123/maxresdefault.jpg" {
		t.Fatalf("image = %q, want generated thumbnail", got)
	}
}

func TestYouTubeExtractorIgnoresNonStringSchemaStrings(t *testing.T) {
	t.Parallel()

	schema := map[string]any{
		"@type":       "VideoObject",
		"name":        42,
		"description": []string{"not", "a", "string"},
		"author":      []any{"Channel"},
		"uploadDate":  20260519,
	}
	doc := newTestDocument(t, `<html><head><title>DOM Video - YouTube</title></head><body><div id="description">DOM description</div></body></html>`)
	extractor := NewYouTubeExtractor(doc, "https://youtube.com/watch?v=dom123", schema)

	result := extractor.Extract()

	if got := result.Variables["title"]; got != "DOM Video" {
		t.Fatalf("title = %q, want DOM fallback", got)
	}
	if got := result.Variables["author"]; got != "" {
		t.Fatalf("author = %q, want empty string", got)
	}
	if got := result.Variables["published"]; got != "" {
		t.Fatalf("published = %q, want empty string", got)
	}
	if !strings.Contains(result.ContentHTML, "DOM description") {
		t.Fatalf("ContentHTML = %q, want DOM description fallback", result.ContentHTML)
	}
}
