package extractors

import (
	"strings"
	"testing"
)

type stubConversationExtractor struct {
	*ConversationExtractorBase
	messages  []ConversationMessage
	metadata  ConversationMetadata
	footnotes []Footnote
}

func (s *stubConversationExtractor) CanExtract() bool {
	return true
}

func (s *stubConversationExtractor) Extract() *ExtractorResult {
	return s.ExtractWithDefuddle(s)
}

func (s *stubConversationExtractor) Name() string {
	return "StubConversationExtractor"
}

func (s *stubConversationExtractor) ExtractMessages() []ConversationMessage {
	return s.messages
}

func (s *stubConversationExtractor) GetMetadata() ConversationMetadata {
	return s.metadata
}

func (s *stubConversationExtractor) GetFootnotes() []Footnote {
	return s.footnotes
}

func TestConversationExtractorBaseCreateContentHTML(t *testing.T) {
	t.Parallel()

	base := NewConversationExtractorBase(newTestDocument(t, `<html><body></body></html>`), "https://claude.ai/share/test", nil)

	html := base.CreateContentHTML([]ConversationMessage{
		{
			Author:    "User",
			Content:   "Hello there",
			Timestamp: "2026-04-21",
			Metadata:  map[string]any{"model": "claude"},
		},
		{
			Author:  "Assistant",
			Content: "<p>Already wrapped</p>",
		},
	}, []Footnote{{URL: "https://example.com/source", Text: "Source"}})

	if !strings.Contains(html, `data-model="claude"`) {
		t.Fatalf("CreateContentHTML() = %q, want metadata attribute", html)
	}
	if !strings.Contains(html, `<p>Hello there</p>`) {
		t.Fatalf("CreateContentHTML() = %q, want wrapped plain text", html)
	}
	if strings.Contains(html, `<p><p>Already wrapped</p></p>`) {
		t.Fatalf("CreateContentHTML() double wrapped paragraph content: %q", html)
	}
	if !strings.Contains(html, `id="fn:1"`) {
		t.Fatalf("CreateContentHTML() = %q, want footnote section", html)
	}
}

func TestConversationExtractorBaseExtractWithDefuddle(t *testing.T) {
	t.Parallel()

	base := NewConversationExtractorBase(newTestDocument(t, `<html><body></body></html>`), "https://claude.ai/share/test", nil)
	extractor := &stubConversationExtractor{
		ConversationExtractorBase: base,
		messages: []ConversationMessage{
			{Author: "User", Content: "Hello"},
			{Author: "Assistant", Content: "Hi"},
		},
		metadata: ConversationMetadata{
			Title: "Test Conversation",
			Site:  "Claude",
		},
	}

	result := base.ExtractWithDefuddle(extractor)
	if result == nil {
		t.Fatal("ExtractWithDefuddle() returned nil")
	}

	messageCount, ok := result.ExtractedContent["messageCount"].(string)
	if !ok || messageCount != "2" {
		t.Fatalf("ExtractedContent[messageCount] = %#v, want %q", result.ExtractedContent["messageCount"], "2")
	}
	if got := result.Variables["title"]; got != "Test Conversation" {
		t.Fatalf("Variables[title] = %q, want %q", got, "Test Conversation")
	}
	if got := result.Variables["site"]; got != "Claude" {
		t.Fatalf("Variables[site] = %q, want %q", got, "Claude")
	}
	if got := result.Variables["description"]; got != "Claude conversation with 2 messages" {
		t.Fatalf("Variables[description] = %q, want %q", got, "Claude conversation with 2 messages")
	}
	if !strings.Contains(result.ContentHTML, `message-user`) || !strings.Contains(result.ContentHTML, `message-assistant`) {
		t.Fatalf("ContentHTML = %q, want both rendered messages", result.ContentHTML)
	}
}
