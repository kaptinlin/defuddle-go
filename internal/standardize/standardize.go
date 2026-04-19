// Package standardize provides content standardization functionality for the defuddle content extraction system.
// It converts non-semantic HTML elements to semantic ones and applies standardization rules.
package standardize

import (
	"cmp"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kaptinlin/defuddle-go/internal/constants"
	"github.com/kaptinlin/defuddle-go/internal/metadata"
	"golang.org/x/net/html"
)

// Pre-compiled regex patterns used across standardization functions.
var (
	nbspRe             = regexp.MustCompile(`\xA0+`)
	wordCharRe         = regexp.MustCompile(`\w`)
	whitespaceRe       = regexp.MustCompile(`\s+`)
	semanticClassRe    = regexp.MustCompile(`(?:article|main|content|footnote|reference|bibliography)`)
	wrapperClassRe     = regexp.MustCompile(`(?:wrapper|container|layout|row|col|grid|flex|outer|inner|content-area)`)
	emptyTextRe        = regexp.MustCompile(`^[\x{200C}\x{200B}\x{200D}\x{200E}\x{200F}\x{FEFF}\x{A0}\s]*$`)
	threeNewlinesRe    = regexp.MustCompile(`\n{3,}`)
	leadingNewlinesRe  = regexp.MustCompile(`^[\n\r\t]+`)
	trailingNewlinesRe = regexp.MustCompile(`[\n\r\t]+$`)
	spacesAroundNlRe   = regexp.MustCompile(`[ \t]*\n[ \t]*`)
	threeSpacesRe      = regexp.MustCompile(`[ \t]{3,}`)
	onlySpacesRe       = regexp.MustCompile(`^[ ]+$`)
	spaceBeforePunctRe = regexp.MustCompile(`\s+([,.!?:;])`)
	zeroWidthCharsRe   = regexp.MustCompile(`[\x{200C}\x{200B}\x{200D}\x{200E}\x{200F}\x{FEFF}]+`)
	multiNbspRe        = regexp.MustCompile(`(?:\xA0){2,}`)
	blockStartSpaceRe  = regexp.MustCompile(`^[\n\r\t \x{200C}\x{200B}\x{200D}\x{200E}\x{200F}\x{FEFF}\x{A0}]*$`)
	inlineStartSpaceRe = regexp.MustCompile(`^[\n\r\t\x{200C}\x{200B}\x{200D}\x{200E}\x{200F}\x{FEFF}]*$`)
	startsWithPunctRe  = regexp.MustCompile(`^[,.!?:;)\]]`)
	endsWithPunctRe    = regexp.MustCompile(`[,.!?:;(\[]\s*$`)
	orderedListLabelRe = regexp.MustCompile(`^\d+\)`)
)

// StandardizationRule represents element standardization rules
// JavaScript original code:
//
//	interface StandardizationRule {
//		selector: string;
//		element: string;
//		transform?: (el: Element, doc: Document) => Element;
//	}
type StandardizationRule struct {
	Selector  string
	Element   string
	Transform func(el *goquery.Selection, doc *goquery.Document) *goquery.Selection
}

// ELEMENT_STANDARDIZATION_RULES maps selectors to their target HTML element name
// JavaScript original code:
//
//	const ELEMENT_STANDARDIZATION_RULES: StandardizationRule[] = [
//		...mathRules,
//		...codeBlockRules,
//		...headingRules,
//		...imageRules,
//		// Convert divs with paragraph role to actual paragraphs
//		{
//			selector: 'div[data-testid^="paragraph"], div[role="paragraph"]',
//			element: 'p',
//			transform: (el: Element, doc: Document): Element => { ... }
//		},
//		// Convert divs with list roles to actual lists
//		{
//			selector: 'div[role="list"]',
//			element: 'ul',
//			transform: (el: Element, doc: Document): Element => { ... }
//		},
//		{
//			selector: 'div[role="listitem"]',
//			element: 'li',
//			transform: (el: Element, doc: Document): Element => { ... }
//		}
//	];
var elementStandardizationRules = []StandardizationRule{
	// Convert divs with paragraph role to actual paragraphs
	{
		Selector: `div[data-testid^="paragraph"], div[role="paragraph"]`,
		Element:  "p",
		Transform: func(el *goquery.Selection, _ *goquery.Document) *goquery.Selection {
			// Get the inner HTML and attributes
			html, _ := el.Html()

			// Build new paragraph HTML
			var newHTML strings.Builder
			newHTML.WriteString("<p")

			// Copy allowed attributes (except role)
			if el.Length() > 0 {
				node := el.Get(0)
				for _, attr := range node.Attr {
					if constants.IsAllowedAttribute(attr.Key) && attr.Key != "role" {
						newHTML.WriteString(` ` + attr.Key + `="` + attr.Val + `"`)
					}
				}
			}

			newHTML.WriteString(">" + html + "</p>")

			// Replace the element with the new HTML
			el.ReplaceWithHtml(newHTML.String())

			// Return nil to indicate we handled the replacement
			return nil
		},
	},
	// Convert divs with list roles to actual lists
	{
		Selector:  `div[role="list"]`,
		Element:   "ul",
		Transform: transformListElement,
	},
	{
		Selector:  `div[role="listitem"]`,
		Element:   "li",
		Transform: transformListItemElement,
	},
}

// Content standardizes and cleans up the main content element
// JavaScript original code:
//
//	export function standardizeContent(element: Element, metadata: DefuddleMetadata, doc: Document, debug: boolean = false): void {
//		standardizeSpaces(element);
//
//		// Remove HTML comments
//		removeHTMLComments(element);
//
//		// Handle H1 elements - remove first one and convert others to H2
//		standardizeHeadings(element, metadata.title, doc);
//
//		// Standardize footnotes and citations
//		standardizeFootnotes(element);
//
//		// Convert embedded content to standard formats
//		standardizeElements(element, doc);
//
//		// If not debug mode, do the full cleanup
//		if (!debug) {
//			// First pass of div flattening
//			flattenWrapperElements(element, doc);
//
//			// Strip unwanted attributes
//			stripUnwantedAttributes(element, debug);
//
//			// Remove empty elements
//			removeEmptyElements(element);
//
//			// Remove trailing headings
//			removeTrailingHeadings(element);
//
//			// Final pass of div flattening after cleanup operations
//			flattenWrapperElements(element, doc);
//
//			// Standardize consecutive br elements
//			stripExtraBrElements(element);
//
//			// Clean up empty lines
//			removeEmptyLines(element, doc);
//		} else {
//			// In debug mode, still do basic cleanup but preserve structure
//			stripUnwantedAttributes(element, debug);
//			removeTrailingHeadings(element);
//			stripExtraBrElements(element);
//			logDebug('Debug mode: Skipping div flattening to preserve structure');
//		}
//	}
func Content(element *goquery.Selection, metadata *metadata.Metadata, doc *goquery.Document, debug bool) {
	standardizeSpaces(element)

	// Remove HTML comments
	removeHTMLComments(element)

	// Handle H1 elements - remove first one and convert others to H2
	standardizeHeadings(element, metadata.Title, doc)

	// Standardize footnotes and citations
	standardizeFootnotes(element)

	// Convert embedded content to standard formats
	standardizeElements(element, doc)

	// If not debug mode, do the full cleanup
	if !debug {
		// First pass of div flattening
		flattenWrapperElements(element, doc)

		// Strip unwanted attributes
		stripUnwantedAttributes(element, debug)

		// Remove empty elements
		removeEmptyElements(element)

		// Remove trailing headings
		removeTrailingHeadings(element)

		// Final pass of div flattening after cleanup operations
		flattenWrapperElements(element, doc)

		// Standardize consecutive br elements
		stripExtraBrElements(element)

		// Clean up empty lines
		removeEmptyLines(element, doc)
	} else {
		// In debug mode, still do basic cleanup but preserve structure
		stripUnwantedAttributes(element, debug)
		removeTrailingHeadings(element)
		stripExtraBrElements(element)
		// Debug mode: Skipping div flattening to preserve structure
	}
}

// standardizeSpaces normalizes whitespace in text content
// JavaScript original code:
//
//	function standardizeSpaces(element: Element): void {
//		const processNode = (node: Node) => {
//			// Skip pre and code elements
//			if (isElement(node)) {
//				const tag = (node as Element).tagName.toLowerCase();
//				if (tag === 'pre' || tag === 'code') {
//					return;
//				}
//			}
//
//			// Process text nodes
//			if (isTextNode(node)) {
//				const text = node.textContent || '';
//				// Replace &nbsp; with regular spaces, except when it's a single &nbsp; between words
//				const newText = text.replace(/\xA0+/g, (match) => {
//					// If it's a single &nbsp; between word characters, preserve it
//					if (match.length === 1) {
//						const prev = node.previousSibling?.textContent?.slice(-1);
//						const next = node.nextSibling?.textContent?.charAt(0);
//						if (prev?.match(/\w/) && next?.match(/\w/)) {
//							return '\xA0';
//						}
//					}
//					return ' '.repeat(match.length);
//				});
//
//				if (newText !== text) {
//					node.textContent = newText;
//				}
//			}
//
//			// Process children recursively
//			if (node.hasChildNodes()) {
//				Array.from(node.childNodes).forEach(processNode);
//			}
//		};
//
//		processNode(element);
//	}
func standardizeSpaces(element *goquery.Selection) {
	var processNode func(node *html.Node)
	processNode = func(node *html.Node) {
		// Skip pre and code elements
		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)
			if tag == "pre" || tag == "code" {
				return
			}
		}

		// Process text nodes
		if node.Type == html.TextNode {
			text := node.Data
			// Replace &nbsp; with regular spaces, except when it's a single &nbsp; between words
			newText := nbspRe.ReplaceAllStringFunc(text, func(match string) string {
				// If it's a single &nbsp; between word characters, preserve it
				if len(match) == 1 {
					// Check previous sibling
					var prev string
					if node.PrevSibling != nil && node.PrevSibling.Type == html.TextNode {
						prevText := node.PrevSibling.Data
						if len(prevText) > 0 {
							prev = string(prevText[len(prevText)-1])
						}
					}

					// Check next sibling
					var next string
					if node.NextSibling != nil && node.NextSibling.Type == html.TextNode {
						nextText := node.NextSibling.Data
						if len(nextText) > 0 {
							next = string(nextText[0])
						}
					}

					// If between word characters, preserve the &nbsp;
					if wordCharRe.MatchString(prev) && wordCharRe.MatchString(next) {
						return "\xA0"
					}
				}
				return strings.Repeat(" ", len(match))
			})

			if newText != text {
				node.Data = newText
			}
		}

		// Process children recursively
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			processNode(child)
		}
	}

	// Process all nodes in the selection
	element.Each(func(_ int, sel *goquery.Selection) {
		if sel.Length() > 0 {
			processNode(sel.Get(0))
		}
	})
}

// removeHtmlComments removes HTML comments from the element
// JavaScript original code:
//
//	function removeHtmlComments(element: Element): void {
//		const walker = document.createTreeWalker(
//			element,
//			NodeFilter.SHOW_COMMENT,
//			null,
//			false
//		);
//
//		const commentsToRemove: Comment[] = [];
//		let node: Comment | null;
//		while (node = walker.nextNode() as Comment) {
//			commentsToRemove.push(node);
//		}
//
//		commentsToRemove.forEach(comment => {
//			comment.remove();
//		});
//	}
func removeHTMLComments(_ *goquery.Selection) {
	// goquery automatically handles comment removal during parsing
	// This is a no-op in Go as comments are not preserved in the DOM tree
}

// standardizeHeadings handles H1 elements and converts them appropriately
// JavaScript original code:
//
//	function standardizeHeadings(element: Element, title: string, doc: Document): void {
//		const normalizeText = (text: string): string => {
//			return text
//				.replace(/\u00A0/g, ' ') // Convert non-breaking spaces to regular spaces
//				.replace(/\s+/g, ' ') // Normalize all whitespace to single spaces
//				.trim()
//				.toLowerCase();
//		};
//
//		const h1s = element.getElementsByTagName('h1');
//
//		Array.from(h1s).forEach(h1 => {
//			const h2 = doc.createElement('h2');
//			h2.innerHTML = h1.innerHTML;
//			// Copy allowed attributes
//			Array.from(h1.attributes).forEach(attr => {
//				if (ALLOWED_ATTRIBUTES.has(attr.name)) {
//					h2.setAttribute(attr.name, attr.value);
//				}
//			});
//			h1.parentNode?.replaceChild(h2, h1);
//		});
//
//		// Remove first H2 if it matches title
//		const h2s = element.getElementsByTagName('h2');
//		if (h2s.length > 0) {
//			const firstH2 = h2s[0];
//			const firstH2Text = normalizeText(firstH2.textContent || '');
//			const normalizedTitle = normalizeText(title);
//			if (normalizedTitle && normalizedTitle === firstH2Text) {
//				firstH2.remove();
//			}
//		}
//	}
func standardizeHeadings(element *goquery.Selection, title string, _ *goquery.Document) {
	normalizeText := func(text string) string {
		// Convert non-breaking spaces to regular spaces
		text = strings.ReplaceAll(text, "\u00A0", " ")
		// Normalize all whitespace to single spaces
		text = whitespaceRe.ReplaceAllString(text, " ")
		// Trim and convert to lowercase
		return strings.ToLower(strings.TrimSpace(text))
	}

	// Convert all H1s to H2s
	element.Find("h1").Each(func(_ int, h1 *goquery.Selection) {
		html, _ := h1.Html()

		// Create new H2 element
		var newH2 strings.Builder
		newH2.WriteString("<h2")

		// Copy allowed attributes
		if h1.Length() > 0 {
			node := h1.Get(0)
			for _, attr := range node.Attr {
				if constants.IsAllowedAttribute(attr.Key) {
					newH2.WriteString(` ` + attr.Key + `="` + attr.Val + `"`)
				}
			}
		}

		newH2.WriteString(">" + html + "</h2>")
		h1.ReplaceWithHtml(newH2.String())
	})

	// Remove first H2 if it matches title
	firstH2 := element.Find("h2").First()
	if firstH2.Length() > 0 {
		firstH2Text := normalizeText(firstH2.Text())
		normalizedTitle := normalizeText(title)
		if normalizedTitle != "" && normalizedTitle == firstH2Text {
			firstH2.Remove()
		}
	}
}

// standardizeFootnotes processes footnotes and citations
// JavaScript original code:
//
//	export function standardizeFootnotes(element: Element): void {
//		// Remove footnote back-references
//		const backRefs = element.querySelectorAll(FOOTNOTE_BACK_REFERENCES);
//		backRefs.forEach(ref => ref.remove());
//
//		// Process inline footnote references
//		const inlineRefs = element.querySelectorAll(FOOTNOTE_INLINE_REFERENCES);
//		inlineRefs.forEach(ref => {
//			// Convert to superscript if not already
//			if (ref.tagName.toLowerCase() !== 'sup') {
//				const sup = ref.ownerDocument.createElement('sup');
//				sup.innerHTML = ref.innerHTML;
//				ref.replaceWith(sup);
//			}
//		});
//	}
func standardizeFootnotes(element *goquery.Selection) {
	// Remove footnote back-references
	backRefSelectors := []string{
		`a[href^="#"][class*="anchor"]`,
		`a[href^="#"][class*="ref"]`,
		`a[class*="footnote-backref"]`,
		`.footnote-backref`,
	}

	for _, selector := range backRefSelectors {
		element.Find(selector).Remove()
	}

	// Process inline footnote references
	footnoteSelectors := constants.GetFootnoteInlineReferences()
	for _, selector := range footnoteSelectors {
		element.Find(selector).Each(func(_ int, ref *goquery.Selection) {
			// Convert to superscript if not already
			if goquery.NodeName(ref) != "sup" {
				html, _ := ref.Html()
				ref.ReplaceWithHtml("<sup>" + html + "</sup>")
			}
		})
	}
}

// standardizeElements converts embedded content to standard formats
// JavaScript original code:
//
//	function standardizeElements(element: Element, doc: Document): void {
//		ELEMENT_STANDARDIZATION_RULES.forEach(rule => {
//			const elements = element.querySelectorAll(rule.selector);
//			elements.forEach(el => {
//				try {
//					let newElement: Element;
//					if (rule.transform) {
//						newElement = rule.transform(el, doc);
//					} else {
//						newElement = doc.createElement(rule.element);
//						newElement.innerHTML = el.innerHTML;
//
//						// Copy allowed attributes
//						Array.from(el.attributes).forEach(attr => {
//							if (ALLOWED_ATTRIBUTES.has(attr.name)) {
//								newElement.setAttribute(attr.name, attr.value);
//							}
//						});
//					}
//
//					el.replaceWith(newElement);
//				} catch (e) {
//					console.warn('Failed to standardize element:', e);
//				}
//			});
//		});
//	}
func standardizeElements(element *goquery.Selection, doc *goquery.Document) {
	processedCount := 0

	// Process each standardization rule
	for _, rule := range elementStandardizationRules {
		element.Find(rule.Selector).Each(func(_ int, el *goquery.Selection) {
			if rule.Transform != nil {
				// Use custom transform function
				newElement := rule.Transform(el, doc)
				if newElement != nil && newElement.Length() > 0 {
					// Get the HTML of the new element and replace
					newHTML, err := goquery.OuterHtml(newElement)
					if err == nil {
						el.ReplaceWithHtml(newHTML)
						processedCount++
					}
				}
			} else {
				// Default transformation
				html, _ := el.Html()
				var newElementHTML strings.Builder
				newElementHTML.WriteString("<" + rule.Element)

				// Copy allowed attributes
				if el.Length() > 0 {
					node := el.Get(0)
					for _, attr := range node.Attr {
						if constants.IsAllowedAttribute(attr.Key) {
							newElementHTML.WriteString(` ` + attr.Key + `="` + attr.Val + `"`)
						}
					}
				}

				newElementHTML.WriteString(">" + html + "</" + rule.Element + ">")
				el.ReplaceWithHtml(newElementHTML.String())
				processedCount++
			}
		})
	}

	// Convert lite-youtube elements
	element.Find("lite-youtube").Each(func(_ int, el *goquery.Selection) {
		videoID, exists := el.Attr("videoid")
		if !exists || videoID == "" {
			return
		}

		videoTitle, _ := el.Attr("videotitle")
		if videoTitle == "" {
			videoTitle = "YouTube video player"
		}

		iframeHTML := `<iframe width="560" height="315" ` +
			`src="https://www.youtube.com/embed/` + videoID + `" ` +
			`title="` + videoTitle + `" ` +
			`frameborder="0" ` +
			`allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" ` +
			`allowfullscreen></iframe>`

		el.ReplaceWithHtml(iframeHTML)
		processedCount++
	})

	slog.Debug("Converted embedded elements", "count", processedCount)
}

// flattenWrapperElements removes unnecessary wrapper divs
// JavaScript original code:
//
//	function flattenWrapperElements(element: Element, doc: Document): void {
//		let processedCount = 0;
//		const startTime = Date.now();
//
//		// Process in batches to maintain performance
//		let keepProcessing = true;
//
//		// Helper function to check if an element directly contains inline content
//		// This helps prevent unwrapping divs that visually act as paragraphs.
//		function hasDirectInlineContent(el: Element): boolean {
//			for (const child of el.childNodes) {
//				// Check for non-empty text nodes
//				if (isTextNode(child) && child.textContent?.trim()) {
//					return true;
//				}
//				// Check for element nodes that are considered inline
//				if (isElement(child) && INLINE_ELEMENTS.has(child.nodeName.toLowerCase())) {
//					return true;
//				}
//			}
//			return false;
//		}
//
//		const shouldPreserveElement = (el: Element): boolean => {
//			const tagName = el.tagName.toLowerCase();
//
//			// Check if element should be preserved
//			if (PRESERVE_ELEMENTS.has(tagName)) return true;
//
//			// Check for semantic roles
//			const role = el.getAttribute('role');
//			if (role && ['article', 'main', 'navigation', 'banner', 'contentinfo'].includes(role)) {
//				return true;
//			}
//
//			// Check for semantic classes
//			const className = el.className;
//			if (typeof className === 'string' && className.toLowerCase().match(/(?:article|main|content|footnote|reference|bibliography)/)) {
//				return true;
//			}
//
//			// Check if element contains mixed content types that should be preserved
//			const children = Array.from(el.children);
//			const hasPreservedElements = children.some(child =>
//				PRESERVE_ELEMENTS.has(child.tagName.toLowerCase()) ||
//				child.getAttribute('role') === 'article' ||
//				(child.className && typeof child.className === 'string' &&
//					child.className.toLowerCase().match(/(?:article|main|content|footnote|reference|bibliography)/))
//			);
//			if (hasPreservedElements) return true;
//
//			return false;
//		};
//
//		const isWrapperElement = (el: Element): boolean => {
//			// If it directly contains inline content, it's NOT a wrapper
//			if (hasDirectInlineContent(el)) {
//				return false;
//			}
//
//			// Check if it's just empty space
//			if (!el.textContent?.trim()) return true;
//
//			// Check if it only contains other block elements
//			const children = Array.from(el.children);
//			if (children.length === 0) return true;
//
//			// Check if all children are block elements
//			const allBlockElements = children.every(child => {
//				const tag = child.tagName.toLowerCase();
//				return BLOCK_ELEMENTS.includes(tag) ||
//					   tag === 'p' || tag === 'h1' || tag === 'h2' ||
//					   tag === 'h3' || tag === 'h4' || tag === 'h5' || tag === 'h6' ||
//					   tag === 'ul' || tag === 'ol' || tag === 'pre' || tag === 'blockquote' ||
//					   tag === 'figure';
//			});
//			if (allBlockElements) return true;
//
//			// Check for common wrapper patterns
//			const className = el.className.toLowerCase();
//			const isWrapper = /(?:wrapper|container|layout|row|col|grid|flex|outer|inner|content-area)/i.test(className);
//			if (isWrapper) return true;
//
//			// Check if it has excessive whitespace or empty text nodes
//			const textNodes = Array.from(el.childNodes).filter(node =>
//				isTextNode(node) && node.textContent?.trim()
//			);
//			if (textNodes.length === 0) return true;
//
//			// Check if it only contains block elements
//			const hasOnlyBlockElements = children.length > 0 && !children.some(child => {
//				const tag = child.tagName.toLowerCase();
//				return INLINE_ELEMENTS.has(tag);
//			});
//			if (hasOnlyBlockElements) return true;
//
//			return false;
//		};
//
//		// ... (complex processing logic continues)
//	}
func flattenWrapperElements(element *goquery.Selection, _ *goquery.Document) {
	processedCount := 0
	startTime := time.Now()

	// Process in batches to maintain performance
	keepProcessing := true

	// Helper function to check if an element directly contains inline content
	hasDirectInlineContent := func(el *goquery.Selection) bool {
		hasInlineContent := false
		el.Contents().Each(func(_ int, child *goquery.Selection) {
			if goquery.NodeName(child) == "#text" {
				text := strings.TrimSpace(child.Text())
				if text != "" {
					hasInlineContent = true
				}
			} else {
				tagName := goquery.NodeName(child)
				inlineElements := constants.GetInlineElements()
				if slices.Contains(inlineElements, tagName) {
					hasInlineContent = true
				}
			}
		})
		return hasInlineContent
	}

	shouldPreserveElement := func(el *goquery.Selection) bool {
		tagName := goquery.NodeName(el)

		// Check if element should be preserved
		if constants.IsPreserveElement(tagName) {
			return true
		}

		// Check for semantic roles
		role, _ := el.Attr("role")
		semanticRoles := []string{"article", "main", "navigation", "banner", "contentinfo"}
		if slices.Contains(semanticRoles, role) {
			return true
		}

		// Check for semantic classes
		className := strings.ToLower(el.AttrOr("class", ""))
		if semanticClassRe.MatchString(className) {
			return true
		}

		// Check if element contains mixed content types that should be preserved
		hasPreservedElements := false
		el.Children().Each(func(_ int, child *goquery.Selection) {
			childTag := goquery.NodeName(child)
			childRole, _ := child.Attr("role")
			childClass := strings.ToLower(child.AttrOr("class", ""))

			if constants.IsPreserveElement(childTag) ||
				childRole == "article" ||
				semanticClassRe.MatchString(childClass) {
				hasPreservedElements = true
			}
		})

		return hasPreservedElements
	}

	isWrapperElement := func(el *goquery.Selection) bool {
		// If it directly contains inline content, it's NOT a wrapper
		if hasDirectInlineContent(el) {
			return false
		}

		// Check if it's just empty space
		text := strings.TrimSpace(el.Text())
		if text == "" {
			return true
		}

		// Check if it only contains other block elements
		children := el.Children()
		if children.Length() == 0 {
			return true
		}

		// Check if all children are block elements
		allBlockElements := true
		blockElements := constants.GetBlockElements()
		additionalBlocks := []string{"p", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "pre", "blockquote", "figure"}

		children.Each(func(_ int, child *goquery.Selection) {
			tag := goquery.NodeName(child)
			isBlock := slices.Contains(blockElements, tag)

			// Check additional block elements
			if !isBlock {
				if slices.Contains(additionalBlocks, tag) {
					isBlock = true
				}
			}

			if !isBlock {
				allBlockElements = false
			}
		})

		if allBlockElements {
			return true
		}

		// Check for common wrapper patterns
		className := strings.ToLower(el.AttrOr("class", ""))
		if wrapperClassRe.MatchString(className) {
			return true
		}

		// Check if it has excessive whitespace or empty text nodes
		hasTextContent := false
		el.Contents().Each(func(_ int, child *goquery.Selection) {
			if goquery.NodeName(child) == "#text" {
				childText := strings.TrimSpace(child.Text())
				if childText != "" {
					hasTextContent = true
				}
			}
		})

		if !hasTextContent {
			return true
		}

		// Check if it only contains block elements (different check)
		hasOnlyBlockElements := children.Length() > 0
		inlineElements := constants.GetInlineElements()

		children.Each(func(_ int, child *goquery.Selection) {
			tag := goquery.NodeName(child)
			if slices.Contains(inlineElements, tag) {
				hasOnlyBlockElements = false
			}
		})

		return hasOnlyBlockElements
	}

	// Function to process a single element
	processElement := func(el *goquery.Selection) bool {
		// Skip processing if element has been removed or should be preserved
		if el.Length() == 0 || shouldPreserveElement(el) {
			return false
		}

		tagName := goquery.NodeName(el)

		// Case 1: Element is truly empty (no text content, no child elements) and not self-closing
		allowedEmptyElements := constants.GetAllowedEmptyElements()
		isAllowedEmpty := slices.Contains(allowedEmptyElements, tagName)

		if !isAllowedEmpty && el.Children().Length() == 0 && strings.TrimSpace(el.Text()) == "" {
			el.Remove()
			processedCount++
			return true
		}

		// Case 2: Top-level element - be more aggressive
		if el.Parent().Length() > 0 && el.Parent().Get(0) == element.Get(0) {
			children := el.Children()
			hasOnlyBlockElements := children.Length() > 0
			inlineElements := constants.GetInlineElements()

			children.Each(func(_ int, child *goquery.Selection) {
				tag := goquery.NodeName(child)
				if slices.Contains(inlineElements, tag) {
					hasOnlyBlockElements = false
				}
			})

			if hasOnlyBlockElements {
				html, _ := el.Html()
				el.ReplaceWithHtml(html)
				processedCount++
				return true
			}
		}

		// Case 3: Wrapper element - merge up aggressively
		if isWrapperElement(el) {
			// Special case: if element only contains block elements, merge them up
			children := el.Children()
			onlyBlockElements := true
			inlineElements := constants.GetInlineElements()

			children.Each(func(_ int, child *goquery.Selection) {
				tag := goquery.NodeName(child)
				if slices.Contains(inlineElements, tag) {
					onlyBlockElements = false
				}
			})

			if onlyBlockElements {
				html, _ := el.Html()
				el.ReplaceWithHtml(html)
				processedCount++
				return true
			}

			// Otherwise handle as normal wrapper
			html, _ := el.Html()
			el.ReplaceWithHtml(html)
			processedCount++
			return true
		}

		// Case 4: Element only contains text and/or inline elements - convert to paragraph
		hasOnlyInlineOrText := true
		hasContent := false
		inlineElements := constants.GetInlineElements()

		el.Contents().Each(func(_ int, child *goquery.Selection) {
			if goquery.NodeName(child) == "#text" {
				text := strings.TrimSpace(child.Text())
				if text != "" {
					hasContent = true
				}
			} else {
				tag := goquery.NodeName(child)
				isInline := slices.Contains(inlineElements, tag)
				if !isInline {
					hasOnlyInlineOrText = false
				}
			}
		})

		if hasOnlyInlineOrText && hasContent {
			html, _ := el.Html()
			el.ReplaceWithHtml("<p>" + html + "</p>")
			processedCount++
			return true
		}

		// Case 5: Element has single child - unwrap only if child is block-level
		children := el.Children()
		if children.Length() == 1 {
			child := children.First()
			childTag := goquery.NodeName(child)

			// Only unwrap if the single child is a block element and not preserved
			blockElements := constants.GetBlockElements()
			isBlockChild := slices.Contains(blockElements, childTag)

			if isBlockChild && !shouldPreserveElement(child) {
				childHTML, _ := child.Html()
				el.ReplaceWithHtml("<" + childTag + ">" + childHTML + "</" + childTag + ">")
				processedCount++
				return true
			}
		}

		// Case 6: Deeply nested element - merge up
		nestingDepth := 0
		parent := el.Parent()
		blockElements := constants.GetBlockElements()

		for parent.Length() > 0 {
			parentTag := goquery.NodeName(parent)
			if slices.Contains(blockElements, parentTag) {
				nestingDepth++
			}
			parent = parent.Parent()
		}

		// Only unwrap if nested AND does not contain direct inline content
		if nestingDepth > 0 && !hasDirectInlineContent(el) {
			html, _ := el.Html()
			el.ReplaceWithHtml(html)
			processedCount++
			return true
		}

		return false
	}

	// First pass: Process top-level wrapper elements
	processTopLevelElements := func() bool {
		modified := false
		blockElements := constants.GetBlockElements()

		element.Children().Each(func(_ int, el *goquery.Selection) {
			tag := goquery.NodeName(el)
			isBlock := slices.Contains(blockElements, tag)

			if isBlock && processElement(el) {
				modified = true
			}
		})

		return modified
	}

	// Second pass: Process remaining wrapper elements from deepest to shallowest
	processRemainingElements := func() bool {
		modified := false
		blockElements := constants.GetBlockElements()
		blockSelector := strings.Join(blockElements, ",")

		// Get all wrapper elements and sort by depth (deepest first)
		var allElements []*goquery.Selection
		element.Find(blockSelector).Each(func(_ int, el *goquery.Selection) {
			allElements = append(allElements, el)
		})

		// Sort by depth descending (deepest first)
		slices.SortFunc(allElements, func(a, b *goquery.Selection) int {
			return cmp.Compare(b.Parents().Length(), a.Parents().Length())
		})

		for _, el := range allElements {
			if processElement(el) {
				modified = true
			}
		}

		return modified
	}

	// Final cleanup pass - aggressively flatten remaining wrapper elements
	finalCleanup := func() bool {
		modified := false
		blockElements := constants.GetBlockElements()
		blockSelector := strings.Join(blockElements, ",")

		element.Find(blockSelector).Each(func(_ int, el *goquery.Selection) {
			// Check if element only contains paragraphs
			children := el.Children()
			onlyParagraphs := children.Length() > 0

			children.Each(func(_ int, child *goquery.Selection) {
				if goquery.NodeName(child) != "p" {
					onlyParagraphs = false
				}
			})

			// Unwrap if it only contains paragraphs OR is a non-preserved wrapper element
			if onlyParagraphs || (!shouldPreserveElement(el) && isWrapperElement(el)) {
				html, _ := el.Html()
				el.ReplaceWithHtml(html)
				processedCount++
				modified = true
			}
		})

		return modified
	}

	// Execute all passes until no more changes
	for keepProcessing {
		keepProcessing = false
		if processTopLevelElements() {
			keepProcessing = true
		}
		if processRemainingElements() {
			keepProcessing = true
		}
		if finalCleanup() {
			keepProcessing = true
		}
	}

	endTime := time.Now()
	processingTime := float64(endTime.Sub(startTime).Nanoseconds()) / 1e6 // Convert to milliseconds
	slog.Debug("Flattened wrapper elements",
		"count", processedCount,
		"processingTime", processingTime)
}

// stripUnwantedAttributes removes unwanted attributes from elements
// JavaScript original code:
//
//	function stripUnwantedAttributes(element: Element, debug: boolean): void {
//		let attributeCount = 0;
//
//		const processElement = (el: Element) => {
//			// Skip SVG elements - preserve all their attributes
//			if (el.tagName.toLowerCase() === 'svg' || el.namespaceURI === 'http://www.w3.org/2000/svg') {
//				return;
//			}
//
//			const attributes = Array.from(el.attributes);
//			const tag = el.tagName.toLowerCase();
//
//			attributes.forEach(attr => {
//				const attrName = attr.name.toLowerCase();
//				const attrValue = attr.value;
//
//				// Special cases for preserving specific attributes
//				if (
//					// Preserve footnote IDs
//					(attrName === 'id' && (
//						attrValue.startsWith('fnref:') || // Footnote reference
//						attrValue.startsWith('fn:') || // Footnote content
//						attrValue === 'footnotes' // Footnotes container
//					)) ||
//					// Preserve code block language classes and footnote backref class
//					(attrName === 'class' && (
//						(tag === 'code' && attrValue.startsWith('language-')) ||
//						attrValue === 'footnote-backref'
//					))
//				) {
//					return;
//				}
//
//				// In debug mode, allow debug attributes and data- attributes
//				if (debug) {
//					if (!ALLOWED_ATTRIBUTES.has(attrName) &&
//						!ALLOWED_ATTRIBUTES_DEBUG.has(attrName) &&
//						!attrName.startsWith('data-')) {
//						el.removeAttribute(attr.name);
//						attributeCount++;
//					}
//				} else {
//					// In normal mode, only allow standard attributes
//					if (!ALLOWED_ATTRIBUTES.has(attrName)) {
//						el.removeAttribute(attr.name);
//						attributeCount++;
//					}
//				}
//			});
//		};
//
//		processElement(element);
//		element.querySelectorAll('*').forEach(processElement);
//
//		logDebug('Stripped attributes:', attributeCount);
//	}
func stripUnwantedAttributes(element *goquery.Selection, debug bool) {
	attributeCount := 0

	processElement := func(el *goquery.Selection) {
		if el.Length() == 0 {
			return
		}

		node := el.Get(0)

		// Skip SVG elements - preserve all their attributes
		tagName := strings.ToLower(node.Data)
		if tagName == "svg" || node.Namespace == "http://www.w3.org/2000/svg" {
			return
		}

		// Get all attributes and process them
		var attributesToRemove []string
		for _, attr := range node.Attr {
			attrName := strings.ToLower(attr.Key)
			attrValue := attr.Val

			// Special cases for preserving specific attributes
			preserveAttribute := false

			// Preserve footnote IDs
			if attrName == "id" && (strings.HasPrefix(attrValue, "fnref:") || // Footnote reference
				strings.HasPrefix(attrValue, "fn:") || // Footnote content
				attrValue == "footnotes") { // Footnotes container
				preserveAttribute = true
			}

			// Preserve code block language classes and footnote backref class
			if attrName == "class" && ((tagName == "code" && strings.HasPrefix(attrValue, "language-")) ||
				attrValue == "footnote-backref") {
				preserveAttribute = true
			}

			if preserveAttribute {
				continue
			}

			// In debug mode, allow debug attributes and data- attributes
			if debug {
				if !constants.IsAllowedAttribute(attrName) &&
					!constants.IsAllowedAttributeDebug(attrName) &&
					!strings.HasPrefix(attrName, "data-") {
					attributesToRemove = append(attributesToRemove, attr.Key)
					attributeCount++
				}
			} else {
				// In normal mode, only allow standard attributes
				if !constants.IsAllowedAttribute(attrName) {
					attributesToRemove = append(attributesToRemove, attr.Key)
					attributeCount++
				}
			}
		}

		// Remove unwanted attributes
		for _, attrName := range attributesToRemove {
			el.RemoveAttr(attrName)
		}
	}

	processElement(element)
	element.Find("*").Each(func(_ int, el *goquery.Selection) {
		processElement(el)
	})

	slog.Debug("Stripped attributes", "count", attributeCount)
}

// removeEmptyElements removes empty elements that don't contribute content
// JavaScript original code:
//
//	function removeEmptyElements(element: Element): void {
//		let removedCount = 0;
//		let iterations = 0;
//		let keepRemoving = true;
//
//		while (keepRemoving) {
//			iterations++;
//			keepRemoving = false;
//			// Get all elements without children, working from deepest first
//			const emptyElements = Array.from(element.getElementsByTagName('*')).filter(el => {
//				if (ALLOWED_EMPTY_ELEMENTS.has(el.tagName.toLowerCase())) {
//					return false;
//				}
//
//				// Check if element has only whitespace or &nbsp;
//				const textContent = el.textContent || '';
//				const hasOnlyWhitespace = textContent.trim().length === 0;
//				const hasNbsp = textContent.includes('\u00A0'); // Unicode non-breaking space
//
//				// Check if element has no meaningful children
//				const hasNoChildren = !el.hasChildNodes() ||
//					(Array.from(el.childNodes).every(node => {
//						if (isTextNode(node)) { // TEXT_NODE
//							const nodeText = node.textContent || '';
//							return nodeText.trim().length === 0 && !nodeText.includes('\u00A0');
//						}
//						return false;
//					}));
//
//				// Special case: Check for divs that only contain spans with commas
//				if (el.tagName.toLowerCase() === 'div') {
//					const children = Array.from(el.children);
//					const hasOnlyCommaSpans = children.length > 0 && children.every(child => {
//						if (child.tagName.toLowerCase() !== 'span') return false;
//						const content = child.textContent?.trim() || '';
//						return content === ',' || content === '' || content === ' ';
//					});
//					if (hasOnlyCommaSpans) return true;
//				}
//
//				return hasOnlyWhitespace && !hasNbsp && hasNoChildren;
//			});
//
//			if (emptyElements.length > 0) {
//				emptyElements.forEach(el => {
//					el.remove();
//					removedCount++;
//				});
//				keepRemoving = true;
//			}
//		}
//
//		logDebug('Removed empty elements:', removedCount, 'iterations:', iterations);
//	}
func removeEmptyElements(element *goquery.Selection) {
	removedCount := 0
	iterations := 0
	keepRemoving := true

	for keepRemoving {
		iterations++
		keepRemoving = false

		// Get all elements and filter for empty ones, working from deepest first
		var emptyElements []*goquery.Selection

		element.Find("*").Each(func(_ int, el *goquery.Selection) {
			tagName := strings.ToLower(goquery.NodeName(el))

			// Skip allowed empty elements
			if constants.IsAllowedEmptyElement(tagName) {
				return
			}

			// Check if element has only whitespace or &nbsp;
			textContent := el.Text()
			hasOnlyWhitespace := strings.TrimSpace(textContent) == ""
			hasNbsp := strings.Contains(textContent, "\u00A0") // Unicode non-breaking space

			// Check if element has no meaningful children
			hasNoChildren := true
			el.Contents().Each(func(_ int, child *goquery.Selection) {
				if goquery.NodeName(child) == "#text" {
					nodeText := child.Text()
					if strings.TrimSpace(nodeText) != "" || strings.Contains(nodeText, "\u00A0") {
						hasNoChildren = false
					}
				} else {
					hasNoChildren = false
				}
			})

			// If no child nodes at all, it's definitely empty
			if el.Contents().Length() == 0 {
				hasNoChildren = true
			}

			// Special case: Check for divs that only contain spans with commas
			if tagName == "div" {
				children := el.Children()
				if children.Length() > 0 {
					hasOnlyCommaSpans := true
					children.Each(func(_ int, child *goquery.Selection) {
						childTag := strings.ToLower(goquery.NodeName(child))
						if childTag != "span" {
							hasOnlyCommaSpans = false
							return
						}
						content := strings.TrimSpace(child.Text())
						if content != "," && content != "" && content != " " {
							hasOnlyCommaSpans = false
							return
						}
					})
					if hasOnlyCommaSpans {
						emptyElements = append(emptyElements, el)
						return
					}
				}
			}

			// Element is empty if it has only whitespace, no &nbsp;, and no meaningful children
			if hasOnlyWhitespace && !hasNbsp && hasNoChildren {
				emptyElements = append(emptyElements, el)
			}
		})

		// Remove empty elements
		if len(emptyElements) > 0 {
			for _, el := range emptyElements {
				el.Remove()
				removedCount++
			}
			keepRemoving = true
		}
	}

	slog.Debug("Removed empty elements",
		"count", removedCount,
		"iterations", iterations)
}

// removeTrailingHeadings removes headings at the end of content
// JavaScript original code:
//
//	function removeTrailingHeadings(element: Element): void {
//		const hasContentAfter = (el: Element): boolean => {
//			let sibling = el.nextElementSibling;
//			while (sibling) {
//				const text = sibling.textContent?.trim() || '';
//				if (text.length > 0) {
//					return true;
//				}
//				sibling = sibling.nextElementSibling;
//			}
//			return false;
//		};
//
//		const headings = element.querySelectorAll('h1, h2, h3, h4, h5, h6');
//		headings.forEach(heading => {
//			if (!hasContentAfter(heading)) {
//				heading.remove();
//			}
//		});
//	}
func removeTrailingHeadings(element *goquery.Selection) {
	hasContentAfter := func(el *goquery.Selection) bool {
		siblings := el.NextAll()
		hasContent := false
		siblings.Each(func(_ int, sibling *goquery.Selection) {
			text := strings.TrimSpace(sibling.Text())
			if text != "" {
				hasContent = true
			}
		})
		return hasContent
	}

	element.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, heading *goquery.Selection) {
		if !hasContentAfter(heading) {
			heading.Remove()
		}
	})
}

// stripExtraBrElements removes excessive br elements
// JavaScript original code:
//
//	function stripExtraBrElements(element: Element): void {
//		// Remove more than 2 consecutive br elements
//		const processBrs = () => {
//			const brs = Array.from(element.querySelectorAll('br'));
//			let consecutiveCount = 0;
//			let toRemove: Element[] = [];
//
//			brs.forEach((br, index) => {
//				const nextSibling = br.nextElementSibling;
//				if (nextSibling && nextSibling.tagName.toLowerCase() === 'br') {
//					consecutiveCount++;
//					if (consecutiveCount >= 2) {
//						toRemove.push(br);
//					}
//				} else {
//					consecutiveCount = 0;
//				}
//			});
//
//			toRemove.forEach(br => br.remove());
//		};
//
//		processBrs();
//	}
func stripExtraBrElements(element *goquery.Selection) {
	// Remove more than 2 consecutive br elements
	var toRemove []*goquery.Selection
	consecutiveCount := 0

	element.Find("br").Each(func(_ int, br *goquery.Selection) {
		next := br.Next()
		if next.Length() > 0 && goquery.NodeName(next) == "br" {
			consecutiveCount++
			if consecutiveCount >= 2 {
				toRemove = append(toRemove, br)
			}
		} else {
			consecutiveCount = 0
		}
	})

	for _, br := range toRemove {
		br.Remove()
	}
}

// removeEmptyLines removes empty lines and excessive whitespace
// JavaScript original code:
//
//	function removeEmptyLines(element: Element, doc: Document): void {
//		let removedCount = 0;
//		const startTime = Date.now();
//
//		// First pass: remove empty text nodes
//		const removeEmptyTextNodes = (node: Node) => {
//			// Skip if inside pre or code
//			if (isElement(node)) {
//				const tag = (node as Element).tagName.toLowerCase();
//				if (tag === 'pre' || tag === 'code') {
//					return;
//				}
//			}
//
//			// Process children first (depth-first)
//			const children = Array.from(node.childNodes);
//			children.forEach(removeEmptyTextNodes);
//
//			// Then handle this node
//			if (isTextNode(node)) {
//				const text = node.textContent || '';
//				// If it's completely empty or just special characters/whitespace, remove it
//				if (!text || text.match(/^[\u200C\u200B\u200D\u200E\u200F\uFEFF\xA0\s]*$/)) {
//					node.parentNode?.removeChild(node);
//					removedCount++;
//				} else {
//					// Clean up the text content while preserving important spaces
//					const newText = text
//						.replace(/\n{3,}/g, '\n\n') // More than 2 newlines -> 2 newlines
//						.replace(/^[\n\r\t]+/, '') // Remove leading newlines/tabs (preserve spaces)
//						.replace(/[\n\r\t]+$/, '') // Remove trailing newlines/tabs (preserve spaces)
//						.replace(/[ \t]*\n[ \t]*/g, '\n') // Remove spaces around newlines
//						.replace(/[ \t]{3,}/g, ' ') // 3+ spaces -> 1 space
//						.replace(/^[ ]+$/, ' ') // Multiple spaces between elements -> single space
//						.replace(/\s+([,.!?:;])/g, '$1') // Remove spaces before punctuation
//						// Clean up zero-width characters and multiple non-breaking spaces
//						.replace(/[\u200C\u200B\u200D\u200E\u200F\uFEFF]+/g, '')
//						.replace(/(?:\xA0){2,}/g, '\xA0'); // Multiple &nbsp; -> single &nbsp;
//
//					if (newText !== text) {
//						node.textContent = newText;
//						removedCount += text.length - newText.length;
//					}
//				}
//			}
//		};
//
//		// Second pass: clean up empty elements and normalize spacing
//		const cleanupEmptyElements = (node: Node) => {
//			if (!isElement(node)) return;
//
//			// Skip pre and code elements
//			const tag = node.tagName.toLowerCase();
//			if (tag === 'pre' || tag === 'code') {
//				return;
//			}
//
//			// Process children first (depth-first)
//			Array.from(node.childNodes)
//				.filter(isElement)
//				.forEach(cleanupEmptyElements);
//
//			// Then normalize this element's whitespace
//			node.normalize(); // Combine adjacent text nodes
//
//			// Special handling for block elements
//			const isBlockElement = getComputedStyle(node)?.display === 'block';
//
//			// Only remove empty text nodes at the start and end if they contain just newlines/tabs
//			// For block elements, also remove spaces
//			const startPattern = isBlockElement ? /^[\n\r\t \u200C\u200B\u200D\u200E\u200F\uFEFF\xA0]*$/ : /^[\n\r\t\u200C\u200B\u200D\u200E\u200F\uFEFF]*$/;
//			const endPattern = isBlockElement ? /^[\n\r\t \u200C\u200B\u200D\u200E\u200F\uFEFF\xA0]*$/ : /^[\n\r\t\u200C\u200B\u200D\u200E\u200F\uFEFF]*$/;
//
//			while (node.firstChild &&
//				   isTextNode(node.firstChild) &&
//				   (node.firstChild.textContent || '').match(startPattern)) {
//				node.removeChild(node.firstChild);
//				removedCount++;
//			}
//
//			while (node.lastChild &&
//				   isTextNode(node.lastChild) &&
//				   (node.lastChild.textContent || '').match(endPattern)) {
//				node.removeChild(node.lastChild);
//				removedCount++;
//			}
//
//			// Ensure there's a space between inline elements if needed
//			if (!isBlockElement) {
//				const children = Array.from(node.childNodes);
//				for (let i = 0; i < children.length - 1; i++) {
//					const current = children[i];
//					const next = children[i + 1];
//
//					// Only add space between elements or between element and text
//					if (isElement(current) || isElement(next)) {
//						// Get the text content
//						const nextContent = next.textContent || '';
//						const currentContent = current.textContent || '';
//
//						// Don't add space if:
//						// 1. Next content starts with punctuation or closing parenthesis
//						// 2. Current content ends with punctuation or opening parenthesis
//						// 3. There's already a space
//						const nextStartsWithPunctuation = nextContent.match(/^[,.!?:;)\]]/);
//						const currentEndsWithPunctuation = currentContent.match(/[,.!?:;(\[]\s*$/);
//
//						const hasSpace = (isTextNode(current) &&
//										(current.textContent || '').endsWith(' ')) ||
//										(isTextNode(next) &&
//										(next.textContent || '').startsWith(' '));
//
//						// Only add space if none of the above conditions are true
//						if (!nextStartsWithPunctuation &&
//							!currentEndsWithPunctuation &&
//							!hasSpace) {
//							const space = doc.createTextNode(' ');
//							node.insertBefore(space, next);
//						}
//					}
//				}
//			}
//		};
//
//		// Run both passes
//		removeEmptyTextNodes(element);
//		cleanupEmptyElements(element);
//
//		const endTime = Date.now();
//		logDebug('Removed empty lines:', {
//			charactersRemoved: removedCount,
//			processingTime: `${(endTime - startTime).toFixed(2)}ms`
//		});
//	}
func removeEmptyLines(element *goquery.Selection, _ *goquery.Document) {
	removedCount := 0
	startTime := time.Now()

	// First pass: remove empty text nodes and clean up text content
	var removeEmptyTextNodes func(node *html.Node)
	removeEmptyTextNodes = func(node *html.Node) {
		// Skip if inside pre or code
		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)
			if tag == "pre" || tag == "code" {
				return
			}
		}

		// Process children first (depth-first)
		var children []*html.Node
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			children = append(children, child)
		}
		for _, child := range children {
			removeEmptyTextNodes(child)
		}

		// Then handle this node
		if node.Type == html.TextNode {
			text := node.Data
			// If it's completely empty or just special characters/whitespace, remove it
			if text == "" || emptyTextRe.MatchString(text) {
				if node.Parent != nil {
					node.Parent.RemoveChild(node)
					removedCount++
				}
			} else {
				// Clean up the text content while preserving important spaces
				newText := text

				// More than 2 newlines -> 2 newlines
				newText = threeNewlinesRe.ReplaceAllString(newText, "\n\n")

				// Remove leading newlines/tabs (preserve spaces)
				newText = leadingNewlinesRe.ReplaceAllString(newText, "")

				// Remove trailing newlines/tabs (preserve spaces)
				newText = trailingNewlinesRe.ReplaceAllString(newText, "")

				// Remove spaces around newlines
				newText = spacesAroundNlRe.ReplaceAllString(newText, "\n")

				// 3+ spaces -> 1 space
				newText = threeSpacesRe.ReplaceAllString(newText, " ")

				// Multiple spaces between elements -> single space
				newText = onlySpacesRe.ReplaceAllString(newText, " ")

				// Remove spaces before punctuation
				newText = spaceBeforePunctRe.ReplaceAllString(newText, "$1")

				// Clean up zero-width characters and multiple non-breaking spaces
				newText = zeroWidthCharsRe.ReplaceAllString(newText, "")
				newText = multiNbspRe.ReplaceAllString(newText, "\xA0")

				if newText != text {
					node.Data = newText
					removedCount += len(text) - len(newText)
				}
			}
		}
	}

	// Second pass: clean up empty elements and normalize spacing
	var cleanupEmptyElements func(node *html.Node)
	cleanupEmptyElements = func(node *html.Node) {
		if node.Type != html.ElementNode {
			return
		}

		// Skip pre and code elements
		tag := strings.ToLower(node.Data)
		if tag == "pre" || tag == "code" {
			return
		}

		// Process children first (depth-first)
		var children []*html.Node
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				children = append(children, child)
			}
		}
		for _, child := range children {
			cleanupEmptyElements(child)
		}

		// Determine if this is a block element (simplified check)
		blockElements := constants.GetBlockElements()
		isBlockElement := slices.Contains(blockElements, tag)

		// Additional block elements
		additionalBlocks := []string{"p", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "pre", "blockquote", "figure"}
		if !isBlockElement {
			if slices.Contains(additionalBlocks, tag) {
				isBlockElement = true
			}
		}

		// Only remove empty text nodes at the start and end if they contain just newlines/tabs
		// For block elements, also remove spaces
		var startPattern, endPattern *regexp.Regexp
		if isBlockElement {
			startPattern = blockStartSpaceRe
			endPattern = blockStartSpaceRe
		} else {
			startPattern = inlineStartSpaceRe
			endPattern = inlineStartSpaceRe
		}

		// Remove empty text nodes at start
		for node.FirstChild != nil &&
			node.FirstChild.Type == html.TextNode &&
			startPattern.MatchString(node.FirstChild.Data) {
			node.RemoveChild(node.FirstChild)
			removedCount++
		}

		// Remove empty text nodes at end
		for node.LastChild != nil &&
			node.LastChild.Type == html.TextNode &&
			endPattern.MatchString(node.LastChild.Data) {
			node.RemoveChild(node.LastChild)
			removedCount++
		}

		// Ensure there's a space between inline elements if needed
		if !isBlockElement {
			var nodeChildren []*html.Node
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				nodeChildren = append(nodeChildren, child)
			}

			for i := range len(nodeChildren) - 1 {
				current := nodeChildren[i]
				next := nodeChildren[i+1]

				// Only add space between elements or between element and text
				if current.Type == html.ElementNode || next.Type == html.ElementNode {
					// Get the text content (simplified)
					var nextContent, currentContent string
					if next.Type == html.TextNode {
						nextContent = next.Data
					}
					if current.Type == html.TextNode {
						currentContent = current.Data
					}

					// Don't add space if:
					// 1. Next content starts with punctuation or closing parenthesis
					// 2. Current content ends with punctuation or opening parenthesis
					// 3. There's already a space
					nextStartsWithPunctuation := startsWithPunctRe.MatchString(nextContent)
					currentEndsWithPunctuation := endsWithPunctRe.MatchString(currentContent)

					hasSpace := (current.Type == html.TextNode && strings.HasSuffix(current.Data, " ")) ||
						(next.Type == html.TextNode && strings.HasPrefix(next.Data, " "))

					// Only add space if none of the above conditions are true
					if !nextStartsWithPunctuation &&
						!currentEndsWithPunctuation &&
						!hasSpace {
						space := &html.Node{
							Type: html.TextNode,
							Data: " ",
						}
						node.InsertBefore(space, next)
					}
				}
			}
		}
	}

	// Run both passes
	element.Each(func(_ int, sel *goquery.Selection) {
		if sel.Length() > 0 {
			removeEmptyTextNodes(sel.Get(0))
		}
	})

	element.Each(func(_ int, sel *goquery.Selection) {
		if sel.Length() > 0 {
			cleanupEmptyElements(sel.Get(0))
		}
	})

	endTime := time.Now()
	processingTime := float64(endTime.Sub(startTime).Nanoseconds()) / 1e6 // Convert to milliseconds
	slog.Debug("Removed empty lines",
		"charactersRemoved", removedCount,
		"processingTime", processingTime)
}

// transformListElement converts div[role="list"] to actual lists with complex nested handling
// JavaScript original code: (complex transform function from ELEMENT_STANDARDIZATION_RULES)
func transformListElement(el *goquery.Selection, doc *goquery.Document) *goquery.Selection {
	// First determine if this is an ordered list
	firstItem := el.Find(`div[role="listitem"] .label`).First()
	label := strings.TrimSpace(firstItem.Text())
	isOrdered := orderedListLabelRe.MatchString(label)

	// Create the appropriate list type
	listTag := "ul"
	if isOrdered {
		listTag = "ol"
	}

	// Create new list element
	newList := doc.Find("body").AppendHtml("<" + listTag + "></" + listTag + ">").Find(listTag).Last()

	// Process each list item
	el.Find(`div[role="listitem"]`).Each(func(_ int, item *goquery.Selection) {
		li := doc.Find("body").AppendHtml("<li></li>").Find("li").Last()
		content := item.Find(".content").First()

		if content.Length() > 0 {
			// Convert any paragraph divs inside content
			content.Find(`div[role="paragraph"]`).Each(func(_ int, div *goquery.Selection) {
				pHTML, _ := div.Html()
				div.ReplaceWithHtml("<p>" + pHTML + "</p>")
			})

			// Convert any nested lists recursively
			content.Find(`div[role="list"]`).Each(func(_ int, nestedList *goquery.Selection) {
				firstNestedItem := nestedList.Find(`div[role="listitem"] .label`).First()
				nestedLabel := strings.TrimSpace(firstNestedItem.Text())
				isNestedOrdered := orderedListLabelRe.MatchString(nestedLabel)

				nestedListTag := "ul"
				if isNestedOrdered {
					nestedListTag = "ol"
				}

				newNestedList := doc.Find("body").AppendHtml("<" + nestedListTag + "></" + nestedListTag + ">").Find(nestedListTag).Last()

				// Process nested items
				nestedList.Find(`div[role="listitem"]`).Each(func(_ int, nestedItem *goquery.Selection) {
					nestedLi := doc.Find("body").AppendHtml("<li></li>").Find("li").Last()
					nestedContent := nestedItem.Find(".content").First()

					if nestedContent.Length() > 0 {
						// Convert paragraph divs in nested items
						nestedContent.Find(`div[role="paragraph"]`).Each(func(_ int, div *goquery.Selection) {
							pHTML, _ := div.Html()
							div.ReplaceWithHtml("<p>" + pHTML + "</p>")
						})
						contentHTML, _ := nestedContent.Html()
						nestedLi.SetHtml(contentHTML)
					}

					newNestedList.AppendSelection(nestedLi)
				})

				nestedList.ReplaceWithSelection(newNestedList)
			})

			contentHTML, _ := content.Html()
			li.SetHtml(contentHTML)
		}

		newList.AppendSelection(li)
	})

	return newList
}

// transformListItemElement converts div[role="listitem"] to li elements
// JavaScript original code: (transform function for listitem)
func transformListItemElement(el *goquery.Selection, _ *goquery.Document) *goquery.Selection {
	content := el.Find(".content").First()
	if content.Length() == 0 {
		return el
	}

	// Convert any paragraph divs inside content
	content.Find(`div[role="paragraph"]`).Each(func(_ int, div *goquery.Selection) {
		pHTML, _ := div.Html()
		div.ReplaceWithHtml("<p>" + pHTML + "</p>")
	})

	return content
}
