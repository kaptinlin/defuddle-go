package extractors

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Pre-compiled regex patterns for Claude extraction.
var claudeTitleSuffixRe = regexp.MustCompile(` - Claude$`)

// ClaudeExtractor handles Claude conversation content extraction
// TypeScript original code:
// import { ConversationExtractor } from './_conversation';
// import { ConversationMessage, ConversationMetadata } from '../types/extractors';
//
//	export class ClaudeExtractor extends ConversationExtractor {
//		private articles: NodeListOf<Element> | null;
//
//		constructor(document: Document, url: string) {
//			super(document, url);
//			// Find all message blocks - both user and assistant messages
//			this.articles = document.querySelectorAll('div[data-testid="user-message"], div[data-testid="assistant-message"], div.font-claude-message');
//		}
//
//		canExtract(): boolean {
//			return !!this.articles && this.articles.length > 0;
//		}
//
//		protected extractMessages(): ConversationMessage[] {
//			const messages: ConversationMessage[] = [];
//
//			if (!this.articles) return messages;
//
//			this.articles.forEach((article) => {
//				let role: string;
//				let content: string;
//
//				if (article.hasAttribute('data-testid')) {
//					// Handle user messages
//					if (article.getAttribute('data-testid') === 'user-message') {
//						role = 'you';
//						content = article.innerHTML;
//					}
//					// Skip non-message elements
//					else {
//						return;
//					}
//				} else if (article.classList.contains('font-claude-message')) {
//					// Handle Claude messages
//					role = 'assistant';
//					content = article.innerHTML;
//				} else {
//					// Skip unknown elements
//					return;
//				}
//
//				if (content) {
//					messages.push({
//						author: role === 'you' ? 'You' : 'Claude',
//						content: content.trim(),
//						metadata: {
//							role: role
//						}
//					});
//				}
//			});
//
//			return messages;
//		}
//
//		protected getMetadata(): ConversationMetadata {
//			const title = this.getTitle();
//			const messages = this.extractMessages();
//
//			return {
//				title,
//				site: 'Claude',
//				url: this.url,
//				messageCount: messages.length,
//				description: `Claude conversation with ${messages.length} messages`
//			};
//		}
//
//		private getTitle(): string {
//			// Try to get the page title first
//			const pageTitle = this.document.title?.trim();
//			if (pageTitle && pageTitle !== 'Claude') {
//				// Remove ' - Claude' suffix if present
//				return pageTitle.replace(/ - Claude$/, '');
//			}
//
//			// Try to get title from header
//			const headerTitle = this.document.querySelector('header .font-tiempos')?.textContent?.trim();
//			if (headerTitle) {
//				return headerTitle;
//			}
//
//			// Fall back to first user message
//			const firstUserMessage = this.articles?.item(0)?.querySelector('[data-testid="user-message"]');
//			if (firstUserMessage) {
//				const text = firstUserMessage.textContent || '';
//				// Truncate to first 50 characters if longer
//				return text.length > 50 ? text.slice(0, 50) + '...' : text;
//			}
//
//			return 'Claude Conversation';
//		}
//	}
type ClaudeExtractor struct {
	*ConversationExtractorBase
	articles *goquery.Selection
}

// NewClaudeExtractor creates a new Claude extractor
// TypeScript original code:
//
//	constructor(document: Document, url: string) {
//		super(document, url);
//		// Find all message blocks - both user and assistant messages
//		this.articles = document.querySelectorAll('div[data-testid="user-message"], div[data-testid="assistant-message"], div.font-claude-message');
//	}
func NewClaudeExtractor(document *goquery.Document, urlStr string, schemaOrgData any) *ClaudeExtractor {
	// Primary selectors from TypeScript reference
	articles := document.Find(`div[data-testid="user-message"], div[data-testid="assistant-message"], div.font-claude-message`)

	// Fallback selectors if primary ones don't work
	if articles.Length() == 0 {
		slog.Debug("Claude extractor: trying fallback selectors")

		fallbackSelectors := []string{
			`div[data-testid*="message"]`,
			`.message`,
			`div[class*="message"]`,
			`div[class*="chat"]`,
			`div[role="article"]`,
			`article`,
		}

		for _, selector := range fallbackSelectors {
			articles = document.Find(selector)
			if articles.Length() > 0 {
				slog.Debug("Claude extractor: found articles with fallback", "selector", selector, "count", articles.Length())
				break
			}
		}
	}

	slog.Debug("Claude extractor initialized", "articlesFound", articles.Length(), "url", urlStr)

	return &ClaudeExtractor{
		ConversationExtractorBase: NewConversationExtractorBase(document, urlStr, schemaOrgData),
		articles:                  articles,
	}
}

// CanExtract checks if the extractor can extract content
// TypeScript original code:
//
//	canExtract(): boolean {
//		return !!this.articles && this.articles.length > 0;
//	}
func (c *ClaudeExtractor) CanExtract() bool {
	canExtract := c.articles.Length() > 0
	slog.Debug("Claude extractor can extract check", "canExtract", canExtract, "articlesCount", c.articles.Length())
	return canExtract
}

// Name returns the name of the extractor
func (c *ClaudeExtractor) Name() string {
	return "ClaudeExtractor"
}

// Extract extracts the Claude conversation
// TypeScript original code:
//
//	extract(): ExtractorResult {
//		return this.extractWithDefuddle(this);
//	}
func (c *ClaudeExtractor) Extract() *ExtractorResult {
	slog.Debug("Claude extractor starting extraction", "url", c.url)
	return c.ExtractWithDefuddle(c)
}

// ExtractMessages extracts conversation messages
// TypeScript original code:
//
//	protected extractMessages(): ConversationMessage[] {
//		const messages: ConversationMessage[] = [];
//
//		if (!this.articles) return messages;
//
//		this.articles.forEach((article) => {
//			let role: string;
//			let content: string;
//
//			if (article.hasAttribute('data-testid')) {
//				// Handle user messages
//				if (article.getAttribute('data-testid') === 'user-message') {
//					role = 'you';
//					content = article.innerHTML;
//				}
//				// Skip non-message elements
//				else {
//					return;
//				}
//			} else if (article.classList.contains('font-claude-message')) {
//				// Handle Claude messages
//				role = 'assistant';
//				content = article.innerHTML;
//			} else {
//				// Skip unknown elements
//				return;
//			}
//
//			if (content) {
//				messages.push({
//					author: role === 'you' ? 'You' : 'Claude',
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
func (c *ClaudeExtractor) ExtractMessages() []ConversationMessage {
	var messages []ConversationMessage

	if c.articles.Length() == 0 {
		slog.Debug("No articles found for Claude extraction")
		return messages
	}

	c.articles.Each(func(i int, article *goquery.Selection) {
		var role string
		var content string

		// Check if element has data-testid attribute
		testid, exists := article.Attr("data-testid")
		if !exists {
			return
		}

		switch testid {
		case "user-message":
			role = "you"
			content, _ = article.Html()
		case "assistant-message":
			role = "assistant"
			content, _ = article.Html()
		default:
			slog.Debug("Skipping non-message element", "testid", testid, "index", i)
			return
		}

		if strings.TrimSpace(content) != "" {
			var author string
			if role == "you" {
				author = "You"
			} else {
				author = "Claude"
			}

			messages = append(messages, ConversationMessage{
				Author:  author,
				Content: strings.TrimSpace(content),
				Metadata: map[string]any{
					"role": role,
				},
			})
		} else {
			slog.Debug("Empty content found", "role", role, "index", i)
		}
	})

	slog.Debug("Claude messages extracted", "messageCount", len(messages))
	return messages
}

// GetFootnotes returns the conversation footnotes
// TypeScript original code:
//
//	protected getFootnotes(): Footnote[] {
//		return [];
//	}
func (c *ClaudeExtractor) GetFootnotes() []Footnote {
	// Claude extractor doesn't process footnotes in the TypeScript version
	return []Footnote{}
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
//			site: 'Claude',
//			url: this.url,
//			messageCount: messages.length,
//			description: `Claude conversation with ${messages.length} messages`
//		};
//	}
func (c *ClaudeExtractor) GetMetadata() ConversationMetadata {
	title := c.getTitle()
	messages := c.ExtractMessages()

	return ConversationMetadata{
		Title:        title,
		Site:         "Claude",
		URL:          c.url,
		MessageCount: len(messages),
		Description:  fmt.Sprintf("Claude conversation with %d messages", len(messages)),
	}
}

// getTitle extracts the conversation title
// TypeScript original code:
//
//	private getTitle(): string {
//		// Try to get the page title first
//		const pageTitle = this.document.title?.trim();
//		if (pageTitle && pageTitle !== 'Claude') {
//			// Remove ' - Claude' suffix if present
//			return pageTitle.replace(/ - Claude$/, '');
//		}
//
//		// Try to get title from header
//		const headerTitle = this.document.querySelector('header .font-tiempos')?.textContent?.trim();
//		if (headerTitle) {
//			return headerTitle;
//		}
//
//		// Fall back to first user message
//		const firstUserMessage = this.articles?.item(0)?.querySelector('[data-testid="user-message"]');
//		if (firstUserMessage) {
//			const text = firstUserMessage.textContent || '';
//			// Truncate to first 50 characters if longer
//			return text.length > 50 ? text.slice(0, 50) + '...' : text;
//		}
//
//		return 'Claude Conversation';
//	}
func (c *ClaudeExtractor) getTitle() string {
	// Try to get the page title first
	pageTitle := strings.TrimSpace(c.document.Find("title").Text())
	if pageTitle != "" && pageTitle != "Claude" {
		// Remove ' - Claude' suffix if present
		return claudeTitleSuffixRe.ReplaceAllString(pageTitle, "")
	}

	// Try to get title from header
	headerTitle := strings.TrimSpace(c.document.Find("header .font-tiempos").Text())
	if headerTitle != "" {
		return headerTitle
	}

	// Fall back to first user message
	firstUserMessage := c.articles.First().Find(`[data-testid="user-message"]`)
	if firstUserMessage.Length() > 0 {
		text := firstUserMessage.Text()
		// Truncate to first 50 characters if longer
		if len(text) > 50 {
			return text[:50] + "..."
		}
		return text
	}

	// Try to fall back to any first message
	if c.articles.Length() > 0 {
		firstMessage := c.articles.First()
		text := strings.TrimSpace(firstMessage.Text())
		if text != "" {
			// Truncate to first 50 characters if longer
			if len(text) > 50 {
				return text[:50] + "..."
			}
			return text
		}
	}

	return "Claude Conversation"
}
