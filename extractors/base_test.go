package extractors

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func newTestDocument(t *testing.T, html string) *goquery.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("goquery.NewDocumentFromReader() error = %v", err)
	}

	return doc
}

func TestExtractorBaseGettersAndHelpers(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><body><div id="root" data-id="123"><span>Hello</span></div></body></html>`)
	base := NewExtractorBase(doc, "https://example.com/path", map[string]any{"type": "example"})

	if base.GetDocument() != doc {
		t.Fatal("GetDocument() did not return the original document")
	}

	if got := base.GetURL(); got != "https://example.com/path" {
		t.Fatalf("GetURL() = %q, want %q", got, "https://example.com/path")
	}

	schema, ok := base.GetSchemaOrgData().(map[string]any)
	if !ok || schema["type"] != "example" {
		t.Fatalf("GetSchemaOrgData() = %#v, want schema map", base.GetSchemaOrgData())
	}

	root := doc.Find("#root").First()
	if got := base.GetTextContent(root.Find("span").First()); got != "Hello" {
		t.Fatalf("GetTextContent() = %q, want %q", got, "Hello")
	}

	if got := base.GetHTMLContent(root); !strings.Contains(got, `<span>Hello</span>`) {
		t.Fatalf("GetHTMLContent() = %q, want span HTML", got)
	}

	if got := base.GetAttribute(root, "data-id"); got != "123" {
		t.Fatalf("GetAttribute() = %q, want %q", got, "123")
	}

	empty := doc.Find(".missing")
	if got := base.GetTextContent(empty); got != "" {
		t.Fatalf("GetTextContent(empty) = %q, want empty string", got)
	}
	if got := base.GetHTMLContent(empty); got != "" {
		t.Fatalf("GetHTMLContent(empty) = %q, want empty string", got)
	}
	if got := base.GetAttribute(empty, "data-id"); got != "" {
		t.Fatalf("GetAttribute(empty) = %q, want empty string", got)
	}
}
