package markdown

import (
	"strings"
	"testing"
)

func TestConvertHTMLConvertsAndCleansWhitespace(t *testing.T) {
	t.Parallel()

	got, err := ConvertHTML("<p>First</p>\n\n\n<p>Second</p>")
	if err != nil {
		t.Fatalf("ConvertHTML() error = %v", err)
	}
	if strings.TrimSpace(got) != got {
		t.Fatalf("ConvertHTML() = %q, want trimmed output", got)
	}
	if strings.Contains(got, "\n\n\n") {
		t.Fatalf("ConvertHTML() = %q, want excessive newlines removed", got)
	}
	if !strings.Contains(got, "First") || !strings.Contains(got, "Second") {
		t.Fatalf("ConvertHTML() = %q, want both paragraphs converted", got)
	}
}

func TestConvertHTMLEmptyInput(t *testing.T) {
	t.Parallel()

	got, err := ConvertHTML("")
	if err != nil {
		t.Fatalf("ConvertHTML() error = %v", err)
	}
	if got != "" {
		t.Fatalf("ConvertHTML(\"\") = %q, want empty string", got)
	}
}

func TestConvertHTMLPreservesReadableMarkdown(t *testing.T) {
	t.Parallel()

	got, err := ConvertHTML(`<article>
		<h1>Example</h1>
		<p>Read the <a href="https://example.com/docs">docs</a>.</p>
		<blockquote>Quoted text</blockquote>
		<ul><li>First</li><li>Second</li></ul>
		<pre><code class="language-go">fmt.Println("hi")</code></pre>
		<img src="/cover.png" alt="Cover image">
	</article>`)
	if err != nil {
		t.Fatalf("ConvertHTML() error = %v", err)
	}

	checks := []string{
		"# Example",
		"[docs](https://example.com/docs)",
		"> Quoted text",
		"- First",
		"- Second",
		"fmt.Println",
		"![Cover image](/cover.png)",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("ConvertHTML() = %q, want %q", got, check)
		}
	}
}
