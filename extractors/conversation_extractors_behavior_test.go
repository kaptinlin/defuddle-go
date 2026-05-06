package extractors

import (
	"strings"
	"testing"
)

func TestChatGPTExtractorFallsBackToFirstUserMessageTitleAndUnknownRole(t *testing.T) {
	t.Parallel()

	longQuestion := "Explain how readable extraction handles nested inline content in detail for reviewers"
	doc := newTestDocument(t, `<html><head><title>ChatGPT</title></head><body>
		<article data-testid="conversation-turn-1"><h5 class="sr-only">You:</h5><div class="text-message">`+longQuestion+`</div></article>
	</body></html>`)
	extractor := NewChatGPTExtractor(doc, "https://chatgpt.com/share/fallback", nil)

	messages := extractor.ExtractMessages()
	if len(messages) != 1 {
		t.Fatalf("messages length = %d, want 1", len(messages))
	}
	if got := messages[0].Metadata["role"]; got != "unknown" {
		t.Fatalf("message role = %#v, want unknown", got)
	}
	if strings.Contains(messages[0].Content, "sr-only") {
		t.Fatalf("message content = %q, want screen-reader heading removed", messages[0].Content)
	}

	metadata := extractor.GetMetadata()
	if got, want := metadata.Title, longQuestion[:50]+"..."; got != want {
		t.Fatalf("metadata title = %q, want %q", got, want)
	}
	if got := metadata.MessageCount; got != 1 {
		t.Fatalf("metadata message count = %d, want 1", got)
	}
}

func TestClaudeExtractorUsesHeaderTitleAndSkipsUnknownMessageBlocks(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><head><title>Claude</title></head><body>
		<header><div class="font-tiempos">Header conversation title</div></header>
		<div data-testid="user-message"><p>Visible user message.</p></div>
		<div data-testid="tool-output"><p>Hidden tool output.</p></div>
		<div data-testid="assistant-message"><p>Visible assistant message.</p></div>
	</body></html>`)
	extractor := NewClaudeExtractor(doc, "https://claude.ai/share/header", nil)

	messages := extractor.ExtractMessages()
	if len(messages) != 2 {
		t.Fatalf("messages length = %d, want 2", len(messages))
	}
	if got := messages[0].Author; got != "You" {
		t.Fatalf("first author = %q, want You", got)
	}
	if got := messages[1].Author; got != "Claude" {
		t.Fatalf("second author = %q, want Claude", got)
	}
	for _, message := range messages {
		if strings.Contains(message.Content, "Hidden tool output") {
			t.Fatalf("message content = %q, want tool output skipped", message.Content)
		}
	}

	metadata := extractor.GetMetadata()
	if got := metadata.Title; got != "Header conversation title" {
		t.Fatalf("metadata title = %q, want header title", got)
	}
}

func TestGrokExtractorFallsBackToFirstUserMessageTitleAndLeavesNonHTTPLinks(t *testing.T) {
	t.Parallel()

	question := "How should we decide whether additional tests are useful or just coverage noise?"
	doc := newTestDocument(t, `<html><head><title>Grok by xAI</title></head><body>
		<div class="relative group flex flex-col justify-center w-full items-end"><div class="message-bubble">`+question+`</div></div>
		<div class="relative group flex flex-col justify-center w-full items-start"><div class="message-bubble"><p>Use behavior. <a href="#local">local note</a> <a href="mailto:test@example.com">mail</a></p></div></div>
	</body></html>`)
	extractor := NewGrokExtractor(doc, "https://grok.x.ai/share/title", nil)

	messages := extractor.ExtractMessages()
	if len(messages) != 2 {
		t.Fatalf("messages length = %d, want 2", len(messages))
	}
	if got := len(extractor.GetFootnotes()); got != 0 {
		t.Fatalf("footnotes length = %d, want 0 for non-HTTP links", got)
	}
	if !strings.Contains(messages[1].Content, `href="#local"`) || !strings.Contains(messages[1].Content, `mailto:test@example.com`) {
		t.Fatalf("assistant content = %q, want non-HTTP links preserved", messages[1].Content)
	}

	metadata := extractor.GetMetadata()
	if got, want := metadata.Title, question[:50]+"..."; got != want {
		t.Fatalf("metadata title = %q, want %q", got, want)
	}
}

func TestGeminiExtractorUsesPageTitleExtendedResponseAndDomainOnlySources(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><head><title>Independent research notes</title></head><body>
		<browse-item><a href="https://example.com/domain-only"><span class="domain">example.com</span></a></browse-item>
		<div class="conversation-container">
			<user-query><div class="query-text">Compare parser options.</div></user-query>
			<model-response>
				<div class="model-response-text"><div class="markdown"><p>Regular response</p></div></div>
				<div id="extended-response-markdown-content"><p>Extended response</p></div>
			</model-response>
		</div>
	</body></html>`)
	extractor := NewGeminiExtractor(doc, "https://gemini.google.com/app/extended", nil)

	messages := extractor.ExtractMessages()
	if len(messages) != 2 {
		t.Fatalf("messages length = %d, want 2", len(messages))
	}
	if !strings.Contains(messages[1].Content, "Extended response") {
		t.Fatalf("assistant content = %q, want extended response", messages[1].Content)
	}
	if strings.Contains(messages[1].Content, "Regular response") {
		t.Fatalf("assistant content = %q, want regular response ignored when extended content exists", messages[1].Content)
	}

	footnotes := extractor.GetFootnotes()
	if len(footnotes) != 1 {
		t.Fatalf("footnotes length = %d, want 1", len(footnotes))
	}
	if got := footnotes[0].Text; got != "example.com" {
		t.Fatalf("footnote text = %q, want domain-only text", got)
	}

	metadata := extractor.GetMetadata()
	if got := metadata.Title; got != "Independent research notes" {
		t.Fatalf("metadata title = %q, want page title", got)
	}
	if got := metadata.MessageCount; got != 2 {
		t.Fatalf("metadata message count = %d, want cached extracted count", got)
	}
}
