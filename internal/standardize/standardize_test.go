package standardize

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	internalmetadata "github.com/kaptinlin/defuddle-go/internal/metadata"
)

func newStandardizeDocument(t *testing.T, html string) *goquery.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("goquery.NewDocumentFromReader() error = %v", err)
	}

	return doc
}

func TestContentStandardizesSemanticStructure(t *testing.T) {
	t.Parallel()

	doc := newStandardizeDocument(t, `<html><body><article>
		<h1>Example Title</h1>
		<div role="paragraph" id="intro">Intro text</div>
		<div role="list" id="steps">
			<div role="listitem">
				<span class="label">1)</span>
				<div class="content"><div role="paragraph">First item</div></div>
			</div>
		</div>
		<p>Body<a class="footnote-backref" href="#fnref:1">↩</a></p>
		<h3>Trailing heading</h3>
	</article></body></html>`)
	article := doc.Find("article").First()

	Content(article, &internalmetadata.Metadata{Title: "Example Title"}, doc, false)

	if article.Find("h1, h2, h3").Length() != 0 {
		t.Fatalf("Content() left headings behind: %q", article.Text())
	}
	if !strings.Contains(article.Text(), "Intro text") {
		t.Fatalf("Content() removed paragraph text: %q", article.Text())
	}
	if !strings.Contains(article.Text(), "First item") {
		t.Fatalf("Content() removed list item text: %q", article.Text())
	}
	if article.Find("ol li").Length() != 1 {
		t.Fatalf("Content() did not convert role list to ordered list")
	}
	if article.Find(".footnote-backref").Length() != 0 {
		t.Fatal("Content() did not remove footnote back-reference")
	}
	if article.Find(`[role]`).Length() != 0 {
		t.Fatal("Content() left role attributes behind")
	}
}

func TestContentDebugModePreservesWrapperDivs(t *testing.T) {
	t.Parallel()

	doc := newStandardizeDocument(t, `<html><body><article><div class="wrapper"><p>Wrapped text</p></div></article></body></html>`)
	article := doc.Find("article").First()

	Content(article, &internalmetadata.Metadata{}, doc, true)

	if article.Find("div").Length() == 0 {
		t.Fatal("Content() in debug mode removed wrapper divs")
	}
}
