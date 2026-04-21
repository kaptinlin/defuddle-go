package extractors

import (
	"strings"
	"testing"
)

func TestYouTubeExtractorExtractsFromShortURLAndDOMFallbacks(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><head><title>Demo Video - YouTube</title></head><body><div id="description">Line one
Line two</div></body></html>`)
	extractor := NewYouTubeExtractor(doc, "https://youtu.be/abc123", nil)

	result := extractor.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}
	if !strings.Contains(result.ContentHTML, `src="https://www.youtube.com/embed/abc123"`) {
		t.Fatalf("ContentHTML = %q, want iframe with short URL video ID", result.ContentHTML)
	}
	if !strings.Contains(result.ContentHTML, `Line one<br>Line two`) {
		t.Fatalf("ContentHTML = %q, want formatted description with <br>", result.ContentHTML)
	}
	if got := result.Variables["title"]; got != "Demo Video" {
		t.Fatalf("Variables[title] = %q, want %q", got, "Demo Video")
	}
	if got := result.Variables["image"]; got != "https://img.youtube.com/vi/abc123/maxresdefault.jpg" {
		t.Fatalf("Variables[image] = %q, want generated thumbnail", got)
	}
	if got := result.ExtractedContent["videoId"]; got != "abc123" {
		t.Fatalf("ExtractedContent[videoId] = %#v, want %q", got, "abc123")
	}
}

func TestYouTubeExtractorFallsBackWhenVideoIDIsMissing(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><head><title>Ignored - YouTube</title></head><body></body></html>`)
	schema := []any{
		map[string]any{
			"@type":       "VideoObject",
			"name":        "Test Video",
			"description": strings.Repeat("word ", 60),
			"author":      "Test Author",
			"uploadDate":  "2026-04-21",
			"thumbnailUrl": []any{
				"https://cdn.example.com/thumb.jpg",
			},
		},
	}

	extractor := NewYouTubeExtractor(doc, "https://youtube.com/watch?v=", schema)
	result := extractor.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}
	if strings.Contains(result.ContentHTML, `<iframe`) {
		t.Fatalf("ContentHTML = %q, want description-only fallback when video ID is empty", result.ContentHTML)
	}
	if got := result.Variables["title"]; got != "Test Video" {
		t.Fatalf("Variables[title] = %q, want %q", got, "Test Video")
	}
	if got := result.Variables["author"]; got != "Test Author" {
		t.Fatalf("Variables[author] = %q, want %q", got, "Test Author")
	}
	if got := result.Variables["image"]; got != "https://cdn.example.com/thumb.jpg" {
		t.Fatalf("Variables[image] = %q, want %q", got, "https://cdn.example.com/thumb.jpg")
	}
	if got := result.Variables["published"]; got != "2026-04-21" {
		t.Fatalf("Variables[published] = %q, want %q", got, "2026-04-21")
	}
	if got := result.ExtractedContent["videoId"]; got != "" {
		t.Fatalf("ExtractedContent[videoId] = %#v, want empty string", got)
	}
	if got := result.Variables["description"]; len(got) > 200 {
		t.Fatalf("Variables[description] length = %d, want <= 200", len(got))
	}
}
