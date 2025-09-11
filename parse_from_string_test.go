package defuddle

import (
	"context"
	"testing"
)

func TestParseFromString(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
	<meta name="description" content="This is a test page">
</head>
<body>
	<h1>Main Heading</h1>
	<p>This is the main content of the test page.</p>
	<p>Another paragraph with more content.</p>
</body>
</html>
`

	options := &Options{
		Markdown: true,
		URL:      "https://example.com/test",
	}

	result, err := ParseFromString(context.Background(), html, options)
	if err != nil {
		t.Fatalf("ParseFromString failed: %v", err)
	}

	// Check basic fields
	if result.Title == "" {
		t.Error("Expected title to be extracted")
	}

	if result.Content == "" {
		t.Error("Expected content to be extracted")
	}

	if result.ContentMarkdown == nil || *result.ContentMarkdown == "" {
		t.Error("Expected markdown content to be generated")
	}

	// Check that domain is extracted
	if result.Domain != "example.com" {
		t.Errorf("Expected domain to be 'example.com', got '%s'", result.Domain)
	}

	t.Logf("Title: %s", result.Title)
	t.Logf("Content length: %d", len(result.Content))
	t.Logf("Markdown length: %d", len(*result.ContentMarkdown))
}

func TestParseFromStringWithoutOptions(t *testing.T) {
	html := `<html><body><h1>Simple Test</h1><p>Content</p></body></html>`

	result, err := ParseFromString(context.Background(), html, nil)
	if err != nil {
		t.Fatalf("ParseFromString with nil options failed: %v", err)
	}

	if result.Content == "" {
		t.Error("Expected content to be extracted even with nil options")
	}
}
