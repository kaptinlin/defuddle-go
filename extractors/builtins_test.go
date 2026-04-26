package extractors

import (
	"strings"
	"testing"
)

func TestGitHubExtractorExtractsIssueAndComments(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><head><title>kaptinlin/defuddle-go: Test issue</title></head><body>
		<meta name="expected-hostname" content="github.com">
		<div data-testid="issue-title">Test issue</div>
		<div data-testid="issue-viewer-issue-container">
			<a data-testid="issue-body-header-author">alice</a>
			<relative-time datetime="2026-04-21T12:00:00Z"></relative-time>
			<div data-testid="issue-body-viewer"><div class="markdown-body"><p>Issue body</p><task-lists><li>task</li></task-lists></div></div>
		</div>
		<div data-wrapper-timeline-id="comment-1">
			<div class="react-issue-comment">
				<a data-testid="avatar-link">bob</a>
				<relative-time datetime="2026-04-22T12:00:00Z"></relative-time>
				<div class="markdown-body"><p>Comment body</p></div>
			</div>
		</div>
	</body></html>`)
	extractor := NewGitHubExtractor(doc, "https://github.com/kaptinlin/defuddle-go/issues/123", nil)

	if !extractor.CanExtract() {
		t.Fatal("CanExtract() = false, want true")
	}
	result := extractor.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}
	if !strings.Contains(result.ContentHTML, "Issue body") || !strings.Contains(result.ContentHTML, "Comment body") {
		t.Fatalf("ContentHTML = %q, want issue and comment bodies", result.ContentHTML)
	}
	if got := result.ExtractedContent["owner"]; got != "kaptinlin" {
		t.Fatalf("ExtractedContent[owner] = %#v, want %q", got, "kaptinlin")
	}
	if got := result.ExtractedContent["repository"]; got != "defuddle-go" {
		t.Fatalf("ExtractedContent[repository] = %#v, want %q", got, "defuddle-go")
	}
	if got := result.ExtractedContent["issueNumber"]; got != "123" {
		t.Fatalf("ExtractedContent[issueNumber] = %#v, want %q", got, "123")
	}
	if got := result.Variables["site"]; got != "GitHub - kaptinlin/defuddle-go" {
		t.Fatalf("Variables[site] = %q, want GitHub repo site", got)
	}
}

func TestRedditExtractorExtractsPostAndNestedComments(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><body>
		<h1>Reddit title</h1>
		<shreddit-post author="poster"><div slot="text-body"><p>Post body</p></div><div id="post-image"><img src="post.jpg"></div></shreddit-post>
		<shreddit-comment author="commenter" score="7" permalink="/r/golang/comments/abc/test/comment1" depth="0"><faceplate-timeago ts="1776844800"></faceplate-timeago><div slot="comment"><p>First comment</p></div></shreddit-comment>
		<shreddit-comment author="reply" score="3" permalink="/r/golang/comments/abc/test/comment2" depth="1"><div slot="comment"><p>Nested reply</p></div></shreddit-comment>
	</body></html>`)
	extractor := NewRedditExtractor(doc, "https://www.reddit.com/r/golang/comments/abc/test_post/", nil)

	if !extractor.CanExtract() {
		t.Fatal("CanExtract() = false, want true")
	}
	result := extractor.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}
	for _, want := range []string{"Post body", "First comment", "Nested reply", `<div class="reddit-comments">`} {
		if !strings.Contains(result.ContentHTML, want) {
			t.Fatalf("ContentHTML = %q, want %q", result.ContentHTML, want)
		}
	}
	if got := result.ExtractedContent["postId"]; got != "abc" {
		t.Fatalf("ExtractedContent[postId] = %#v, want %q", got, "abc")
	}
	if got := result.ExtractedContent["subreddit"]; got != "golang" {
		t.Fatalf("ExtractedContent[subreddit] = %#v, want %q", got, "golang")
	}
	if got := result.Variables["author"]; got != "poster" {
		t.Fatalf("Variables[author] = %q, want %q", got, "poster")
	}
}

func TestTwitterExtractorExtractsThreadTextMediaAndMetadata(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><body><main role="main">
		<article data-testid="tweet">
			<div data-testid="User-Name"><a>Alice Example</a><a>alice</a></div>
			<a href="/alice/status/123"><time datetime="2026-04-21T12:00:00Z"></time></a>
			<div data-testid="tweetText"><span>Hello</span> <a href="/bob">@bob</a></div>
			<img src="https://pbs.twimg.com/media/photo.jpg?format=jpg&name=small" alt=" A photo ">
		</article>
		<article data-testid="tweet">
			<div data-testid="User-Name"><a>Alice Example</a><a>@alice</a></div>
			<div data-testid="tweetText">Thread reply</div>
		</article>
	</main></body></html>`)
	extractor := NewTwitterExtractor(doc, "https://x.com/alice/status/123", nil)

	if !extractor.CanExtract() {
		t.Fatal("CanExtract() = false, want true")
	}
	result := extractor.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}
	for _, want := range []string{"tweet-thread", "Hello @bob", "Thread reply", "name=large"} {
		if !strings.Contains(result.ContentHTML, want) {
			t.Fatalf("ContentHTML = %q, want %q", result.ContentHTML, want)
		}
	}
	if got := result.ExtractedContent["tweetId"]; got != "123" {
		t.Fatalf("ExtractedContent[tweetId] = %#v, want %q", got, "123")
	}
	if got := result.Variables["author"]; got != "@alice" {
		t.Fatalf("Variables[author] = %q, want %q", got, "@alice")
	}
	if got := result.Variables["site"]; got != "X (Twitter)" {
		t.Fatalf("Variables[site] = %q, want X site", got)
	}
}

func TestChatGPTExtractorExtractsMessagesAndFootnotes(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><head><title>Research chat</title></head><body>
		<article data-testid="conversation-turn-1" data-message-author-role="user"><h5 class="sr-only">You:</h5><div class="text-message">What is Go?</div></article>
		<article data-testid="conversation-turn-2" data-message-author-role="assistant"><h6 class="sr-only">ChatGPT:</h6><p>Go is a language <span><a href="https://example.com/page#:~:text=Go,language" target="_blank" rel="noopener">source</a></span></p><p>   </p><span data-state="closed">copy</span></article>
	</body></html>`)
	extractor := NewChatGPTExtractor(doc, "https://chatgpt.com/share/test", nil)

	if !extractor.CanExtract() {
		t.Fatal("CanExtract() = false, want true")
	}
	result := extractor.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}
	if !strings.Contains(result.ContentHTML, "What is Go?") || !strings.Contains(result.ContentHTML, `id="fn:1"`) {
		t.Fatalf("ContentHTML = %q, want messages and footnotes", result.ContentHTML)
	}
	if strings.Contains(result.ContentHTML, "copy") {
		t.Fatalf("ContentHTML = %q, want closed controls removed", result.ContentHTML)
	}
	if got := result.ExtractedContent["messageCount"]; got != "2" {
		t.Fatalf("ExtractedContent[messageCount] = %#v, want %q", got, "2")
	}
	if got := result.Variables["title"]; got != "Research chat" {
		t.Fatalf("Variables[title] = %q, want page title", got)
	}
}

func TestClaudeExtractorExtractsUserAndAssistantMessages(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><head><title>Plan discussion - Claude</title></head><body>
		<div data-testid="user-message"><p>Please draft a plan.</p></div>
		<div data-testid="assistant-message"><p>Here is the plan.</p></div>
	</body></html>`)
	extractor := NewClaudeExtractor(doc, "https://claude.ai/share/test", nil)

	if !extractor.CanExtract() {
		t.Fatal("CanExtract() = false, want true")
	}
	result := extractor.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}
	if !strings.Contains(result.ContentHTML, "Please draft a plan.") || !strings.Contains(result.ContentHTML, "Here is the plan.") {
		t.Fatalf("ContentHTML = %q, want both Claude messages", result.ContentHTML)
	}
	if got := result.ExtractedContent["messageCount"]; got != "2" {
		t.Fatalf("ExtractedContent[messageCount] = %#v, want %q", got, "2")
	}
	if got := result.Variables["title"]; got != "Plan discussion" {
		t.Fatalf("Variables[title] = %q, want suffix-trimmed title", got)
	}
}

func TestGrokExtractorExtractsMessagesAndDeduplicatesFootnotes(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><head><title>Grok exchange - Grok</title></head><body>
		<div class="relative group flex flex-col justify-center w-full items-end"><div class="message-bubble">User question</div></div>
		<div class="relative group flex flex-col justify-center w-full items-start"><div class="message-bubble"><p>Answer with <a href="https://example.com/a">source</a> and <a href="https://example.com/a">again</a>.</p><div class="relative border border-border-l1 bg-surface-base">artifact</div></div></div>
	</body></html>`)
	extractor := NewGrokExtractor(doc, "https://grok.x.ai/share/test", nil)

	if !extractor.CanExtract() {
		t.Fatal("CanExtract() = false, want true")
	}
	result := extractor.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}
	if !strings.Contains(result.ContentHTML, "User question") || !strings.Contains(result.ContentHTML, "source") || !strings.Contains(result.ContentHTML, `id="fn:1"`) {
		t.Fatalf("ContentHTML = %q, want user, assistant, and footnote references", result.ContentHTML)
	}
	if strings.Contains(result.ContentHTML, "artifact") {
		t.Fatalf("ContentHTML = %q, want DeepSearch artifact removed", result.ContentHTML)
	}
	if got := result.ExtractedContent["messageCount"]; got != "2" {
		t.Fatalf("ExtractedContent[messageCount] = %#v, want %q", got, "2")
	}
	if got := result.Variables["site"]; got != "Grok" {
		t.Fatalf("Variables[site] = %q, want Grok", got)
	}
}

func TestGeminiExtractorExtractsMessagesSourcesAndKeepsTableContent(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><head><title>Gemini</title></head><body>
		<div class="title-text">Research title</div>
		<browse-item><a href="https://example.com/source"><span class="domain">example.com</span><span class="title">Source title</span></a></browse-item>
		<div class="conversation-container">
			<user-query><div class="query-text">Summarize this</div></user-query>
			<model-response><div class="model-response-text"><div class="markdown"><div class="table-content">Table body</div></div></div></model-response>
		</div>
	</body></html>`)
	extractor := NewGeminiExtractor(doc, "https://gemini.google.com/app/test", nil)

	if !extractor.CanExtract() {
		t.Fatal("CanExtract() = false, want true")
	}
	result := extractor.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}
	if !strings.Contains(result.ContentHTML, "Summarize this") || !strings.Contains(result.ContentHTML, "Table body") || !strings.Contains(result.ContentHTML, `id="fn:1"`) {
		t.Fatalf("ContentHTML = %q, want messages and source footnotes", result.ContentHTML)
	}
	if strings.Contains(result.ContentHTML, "table-content") {
		t.Fatalf("ContentHTML = %q, want table-content class removed but content preserved", result.ContentHTML)
	}
	if got := result.ExtractedContent["messageCount"]; got != "2" {
		t.Fatalf("ExtractedContent[messageCount] = %#v, want %q", got, "2")
	}
	if got := result.Variables["title"]; got != "Research title" {
		t.Fatalf("Variables[title] = %q, want research title", got)
	}
}

func TestHackerNewsExtractorExtractsPostAndNestedComments(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><body><table class="fatitem">
		<tr class="athing" id="123"><td class="title"><span class="titleline"><a href="https://example.com/article">Example article</a></span></td></tr>
		<tr><td class="subtext"><span class="score">42 points</span> by <a class="hnuser">alice</a> <span class="age" title="2026-04-21T12:00:00Z"><a>1 day ago</a></span></td></tr>
		<tr><td><div class="toptext"><p>Post text</p></div></td></tr>
	</table>
	<table>
		<tr class="comtr" id="456"><td class="ind"><img width="0"></td><td><span class="comhead"><a class="hnuser">bob</a> <span class="age" title="2026-04-22T12:00:00Z"></span></span><div class="commtext"><p>First comment</p></div></td></tr>
		<tr class="comtr" id="457"><td class="ind"><img width="40"></td><td><span class="comhead"><a class="hnuser">carol</a> <span class="age" title="2026-04-22T13:00:00Z"></span></span><div class="commtext"><p>Nested comment</p></div></td></tr>
	</table></body></html>`)
	extractor := NewHackerNewsExtractor(doc, "https://news.ycombinator.com/item?id=123", nil)

	if !extractor.CanExtract() {
		t.Fatal("CanExtract() = false, want true")
	}
	result := extractor.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}
	for _, want := range []string{"https://example.com/article", "Post text", "First comment", "Nested comment", "<blockquote>"} {
		if !strings.Contains(result.ContentHTML, want) {
			t.Fatalf("ContentHTML = %q, want %q", result.ContentHTML, want)
		}
	}
	if got := result.ExtractedContent["postId"]; got != "123" {
		t.Fatalf("ExtractedContent[postId] = %#v, want %q", got, "123")
	}
	if got := result.ExtractedContent["postAuthor"]; got != "alice" {
		t.Fatalf("ExtractedContent[postAuthor] = %#v, want %q", got, "alice")
	}
	if got := result.Variables["title"]; got != "Example article" {
		t.Fatalf("Variables[title] = %q, want article title", got)
	}
	if got := result.Variables["published"]; got != "2026-04-21" {
		t.Fatalf("Variables[published] = %q, want post date", got)
	}
}

func TestHackerNewsExtractorExtractsCommentPage(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><body><table class="fatitem">
		<tr><td class="navs"><a href="item?id=100&parent=456">parent</a></td></tr>
		<tr class="comtr" id="456"><td class="ind"><img width="0"></td><td><div class="comment"><span class="score">3 points</span> <a class="hnuser">commenter</a> <span class="age" title="2026-04-22T12:00:00Z"></span><div class="commtext"><p>Main comment content that is long enough for a title preview.</p></div></div></td></tr>
	</table></body></html>`)
	extractor := NewHackerNewsExtractor(doc, "https://news.ycombinator.com/item?id=456", nil)

	if !extractor.CanExtract() {
		t.Fatal("CanExtract() = false, want true")
	}
	result := extractor.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}
	for _, want := range []string{"main-comment", "commenter", "Main comment content", `href="https://news.ycombinator.com/item?id=100&parent=456"`} {
		if !strings.Contains(result.ContentHTML, want) {
			t.Fatalf("ContentHTML = %q, want %q", result.ContentHTML, want)
		}
	}
	if got := result.ExtractedContent["postId"]; got != "456" {
		t.Fatalf("ExtractedContent[postId] = %#v, want %q", got, "456")
	}
	if got := result.Variables["title"]; !strings.HasPrefix(got, "Comment by commenter: Main comment content") {
		t.Fatalf("Variables[title] = %q, want comment preview title", got)
	}
	if got := result.Variables["description"]; got != "Comment by commenter on Hacker News" {
		t.Fatalf("Variables[description] = %q, want comment page description", got)
	}
}
