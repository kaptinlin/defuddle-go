// Package elements provides enhanced element processing functionality
// This module handles heading processing including navigation element removal,
// anchor link cleanup, and text content extraction
package elements

import (
	"log/slog"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

/*
TypeScript source code (headings.ts, 105 lines):

This module provides heading processing functionality including:
- Navigation element removal from headings
- Anchor link and button cleanup
- Text content extraction and normalization
- Heading structure simplification

Original TypeScript implementation:
export const headingRules = [
    // Simplify headings by removing internal navigation elements
	{
		selector: 'h1, h2, h3, h4, h5, h6',
		element: 'keep',
		transform: (el: Element): Element => {
			// Get document from element's owner document
			const doc = el.ownerDocument;
			if (!doc) {
				console.warn('No document available');
				return el;
			}

			// Create new heading of same level
			const newHeading = doc.createElement(el.tagName);

			// Copy allowed attributes from original heading
			Array.from(el.attributes).forEach(attr => {
				if (ALLOWED_ATTRIBUTES.has(attr.name)) {
					newHeading.setAttribute(attr.name, attr.value);
				}
			});

			// Clone the element so we can modify it without affecting the original
			const clone = el.cloneNode(true) as Element;

			// First extract text from navigation elements before removing them
			const navigationText = new Map<Element, string>();

			// Find all navigation elements and store their text content
			Array.from(clone.querySelectorAll('*')).forEach(child => {
				let shouldRemove = false;

				if (child.tagName.toLowerCase() === 'a') {
					const href = child.getAttribute('href');
					if (href?.includes('#') || href?.startsWith('#')) {
						navigationText.set(child, child.textContent?.trim() || '');
						shouldRemove = true;
					}
				}
				if (child.classList.contains('anchor')) {
					navigationText.set(child, child.textContent?.trim() || '');
					shouldRemove = true;
				}
				if (child.tagName.toLowerCase() === 'button') {
					shouldRemove = true;
				}
				if ((child.tagName.toLowerCase() === 'span' || child.tagName.toLowerCase() === 'div') &&
					child.querySelector('a[href^="#"]')) {
					const anchor = child.querySelector('a[href^="#"]');
					if (anchor) {
						navigationText.set(child, anchor.textContent?.trim() || '');
					}
					shouldRemove = true;
				}

				if (shouldRemove) {
					// If this element contains the only text content of its parent,
					// store its text to be used for the parent
					const parent = child.parentElement;
					if (parent && parent !== clone &&
						parent.textContent?.trim() === child.textContent?.trim()) {
						navigationText.set(parent, child.textContent?.trim() || '');
					}
				}
			});

			// Remove navigation elements
			const toRemove = Array.from(clone.querySelectorAll('*')).filter(child => {
				if (child.tagName.toLowerCase() === 'a') {
					const href = child.getAttribute('href');
					return href?.includes('#') || href?.startsWith('#');
				}
				if (child.classList.contains('anchor')) {
					return true;
				}
				if (child.tagName.toLowerCase() === 'button') {
					return true;
				}
				if ((child.tagName.toLowerCase() === 'span' || child.tagName.toLowerCase() === 'div') &&
					child.querySelector('a[href^="#"]')) {
					return true;
				}
				return false;
			});

			toRemove.forEach(element => element.remove());

			// Get the text content after removing navigation elements
			let textContent = clone.textContent?.trim() || '';

			// If we lost all text content but had navigation text, use that instead
			if (!textContent && navigationText.size > 0) {
				textContent = Array.from(navigationText.values())[0];
			}

			// Set the clean text content
			newHeading.textContent = textContent;

			return newHeading;
		}
	}
];
*/

// HeadingProcessor handles heading processing and enhancement
// TypeScript original code:
// export const headingRules = [
//
//	{
//	  selector: 'h1, h2, h3, h4, h5, h6',
//	  element: 'keep',
//	  transform: (el: Element): Element => {
//	    // Processing logic here
//	  }
//	}
//
// ];
type HeadingProcessor struct {
	doc *goquery.Document
}

// HeadingProcessingOptions contains options for heading processing
// TypeScript original code:
//
//	interface HeadingOptions {
//	  removeNavigation?: boolean;
//	  preserveStructure?: boolean;
//	  allowedAttributes?: string[];
//	}
type HeadingProcessingOptions struct {
	RemoveNavigation  bool
	PreserveStructure bool
	AllowedAttributes []string
}

// DefaultHeadingProcessingOptions returns default options for heading processing
// TypeScript original code:
//
//	const defaultOptions: HeadingOptions = {
//	  removeNavigation: true,
//	  preserveStructure: true,
//	  allowedAttributes: ['id', 'class', 'data-*']
//	};
func DefaultHeadingProcessingOptions() *HeadingProcessingOptions {
	return &HeadingProcessingOptions{
		RemoveNavigation:  true,
		PreserveStructure: true,
		AllowedAttributes: []string{"id", "class"},
	}
}

// NewHeadingProcessor creates a new heading processor
// TypeScript original code:
//
//	class HeadingProcessor {
//	  constructor(private document: Document) {}
//	}
func NewHeadingProcessor(doc *goquery.Document) *HeadingProcessor {
	return &HeadingProcessor{
		doc: doc,
	}
}

// ProcessHeadings processes all headings in the document
// TypeScript original code:
// export const headingRules = [
//
//	{
//	  selector: 'h1, h2, h3, h4, h5, h6',
//	  element: 'keep',
//	  transform: (el: Element): Element => {
//	    // Processing logic
//	  }
//	}
//
// ];
func (p *HeadingProcessor) ProcessHeadings(options *HeadingProcessingOptions) {
	if options == nil {
		options = DefaultHeadingProcessingOptions()
	}

	slog.Debug("processing headings", "removeNavigation", options.RemoveNavigation, "preserveStructure", options.PreserveStructure)

	var processedCount int
	p.doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
		p.processHeading(s, options)
		processedCount++
	})

	slog.Info("headings processed", "count", processedCount)
}

// processHeading processes a single heading element
// TypeScript original code:
//
//	transform: (el: Element): Element => {
//	  // Get document from element's owner document
//	  const doc = el.ownerDocument;
//	  if (!doc) {
//	    console.warn('No document available');
//	    return el;
//	  }
//
//	  // Create new heading of same level
//	  const newHeading = doc.createElement(el.tagName);
//
//	  // Copy allowed attributes from original heading
//	  Array.from(el.attributes).forEach(attr => {
//	    if (ALLOWED_ATTRIBUTES.has(attr.name)) {
//	      newHeading.setAttribute(attr.name, attr.value);
//	    }
//	  });
//
//	  // Clone the element so we can modify it without affecting the original
//	  const clone = el.cloneNode(true) as Element;
//	  // Processing logic...
//	}
func (p *HeadingProcessor) processHeading(s *goquery.Selection, options *HeadingProcessingOptions) {
	slog.Debug("processing individual heading", "tag", goquery.NodeName(s))

	if !options.RemoveNavigation {
		return
	}

	// Clone the heading for processing
	clone := s.Clone()

	// Extract navigation text before removing elements
	navigationTexts := p.extractNavigationTexts(clone)

	// Remove navigation elements
	p.removeNavigationElements(clone)

	// Get cleaned text content
	textContent := strings.TrimSpace(clone.Text())

	// If we lost all text content but had navigation text, use that instead
	if textContent == "" && len(navigationTexts) > 0 {
		textContent = navigationTexts[0]
	}

	// Create new heading with cleaned content
	if options.PreserveStructure {
		p.replaceHeadingContent(s, textContent, options)
	} else {
		s.SetText(textContent)
	}

	slog.Debug("cleaned heading", "originalLength", len(s.Text()), "cleanedLength", len(textContent))
}

// extractNavigationTexts extracts text from navigation elements before removal
// TypeScript original code:
// // First extract text from navigation elements before removing them
// const navigationText = new Map<Element, string>();
//
// // Find all navigation elements and store their text content
//
//	Array.from(clone.querySelectorAll('*')).forEach(child => {
//	  let shouldRemove = false;
//
//	  if (child.tagName.toLowerCase() === 'a') {
//	    const href = child.getAttribute('href');
//	    if (href?.includes('#') || href?.startsWith('#')) {
//	      navigationText.set(child, child.textContent?.trim() || '');
//	      shouldRemove = true;
//	    }
//	  }
//	  if (child.classList.contains('anchor')) {
//	    navigationText.set(child, child.textContent?.trim() || '');
//	    shouldRemove = true;
//	  }
//	  if (child.tagName.toLowerCase() === 'button') {
//	    shouldRemove = true;
//	  }
//	  if ((child.tagName.toLowerCase() === 'span' || child.tagName.toLowerCase() === 'div') &&
//	    child.querySelector('a[href^="#"]')) {
//	    const anchor = child.querySelector('a[href^="#"]');
//	    if (anchor) {
//	      navigationText.set(child, anchor.textContent?.trim() || '');
//	    }
//	    shouldRemove = true;
//	  }
//
//	  if (shouldRemove) {
//	    // If this element contains the only text content of its parent,
//	    // store its text to be used for the parent
//	    const parent = child.parentElement;
//	    if (parent && parent !== clone &&
//	      parent.textContent?.trim() === child.textContent?.trim()) {
//	      navigationText.set(parent, child.textContent?.trim() || '');
//	    }
//	  }
//	});
func (p *HeadingProcessor) extractNavigationTexts(s *goquery.Selection) []string {
	var navigationTexts []string
	textMap := make(map[string]bool) // To avoid duplicates

	s.Find("*").Each(func(_ int, child *goquery.Selection) {
		shouldExtract := false
		var extractedText string

		// Check for anchor links with hash
		if child.Is("a") {
			href, hasHref := child.Attr("href")
			if hasHref && (strings.Contains(href, "#") || strings.HasPrefix(href, "#")) {
				extractedText = strings.TrimSpace(child.Text())
				shouldExtract = true
			}
		}

		// Check for anchor class
		if child.HasClass("anchor") {
			extractedText = strings.TrimSpace(child.Text())
			shouldExtract = true
		}

		// Check for buttons
		if child.Is("button") {
			shouldExtract = true // But don't extract text from buttons
		}

		// Check for spans/divs containing anchor links
		if child.Is("span, div") {
			anchor := child.Find("a[href^=\"#\"]").First()
			if anchor.Length() > 0 {
				extractedText = strings.TrimSpace(anchor.Text())
				shouldExtract = true
			}
		}

		// Store navigation text if it's meaningful
		if shouldExtract && extractedText != "" && !textMap[extractedText] {
			navigationTexts = append(navigationTexts, extractedText)
			textMap[extractedText] = true

			// Also check parent-child text relationship like TypeScript
			parent := child.Parent()
			childText := strings.TrimSpace(child.Text())
			parentText := strings.TrimSpace(parent.Text())
			if parentText == childText && !textMap[parentText] {
				navigationTexts = append(navigationTexts, parentText)
				textMap[parentText] = true
			}
		}
	})

	return navigationTexts
}

// removeNavigationElements removes navigation elements from heading
// TypeScript original code:
// // Remove navigation elements
//
//	const toRemove = Array.from(clone.querySelectorAll('*')).filter(child => {
//	  if (child.tagName.toLowerCase() === 'a') {
//	    const href = child.getAttribute('href');
//	    return href?.includes('#') || href?.startsWith('#');
//	  }
//	  if (child.classList.contains('anchor')) {
//	    return true;
//	  }
//	  if (child.tagName.toLowerCase() === 'button') {
//	    return true;
//	  }
//	  if ((child.tagName.toLowerCase() === 'span' || child.tagName.toLowerCase() === 'div') &&
//	    child.querySelector('a[href^="#"]')) {
//	    return true;
//	  }
//	  return false;
//	});
//
// toRemove.forEach(element => element.remove());
func (p *HeadingProcessor) removeNavigationElements(s *goquery.Selection) {
	var toRemove []*goquery.Selection

	s.Find("*").Each(func(_ int, child *goquery.Selection) {
		shouldRemove := false

		// Remove anchor links with hash
		if child.Is("a") {
			href, hasHref := child.Attr("href")
			if hasHref && (strings.Contains(href, "#") || strings.HasPrefix(href, "#")) {
				shouldRemove = true
			}
		}

		// Remove elements with anchor class
		if child.HasClass("anchor") {
			shouldRemove = true
		}

		// Remove buttons
		if child.Is("button") {
			shouldRemove = true
		}

		// Remove spans/divs containing anchor links
		if child.Is("span, div") {
			anchor := child.Find("a[href^=\"#\"]")
			if anchor.Length() > 0 {
				shouldRemove = true
			}
		}

		if shouldRemove {
			toRemove = append(toRemove, child)
		}
	})

	// Remove collected elements
	for _, element := range toRemove {
		element.Remove()
	}
}

// replaceHeadingContent replaces heading content while preserving structure
// TypeScript original code:
// // Create new heading of same level
// const newHeading = doc.createElement(el.tagName);
//
// // Copy allowed attributes from original heading
//
//	Array.from(el.attributes).forEach(attr => {
//	  if (ALLOWED_ATTRIBUTES.has(attr.name)) {
//	    newHeading.setAttribute(attr.name, attr.value);
//	  }
//	});
//
// // Set the clean text content
// newHeading.textContent = textContent;
func (p *HeadingProcessor) replaceHeadingContent(s *goquery.Selection, textContent string, options *HeadingProcessingOptions) {
	tagName := goquery.NodeName(s)

	// Build new heading HTML
	var headingHTML strings.Builder
	headingHTML.WriteString("<")
	headingHTML.WriteString(tagName)

	// Copy allowed attributes
	for _, attrName := range options.AllowedAttributes {
		if attrValue, hasAttr := s.Attr(attrName); hasAttr {
			headingHTML.WriteString(" ")
			headingHTML.WriteString(attrName)
			headingHTML.WriteString("=\"")
			// Escape attribute value
			escapedValue := strings.ReplaceAll(attrValue, "\"", "&quot;")
			headingHTML.WriteString(escapedValue)
			headingHTML.WriteString("\"")
		}
	}

	headingHTML.WriteString(">")
	// Escape text content
	escapedContent := strings.ReplaceAll(textContent, "&", "&amp;")
	escapedContent = strings.ReplaceAll(escapedContent, "<", "&lt;")
	escapedContent = strings.ReplaceAll(escapedContent, ">", "&gt;")
	headingHTML.WriteString(escapedContent)
	headingHTML.WriteString("</")
	headingHTML.WriteString(tagName)
	headingHTML.WriteString(">")

	// Replace original heading
	s.ReplaceWithHtml(headingHTML.String())
}

// ProcessHeadings processes all headings in the document (public interface)
// TypeScript original code:
//
//	export function processHeadings(doc: Document, options?: HeadingOptions): void {
//	  const processor = new HeadingProcessor(doc);
//	  processor.processAllHeadings(options || defaultOptions);
//	}
func ProcessHeadings(doc *goquery.Document, options *HeadingProcessingOptions) {
	processor := NewHeadingProcessor(doc)
	processor.ProcessHeadings(options)
}
