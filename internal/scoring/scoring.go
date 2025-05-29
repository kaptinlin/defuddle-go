package scoring

import (
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kaptinlin/defuddle-go/internal/constants"
)

// ContentScore represents a scored element
// JavaScript original code:
//
//	export interface ContentScore {
//	  score: number;
//	  element: Element;
//	}
type ContentScore struct {
	Score   float64
	Element *goquery.Selection
}

// ContentScorer provides content scoring functionality
// JavaScript original code:
//
//	export class ContentScorer {
//		private doc: Document;
//		private debug: boolean;
//
//		constructor(doc: Document, debug: boolean = false) {
//			this.doc = doc;
//			this.debug = debug;
//		}
//	}
type ContentScorer struct {
	doc   *goquery.Document
	debug bool
}

// NewContentScorer creates a new ContentScorer instance
func NewContentScorer(doc *goquery.Document, debug bool) *ContentScorer {
	return &ContentScorer{
		doc:   doc,
		debug: debug,
	}
}

// contentIndicators are class/id patterns that indicate content elements
// JavaScript original code:
//
//	const contentIndicators = [
//		'admonition',
//		'article',
//		'content',
//		'entry',
//		'image',
//		'img',
//		'font',
//		'figure',
//		'figcaption',
//		'pre',
//		'main',
//		'post',
//		'story',
//		'table'
//	];
var contentIndicators = []string{
	"admonition",
	"article",
	"content",
	"entry",
	"image",
	"img",
	"font",
	"figure",
	"figcaption",
	"pre",
	"main",
	"post",
	"story",
	"table",
}

// navigationIndicators are text patterns that indicate navigation/non-content
// JavaScript original code:
//
//	const navigationIndicators = [
//		'advertisement',
//		'all rights reserved',
//		'banner',
//		'cookie',
//		'comments',
//		'copyright',
//		'follow me',
//		'follow us',
//		'footer',
//		'header',
//		'homepage',
//		'login',
//		'menu',
//		'more articles',
//		'more like this',
//		'most read',
//		'nav',
//		'navigation',
//		'newsletter',
//		'newsletter',
//		'popular',
//		'privacy',
//		'recommended',
//		'register',
//		'related',
//		'responses',
//		'share',
//		'sidebar',
//		'sign in',
//		'sign up',
//		'signup',
//		'social',
//		'sponsored',
//		'subscribe',
//		'subscribe',
//		'terms',
//		'trending'
//	];
var navigationIndicators = []string{
	"advertisement",
	"all rights reserved",
	"banner",
	"cookie",
	"comments",
	"copyright",
	"follow me",
	"follow us",
	"footer",
	"header",
	"homepage",
	"login",
	"menu",
	"more articles",
	"more like this",
	"most read",
	"nav",
	"navigation",
	"newsletter",
	"popular",
	"privacy",
	"recommended",
	"register",
	"related",
	"responses",
	"share",
	"sidebar",
	"sign in",
	"sign up",
	"signup",
	"social",
	"sponsored",
	"subscribe",
	"terms",
	"trending",
}

// nonContentPatterns are class/id patterns that indicate non-content elements
// JavaScript original code:
//
//	const nonContentPatterns = [
//		'ad',
//		'banner',
//		'cookie',
//		'copyright',
//		'footer',
//		'header',
//		'homepage',
//		'menu',
//		'nav',
//		'newsletter',
//		'popular',
//		'privacy',
//		'recommended',
//		'related',
//		'rights',
//		'share',
//		'sidebar',
//		'social',
//		'sponsored',
//		'subscribe',
//		'terms',
//		'trending',
//		'widget'
//	];
var nonContentPatterns = []string{
	"ad",
	"banner",
	"cookie",
	"copyright",
	"footer",
	"header",
	"homepage",
	"menu",
	"nav",
	"newsletter",
	"popular",
	"privacy",
	"recommended",
	"related",
	"rights",
	"share",
	"sidebar",
	"social",
	"sponsored",
	"subscribe",
	"terms",
	"trending",
	"widget",
}

// ScoreElement scores an element based on various content indicators
// JavaScript original code:
//
//	static scoreElement(element: Element): number {
//		let score = 0;
//
//		// Text density
//		const text = element.textContent || '';
//		const words = text.split(/\s+/).length;
//		score += words;
//
//		// Paragraph ratio
//		const paragraphs = element.getElementsByTagName('p').length;
//		score += paragraphs * 10;
//
//		// Link density (penalize high link density)
//		const links = element.getElementsByTagName('a').length;
//		const linkDensity = links / (words || 1);
//		score -= linkDensity * 5;
//
//		// Image ratio (penalize high image density)
//		const images = element.getElementsByTagName('img').length;
//		const imageDensity = images / (words || 1);
//		score -= imageDensity * 3;
//
//		// Position bonus (center/right elements)
//		try {
//			const style = element.getAttribute('style') || '';
//			const align = element.getAttribute('align') || '';
//			const isRightSide = style.includes('float: right') ||
//							   style.includes('text-align: right') ||
//							   align === 'right';
//			if (isRightSide) score += 5;
//		} catch (e) {
//			// Ignore position if we can't get style
//		}
//
//		// Content indicators
//		const hasDate = /\b(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\s+\d{1,2},?\s+\d{4}\b/i.test(text);
//		if (hasDate) score += 10;
//
//		const hasAuthor = /\b(?:by|written by|author:)\s+[A-Za-z\s]+\b/i.test(text);
//		if (hasAuthor) score += 10;
//
//		// Check for common content classes/attributes
//		const className = element.className.toLowerCase();
//		if (className.includes('content') || className.includes('article') || className.includes('post')) {
//			score += 15;
//		}
//
//		// Check for footnotes/references
//		const hasFootnotes = element.querySelector(FOOTNOTE_INLINE_REFERENCES);
//		if (hasFootnotes) score += 10;
//
//		const hasFootnotesList = element.querySelector(FOOTNOTE_LIST_SELECTORS);
//		if (hasFootnotesList) score += 10;
//
//		// Check for nested tables (penalize)
//		const nestedTables = element.getElementsByTagName('table').length;
//		score -= nestedTables * 5;
//
//		// Additional scoring for table cells
//		if (element.tagName.toLowerCase() === 'td') {
//			// Table cells get a bonus for being in the main content area
//			const parentTable = element.closest('table');
//			if (parentTable) {
//				// Only favor cells in tables that look like old-style content layouts
//				const tableWidth = parseInt(parentTable.getAttribute('width') || '0');
//				const tableAlign = parentTable.getAttribute('align') || '';
//				const tableClass = parentTable.className.toLowerCase();
//				const isTableLayout =
//					tableWidth > 400 || // Common width for main content tables
//					tableAlign === 'center' ||
//					tableClass.includes('content') ||
//					tableClass.includes('article');
//
//				if (isTableLayout) {
//					// Additional checks to ensure this is likely the main content cell
//					const allCells = Array.from(parentTable.getElementsByTagName('td'));
//					const cellIndex = allCells.indexOf(element as HTMLTableCellElement);
//					const isCenterCell = cellIndex > 0 && cellIndex < allCells.length - 1;
//
//					if (isCenterCell) {
//						score += 10;
//					}
//				}
//			}
//		}
//
//		return score;
//	}
func ScoreElement(element *goquery.Selection) float64 {
	score := 0.0

	// Text density
	text := strings.TrimSpace(element.Text())
	words := len(strings.Fields(text))
	score += float64(words)

	// Paragraph ratio
	paragraphs := element.Find("p").Length()
	score += float64(paragraphs) * 10

	// Link density (penalize high link density)
	links := element.Find("a").Length()
	linkDensity := float64(links) / float64(max(words, 1))
	score -= linkDensity * 5

	// Image ratio (penalize high image density)
	images := element.Find("img").Length()
	imageDensity := float64(images) / float64(max(words, 1))
	score -= imageDensity * 3

	// Position bonus (center/right elements)
	style, _ := element.Attr("style")
	align, _ := element.Attr("align")
	isRightSide := strings.Contains(style, "float: right") ||
		strings.Contains(style, "text-align: right") ||
		align == "right"
	if isRightSide {
		score += 5
	}

	// Content indicators
	dateRegex := regexp.MustCompile(`(?i)\b(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\s+\d{1,2},?\s+\d{4}\b`)
	if dateRegex.MatchString(text) {
		score += 10
	}

	authorRegex := regexp.MustCompile(`(?i)\b(?:by|written by|author:)\s+[A-Za-z\s]+\b`)
	if authorRegex.MatchString(text) {
		score += 10
	}

	// Check for common content classes/attributes
	className := strings.ToLower(element.AttrOr("class", ""))
	if strings.Contains(className, "content") ||
		strings.Contains(className, "article") ||
		strings.Contains(className, "post") {
		score += 15
	}

	// Check for footnotes/references
	footnoteSelectors := constants.GetFootnoteInlineReferences()
	for _, selector := range footnoteSelectors {
		if element.Find(selector).Length() > 0 {
			score += 10
			break
		}
	}

	footnoteListSelectors := constants.GetFootnoteListSelectors()
	for _, selector := range footnoteListSelectors {
		if element.Find(selector).Length() > 0 {
			score += 10
			break
		}
	}

	// Check for nested tables (penalize)
	nestedTables := element.Find("table").Length()
	score -= float64(nestedTables) * 5

	// Additional scoring for table cells
	if goquery.NodeName(element) == "td" {
		parentTable := element.Closest("table")
		if parentTable.Length() > 0 {
			// Only favor cells in tables that look like old-style content layouts
			widthStr, _ := parentTable.Attr("width")
			tableWidth := 0
			if widthStr != "" {
				if w, err := strconv.Atoi(widthStr); err == nil {
					tableWidth = w
				}
			}
			tableAlign, _ := parentTable.Attr("align")
			tableClass := strings.ToLower(parentTable.AttrOr("class", ""))

			isTableLayout := tableWidth > 400 || // Common width for main content tables
				tableAlign == "center" ||
				strings.Contains(tableClass, "content") ||
				strings.Contains(tableClass, "article")

			if isTableLayout {
				// Additional checks to ensure this is likely the main content cell
				allCells := parentTable.Find("td")
				cellIndex := -1
				allCells.Each(func(i int, cell *goquery.Selection) {
					if cell.Get(0) == element.Get(0) {
						cellIndex = i
					}
				})

				isCenterCell := cellIndex > 0 && cellIndex < allCells.Length()-1
				if isCenterCell {
					score += 10
				}
			}
		}
	}

	return score
}

// FindBestElement finds the best scoring element from a list
// JavaScript original code:
//
//	static findBestElement(elements: Element[], minScore: number = 50): Element | null {
//		let bestElement: Element | null = null;
//		let bestScore = 0;
//
//		elements.forEach(element => {
//			const score = this.scoreElement(element);
//			if (score > bestScore) {
//				bestScore = score;
//				bestElement = element;
//			}
//		});
//
//		return bestScore > minScore ? bestElement : null;
//	}
func FindBestElement(elements []*goquery.Selection, minScore float64) *goquery.Selection {
	var bestElement *goquery.Selection
	bestScore := 0.0

	for _, element := range elements {
		score := ScoreElement(element)
		if score > bestScore {
			bestScore = score
			bestElement = element
		}
	}

	if bestScore > minScore {
		return bestElement
	}
	return nil
}

// ScoreAndRemove scores blocks and removes those that are likely not content
// JavaScript original code:
//
//	public static scoreAndRemove(doc: Document, debug: boolean = false) {
//		const startTime = Date.now();
//		let removedCount = 0;
//
//		// Track all elements to be removed
//		const elementsToRemove = new Set<Element>();
//
//		// Get all block elements
//		const blockElements = Array.from(doc.querySelectorAll(BLOCK_ELEMENTS.join(',')));
//
//		// Process each block element
//		blockElements.forEach(element => {
//			// Skip elements that are already marked for removal
//			if (elementsToRemove.has(element)) {
//				return;
//			}
//
//			// Skip elements that are likely to be content
//			if (ContentScorer.isLikelyContent(element)) {
//				return;
//			}
//
//			// Score the element based on various criteria
//			const score = ContentScorer.scoreNonContentBlock(element);
//
//			// If the score is below the threshold, mark for removal
//			if (score < 0) {
//				elementsToRemove.add(element);
//				removedCount++;
//			}
//		});
//
//		// Remove all collected elements in a single pass
//		elementsToRemove.forEach(el => el.remove());
//
//		const endTime = Date.now();
//		if (debug) {
//			console.log('Defuddle', 'Removed non-content blocks:', {
//				count: removedCount,
//				processingTime: `${(endTime - startTime).toFixed(2)}ms`
//			});
//		}
//	}
func ScoreAndRemove(doc *goquery.Document, debug bool) {
	startTime := time.Now()
	removedCount := 0

	// Track all elements to be removed
	elementsToRemove := make([]*goquery.Selection, 0, 10) // Pre-allocate with reasonable capacity

	// Get all block elements
	blockElements := constants.GetBlockElements()
	blockSelector := strings.Join(blockElements, ",")

	// Process each block element
	doc.Find(blockSelector).Each(func(i int, element *goquery.Selection) {
		// Skip elements that are likely to be content
		if isLikelyContent(element) {
			return
		}

		// Score the element based on various criteria
		score := scoreNonContentBlock(element)

		// If the score is below the threshold, mark for removal
		if score < 0 {
			elementsToRemove = append(elementsToRemove, element)
			removedCount++
		}
	})

	// Remove all collected elements in a single pass
	for _, el := range elementsToRemove {
		el.Remove()
	}

	endTime := time.Now()
	if debug {
		processingTime := float64(endTime.Sub(startTime).Nanoseconds()) / 1e6 // Convert to milliseconds
		slog.Debug("Removed non-content blocks",
			"count", removedCount,
			"processingTime", processingTime)
	}
}

// isLikelyContent determines if an element is likely to be content
// JavaScript original code:
//
//	private static isLikelyContent(element: Element): boolean {
//		// Check if the element has a role that indicates content
//		const role = element.getAttribute('role');
//		if (role && ['article', 'main', 'contentinfo'].includes(role)) {
//			return true;
//		}
//
//		// Check if the element has a class or id that indicates content
//		const className = element.className.toLowerCase();
//		const id = element.id.toLowerCase();
//
//		for (const indicator of contentIndicators) {
//			if (className.includes(indicator) || id.includes(indicator)) {
//				return true;
//			}
//		}
//
//		// Check if the element has a high text density
//		const text = element.textContent || '';
//		const words = text.split(/\s+/).length;
//		const paragraphs = element.getElementsByTagName('p').length;
//
//		// If the element has a significant amount of text and paragraphs, it's likely content
//		if (words > 50 && paragraphs > 1) {
//			return true;
//		}
//
//		// Check for elements with significant text content, even if they don't have many paragraphs
//		if (words > 100) {
//			return true;
//		}
//
//		// Check for elements with text content and some paragraphs
//		if (words > 30 && paragraphs > 0) {
//			return true;
//		}
//
//		return false;
//	}
func isLikelyContent(element *goquery.Selection) bool {
	// Check if the element has a role that indicates content
	role, _ := element.Attr("role")
	if role != "" {
		contentRoles := []string{"article", "main", "contentinfo"}
		for _, contentRole := range contentRoles {
			if role == contentRole {
				return true
			}
		}
	}

	// Check if the element has a class or id that indicates content
	className := strings.ToLower(element.AttrOr("class", ""))
	id := strings.ToLower(element.AttrOr("id", ""))

	for _, indicator := range contentIndicators {
		if strings.Contains(className, indicator) || strings.Contains(id, indicator) {
			return true
		}
	}

	// Check if the element has a high text density
	text := strings.TrimSpace(element.Text())
	words := len(strings.Fields(text))
	paragraphs := element.Find("p").Length()

	// If the element has a significant amount of text and paragraphs, it's likely content
	if words > 50 && paragraphs > 1 {
		return true
	}

	// Check for elements with significant text content, even if they don't have many paragraphs
	if words > 100 {
		return true
	}

	// Check for elements with text content and some paragraphs
	if words > 30 && paragraphs > 0 {
		return true
	}

	return false
}

// scoreNonContentBlock scores a block element to determine if it's likely not content
// JavaScript original code:
//
//	private static scoreNonContentBlock(element: Element): number {
//		// Skip footnote list elements
//		if (element.querySelector(FOOTNOTE_LIST_SELECTORS)) {
//			return 0;
//		}
//
//		let score = 0;
//
//		// Get text content
//		const text = element.textContent || '';
//		const words = text.split(/\s+/).length;
//
//		// Skip very small elements
//		if (words < 3) {
//			return 0;
//		}
//
//		for (const indicator of navigationIndicators) {
//			if (text.toLowerCase().includes(indicator)) {
//				score -= 10;
//			}
//		}
//
//		// Check for high link density (navigation)
//		const links = element.getElementsByTagName('a').length;
//		const linkDensity = links / (words || 1);
//		if (linkDensity > 0.5) {
//			score -= 15;
//		}
//
//		// Check for list structure (navigation)
//		const lists = element.getElementsByTagName('ul').length + element.getElementsByTagName('ol').length;
//		if (lists > 0 && links > lists * 3) {
//			score -= 10;
//		}
//
//		// Check for specific class patterns that indicate non-content
//		const className = element.className.toLowerCase();
//		const id = element.id.toLowerCase();
//
//		for (const pattern of nonContentPatterns) {
//			if (className.includes(pattern) || id.includes(pattern)) {
//				score -= 8;
//			}
//		}
//
//		return score;
//	}
func scoreNonContentBlock(element *goquery.Selection) float64 {
	// Skip footnote list elements
	footnoteListSelectors := constants.GetFootnoteListSelectors()
	for _, selector := range footnoteListSelectors {
		if element.Find(selector).Length() > 0 {
			return 0
		}
	}

	score := 0.0

	// Get text content
	text := strings.TrimSpace(element.Text())
	words := len(strings.Fields(text))

	// Skip very small elements
	if words < 3 {
		return 0
	}

	// Check for navigation indicators in text
	lowerText := strings.ToLower(text)
	for _, indicator := range navigationIndicators {
		if strings.Contains(lowerText, indicator) {
			score -= 10
		}
	}

	// Check for high link density (navigation)
	links := element.Find("a").Length()
	linkDensity := float64(links) / float64(max(words, 1))
	if linkDensity > 0.5 {
		score -= 15
	}

	// Check for list structure (navigation)
	lists := element.Find("ul").Length() + element.Find("ol").Length()
	if lists > 0 && links > lists*3 {
		score -= 10
	}

	// Check for specific class patterns that indicate non-content
	className := strings.ToLower(element.AttrOr("class", ""))
	id := strings.ToLower(element.AttrOr("id", ""))

	for _, pattern := range nonContentPatterns {
		if strings.Contains(className, pattern) || strings.Contains(id, pattern) {
			score -= 8
		}
	}

	return score
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
