package extractors

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GeminiExtractor handles Gemini conversation content extraction
type GeminiExtractor struct {
	*ConversationExtractorBase
	conversationContainers *goquery.Selection
	footnotes              []Footnote
	messageCount           *int
}

// NewGeminiExtractor creates a new Gemini extractor
// TypeScript original code:
//
//	constructor(document: Document, url: string) {
//		super(document, url);
//		this.conversationContainers = document.querySelectorAll('div.conversation-container');
//		this.footnotes = [];
//	}
func NewGeminiExtractor(document *goquery.Document, urlStr string, schemaOrgData interface{}) *GeminiExtractor {
	return &GeminiExtractor{
		ConversationExtractorBase: NewConversationExtractorBase(document, urlStr, schemaOrgData),
		conversationContainers:    document.Find("div.conversation-container"),
		footnotes:                 make([]Footnote, 0),
		messageCount:              nil,
	}
}

// CanExtract checks if the extractor can extract content
// TypeScript original code:
//
//	canExtract(): boolean {
//		return !!this.conversationContainers && this.conversationContainers.length > 0;
//	}
func (g *GeminiExtractor) CanExtract() bool {
	return g.conversationContainers.Length() > 0
}

// GetName returns the name of the extractor
func (g *GeminiExtractor) GetName() string {
	return "GeminiExtractor"
}

// Extract extracts the Gemini conversation
func (g *GeminiExtractor) Extract() *ExtractorResult {
	return g.ExtractWithDefuddle(g)
}

// ExtractMessages extracts conversation messages
// TypeScript original code:
//
//	protected extractMessages(): ConversationMessage[] {
//		this.messageCount = 0;
//		const messages: ConversationMessage[] = [];
//
//		if (!this.conversationContainers) return messages;
//
//		this.extractSources();
//
//		this.conversationContainers.forEach((container) => {
//			const userQuery = container.querySelector('user-query');
//			if (userQuery) {
//				const queryText = userQuery.querySelector('.query-text');
//				if (queryText) {
//					const content = queryText.innerHTML || '';
//					messages.push({
//						author: 'You',
//						content: content.trim(),
//						metadata: { role: 'user' }
//					});
//				}
//			}
//
//			const modelResponse = container.querySelector('model-response');
//			if (modelResponse) {
//				const regularContent = modelResponse.querySelector('.model-response-text .markdown');
//				const extendedContent = modelResponse.querySelector('#extended-response-markdown-content');
//				const contentElement = extendedContent || regularContent;
//
//				if (contentElement) {
//					let content = contentElement.innerHTML || '';
//
//					const tempDiv = document.createElement('div');
//					tempDiv.innerHTML = content;
//
//					tempDiv.querySelectorAll('.table-content').forEach(el => {
//						// `table-content` is a PARTIAL selector in defuddle (table of contents, will be removed), but a real table in Gemini (should be kept).
//						el.classList.remove('table-content');
//					});
//
//					content = tempDiv.innerHTML;
//
//					messages.push({
//						author: 'Gemini',
//						content: content.trim(),
//						metadata: { role: 'assistant' }
//					});
//				}
//			}
//		});
//		this.messageCount = messages.length;
//		return messages;
//	}
func (g *GeminiExtractor) ExtractMessages() []ConversationMessage {
	count := 0
	g.messageCount = &count
	var messages []ConversationMessage

	g.extractSources()

	g.conversationContainers.Each(func(i int, container *goquery.Selection) {
		// Handle user queries
		userQuery := container.Find("user-query").First()
		if userQuery.Length() > 0 {
			queryText := userQuery.Find(".query-text").First()
			if queryText.Length() > 0 {
				content, _ := queryText.Html()
				if content != "" {
					messages = append(messages, ConversationMessage{
						Author:  "You",
						Content: strings.TrimSpace(content),
						Metadata: map[string]interface{}{
							"role": "user",
						},
					})
				}
			}
		}

		// Handle model responses
		modelResponse := container.Find("model-response").First()
		if modelResponse.Length() > 0 {
			regularContent := modelResponse.Find(".model-response-text .markdown").First()
			extendedContent := modelResponse.Find("#extended-response-markdown-content").First()

			var contentElement *goquery.Selection
			if extendedContent.Length() > 0 {
				contentElement = extendedContent
			} else {
				contentElement = regularContent
			}

			if contentElement.Length() > 0 {
				content, _ := contentElement.Html()

				// Parse content to modify table-content classes
				// table-content is a PARTIAL selector in defuddle (table of contents, will be removed),
				// but a real table in Gemini (should be kept).
				if tempDoc, err := goquery.NewDocumentFromReader(strings.NewReader(content)); err == nil {
					tempDoc.Find(".table-content").Each(func(j int, el *goquery.Selection) {
						el.RemoveClass("table-content")
					})
					content, _ = tempDoc.Html()
				}

				if content != "" {
					messages = append(messages, ConversationMessage{
						Author:  "Gemini",
						Content: strings.TrimSpace(content),
						Metadata: map[string]interface{}{
							"role": "assistant",
						},
					})
				}
			}
		}
	})

	count = len(messages)
	g.messageCount = &count
	return messages
}

// extractSources extracts footnotes from browse items
// TypeScript original code:
//
//	private extractSources(): void {
//		const browseItems = this.document.querySelectorAll('browse-item');
//
//		if (browseItems && browseItems.length > 0) {
//			browseItems.forEach(item => {
//				const link = item.querySelector('a');
//				if (link instanceof HTMLAnchorElement) {
//					const url = link.href;
//					const domain = link.querySelector('.domain')?.textContent?.trim() || '';
//					const title = link.querySelector('.title')?.textContent?.trim() || '';
//
//					if (url && (domain || title)) {
//						this.footnotes.push({
//							url,
//							text: title ? `${domain}: ${title}` : domain
//						});
//					}
//				}
//			});
//		}
//	}
func (g *GeminiExtractor) extractSources() {
	browseItems := g.document.Find("browse-item")

	browseItems.Each(func(i int, item *goquery.Selection) {
		link := item.Find("a").First()
		if link.Length() > 0 {
			url, exists := link.Attr("href")
			if exists && url != "" {
				domain := strings.TrimSpace(link.Find(".domain").Text())
				title := strings.TrimSpace(link.Find(".title").Text())

				if domain != "" || title != "" {
					var text string
					if title != "" {
						text = fmt.Sprintf("%s: %s", domain, title)
					} else {
						text = domain
					}

					g.footnotes = append(g.footnotes, Footnote{
						URL:  url,
						Text: text,
					})
				}
			}
		}
	})
}

// GetFootnotes returns the conversation footnotes
// TypeScript original code:
//
//	protected getFootnotes(): Footnote[] {
//		return this.footnotes;
//	}
func (g *GeminiExtractor) GetFootnotes() []Footnote {
	return g.footnotes
}

// GetMetadata returns conversation metadata
// TypeScript original code:
//
//	protected getMetadata(): ConversationMetadata {
//		const title = this.getTitle();
//		const messageCount = this.messageCount ?? this.extractMessages().length;
//		return {
//			title,
//			site: 'Gemini',
//			url: this.url,
//			messageCount,
//			description: `Gemini conversation with ${messageCount} messages`
//		};
//	}
func (g *GeminiExtractor) GetMetadata() ConversationMetadata {
	title := g.getTitle()
	messageCount := 0
	if g.messageCount != nil {
		messageCount = *g.messageCount
	} else {
		messageCount = len(g.ExtractMessages())
	}

	return ConversationMetadata{
		Title:        title,
		Site:         "Gemini",
		URL:          g.url,
		MessageCount: messageCount,
		Description:  fmt.Sprintf("Gemini conversation with %d messages", messageCount),
	}
}

// getTitle extracts the conversation title
// TypeScript original code:
//
//	private getTitle(): string {
//		const pageTitle = this.document.title?.trim();
//		if (pageTitle && pageTitle !== 'Gemini' && !pageTitle.includes('Gemini')) {
//			return pageTitle;
//		}
//
//		const researchTitle = this.document.querySelector('.title-text')?.textContent?.trim();
//		if (researchTitle) {
//			return researchTitle;
//		}
//
//		const firstUserQuery = this.conversationContainers?.item(0)?.querySelector('.query-text');
//		if (firstUserQuery) {
//			const text = firstUserQuery.textContent || '';
//			return text.length > 50 ? text.slice(0, 50) + '...' : text;
//		}
//
//		return 'Gemini Conversation';
//	}
func (g *GeminiExtractor) getTitle() string {
	pageTitle := strings.TrimSpace(g.document.Find("title").Text())
	if pageTitle != "" && pageTitle != "Gemini" && !strings.Contains(pageTitle, "Gemini") {
		return pageTitle
	}

	researchTitle := strings.TrimSpace(g.document.Find(".title-text").Text())
	if researchTitle != "" {
		return researchTitle
	}

	firstUserQuery := g.conversationContainers.First().Find(".query-text")
	if firstUserQuery.Length() > 0 {
		text := firstUserQuery.Text()
		if len(text) > 50 {
			return text[:50] + "..."
		}
		return text
	}

	return "Gemini Conversation"
}
