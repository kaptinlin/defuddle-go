package extractors

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Pre-compiled regex patterns for ChatGPT extraction.
var (
	chatgptEmptyParagraphRe = regexp.MustCompile(`<p[^>]*>\s*</p>`)
	chatgptCitationRe       = regexp.MustCompile(`(&ZeroWidthSpace;)?(<span[^>]*?>\s*<a[^>]*?href="([^"]+)"[^>]*?target="_blank"[^>]*?rel="noopener"[^>]*?>[\s\S]*?</a>\s*</span>)`)
)

// ChatGPTExtractor handles ChatGPT conversation content extraction
// TypeScript original code:
// import { ConversationExtractor } from './_conversation';
// import { ConversationMessage, ConversationMetadata, Footnote } from '../types/extractors';
//
//	export class ChatGPTExtractor extends ConversationExtractor {
//		private articles: NodeListOf<Element> | null;
//		private footnotes: Footnote[];
//		private footnoteCounter: number;
//
//		constructor(document: Document, url: string) {
//			super(document, url);
//			this.articles = document.querySelectorAll('article[data-testid^="conversation-turn-"]');
//			this.footnotes = [];
//			this.footnoteCounter = 0;
//		}
//
//		canExtract(): boolean {
//			return !!this.articles && this.articles.length > 0;
//		}
//
//		protected extractMessages(): ConversationMessage[] {
//			const messages: ConversationMessage[] = [];
//			this.footnotes = [];
//			this.footnoteCounter = 0;
//
//			if (!this.articles) return messages;
//
//			this.articles.forEach((article) => {
//				// Get the localized author text from the sr-only heading and clean it
//				const authorElement = article.querySelector('h5.sr-only, h6.sr-only');
//				const authorText = authorElement?.textContent
//					?.trim()
//					?.replace(/:\s*$/, '') // Remove colon and any trailing whitespace
//					|| '';
//
//				let currentAuthorRole = '';
//
//				const authorRole = article.getAttribute('data-message-author-role');
//				if (authorRole) {
//					currentAuthorRole = authorRole;
//				}
//
//				let messageContent = article.innerHTML || '';
//				messageContent = messageContent.replace(/\u200B/g, '');
//
//				// Remove specific elements from the message content
//				const tempDiv = document.createElement('div');
//				tempDiv.innerHTML = messageContent;
//				tempDiv.querySelectorAll('h5.sr-only, h6.sr-only, span[data-state="closed"]').forEach(el => el.remove());
//				messageContent = tempDiv.innerHTML;
//
//				// Process inline references using regex to find the containers
//				// Look for spans containing citation links (a[target=_blank][rel=noopener]), replacing entire structure
//				// Also capture optional preceding ZeroWidthSpace
//				const citationPattern = /(&ZeroWidthSpace;)?(<span[^>]*?>\s*<a(?=[^>]*?href="([^"]+)")(?=[^>]*?target="_blank")(?=[^>]*?rel="noopener")[^>]*?>[\s\S]*?<\/a>\s*<\/span>)/gi;
//
//				messageContent = messageContent.replace(citationPattern, (match, zws, spanStructure, url) => {
//					// url is captured group 3
//					let domain = '';
//					let fragmentText = '';
//
//					try {
//						// Extract domain without www.
//						domain = new URL(url).hostname.replace(/^www\./, '');
//
//						// Extract and decode the fragment text if it exists
//						const hashParts = url.split('#:~:text=');
//						if (hashParts.length > 1) {
//							fragmentText = decodeURIComponent(hashParts[1]);
//							fragmentText = fragmentText.replace(/%2C/g, ',');
//
//							const parts = fragmentText.split(',');
//							if (parts.length > 1 && parts[0].trim()) {
//								fragmentText = ` — ${parts[0].trim()}...`;
//							} else if (parts[0].trim()) {
//								fragmentText = ` — ${fragmentText.trim()}`;
//							} else {
//								fragmentText = '';
//							}
//						}
//					} catch (e) {
//						console.error(`Failed to parse URL: ${url}`, e);
//						domain = url;
//					}
//
//					// Check if this URL already exists in our footnotes
//					let footnoteIndex = this.footnotes.findIndex(fn => fn.url === url);
//					let footnoteNumber: number;
//
//					if (footnoteIndex === -1) {
//						this.footnoteCounter++;
//						footnoteNumber = this.footnoteCounter;
//						this.footnotes.push({
//							url,
//							text: `<a href="${url}">${domain}</a>${fragmentText}`
//						});
//					} else {
//						footnoteNumber = footnoteIndex + 1;
//					}
//
//					// Return just the footnote reference, replacing the ZWS (if captured) and the entire span structure
//					return `<sup id="fnref:${footnoteNumber}"><a href="#fn:${footnoteNumber}">${footnoteNumber}</a></sup>`;
//				});
//
//				// Clean up any stray empty paragraph tags
//				messageContent = messageContent
//					.replace(/<p[^>]*>\s*<\/p>/g, '');
//
//				messages.push({
//					author: authorText,
//					content: messageContent.trim(),
//					metadata: {
//						role: currentAuthorRole || 'unknown'
//					}
//				});
//
//			});
//
//			return messages;
//		}
//
//		protected getFootnotes(): Footnote[] {
//			return this.footnotes;
//		}
//
//		protected getMetadata(): ConversationMetadata {
//			const title = this.getTitle();
//			const messages = this.extractMessages();
//
//			return {
//				title,
//				site: 'ChatGPT',
//				url: this.url,
//				messageCount: messages.length,
//				description: `ChatGPT conversation with ${messages.length} messages`
//			};
//		}
//
//		private getTitle(): string {
//			// Try to get the page title first
//			const pageTitle = this.document.title?.trim();
//			if (pageTitle && pageTitle !== 'ChatGPT') {
//				return pageTitle;
//			}
//
//			// Fall back to first user message
//			const firstUserTurn = this.articles?.item(0)?.querySelector('.text-message');
//			if (firstUserTurn) {
//				const text = firstUserTurn.textContent || '';
//				// Truncate to first 50 characters if longer
//				return text.length > 50 ? text.slice(0, 50) + '...' : text;
//			}
//
//			return 'ChatGPT Conversation';
//		}
//	}
type ChatGPTExtractor struct {
	*ConversationExtractorBase
	articles        *goquery.Selection
	footnotes       []Footnote
	footnoteCounter int
}

// NewChatGPTExtractor creates a new ChatGPT extractor
// TypeScript original code:
//
//	constructor(document: Document, url: string) {
//		super(document, url);
//		this.articles = document.querySelectorAll('article[data-testid^="conversation-turn-"]');
//		this.footnotes = [];
//		this.footnoteCounter = 0;
//	}
func NewChatGPTExtractor(document *goquery.Document, urlStr string, schemaOrgData any) *ChatGPTExtractor {
	articles := document.Find(`article[data-testid^="conversation-turn-"]`)
	slog.Debug("ChatGPT extractor initialized", "articlesFound", articles.Length(), "url", urlStr)

	return &ChatGPTExtractor{
		ConversationExtractorBase: NewConversationExtractorBase(document, urlStr, schemaOrgData),
		articles:                  articles,
		footnotes:                 make([]Footnote, 0),
		footnoteCounter:           0,
	}
}

// CanExtract checks if the extractor can extract content
// TypeScript original code:
//
//	canExtract(): boolean {
//		return !!this.articles && this.articles.length > 0;
//	}
func (c *ChatGPTExtractor) CanExtract() bool {
	canExtract := c.articles.Length() > 0
	slog.Debug("ChatGPT extractor can extract check", "canExtract", canExtract, "articlesCount", c.articles.Length())
	return canExtract
}

// Name returns the name of the extractor
func (c *ChatGPTExtractor) Name() string {
	return "ChatGPTExtractor"
}

// Extract extracts the ChatGPT conversation
// TypeScript original code:
//
//	extract(): ExtractorResult {
//		const messages = this.extractMessages();
//		const metadata = this.getMetadata();
//		const footnotes = this.getFootnotes();
//		const rawContentHtml = this.createContentHtml(messages, footnotes);
//		// ... rest of extract method
//	}
func (c *ChatGPTExtractor) Extract() *ExtractorResult {
	slog.Debug("ChatGPT extractor starting extraction", "url", c.url)
	return c.ExtractWithDefuddle(c)
}

// ExtractMessages extracts conversation messages
// TypeScript original code (improved version):
//
//	protected extractMessages(): ConversationMessage[] {
//		const messages: ConversationMessage[] = [];
//		this.footnotes = [];
//		this.footnoteCounter = 0;
//
//		if (!this.articles) return messages;
//
//		this.articles.forEach((article) => {
//			// Get the localized author text from the sr-only heading and clean it
//			const authorElement = article.querySelector('h5.sr-only, h6.sr-only');
//			const authorText = authorElement?.textContent
//				?.trim()
//				?.replace(/:\s*$/, '') // Remove colon and any trailing whitespace
//				|| '';
//
//			let currentAuthorRole = '';
//
//			const authorRole = article.getAttribute('data-message-author-role');
//			if (authorRole) {
//				currentAuthorRole = authorRole;
//			}
//
//			let messageContent = article.innerHTML || '';
//			messageContent = messageContent.replace(/\u200B/g, '');
//
//			// Remove specific elements from the message content
//			const tempDiv = document.createElement('div');
//			tempDiv.innerHTML = messageContent;
//			tempDiv.querySelectorAll('h5.sr-only, h6.sr-only, span[data-state="closed"]').forEach(el => el.remove());
//			messageContent = tempDiv.innerHTML;
//
//			// Process inline references
//			messageContent = this.processFootnotes(messageContent);
//
//			// Clean up any stray empty paragraph tags
//			messageContent = messageContent.replace(/<p[^>]*>\s*<\/p>/g, '');
//
//			messages.push({
//				author: authorText,
//				content: messageContent.trim(),
//				metadata: {
//					role: currentAuthorRole || 'unknown'
//				}
//			});
//		});
//
//		return messages;
//	}
func (c *ChatGPTExtractor) ExtractMessages() []ConversationMessage {
	var messages []ConversationMessage
	c.footnotes = make([]Footnote, 0)
	c.footnoteCounter = 0

	if c.articles.Length() == 0 {
		slog.Debug("No articles found for ChatGPT extraction")
		return messages
	}

	c.articles.Each(func(i int, article *goquery.Selection) {
		// Get the localized author text from the sr-only heading and clean it
		authorElement := article.Find("h5.sr-only, h6.sr-only").First()
		authorText := strings.TrimSpace(authorElement.Text())

		// Remove colon and any trailing whitespace
		authorText = strings.TrimSuffix(strings.TrimSpace(authorText), ":")

		// Get author role from data attribute
		currentAuthorRole, _ := article.Attr("data-message-author-role")
		if currentAuthorRole == "" {
			currentAuthorRole = "unknown"
		}

		// Get message content
		messageContent, _ := article.Html()
		if messageContent == "" {
			slog.Debug("Empty message content found", "index", i)
			return
		}

		// Remove zero-width space characters
		messageContent = strings.ReplaceAll(messageContent, "\u200B", "")

		// Remove specific elements from the message content
		messageContent = c.cleanMessageContent(messageContent)

		// Process inline references using regex to find the containers
		messageContent = c.processFootnotes(messageContent)

		// Clean up any stray empty paragraph tags
		messageContent = chatgptEmptyParagraphRe.ReplaceAllString(messageContent, "")

		if strings.TrimSpace(messageContent) != "" {
			messages = append(messages, ConversationMessage{
				Author:  authorText,
				Content: strings.TrimSpace(messageContent),
				Metadata: map[string]any{
					"role": currentAuthorRole,
				},
			})
		}
	})

	slog.Debug("ChatGPT messages extracted", "messageCount", len(messages), "footnoteCount", len(c.footnotes))
	return messages
}

// cleanMessageContent removes specific elements from message content
// TypeScript original code:
//
//	// Remove specific elements from the message content
//	const tempDiv = document.createElement('div');
//	tempDiv.innerHTML = messageContent;
//	tempDiv.querySelectorAll('h5.sr-only, h6.sr-only, span[data-state="closed"]').forEach(el => el.remove());
//	messageContent = tempDiv.innerHTML;
func (c *ChatGPTExtractor) cleanMessageContent(messageContent string) string {
	// Create a temporary document to manipulate the HTML
	tempDoc, err := goquery.NewDocumentFromReader(strings.NewReader(messageContent))
	if err != nil {
		slog.Warn("Failed to parse message content as HTML", "error", err)
		return messageContent
	}

	// Remove specific elements
	tempDoc.Find(`h5.sr-only, h6.sr-only, span[data-state="closed"]`).Remove()

	// Get the cleaned HTML
	cleanedContent, err := tempDoc.Html()
	if err != nil {
		slog.Warn("Failed to get cleaned HTML content", "error", err)
		return messageContent
	}

	return cleanedContent
}

// GetFootnotes returns the conversation footnotes
// TypeScript original code:
//
//	protected getFootnotes(): Footnote[] {
//		return this.footnotes;
//	}
func (c *ChatGPTExtractor) GetFootnotes() []Footnote {
	return c.footnotes
}

// GetMetadata returns conversation metadata
// TypeScript original code:
//
//	protected getMetadata(): ConversationMetadata {
//		const title = this.getTitle();
//		const messages = this.extractMessages();
//
//		return {
//			title,
//			site: 'ChatGPT',
//			url: this.url,
//			messageCount: messages.length,
//			description: `ChatGPT conversation with ${messages.length} messages`
//		};
//	}
func (c *ChatGPTExtractor) GetMetadata() ConversationMetadata {
	title := c.getTitle()
	messages := c.ExtractMessages()

	return ConversationMetadata{
		Title:        title,
		Site:         "ChatGPT",
		URL:          c.url,
		MessageCount: len(messages),
		Description:  fmt.Sprintf("ChatGPT conversation with %d messages", len(messages)),
	}
}

// getTitle extracts the conversation title
// TypeScript original code:
//
//	private getTitle(): string {
//		// Try to get the page title first
//		const pageTitle = this.document.title?.trim();
//		if (pageTitle && pageTitle !== 'ChatGPT') {
//			return pageTitle;
//		}
//
//		// Fall back to first user message
//		const firstUserTurn = this.articles?.item(0)?.querySelector('.text-message');
//		if (firstUserTurn) {
//			const text = firstUserTurn.textContent || '';
//			// Truncate to first 50 characters if longer
//			return text.length > 50 ? text.slice(0, 50) + '...' : text;
//		}
//
//		return 'ChatGPT Conversation';
//	}
func (c *ChatGPTExtractor) getTitle() string {
	// Try to get the page title first
	pageTitle := strings.TrimSpace(c.document.Find("title").Text())
	if pageTitle != "" && pageTitle != "ChatGPT" {
		return pageTitle
	}

	// Fall back to first user message
	firstUserTurn := c.articles.First().Find(".text-message").First()
	if firstUserTurn.Length() > 0 {
		text := firstUserTurn.Text()
		// Truncate to first 50 characters if longer
		if len(text) > 50 {
			return text[:50] + "..."
		}
		return text
	}

	return "ChatGPT Conversation"
}

// processFootnotes processes footnotes in the content
// TypeScript original code:
//
//	private processFootnotes(content: string): string {
//	  // Find all citation links and replace them with footnotes
//	  const citationPattern = /(&ZeroWidthSpace;)?(<span[^>]*?>\s*<a(?=[^>]*?href="([^"]+)")(?=[^>]*?target="_blank")(?=[^>]*?rel="noopener")[^>]*?>[\s\S]*?<\/a>\s*<\/span>)/g;
//	  let processedContent = content;
//	  let match;
//
//	  while ((match = citationPattern.exec(content)) !== null) {
//	    const fullMatch = match[0];
//	    const url = match[3];
//
//	    // Add to footnotes
//	    this.footnoteCounter++;
//	    this.footnotes.push({
//	      number: this.footnoteCounter,
//	      url: url,
//	      text: `Source ${this.footnoteCounter}`
//	    });
//
//	    // Replace with footnote reference
//	    processedContent = processedContent.replace(fullMatch, `<sup><a href="#footnote-${this.footnoteCounter}">[${this.footnoteCounter}]</a></sup>`);
//	  }
//
//	  return processedContent;
//	}
func (c *ChatGPTExtractor) processFootnotes(content string) string {
	// Simplified pattern without Perl lookaheads
	// Matches: <span...><a href="..." target="_blank" rel="noopener">...</a></span>
	matches := chatgptCitationRe.FindAllStringSubmatch(content, -1)
	processedContent := content

	for _, match := range matches {
		if len(match) >= 4 {
			fullMatch := match[0]
			url := match[3]

			// Add to footnotes
			c.footnoteCounter++
			c.footnotes = append(c.footnotes, Footnote{
				URL:  url,
				Text: fmt.Sprintf("Source %d", c.footnoteCounter),
			})

			// Replace with footnote reference
			replacement := fmt.Sprintf(`<sup><a href="#footnote-%d">[%d]</a></sup>`, c.footnoteCounter, c.footnoteCounter)
			processedContent = strings.Replace(processedContent, fullMatch, replacement, 1)
		}
	}

	return processedContent
}
