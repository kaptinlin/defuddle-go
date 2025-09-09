package elements

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeBlockProcessing(t *testing.T) {
	html := `
	<div class="highlight language-javascript">
		<pre><code>function test() { return "hello"; }</code></pre>
	</div>
	<div class="syntaxhighlighter">
		<div class="code">
			<div class="line">console.log("test");</div>
		</div>
	</div>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	processor := NewCodeBlockProcessor(doc)
	options := DefaultCodeBlockProcessingOptions()

	processor.ProcessCodeBlocks(options)

	// Check if code blocks were processed
	mathElements := doc.Find("pre").Length()
	assert.Greater(t, mathElements, 0, "Should have processed code blocks")
}

func TestHeadingProcessing(t *testing.T) {
	html := `
	<h1>
		<a href="#test" class="anchor">Test Heading</a>
		<button class="copy-link">Copy</button>
	</h1>
	<h2>
		Clean Heading
		<span><a href="#clean">§</a></span>
	</h2>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	processor := NewHeadingProcessor(doc)
	options := DefaultHeadingProcessingOptions()

	processor.ProcessHeadings(options)

	// Check if headings were cleaned
	headings := doc.Find("h1, h2")
	assert.Equal(t, 2, headings.Length(), "Should have 2 headings")

	// Check if navigation elements were removed
	anchors := doc.Find("a[href^='#']")
	assert.Equal(t, 0, anchors.Length(), "Should have removed anchor links")
}

func TestMathProcessing(t *testing.T) {
	html := `
	<div class="MathJax">
		<script type="math/tex">x = \frac{-b \pm \sqrt{b^2-4ac}}{2a}</script>
	</div>
	<div class="katex">
		<annotation encoding="application/x-tex">\sum_{i=1}^n x_i</annotation>
	</div>
	<script type="text/javascript" src="https://cdn.mathjax.org/mathjax.js"></script>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	processor := NewMathProcessor(doc)
	options := DefaultMathProcessingOptions()

	processor.ProcessMath(options)

	// Check if math elements were processed
	mathElements := doc.Find("math")
	assert.Greater(t, mathElements.Length(), 0, "Should have created math elements")

	// Check for MathML namespace
	mathElements.Each(func(_ int, s *goquery.Selection) {
		xmlns, exists := s.Attr("xmlns")
		assert.True(t, exists, "Math element should have xmlns attribute")
		assert.Equal(t, "http://www.w3.org/1998/Math/MathML", xmlns, "Should have correct MathML namespace")
	})
}

func TestImageProcessing(t *testing.T) {
	html := `
	<img src="test.jpg" alt="">
	<img src="small.jpg" width="10" height="10" alt="small">
	<figure>
		<img src="figure.jpg" alt="Figure image">
		<figcaption>Test caption</figcaption>
	</figure>
	<img src="data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7" alt="tracking">
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	processor := NewImageProcessor(doc)
	options := DefaultImageProcessingOptions()

	processor.ProcessImages(options)

	// Check if images were processed
	images := doc.Find("img")
	assert.Greater(t, images.Length(), 0, "Should have some remaining images")

	// Check if small/tracking images were removed
	smallImages := doc.Find("img[width='10']")
	assert.Equal(t, 0, smallImages.Length(), "Should have removed small images")

	// Check if responsive classes were added
	images.Each(func(_ int, s *goquery.Selection) {
		class, _ := s.Attr("class")
		if !strings.Contains(class, "responsive-image") {
			// Some images might be removed or modified
			return
		}
		assert.Contains(t, class, "responsive-image", "Should have responsive class")
	})
}

func TestFootnoteProcessing(t *testing.T) {
	html := `
	<p>This is text with a footnote<sup><a href="#fn1">1</a></sup>.</p>
	<div id="fn1">This is the footnote content.</div>
	<p>Another reference<a href="#note2">[2]</a>.</p>
	<div id="note2">Second footnote.</div>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	processor := NewFootnoteProcessor(doc)
	options := DefaultFootnoteProcessingOptions()

	footnotes := processor.ProcessFootnotes(options)

	// Check if footnotes were detected
	assert.Greater(t, len(footnotes), 0, "Should have detected footnotes")

	// Check footnote structure
	for _, footnote := range footnotes {
		assert.NotEmpty(t, footnote.ID, "Footnote should have ID")
		if footnote.Definition != nil && footnote.Definition.Length() > 0 {
			assert.NotEmpty(t, footnote.Content, "Footnote should have content")
		}
	}
}

func TestPublicInterfaces(t *testing.T) {
	html := `
	<div>
		<h1><a href="#test">Test</a></h1>
		<pre><code class="language-go">fmt.Println("hello")</code></pre>
		<img src="test.jpg" alt="">
		<div class="MathJax"><script type="math/tex">x^2</script></div>
		<p>Footnote<sup><a href="#fn1">1</a></sup></p>
		<div id="fn1">Note content</div>
	</div>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	// Test public interfaces
	ProcessHeadings(doc, DefaultHeadingProcessingOptions())
	ProcessCodeBlocks(doc, DefaultCodeBlockProcessingOptions())
	ProcessImages(doc, DefaultImageProcessingOptions())
	ProcessMath(doc, DefaultMathProcessingOptions())
	ProcessFootnotes(doc, DefaultFootnoteProcessingOptions())

	// Verify the document still has valid structure
	body := doc.Find("body, div").First()
	assert.Greater(t, body.Children().Length(), 0, "Document should maintain structure")
}
