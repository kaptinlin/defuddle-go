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
