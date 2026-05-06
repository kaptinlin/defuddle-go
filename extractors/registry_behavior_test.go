package extractors

import (
	"regexp"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestRegistryClearCacheAllowsUpdatedMappingToApply(t *testing.T) {
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

	if got := registry.FindExtractor(doc, "https://example.com/one", nil); got == nil {
		t.Fatal("FindExtractor() returned nil before cache clear")
	}
	registry.ClearCache()
	if got := registry.FindExtractor(doc, "https://example.com/two", nil); got == nil {
		t.Fatal("FindExtractor() returned nil after cache clear")
	}
	if calls != 2 {
		t.Fatalf("constructor calls = %d, want 2", calls)
	}
}

func TestRegistryPatternMatchingSupportsSubdomainsAndRegexes(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if !registry.matchesPatterns("https://docs.example.com/post", "docs.example.com", []any{"example.com"}) {
		t.Fatal("matchesPatterns() did not match subdomain")
	}
	if !registry.matchesPatterns("https://example.net/articles/42", "example.net", []any{regexp.MustCompile(`/articles/\d+$`)}) {
		t.Fatal("matchesPatterns() did not match regex")
	}
	if registry.matchesPatterns("https://example.net/post", "example.net", []any{42}) {
		t.Fatal("matchesPatterns() matched unsupported pattern type")
	}
}

func TestDefaultRegistryConvenienceFunctions(t *testing.T) {
	// DefaultRegistry is global state, so this test must not run in parallel.
	doc := newTestDocument(t, `<html><body></body></html>`)
	pattern := regexp.MustCompile(`https://coverage-helper\.example/.*`)

	Register(ExtractorMapping{
		Patterns: []any{pattern},
		Extractor: func(doc *goquery.Document, url string, schemaOrgData any) BaseExtractor {
			return &stubRegistryExtractor{ExtractorBase: NewExtractorBase(doc, url, schemaOrgData)}
		},
	})
	ClearCache()

	if got := FindExtractor(doc, "https://coverage-helper.example/article", nil); got == nil {
		t.Fatal("FindExtractor() returned nil for default registry mapping")
	}
}
