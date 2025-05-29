package extractors

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GrokExtractor handles Grok (X.AI) conversation content extraction
// Corresponding to TypeScript class GrokExtractor extends ConversationExtractor
type GrokExtractor struct {
	*ConversationExtractorBase
	messageContainerSelector string
	messageBubbles           *goquery.Selection
	footnotes                []Footnote
	footnoteCounter          int
}

// NewGrokExtractor creates a new Grok extractor
// TypeScript original code:
//
//	constructor(document: Document, url: string) {
//		super(document, url);
//
//		// Note: This selector relies heavily on CSS utility classes and may break if Grok's UI changes.
//		this.messageContainerSelector = '.relative.group.flex.flex-col.justify-center.w-full';
//		this.messageBubbles = document.querySelectorAll(this.messageContainerSelector);
//		this.footnotes = [];
//		this.footnoteCounter = 0;
//	}
func NewGrokExtractor(document *goquery.Document, urlStr string, schemaOrgData interface{}) *GrokExtractor {
	// Note: This selector relies heavily on CSS utility classes and may break if Grok's UI changes.
	messageContainerSelector := ".relative.group.flex.flex-col.justify-center.w-full"

	return &GrokExtractor{
		ConversationExtractorBase: NewConversationExtractorBase(document, urlStr, schemaOrgData),
		messageContainerSelector:  messageContainerSelector,
		messageBubbles:            document.Find(messageContainerSelector),
		footnotes:                 make([]Footnote, 0),
		footnoteCounter:           0,
	}
}

// CanExtract checks if the extractor can extract content
// TypeScript original code:
//
//	canExtract(): boolean {
//		return this.messageBubbles && this.messageBubbles.length > 0;
//	}
func (g *GrokExtractor) CanExtract() bool {
	return g.messageBubbles.Length() > 0
}

// GetName returns the name of the extractor
func (g *GrokExtractor) GetName() string {
	return "GrokExtractor"
}

// Extract extracts the Grok conversation
// TypeScript original code:
//
//	extract(): ExtractorResult {
//		return this.extractWithDefuddle(this);
//	}
func (g *GrokExtractor) Extract() *ExtractorResult {
	return g.ExtractWithDefuddle(g)
}

// ExtractMessages extracts conversation messages
// TypeScript original code:
//
//	protected extractMessages(): ConversationMessage[] {
//		const messages: ConversationMessage[] = [];
//		this.footnotes = [];
//		this.footnoteCounter = 0;
//
//		this.messageBubbles.forEach((container) => {
//			// Note: Relies on layout classes 'items-end' and 'items-start' which might change.
//			const isUserMessage = container.classList.contains('items-end');
//			const isGrokMessage = container.classList.contains('items-start');
//
//			if (!isUserMessage && !isGrokMessage) {
//				return; // Skip elements that aren't clearly user or Grok messages
//			}
//
//			const messageBubble = container.querySelector('.message-bubble');
//			if (!messageBubble) {
//				return; // Skip if the core message bubble isn't found
//			}
//
//			let content: string;
//			let role: string;
//			let author: string;
//
//			if (isUserMessage) {
//				// Assume user message bubble's textContent is the desired content.
//				content = messageBubble.textContent || '';
//				role = 'user';
//				author = 'You';
//			} else if (isGrokMessage) {
//				role = 'assistant';
//				author = 'Grok';
//
//				// Clone the bubble to modify it without affecting the original page
//				const clonedBubble = messageBubble.cloneNode(true) as Element;
//
//				// Remove known non-content elements like the DeepSearch artifact
//				clonedBubble.querySelectorAll('.relative.border.border-border-l1.bg-surface-base').forEach(el => el.remove());
//
//				content = clonedBubble.innerHTML;
//
//				// Process footnotes/links in the cleaned content
//				content = this.processFootnotes(content);
//			}
//
//			if (content && content.trim()) {
//				messages.push({
//					author,
//					content: content.trim(),
//					metadata: { role }
//				});
//			}
//		});
//
//		return messages;
//	}
func (g *GrokExtractor) ExtractMessages() []ConversationMessage {
	var messages []ConversationMessage
	g.footnotes = make([]Footnote, 0)
	g.footnoteCounter = 0

	g.messageBubbles.Each(func(i int, container *goquery.Selection) {
		// Note: Relies on layout classes 'items-end' and 'items-start' which might change.
		isUserMessage := container.HasClass("items-end")
		isGrokMessage := container.HasClass("items-start")

		if !isUserMessage && !isGrokMessage {
			return // Skip elements that aren't clearly user or Grok messages
		}

		messageBubble := container.Find(".message-bubble").First()
		if messageBubble.Length() == 0 {
			return // Skip if the core message bubble isn't found
		}

		var content string
		var role string
		var author string

		if isUserMessage {
			// Assume user message bubble's textContent is the desired content.
			// This is simpler and potentially less brittle than selecting specific spans.
			content = messageBubble.Text()
			role = "user"
			author = "You" // Or potentially extract from an attribute if available later
		} else if isGrokMessage {
			role = "assistant"
			author = "Grok" // Or potentially extract from an attribute if available later

			// Clone the bubble to modify it without affecting the original page
			clonedBubbleHTML, _ := messageBubble.Html()
			clonedDoc, err := goquery.NewDocumentFromReader(strings.NewReader(clonedBubbleHTML))
			if err != nil {
				return
			}

			// Remove known non-content elements like the DeepSearch artifact
			clonedDoc.Find(".relative.border.border-border-l1.bg-surface-base").Remove()
			// Add selectors here for any other known elements to remove (e.g., buttons, toolbars within the bubble)

			clonedContent, _ := clonedDoc.Html()
			content = clonedContent

			// Process footnotes/links in the cleaned content
			content = g.processFootnotes(content)
		}

		if strings.TrimSpace(content) != "" {
			messages = append(messages, ConversationMessage{
				Author:  author,
				Content: strings.TrimSpace(content),
				Metadata: map[string]interface{}{
					"role": role,
				},
			})
		}
	})

	return messages
}

// GetFootnotes returns the conversation footnotes
// TypeScript original code:
//
//	protected getFootnotes(): Footnote[] {
//		return this.footnotes;
//	}
func (g *GrokExtractor) GetFootnotes() []Footnote {
	return g.footnotes
}

// GetMetadata returns conversation metadata
// TypeScript original code:
//
//	protected getMetadata(): ConversationMetadata {
//		const title = this.getTitle();
//		const messageCount = this.messageBubbles.length;
//
//		return {
//			title,
//			site: 'Grok',
//			url: this.url,
//			messageCount, // Use estimated count
//			description: `Grok conversation with ${messageCount} messages`
//		};
//	}
func (g *GrokExtractor) GetMetadata() ConversationMetadata {
	title := g.getTitle()
	messageCount := g.messageBubbles.Length()

	return ConversationMetadata{
		Title:        title,
		Site:         "Grok",
		URL:          g.url,
		MessageCount: messageCount, // Use estimated count
		Description:  fmt.Sprintf("Grok conversation with %d messages", messageCount),
	}
}

// getTitle extracts the conversation title
// TypeScript original code:
//
//	private getTitle(): string {
//		// Try to get the page title first (more reliable)
//		const pageTitle = this.document.title?.trim();
//		if (pageTitle && pageTitle !== 'Grok' && !pageTitle.startsWith('Grok by ')) {
//			// Remove ' - Grok' suffix if present
//			return pageTitle.replace(/\s-\s*Grok$/, '').trim();
//		}
//
//		// Fallback: Find the first user message bubble and use its text content
//		// Note: Still relies on 'items-end' class.
//		const firstUserContainer = this.document.querySelector(`${this.messageContainerSelector}.items-end`);
//		if (firstUserContainer) {
//			const messageBubble = firstUserContainer.querySelector('.message-bubble');
//			if (messageBubble) {
//				const text = messageBubble.textContent?.trim() || '';
//				// Truncate to first 50 characters if longer
//				return text.length > 50 ? text.substring(0, 50) + '...' : text;
//			}
//		}
//
//		return 'Grok Conversation'; // Default fallback
//	}
func (g *GrokExtractor) getTitle() string {
	// Try to get the page title first (more reliable)
	pageTitle := strings.TrimSpace(g.document.Find("title").Text())
	if pageTitle != "" && pageTitle != "Grok" && !strings.HasPrefix(pageTitle, "Grok by ") {
		// Remove ' - Grok' suffix if present
		re := regexp.MustCompile(`\s-\s*Grok$`)
		return strings.TrimSpace(re.ReplaceAllString(pageTitle, ""))
	}

	// Fallback: Find the first user message bubble and use its text content
	// Note: Still relies on 'items-end' class.
	firstUserContainer := g.document.Find(fmt.Sprintf("%s.items-end", g.messageContainerSelector)).First()
	if firstUserContainer.Length() > 0 {
		messageBubble := firstUserContainer.Find(".message-bubble").First()
		if messageBubble.Length() > 0 {
			text := strings.TrimSpace(messageBubble.Text())
			// Truncate to first 50 characters if longer
			if len(text) > 50 {
				return text[:50] + "..."
			}
			return text
		}
	}

	return "Grok Conversation" // Default fallback
}

// processFootnotes processes links in content and converts them to footnotes
// TypeScript original code:
//
//	private processFootnotes(content: string): string {
//		// Regex to find <a> tags, capture href and link text
//		const linkPattern = /<a\s+(?:[^>]*?\s+)?href="([^"]*)"[^>]*>(.*?)<\/a>/g;
//
//		return content.replace(linkPattern, (match, url, linkText) => {
//			// Skip processing for internal anchor links, empty URLs, or non-http(s) protocols
//			if (!url || url.startsWith('#')) {
//				return match;
//			}
//
//			if (!/^https?:\/\//.test(url)) {
//				return match;
//			}
//
//			// Check if this URL already exists in our footnotes
//			let footnoteIndex = this.footnotes.findIndex(fn => fn.url === url);
//
//			if (footnoteIndex === -1) {
//				// Create a new footnote if URL doesn't exist
//				this.footnoteCounter++;
//				footnoteIndex = this.footnoteCounter - 1;
//
//				let domainText = url; // Default to full URL if parsing fails
//				try {
//					const domain = new URL(url).hostname.replace(/^www\./, '');
//					domainText = `<a href="${url}" target="_blank" rel="noopener noreferrer">${domain}</a>`;
//				} catch (e) {
//					domainText = `<a href="${url}" target="_blank" rel="noopener noreferrer">${url}</a>`;
//				}
//
//				this.footnotes.push({
//					url,
//					text: domainText // Store the link HTML directly
//				});
//			}
//
//			const footnoteNumber = footnoteIndex + 1;
//
//			// Return the original link text wrapped with a footnote reference
//			return `${linkText}<sup id="fnref:${footnoteNumber}" class="footnote-ref"><a href="#fn:${footnoteNumber}" class="footnote-link">${footnoteNumber}</a></sup>`;
//		});
//	}
func (g *GrokExtractor) processFootnotes(content string) string {
	// Regex to find <a> tags, capture href and link text
	linkPattern := regexp.MustCompile(`<a\s+(?:[^>]*?\s+)?href="([^"]*)"[^>]*>(.*?)</a>`)

	return linkPattern.ReplaceAllStringFunc(content, func(match string) string {
		matches := linkPattern.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}

		urlStr := matches[1]
		linkText := matches[2]

		// Skip processing for internal anchor links, empty URLs, or non-http(s) protocols
		if urlStr == "" || strings.HasPrefix(urlStr, "#") {
			return match
		}

		httpPattern := regexp.MustCompile(`^https?://`)
		if !httpPattern.MatchString(urlStr) {
			return match
		}

		// Check if this URL already exists in our footnotes
		var footnoteIndex int
		found := false

		for idx, footnote := range g.footnotes {
			if footnote.URL == urlStr {
				footnoteIndex = idx + 1 // 1-based index
				found = true
				break
			}
		}

		if !found {
			// Create a new footnote if URL doesn't exist
			g.footnoteCounter++
			footnoteIndex = g.footnoteCounter

			var domainText string
			if parsedURL, err := url.Parse(urlStr); err == nil {
				domain := strings.TrimPrefix(parsedURL.Hostname(), "www.")
				domainText = fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer">%s</a>`, urlStr, domain)
			} else {
				// Use full URL if parsing fails
				domainText = fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer">%s</a>`, urlStr, urlStr)
			}

			g.footnotes = append(g.footnotes, Footnote{
				URL:  urlStr,
				Text: domainText, // Store the link HTML directly
			})
		}

		// Return the original link text wrapped with a footnote reference
		// Ensure the link text itself is not clickable again if it was part of the original match
		return fmt.Sprintf(`%s<sup id="fnref:%d" class="footnote-ref"><a href="#fn:%d" class="footnote-link">%d</a></sup>`,
			linkText, footnoteIndex, footnoteIndex, footnoteIndex)
	})
}
