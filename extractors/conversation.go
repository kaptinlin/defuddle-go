package extractors

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Precompiled regex for performance
var paragraphRegex = regexp.MustCompile(`<p[^>]*>[\s\S]*?</p>`)

// ConversationMessage represents a single message in a conversation
// Corresponding to TypeScript interface ConversationMessage
type ConversationMessage struct {
	Author    string         `json:"author"`
	Content   string         `json:"content"`
	Timestamp string         `json:"timestamp,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// ConversationMetadata represents metadata about the conversation
// Corresponding to TypeScript interface ConversationMetadata
type ConversationMetadata struct {
	Title        string `json:"title"`
	Site         string `json:"site"`
	URL          string `json:"url"`
	MessageCount int    `json:"messageCount"`
	Description  string `json:"description"`
}

// Footnote represents a footnote in the conversation
// Corresponding to TypeScript interface Footnote
type Footnote struct {
	URL  string `json:"url"`
	Text string `json:"text"`
}

// ConversationExtractor defines the interface for conversation extractors
// TypeScript original code:
//
//	export abstract class ConversationExtractor extends BaseExtractor {
//		protected abstract extractMessages(): ConversationMessage[];
//		protected abstract getMetadata(): ConversationMetadata;
//		protected getFootnotes(): Footnote[] {
//			return [];
//		}
//	}
type ConversationExtractor interface {
	BaseExtractor
	ExtractMessages() []ConversationMessage
	GetMetadata() ConversationMetadata
	GetFootnotes() []Footnote
}

// ConversationExtractorBase provides common functionality for conversation extractors
// Implementation corresponding to TypeScript ConversationExtractor abstract class
type ConversationExtractorBase struct {
	*ExtractorBase
}

// NewConversationExtractorBase creates a new conversation extractor base
// TypeScript original code:
//
//	constructor(document: Document, url: string, schemaOrgData?: any) {
//		super(document, url, schemaOrgData);
//	}
func NewConversationExtractorBase(document *goquery.Document, url string, schemaOrgData any) *ConversationExtractorBase {
	return &ConversationExtractorBase{
		ExtractorBase: NewExtractorBase(document, url, schemaOrgData),
	}
}

// CreateContentHTML creates formatted HTML content from messages and footnotes
// TypeScript original code:
//
//	protected createContentHtml(messages: ConversationMessage[], footnotes: Footnote[]): string {
//		const messagesHtml = messages.map((message, index) => {
//			const timestampHtml = message.timestamp ?
//				`<div class="message-timestamp">${message.timestamp}</div>` : '';
//
//			// Check if content already has paragraph tags
//			const hasParagraphs = /<p[^>]*>[\s\S]*?<\/p>/i.test(message.content);
//			const contentHtml = hasParagraphs ? message.content : `<p>${message.content}</p>`;
//
//			// Add metadata to data attributes
//			const dataAttributes = message.metadata ?
//				Object.entries(message.metadata)
//					.map(([key, value]) => `data-${key}="${value}"`)
//					.join(' ') : '';
//
//			return `
//			<div class="message message-${message.author.toLowerCase()}" ${dataAttributes}>
//				<div class="message-header">
//					<p class="message-author"><strong>${message.author}</strong></p>
//					${timestampHtml}
//				</div>
//				<div class="message-content">
//					${contentHtml}
//				</div>
//			</div>${index < messages.length - 1 ? '\n<hr>' : ''}`;
//		}).join('\n').trim();
//
//		// Add footnotes section if we have any
//		const footnotesHtml = footnotes.length > 0 ? `
//			<div id="footnotes">
//				<ol>
//					${footnotes.map((footnote, index) => `
//						<li class="footnote" id="fn:${index + 1}">
//							<p>
//								<a href="${footnote.url}" target="_blank">${footnote.text}</a>&nbsp;<a href="#fnref:${index + 1}" class="footnote-backref">↩</a>
//							</p>
//						</li>
//					`).join('')}
//				</ol>
//			</div>` : '';
//
//		return `${messagesHtml}\n${footnotesHtml}`.trim();
//	}
func (c *ConversationExtractorBase) CreateContentHTML(messages []ConversationMessage, footnotes []Footnote) string {
	var messagesHTML strings.Builder

	for i, message := range messages {
		timestampHTML := ""
		if message.Timestamp != "" {
			timestampHTML = fmt.Sprintf(`<div class="message-timestamp">%s</div>`, message.Timestamp)
		}

		// Check if content already has paragraph tags
		hasParagraphs := paragraphRegex.MatchString(message.Content)
		contentHTML := message.Content
		if !hasParagraphs {
			contentHTML = fmt.Sprintf("<p>%s</p>", message.Content)
		}

		// Add metadata to data attributes
		var dataAttributes strings.Builder
		if message.Metadata != nil {
			for key, value := range message.Metadata {
				dataAttributes.WriteString(fmt.Sprintf(` data-%s="%v"`, key, value))
			}
		}

		authorLower := strings.ToLower(message.Author)
		messageHTML := fmt.Sprintf(`
			<div class="message message-%s"%s>
				<div class="message-header">
					<p class="message-author"><strong>%s</strong></p>
					%s
				</div>
				<div class="message-content">
					%s
				</div>
			</div>`, authorLower, dataAttributes.String(), message.Author, timestampHTML, contentHTML)

		messagesHTML.WriteString(messageHTML)

		if i < len(messages)-1 {
			messagesHTML.WriteString("\n<hr>")
		}
	}

	// Add footnotes section if we have any
	footnotesHTML := ""
	if len(footnotes) > 0 {
		var footnotesBuilder strings.Builder
		footnotesBuilder.WriteString(`
			<div id="footnotes">
				<ol>`)

		for i, footnote := range footnotes {
			footnoteNum := i + 1
			footnoteHTML := fmt.Sprintf(`
						<li class="footnote" id="fn:%d">
							<p>
								<a href="%s" target="_blank">%s</a>&nbsp;<a href="#fnref:%d" class="footnote-backref">↩</a>
							</p>
						</li>`, footnoteNum, footnote.URL, footnote.Text, footnoteNum)
			footnotesBuilder.WriteString(footnoteHTML)
		}

		footnotesBuilder.WriteString(`
				</ol>
			</div>`)
		footnotesHTML = footnotesBuilder.String()
	}

	result := messagesHTML.String()
	if footnotesHTML != "" {
		result += "\n" + footnotesHTML
	}

	return strings.TrimSpace(result)
}

// ExtractWithDefuddle extracts conversation content similar to TypeScript implementation
// TypeScript original code:
//
//	extract(): ExtractorResult {
//		const messages = this.extractMessages();
//		const metadata = this.getMetadata();
//		const footnotes = this.getFootnotes();
//		const rawContentHtml = this.createContentHtml(messages, footnotes);
//
//		// Create a temporary document to run Defuddle on our content
//		const tempDoc = document.implementation.createHTMLDocument();
//		const container = tempDoc.createElement('article');
//		container.innerHTML = rawContentHtml;
//		tempDoc.body.appendChild(container);
//
//		// Run Defuddle on our formatted content
//		const defuddled = new Defuddle(tempDoc).parse();
//		const contentHtml = defuddled.content;
//
//		return {
//			content: contentHtml,
//			contentHtml: contentHtml,
//			extractedContent: {
//				messageCount: messages.length.toString(),
//			},
//			variables: {
//				title: metadata.title || 'Conversation',
//				site: metadata.site,
//				description: metadata.description || `${metadata.site} conversation with ${messages.length} messages`,
//				wordCount: defuddled.wordCount?.toString() || '',
//			}
//		};
//	}
func (c *ConversationExtractorBase) ExtractWithDefuddle(extractor ConversationExtractor) *ExtractorResult {
	messages := extractor.ExtractMessages()
	metadata := extractor.GetMetadata()
	footnotes := extractor.GetFootnotes()
	rawContentHTML := c.CreateContentHTML(messages, footnotes)

	// In Go implementation, we'll use the raw content directly
	// since we don't have a secondary Defuddle pass like in TypeScript
	contentHTML := rawContentHTML

	description := metadata.Description
	if description == "" {
		description = fmt.Sprintf("%s conversation with %d messages", metadata.Site, len(messages))
	}

	return &ExtractorResult{
		Content:     contentHTML,
		ContentHTML: contentHTML,
		ExtractedContent: map[string]any{
			"messageCount": strconv.Itoa(len(messages)),
		},
		Variables: map[string]string{
			"title":       metadata.Title,
			"site":        metadata.Site,
			"description": description,
		},
	}
}
