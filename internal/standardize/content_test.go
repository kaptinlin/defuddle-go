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

	doc := newStandardizeDocument(t, `<html><body><article id="content" class="root" data-score="17"><div class="wrapper" data-step="keep"><p>Wrapped text</p></div></article></body></html>`)
	article := doc.Find("article").First()

	Content(article, &internalmetadata.Metadata{}, doc, true)

	if article.Find("div").Length() == 0 {
		t.Fatal("Content() in debug mode removed wrapper divs")
	}
	if got := article.AttrOr("id", ""); got != "content" {
		t.Fatalf("Content() in debug mode removed id = %q, want %q", got, "content")
	}
	if got := article.AttrOr("class", ""); got != "root" {
		t.Fatalf("Content() in debug mode removed class = %q, want %q", got, "root")
	}
	if got := article.AttrOr("data-score", ""); got != "17" {
		t.Fatalf("Content() in debug mode removed data-score = %q, want %q", got, "17")
	}
}

func TestContentStripsUnwantedAttributesAndPreservesSpecialCases(t *testing.T) {
	t.Parallel()

	doc := newStandardizeDocument(t, `<html><body><article class="root" data-score="17"><p id="fn:1" data-extra="removed"><a href="https://example.com" onclick="evil()" data-extra="removed">source</a><code class="language-go" onclick="evil()">fmt.Println()</code></p></article></body></html>`)
	article := doc.Find("article").First()

	Content(article, &internalmetadata.Metadata{}, doc, false)

	if _, exists := article.Attr("class"); exists {
		t.Fatal("Content() kept class on article in normal mode")
	}
	if _, exists := article.Attr("data-score"); exists {
		t.Fatal("Content() kept data-score on article in normal mode")
	}
	if article.Find(`[id="fn:1"]`).Length() != 1 {
		t.Fatal("Content() removed footnote id")
	}
	if _, exists := article.Find("p").Attr("data-extra"); exists {
		t.Fatal("Content() kept data-extra on paragraph in normal mode")
	}
	link := article.Find("a").First()
	if got := link.AttrOr("href", ""); got != "https://example.com" {
		t.Fatalf("Content() link href = %q, want preserved href", got)
	}
	if _, exists := link.Attr("onclick"); exists {
		t.Fatal("Content() kept onclick on link")
	}
	code := article.Find("code").First()
	if got := code.AttrOr("class", ""); got != "language-go" {
		t.Fatalf("Content() code class = %q, want language-go", got)
	}
	if _, exists := code.Attr("onclick"); exists {
		t.Fatal("Content() kept onclick on code")
	}
}

func TestContentConvertsLiteYouTubeAndLimitsConsecutiveBreaks(t *testing.T) {
	t.Parallel()

	doc := newStandardizeDocument(t, `<html><body><article><p>Before</p><lite-youtube videoid="abc123" videotitle="Demo video"></lite-youtube><p>After<br><br><br><br>Breaks</p></article></body></html>`)
	article := doc.Find("article").First()

	Content(article, &internalmetadata.Metadata{}, doc, false)

	if article.Find("lite-youtube").Length() != 0 {
		t.Fatal("Content() left lite-youtube element behind")
	}
	iframe := article.Find("iframe").First()
	if iframe.Length() == 0 {
		t.Fatal("Content() did not convert lite-youtube to iframe")
	}
	if got := iframe.AttrOr("src", ""); got != "https://www.youtube.com/embed/abc123" {
		t.Fatalf("Content() iframe src = %q, want YouTube embed URL", got)
	}
	if got := iframe.AttrOr("title", ""); got != "Demo video" {
		t.Fatalf("Content() iframe title = %q, want %q", got, "Demo video")
	}
	if got := article.Find("br").Length(); got != 2 {
		t.Fatalf("Content() kept %d consecutive br elements, want 2", got)
	}
}

func TestContentNormalizesTextButPreservesPreAndCode(t *testing.T) {
	t.Parallel()

	doc := newStandardizeDocument(t, `<html><body><article><p>Alpha   beta&#8204; gamma   , done</p><pre>one&nbsp;&nbsp; two</pre><code>fmt  .Println</code></article></body></html>`)
	article := doc.Find("article").First()

	Content(article, &internalmetadata.Metadata{}, doc, false)

	if got := article.Find("p").First().Text(); got != "Alpha beta gamma, done" {
		t.Fatalf("Content() paragraph text = %q, want normalized text", got)
	}
	if got := article.Find("pre").First().Text(); got != "one   two" {
		t.Fatalf("Content() pre text = %q, want preserved pre text", got)
	}
	if got := article.Find("code").First().Text(); got != "fmt  .Println" {
		t.Fatalf("Content() code text = %q, want preserved code text", got)
	}
}
