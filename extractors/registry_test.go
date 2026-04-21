package extractors

import (
	"testing"

	"github.com/PuerkitoBio/goquery"
)

type stubRegistryExtractor struct {
	*ExtractorBase
}

func (s *stubRegistryExtractor) CanExtract() bool {
	return true
}

func (s *stubRegistryExtractor) Extract() *ExtractorResult {
	return &ExtractorResult{Content: "ok", ContentHTML: "ok"}
}

func (s *stubRegistryExtractor) Name() string {
	return "StubRegistryExtractor"
}

func TestInitializeBuiltinsDoesNotPanic(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("initializeBuiltins() panicked: %v", recovered)
		}
	}()

	registry.initializeBuiltins()

	if len(registry.GetMappings()) == 0 {
		t.Fatal("initializeBuiltins() did not register built-in extractors")
	}
}

func TestRegistryFindExtractorMatchesDomainAndUsesCache(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	calls := 0
	registry.Register(ExtractorMapping{
		Patterns: []any{"example.com"},
		Extractor: func(doc *goquery.Document, url string, schemaOrgData any) BaseExtractor {
			calls++
			return &stubRegistryExtractor{ExtractorBase: NewExtractorBase(doc, url, schemaOrgData)}
		},
	})

	doc := newTestDocument(t, `<html><body></body></html>`)
	first := registry.FindExtractor(doc, "https://www.example.com/post", map[string]any{"k": "v"})
	if first == nil {
		t.Fatal("FindExtractor() returned nil for matching domain")
	}
	if calls != 1 {
		t.Fatalf("constructor calls = %d, want 1 after first match", calls)
	}

	second := registry.FindExtractor(doc, "https://www.example.com/another", nil)
	if second == nil {
		t.Fatal("FindExtractor() returned nil for cached domain")
	}
	if calls != 2 {
		t.Fatalf("constructor calls = %d, want 2 because cache stores constructor", calls)
	}
}

func TestRegistryFindExtractorHandlesRegexMissAndInvalidURL(t *testing.T) {
	t.Parallel()

	registry := NewRegistry().Register(ExtractorMapping{
		Patterns: []any{youtubeWatchPattern},
		Extractor: func(doc *goquery.Document, url string, schemaOrgData any) BaseExtractor {
			return &stubRegistryExtractor{ExtractorBase: NewExtractorBase(doc, url, schemaOrgData)}
		},
	})

	doc := newTestDocument(t, `<html><body></body></html>`)
	if got := registry.FindExtractor(doc, "https://youtube.com/watch?v=abc", nil); got == nil {
		t.Fatal("FindExtractor() returned nil for regex match")
	}
	if got := registry.FindExtractor(doc, "https://nomatch.example.net", nil); got != nil {
		t.Fatalf("FindExtractor() = %#v, want nil for unmatched URL", got)
	}
	if got := registry.FindExtractor(doc, "://bad-url", nil); got != nil {
		t.Fatalf("FindExtractor() = %#v, want nil for invalid URL", got)
	}
}
