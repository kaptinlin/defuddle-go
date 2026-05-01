package scoring

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func newScoringDocument(t *testing.T, html string) *goquery.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("goquery.NewDocumentFromReader() error = %v", err)
	}

	return doc
}

func TestNewContentScorerAndScoreElementFavorMainContent(t *testing.T) {
	t.Parallel()

	doc := newScoringDocument(t, `<html><body>
		<article id="content" class="article content" style="text-align: right">
			<p>By Jane Doe</p>
			<p>Jan 2, 2024 this is the main article body with enough words to look like content and clearly outrank the navigation block.</p>
			<p>Another paragraph adds more substance to the document.</p>
		</article>
		<div id="nav" class="sidebar navigation">
			<a href="/home">Home</a>
			<a href="/news">News</a>
			<a href="/login">Login</a>
		</div>
	</body></html>`)

	if scorer := NewContentScorer(doc, true); scorer == nil {
		t.Fatal("NewContentScorer() returned nil")
	}

	contentScore := ScoreElement(doc.Find("#content").First())
	navScore := ScoreElement(doc.Find("#nav").First())
	if contentScore <= navScore {
		t.Fatalf("ScoreElement(content) = %v, ScoreElement(nav) = %v, want content score to be higher", contentScore, navScore)
	}
}

func TestFindBestElementRespectsThreshold(t *testing.T) {
	t.Parallel()

	doc := newScoringDocument(t, `<html><body>
		<div id="weak">tiny text</div>
		<div id="best" class="content"><p>This block has enough text to be selected as the best element.</p><p>It also has multiple paragraphs.</p></div>
	</body></html>`)

	elements := []*goquery.Selection{
		doc.Find("#weak").First(),
		doc.Find("#best").First(),
	}

	best := FindBestElement(elements, 0)
	if best == nil || best.AttrOr("id", "") != "best" {
		t.Fatalf("FindBestElement() = %#v, want #best", best)
	}
	if got := FindBestElement(elements, 1_000); got != nil {
		t.Fatalf("FindBestElement() with high threshold = %#v, want nil", got)
	}
}

func TestScoreAndRemoveRemovesNavigationButKeepsContent(t *testing.T) {
	t.Parallel()

	doc := newScoringDocument(t, `<html><body>
		<div id="nav" class="sidebar navigation">
			<ul>
				<li><a href="/home">home</a></li>
				<li><a href="/popular">popular</a></li>
				<li><a href="/subscribe">subscribe</a></li>
				<li><a href="/privacy">privacy</a></li>
			</ul>
			<p>menu navigation newsletter related trending popular subscribe privacy</p>
		</div>
		<article id="article" role="article">
			<p>`+strings.Repeat(`useful content `, 25)+`</p>
			<p>This second paragraph keeps the main article clearly content-like.</p>
		</article>
	</body></html>`)

	ScoreAndRemove(doc, false)

	if doc.Find("#nav").Length() != 0 {
		t.Fatalf("ScoreAndRemove() did not remove navigation block: %q", doc.Find("body").Text())
	}
	if doc.Find("#article").Length() != 1 {
		t.Fatal("ScoreAndRemove() removed the main article")
	}
}

func TestScoreAndRemoveKeepsFootnotesAndOldTableContent(t *testing.T) {
	t.Parallel()

	doc := newScoringDocument(t, `<html><body>
		<table width="640" align="center"><tr>
			<td id="left-nav"><a href="/home">home</a></td>
			<td id="story"><p>By Jane Doe</p><p>`+strings.Repeat(`old table article content `, 12)+`</p><a href="#fn1" class="footnote-ref">1</a></td>
			<td id="right-nav"><a href="/ads">ads</a></td>
		</tr></table>
		<section id="notes"><ol class="footnotes"><li id="fn1">Footnote source</li></ol></section>
		<div id="related" class="related"><a href="/one">one</a><a href="/two">two</a><a href="/three">three</a><p>related navigation subscribe popular</p></div>
	</body></html>`)

	leftScore := ScoreElement(doc.Find("#left-nav").First())
	storyScore := ScoreElement(doc.Find("#story").First())
	if storyScore <= leftScore {
		t.Fatalf("ScoreElement(story) = %v, left nav = %v, want center table content favored", storyScore, leftScore)
	}

	ScoreAndRemove(doc, false)

	if doc.Find("#story").Length() != 1 {
		t.Fatal("ScoreAndRemove() removed center table story content")
	}
	if doc.Find("#notes").Length() != 1 {
		t.Fatal("ScoreAndRemove() removed footnotes")
	}
	if doc.Find("#related").Length() != 0 {
		t.Fatal("ScoreAndRemove() kept dense related-links block")
	}
}
