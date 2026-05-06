package elements

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessImagesDropsSmallAndDecorativeImages(t *testing.T) {
	t.Parallel()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<article>
			<img src="/analytics/pixel.gif" alt="tracking" width="1" height="1">
			<img src="icon.png" class="decorative-icon" alt="icon" width="32" height="32">
		</article>`))
	require.NoError(t, err)

	ProcessImages(doc, DefaultImageProcessingOptions())

	assert.Equal(t, 0, doc.Find("img[src*='pixel.gif']").Length())
	assert.Equal(t, 0, doc.Find("img.decorative-icon").Length())
}

func TestProcessImagesGeneratesReadableAltTextAndCaptionFromContext(t *testing.T) {
	t.Parallel()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<article>
			<h2>Launch Event Gallery</h2>
			<figure><img src="launch-event-photo.jpg" alt="image" width="960"></figure>
		</article>`))
	require.NoError(t, err)

	ProcessImages(doc, DefaultImageProcessingOptions())

	img := doc.Find("figure img").First()
	assert.Equal(t, "Launch Event Gallery", img.AttrOr("alt", ""))
	assert.Contains(t, img.AttrOr("class", ""), "responsive-image")
	assert.Equal(t, "eager", img.AttrOr("loading", ""))
	assert.Equal(t, "Launch Event Gallery", strings.TrimSpace(doc.Find("figcaption").Text()))
	assert.Contains(t, doc.Find("figure").AttrOr("class", ""), "large-image")
}

func TestProcessMathCleansScriptsAndPreservesLatexAsMathML(t *testing.T) {
	t.Parallel()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<div class="math-display">
			<span class="MathJax_Preview">preview</span>
			<span class="MathJax"><script type="math/tex">x^2 + y^2</script></span>
			<script type="text/javascript" src="/mathjax.js"></script>
		</div>`))
	require.NoError(t, err)

	ProcessMath(doc, DefaultMathProcessingOptions())

	math := doc.Find("math").First()
	require.Equal(t, 1, math.Length())
	assert.Equal(t, "http://www.w3.org/1998/Math/MathML", math.AttrOr("xmlns", ""))
	assert.Equal(t, "block", math.AttrOr("display", ""))
	assert.Contains(t, math.AttrOr("data-latex", ""), "x^2 + y^2")
	assert.Equal(t, 0, doc.Find(".MathJax_Preview, script[src*='mathjax']").Length())
}

func TestProcessMathPreservesExistingMathMLContent(t *testing.T) {
	t.Parallel()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<div><math display="inline"><mi>x</mi><mo>=</mo><mn>1</mn></math></div>`))
	require.NoError(t, err)

	ProcessMath(doc, DefaultMathProcessingOptions())

	math := doc.Find("math").First()
	require.Equal(t, 1, math.Length())
	assert.Equal(t, "inline", math.AttrOr("display", ""))
	assert.Equal(t, "x=1", strings.TrimSpace(math.Text()))
}

func TestProcessImagesDropsDecorativeClassWithoutSizeHint(t *testing.T) {
	t.Parallel()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<article>
			<img src="badge.png" class="profile-avatar-large" alt="Author badge">
			<img src="hero.jpg" class="article-photo" alt="Launch photo">
		</article>`))
	require.NoError(t, err)

	ProcessImages(doc, DefaultImageProcessingOptions())

	assert.Equal(t, 0, doc.Find("img.profile-avatar-large").Length())
	assert.Equal(t, 1, doc.Find("img.article-photo").Length())
}
