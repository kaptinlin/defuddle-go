package standardize

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	internalmetadata "github.com/kaptinlin/defuddle-go/internal/metadata"
)

func TestContentPreservesSemanticContainersAndInlineSpacing(t *testing.T) {
	t.Parallel()

	doc := newStandardizeDocument(t, `<html><body><article>
		<div class="content-card"><p>Preserved semantic content</p></div>
		<p>Read<strong>bold</strong><em>emphasis</em><span>, punctuation</span></p>
	</article></body></html>`)
	article := doc.Find("article").First()

	Content(article, &internalmetadata.Metadata{}, doc, false)

	if !strings.Contains(article.Text(), "Preserved semantic content") {
		t.Fatalf("Content() removed semantic content: %s", article.Text())
	}
	if got := article.Find("p").Last().Text(); got != "Read bold emphasis, punctuation" {
		t.Fatalf("Content() inline text = %q, want readable spacing", got)
	}
}

func TestContentConvertsUnorderedRoleListsAndBareListItems(t *testing.T) {
	t.Parallel()

	doc := newStandardizeDocument(t, `<html><body><article>
		<div role="list">
			<div role="listitem"><div class="content"><div role="paragraph">Alpha item</div></div></div>
			<div role="listitem"><div class="content"><div role="paragraph">Beta item</div></div></div>
		</div>
		<div role="listitem"><div class="content"><div role="paragraph">Loose item</div></div></div>
	</article></body></html>`)
	article := doc.Find("article").First()

	Content(article, &internalmetadata.Metadata{}, doc, false)

	if article.Find("ul > li").Length() != 2 {
		t.Fatalf("Content() did not convert unordered role list: %s", article.Text())
	}
	if article.Find("ol").Length() != 0 {
		t.Fatalf("Content() created ordered list for unlabeled items")
	}
	if !strings.Contains(article.Text(), "Loose item") {
		t.Fatalf("Content() removed bare list item content: %s", article.Text())
	}
}

func TestContentConvertsNestedRoleLists(t *testing.T) {
	t.Parallel()

	doc := newStandardizeDocument(t, `<html><body><article>
		<div role="list">
			<div role="listitem">
				<span class="label">1)</span>
				<div class="content">
					<div role="paragraph">Parent item</div>
					<div role="list">
						<div role="listitem"><span class="label">a)</span><div class="content"><div role="paragraph">Nested bullet</div></div></div>
					</div>
				</div>
			</div>
		</div>
	</article></body></html>`)
	article := doc.Find("article").First()

	Content(article, &internalmetadata.Metadata{}, doc, false)

	if article.Find("ol > li").Length() == 0 {
		t.Fatalf("Content() did not create ordered parent list")
	}
	if article.Find("ul li").Length() == 0 {
		t.Fatalf("Content() did not create nested unordered list")
	}
	if !strings.Contains(article.Text(), "Parent item") || !strings.Contains(article.Text(), "Nested bullet") {
		t.Fatalf("Content() removed nested list text: %s", article.Text())
	}
}

func TestContentRemovesOnlyHeadingsWithoutFollowingContent(t *testing.T) {
	t.Parallel()

	doc := newStandardizeDocument(t, `<html><body><article>
		<h2>Section with body</h2><p>Body text</p><h3>Dangling heading</h3>
	</article></body></html>`)
	article := doc.Find("article").First()

	Content(article, &internalmetadata.Metadata{}, doc, false)

	if !strings.Contains(article.Text(), "Section with body") {
		t.Fatalf("Content() removed heading that had following content: %s", article.Text())
	}
	if strings.Contains(article.Text(), "Dangling heading") {
		t.Fatalf("Content() kept trailing heading: %s", article.Text())
	}
}

func TestRemoveEmptyLinesPreservesCodeAndCleansTextNodes(t *testing.T) {
	t.Parallel()

	doc := newStandardizeDocument(t, `<html><body><article>
		<p>
			Alpha   beta   , gamma
		</p>
		<span>One</span><span>Two</span><span>.</span>
		<pre>
			keep   spacing
		</pre>
	</article></body></html>`)
	article := doc.Find("article").First()

	removeEmptyLines(article, doc)

	if got := strings.TrimSpace(article.Find("p").Text()); got != "Alpha beta, gamma" {
		t.Fatalf("removeEmptyLines() paragraph = %q, want cleaned text", got)
	}
	if got := article.Text(); !strings.Contains(got, "One Two.") {
		t.Fatalf("removeEmptyLines() did not add readable inline spacing: %q", got)
	}
	if got := article.Find("pre").Text(); !strings.Contains(got, "keep   spacing") {
		t.Fatalf("removeEmptyLines() changed pre text: %q", got)
	}
}

func TestTransformListItemWithoutContentIsLeftUntouched(t *testing.T) {
	t.Parallel()

	doc := newStandardizeDocument(t, `<html><body><div role="listitem">Plain item</div></body></html>`)
	item := doc.Find(`[role="listitem"]`).First()

	got := transformListItemElement(item, doc)

	if goquery.NodeName(got) != "div" {
		t.Fatalf("transformListItemElement() node = %q, want original div", goquery.NodeName(got))
	}
	if got.Text() != "Plain item" {
		t.Fatalf("transformListItemElement() text = %q, want original text", got.Text())
	}
}
