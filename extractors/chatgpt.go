package extractors

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
func NewChatGPTExtractor(document *goquery.Document, urlStr string, schemaOrgData interface{}) *ChatGPTExtractor {
	return &ChatGPTExtractor{
		ConversationExtractorBase: NewConversationExtractorBase(document, urlStr, schemaOrgData),
		articles:                  document.Find(`article[data-testid^="conversation-turn-"]`),
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
	return c.articles.Length() > 0
}

// GetName returns the name of the extractor
func (c *ChatGPTExtractor) GetName() string {
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
	return c.ExtractWithDefuddle(c)
}

// ExtractMessages extracts conversation messages
// TypeScript original code:
//
//	protected extractMessages(): ConversationMessage[] {
//		const messages: ConversationMessage[] = [];
//		this.footnotes = [];
//		this.footnoteCounter = 0;
//
//		if (!this.articles) return messages;
//
//		this.articles.forEach((article) => {
//			const roleAttr = article.getAttribute('data-message-author-role');
//			if (!roleAttr) return;
//
//			let content: string;
//			let role: string;
//
//			if (roleAttr === 'user') {
//				const textMessage = article.querySelector('.text-message');
//				content = textMessage ? textMessage.innerHTML : '';
//				role = 'you';
//			} else if (roleAttr === 'assistant') {
//				const messageContent = article.querySelector('.message-content');
//				content = messageContent ? messageContent.innerHTML : '';
//				role = 'assistant';
//
//				// Process footnotes for assistant messages
//				content = this.processFootnotes(content);
//			} else {
//				return;
//			}
//
//			if (content) {
//				messages.push({
//					author: role === 'you' ? 'You' : 'ChatGPT',
//					content: content.trim(),
//					metadata: {
//						role: role
//					}
//				});
//			}
//		});
//
//		return messages;
//	}
func (c *ChatGPTExtractor) ExtractMessages() []ConversationMessage {
	var messages []ConversationMessage
	c.footnotes = make([]Footnote, 0)
	c.footnoteCounter = 0

	c.articles.Each(func(i int, article *goquery.Selection) {
		roleAttr, exists := article.Attr("data-message-author-role")
		if !exists {
			return
		}

		var content string
		var role string

		if roleAttr == "user" {
			textMessage := article.Find(".text-message").First()
			if textMessage.Length() > 0 {
				content, _ = textMessage.Html()
			}
			role = "you"
		} else if roleAttr == "assistant" {
			messageContent := article.Find(".message-content").First()
			if messageContent.Length() > 0 {
				content, _ = messageContent.Html()
			}
			role = "assistant"

			// Process footnotes for assistant messages
			content = c.processFootnotes(content)
		} else {
			return
		}

		if content != "" {
			var author string
			if role == "you" {
				author = "You"
			} else {
				author = "ChatGPT"
			}

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
//		const pageTitle = this.document.title?.trim();
//		if (pageTitle && pageTitle !== 'ChatGPT') {
//			return pageTitle;
//		}
//
//		const firstUserMessage = this.articles?.item(0)?.querySelector('.text-message');
//		if (firstUserMessage) {
//			const text = firstUserMessage.textContent || '';
//			return text.length > 50 ? text.slice(0, 50) + '...' : text;
//		}
//
//		return 'ChatGPT Conversation';
//	}
func (c *ChatGPTExtractor) getTitle() string {
	pageTitle := strings.TrimSpace(c.document.Find("title").Text())
	if pageTitle != "" && pageTitle != "ChatGPT" {
		return pageTitle
	}

	firstUserMessage := c.articles.First().Find(".text-message")
	if firstUserMessage.Length() > 0 {
		text := firstUserMessage.Text()
		if len(text) > 50 {
			return text[:50] + "..."
		}
		return text
	}

	return "ChatGPT Conversation"
}

// processFootnotes processes citation links and converts them to footnotes
// TypeScript original code:
//
//	private processFootnotes(content: string): string {
//		const citationPattern = /(&ZeroWidthSpace;)?<span[^>]*><a\s+href="([^"]*)"[^>]*>([^<]*)</a></span>/g;
//
//		return content.replace(citationPattern, (match, zws, url, linkText) => {
//			if (!url || url.startsWith('#')) {
//				return match;
//			}
//
//			let footnote = this.footnotes.find(fn => fn.url === url);
//			let footnoteIndex: number;
//
//			if (!footnote) {
//				this.footnoteCounter++;
//				footnoteIndex = this.footnoteCounter;
//
//				let decodedText = this.decodeFragmentText(linkText);
//				if (!decodedText || decodedText.trim() === '') {
//					try {
//						const domain = new URL(url).hostname.replace(/^www\./, '');
//						decodedText = domain;
//					} catch (e) {
//						decodedText = 'Link';
//					}
//				}
//
//				this.footnotes.push({
//					url,
//					text: decodedText
//				});
//			} else {
//				footnoteIndex = this.footnotes.findIndex(fn => fn.url === url) + 1;
//			}
//
//			return `<sup id="fnref:${footnoteIndex}" class="footnote-ref"><a href="#fn:${footnoteIndex}" class="footnote-link">${footnoteIndex}</a></sup>`;
//		});
//	}
func (c *ChatGPTExtractor) processFootnotes(content string) string {
	citationPattern := regexp.MustCompile(`(&ZeroWidthSpace;)?<span[^>]*><a\s+href="([^"]*)"[^>]*>([^<]*)</a></span>`)

	return citationPattern.ReplaceAllStringFunc(content, func(match string) string {
		matches := citationPattern.FindStringSubmatch(match)
		if len(matches) < 4 {
			return match
		}

		urlStr := matches[2]
		linkText := matches[3]

		if urlStr == "" || strings.HasPrefix(urlStr, "#") {
			return match
		}

		var footnoteIndex int
		found := false

		for idx, footnote := range c.footnotes {
			if footnote.URL == urlStr {
				footnoteIndex = idx + 1
				found = true
				break
			}
		}

		if !found {
			c.footnoteCounter++
			footnoteIndex = c.footnoteCounter

			decodedText := c.decodeFragmentText(linkText)
			if decodedText == "" || strings.TrimSpace(decodedText) == "" {
				if parsedURL, err := url.Parse(urlStr); err == nil {
					domain := strings.TrimPrefix(parsedURL.Hostname(), "www.")
					decodedText = domain
				} else {
					decodedText = "Link"
				}
			}

			c.footnotes = append(c.footnotes, Footnote{
				URL:  urlStr,
				Text: decodedText,
			})
		}

		return fmt.Sprintf(`<sup id="fnref:%d" class="footnote-ref"><a href="#fn:%d" class="footnote-link">%d</a></sup>`, footnoteIndex, footnoteIndex, footnoteIndex)
	})
}

// decodeFragmentText decodes URL fragment text
// TypeScript original code:
//
//	private decodeFragmentText(text: string): string {
//		try {
//			return decodeURIComponent(text);
//		} catch (e) {
//			return text;
//		}
//	}
func (c *ChatGPTExtractor) decodeFragmentText(text string) string {
	decoded, err := url.QueryUnescape(text)
	if err != nil {
		return text
	}
	return decoded
}
