package extractors

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// HackerNewsExtractor handles Hacker News content extraction
// Corresponding to TypeScript class HackerNewsExtractor extends BaseExtractor
type HackerNewsExtractor struct {
	*ExtractorBase
	mainPost      *goquery.Selection
	isCommentPage bool
	mainComment   *goquery.Selection
}

// NewHackerNewsExtractor creates a new HackerNews extractor
// TypeScript original code:
//
//	constructor(document: Document, url: string) {
//		super(document, url);
//
//		// Find the main post element
//		this.mainPost = document.querySelector('.fatitem');
//
//		// Detect if this is a comment page
//		this.isCommentPage = this.detectCommentPage();
//
//		// Find main comment if on a comment page
//		if (this.isCommentPage) {
//			this.mainComment = this.findMainComment();
//		}
//	}
func NewHackerNewsExtractor(document *goquery.Document, url string, schemaOrgData interface{}) *HackerNewsExtractor {
	extractor := &HackerNewsExtractor{
		ExtractorBase: NewExtractorBase(document, url, schemaOrgData),
	}

	// Find the main post element
	extractor.mainPost = document.Find(".fatitem").First()

	// Detect if this is a comment page
	extractor.isCommentPage = extractor.detectCommentPage()

	// Find main comment if on a comment page
	if extractor.isCommentPage {
		extractor.mainComment = extractor.findMainComment()
	}

	return extractor
}

// detectCommentPage checks if we're on a comment page
// TypeScript original code:
//
//	private detectCommentPage(): boolean {
//		if (!this.mainPost) return false;
//
//		// Check if we're on a comment page by looking for a parent link in the navigation
//		const parentLink = this.mainPost.querySelector('.navs a[href*="parent"]');
//		return !!parentLink;
//	}
func (h *HackerNewsExtractor) detectCommentPage() bool {
	if h.mainPost.Length() == 0 {
		return false
	}

	// Check if we're on a comment page by looking for a parent link in the navigation
	parentLink := h.mainPost.Find(`.navs a[href*="parent"]`)
	return parentLink.Length() > 0
}

// findMainComment finds the main comment on a comment page
// TypeScript original code:
//
//	private findMainComment(): Element | null {
//		if (!this.mainPost) return null;
//
//		// The main comment is the first comment in the fatitem
//		const comment = this.mainPost.querySelector('.comment');
//		return comment;
//	}
func (h *HackerNewsExtractor) findMainComment() *goquery.Selection {
	if h.mainPost.Length() == 0 {
		return nil
	}

	// The main comment is the first comment in the fatitem
	comment := h.mainPost.Find(".comment").First()
	if comment.Length() > 0 {
		return comment
	}

	return nil
}

// CanExtract checks if the extractor can extract content
// TypeScript original code:
//
//	canExtract(): boolean {
//		return !!this.mainPost;
//	}
func (h *HackerNewsExtractor) CanExtract() bool {
	return h.mainPost.Length() > 0
}

// GetName returns the name of the extractor
func (h *HackerNewsExtractor) GetName() string {
	return "HackerNewsExtractor"
}

// Extract extracts the HackerNews content
// TypeScript original code:
//
//	extract(): ExtractorResult {
//		const postContent = this.getPostContent();
//		const comments = this.extractComments();
//
//		const contentHtml = this.createContentHtml(postContent, comments);
//		const postTitle = this.getPostTitle();
//		const postAuthor = this.getPostAuthor();
//		const description = this.createDescription();
//		const published = this.getPostDate();
//
//		return {
//			content: contentHtml,
//			contentHtml: contentHtml,
//			extractedContent: {
//				postId: this.getPostId(),
//				postAuthor: postAuthor
//			},
//			variables: {
//				title: postTitle,
//				author: postAuthor,
//				site: 'Hacker News',
//				description: description,
//				published: published
//			}
//		};
//	}
func (h *HackerNewsExtractor) Extract() *ExtractorResult {
	postContent := h.getPostContent()
	comments := h.extractComments()

	contentHTML := h.createContentHTML(postContent, comments)
	postTitle := h.getPostTitle()
	postAuthor := h.getPostAuthor()
	description := h.createDescription()
	published := h.getPostDate()

	return &ExtractorResult{
		Content:     contentHTML,
		ContentHTML: contentHTML,
		ExtractedContent: map[string]interface{}{
			"postId":     h.getPostID(),
			"postAuthor": postAuthor,
		},
		Variables: map[string]string{
			"title":       postTitle,
			"author":      postAuthor,
			"site":        "Hacker News",
			"description": description,
			"published":   published,
		},
	}
}

// createContentHTML creates the formatted HTML content
// TypeScript original code:
//
//	private createContentHtml(postContent: string, comments: string): string {
//		let content = `<div class="hackernews-post">
//			<div class="post-content">
//				${postContent}
//			</div>
//		</div>`;
//
//		if (comments) {
//			content += `
//		<hr>
//		<h2>Comments</h2>
//		<div class="hackernews-comments">
//			${comments}
//		</div>`;
//		}
//
//		return content.trim();
//	}
func (h *HackerNewsExtractor) createContentHTML(postContent, comments string) string {
	content := fmt.Sprintf(`<div class="hackernews-post">
	<div class="post-content">
		%s
	</div>
</div>`, postContent)

	if comments != "" {
		content += fmt.Sprintf(`
<hr>
<h2>Comments</h2>
<div class="hackernews-comments">
	%s
</div>`, comments)
	}

	return strings.TrimSpace(content)
}

// getPostContent extracts the main post content
// TypeScript original code:
//
//	private getPostContent(): string {
//		if (!this.mainPost) return '';
//
//		// If this is a comment page, use the comment as the main content
//		if (this.isCommentPage && this.mainComment) {
//			const author = this.mainComment.querySelector('.hnuser')?.textContent || '[deleted]';
//			const commentText = this.mainComment.querySelector('.commtext')?.innerHTML || '';
//
//			const timeElement = this.mainComment.querySelector('.age');
//			const timestamp = timeElement?.getAttribute('title') || '';
//			const date = timestamp ? timestamp.split('T')[0] : '';
//
//			const points = this.mainComment.querySelector('.score')?.textContent?.trim() || '';
//			const parentUrl = this.mainPost.querySelector('.navs a[href*="parent"]')?.getAttribute('href') || '';
//
//			let content = '<div class="comment main-comment">';
//			content += '<div class="comment-metadata">';
//			content += `<span class="comment-author"><strong>${author}</strong></span> •`;
//			content += ` <span class="comment-date">${date}</span>`;
//
//			if (points) {
//				content += ` • <span class="comment-points">${points}</span>`;
//			}
//
//			if (parentUrl) {
//				content += ` • <a href="https://news.ycombinator.com/${parentUrl}" class="parent-link">parent</a>`;
//			}
//
//			content += '</div>';
//			content += `<div class="comment-content">${commentText}</div>`;
//			content += '</div>';
//
//			return content;
//		}
//
//		// Otherwise handle regular post content
//		const titleRow = this.mainPost.querySelector('tr.athing');
//		const url = titleRow?.querySelector('.titleline a')?.getAttribute('href') || '';
//
//		let content = '';
//		if (url) {
//			content += `<p><a href="${url}" target="_blank">${url}</a></p>`;
//		}
//
//		const text = this.mainPost.querySelector('.toptext');
//		if (text) {
//			content += `<div class="post-text">${text.innerHTML}</div>`;
//		}
//
//		return content;
//	}
func (h *HackerNewsExtractor) getPostContent() string {
	if h.mainPost.Length() == 0 {
		return ""
	}

	// If this is a comment page, use the comment as the main content
	if h.isCommentPage && h.mainComment != nil && h.mainComment.Length() > 0 {
		author := h.mainComment.Find(".hnuser").Text()
		if author == "" {
			author = "[deleted]"
		}

		commentText, _ := h.mainComment.Find(".commtext").Html()

		timeElement := h.mainComment.Find(".age")
		timestamp, _ := timeElement.Attr("title")
		date := ""
		if timestamp != "" {
			parts := strings.Split(timestamp, "T")
			if len(parts) > 0 {
				date = parts[0]
			}
		}

		points := strings.TrimSpace(h.mainComment.Find(".score").Text())

		parentUrl, _ := h.mainPost.Find(`.navs a[href*="parent"]`).Attr("href")

		var content strings.Builder
		content.WriteString(`<div class="comment main-comment">`)
		content.WriteString(`<div class="comment-metadata">`)
		content.WriteString(fmt.Sprintf(`<span class="comment-author"><strong>%s</strong></span> •`, author))
		content.WriteString(fmt.Sprintf(` <span class="comment-date">%s</span>`, date))

		if points != "" {
			content.WriteString(fmt.Sprintf(` • <span class="comment-points">%s</span>`, points))
		}

		if parentUrl != "" {
			content.WriteString(fmt.Sprintf(` • <a href="https://news.ycombinator.com/%s" class="parent-link">parent</a>`, parentUrl))
		}

		content.WriteString(`</div>`)
		content.WriteString(fmt.Sprintf(`<div class="comment-content">%s</div>`, commentText))
		content.WriteString(`</div>`)

		return content.String()
	}

	// Otherwise handle regular post content
	titleRow := h.mainPost.Find("tr.athing").First()
	url, _ := titleRow.Find(".titleline a").Attr("href")

	var content strings.Builder
	if url != "" {
		content.WriteString(fmt.Sprintf(`<p><a href="%s" target="_blank">%s</a></p>`, url, url))
	}

	text := h.mainPost.Find(".toptext")
	if text.Length() > 0 {
		textHTML, _ := text.Html()
		content.WriteString(fmt.Sprintf(`<div class="post-text">%s</div>`, textHTML))
	}

	return content.String()
}

// extractComments extracts all comments
// TypeScript original code:
//
//	private extractComments(): string {
//		const comments = Array.from(this.document.querySelectorAll('tr.comtr'));
//		return this.processComments(comments);
//	}
func (h *HackerNewsExtractor) extractComments() string {
	var comments []*goquery.Selection
	h.document.Find("tr.comtr").Each(func(i int, s *goquery.Selection) {
		comments = append(comments, s)
	})

	return h.processComments(comments)
}

// processComments processes the comments with proper nesting
// TypeScript original code:
//
//	private processComments(comments: Element[]): string {
//		let html = '';
//		const processedIds = new Set<string>();
//		let currentDepth = -1;
//		const blockquoteStack: number[] = [];
//
//		for (const comment of comments) {
//			const id = comment.getAttribute('id');
//			if (!id || processedIds.has(id)) continue;
//			processedIds.add(id);
//
//			const indentImg = comment.querySelector('.ind img');
//			const indentWidth = parseInt(indentImg?.getAttribute('width') || '0', 10);
//			const depth = indentWidth / 40;
//
//			const commentText = comment.querySelector('.commtext');
//			if (!commentText) continue;
//
//			const author = comment.querySelector('.hnuser')?.textContent || '[deleted]';
//			const timeElement = comment.querySelector('.age');
//			const points = comment.querySelector('.score')?.textContent?.trim() || '';
//
//			const commentUrl = `https://news.ycombinator.com/item?id=${id}`;
//
//			const timestamp = timeElement?.getAttribute('title') || '';
//			const date = timestamp ? timestamp.split('T')[0] : '';
//
//			// For top-level comments, close all previous blockquotes and start fresh
//			if (depth === 0) {
//				while (blockquoteStack.length > 0) {
//					html += '</blockquote>';
//					blockquoteStack.pop();
//				}
//				html += '<blockquote>';
//				blockquoteStack.push(0);
//				currentDepth = 0;
//			} else {
//				// If we're moving back up the tree
//				if (depth < currentDepth) {
//					while (blockquoteStack.length > 0 && blockquoteStack[blockquoteStack.length - 1] >= depth) {
//						html += '</blockquote>';
//						blockquoteStack.pop();
//					}
//				} else if (depth > currentDepth) {
//					// If we're going deeper
//					html += '<blockquote>';
//					blockquoteStack.push(depth);
//				}
//			}
//
//			const commentContent = commentText.innerHTML;
//
//			html += '<div class="comment">';
//			html += '<div class="comment-metadata">';
//			html += `<span class="comment-author"><strong>${author}</strong></span> •`;
//			html += ` <a href="${commentUrl}" class="comment-link">${date}</a>`;
//
//			if (points) {
//				html += ` • <span class="comment-points">${points}</span>`;
//			}
//
//			html += '</div>';
//			html += `<div class="comment-content">${commentContent}</div>`;
//			html += '</div>';
//
//			currentDepth = depth;
//		}
//
//		// Close any remaining blockquotes
//		while (blockquoteStack.length > 0) {
//			html += '</blockquote>';
//			blockquoteStack.pop();
//		}
//
//		return html;
//	}
func (h *HackerNewsExtractor) processComments(comments []*goquery.Selection) string {
	var html strings.Builder
	processedIDs := make(map[string]bool)
	currentDepth := -1
	var blockquoteStack []int

	for _, comment := range comments {
		id, exists := comment.Attr("id")
		if !exists || id == "" || processedIDs[id] {
			continue
		}
		processedIDs[id] = true

		indentImg := comment.Find(".ind img")
		indentWidth, _ := indentImg.Attr("width")
		indent, _ := strconv.Atoi(indentWidth)
		depth := indent / 40

		commentText := comment.Find(".commtext")
		if commentText.Length() == 0 {
			continue
		}

		author := comment.Find(".hnuser").Text()
		if author == "" {
			author = "[deleted]"
		}

		timeElement := comment.Find(".age")
		points := strings.TrimSpace(comment.Find(".score").Text())

		// Get the comment URL
		commentURL := fmt.Sprintf("https://news.ycombinator.com/item?id=%s", id)

		// Get the timestamp from the title attribute and extract the date portion
		timestamp, _ := timeElement.Attr("title")
		date := ""
		if timestamp != "" {
			parts := strings.Split(timestamp, "T")
			if len(parts) > 0 {
				date = parts[0]
			}
		}

		// For top-level comments, close all previous blockquotes and start fresh
		if depth == 0 {
			for len(blockquoteStack) > 0 {
				html.WriteString("</blockquote>")
				blockquoteStack = blockquoteStack[:len(blockquoteStack)-1]
			}
			html.WriteString("<blockquote>")
			blockquoteStack = []int{0}
		} else {
			// If we're moving back up the tree
			if depth < currentDepth {
				for len(blockquoteStack) > 0 && blockquoteStack[len(blockquoteStack)-1] >= depth {
					html.WriteString("</blockquote>")
					blockquoteStack = blockquoteStack[:len(blockquoteStack)-1]
				}
			} else if depth > currentDepth {
				// If we're going deeper
				html.WriteString("<blockquote>")
				blockquoteStack = append(blockquoteStack, depth)
			}
		}

		commentContent, _ := commentText.Html()

		html.WriteString(`<div class="comment">`)
		html.WriteString(`<div class="comment-metadata">`)
		html.WriteString(fmt.Sprintf(`<span class="comment-author"><strong>%s</strong></span> •`, author))
		html.WriteString(fmt.Sprintf(` <a href="%s" class="comment-link">%s</a>`, commentURL, date))

		if points != "" {
			html.WriteString(fmt.Sprintf(` • <span class="comment-points">%s</span>`, points))
		}

		html.WriteString(`</div>`)
		html.WriteString(fmt.Sprintf(`<div class="comment-content">%s</div>`, commentContent))
		html.WriteString(`</div>`)

		currentDepth = depth
	}

	// Close any remaining blockquotes
	for len(blockquoteStack) > 0 {
		html.WriteString("</blockquote>")
		blockquoteStack = blockquoteStack[:len(blockquoteStack)-1]
	}

	return html.String()
}

// getPostID extracts the post ID from the URL
// TypeScript original code:
//
//	private getPostId(): string {
//		const match = this.url.match(/id=(\d+)/);
//		return match ? match[1] : '';
//	}
func (h *HackerNewsExtractor) getPostID() string {
	re := regexp.MustCompile(`id=(\d+)`)
	matches := re.FindStringSubmatch(h.url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// getPostTitle extracts the post title
// TypeScript original code:
//
//	private getPostTitle(): string {
//		if (this.isCommentPage && this.mainComment) {
//			const author = this.mainComment.querySelector('.hnuser')?.textContent || '[deleted]';
//			const commentText = this.mainComment.querySelector('.commtext')?.textContent?.trim() || '';
//
//			// Use first 50 characters of comment as title
//			const preview = commentText.length > 50 ? commentText.substring(0, 50) + '...' : commentText;
//
//			return `Comment by ${author}: ${preview}`;
//		}
//
//		if (!this.mainPost) return '';
//
//		return this.mainPost.querySelector('.titleline')?.textContent?.trim() || '';
//	}
func (h *HackerNewsExtractor) getPostTitle() string {
	if h.isCommentPage && h.mainComment != nil && h.mainComment.Length() > 0 {
		author := h.mainComment.Find(".hnuser").Text()
		if author == "" {
			author = "[deleted]"
		}

		commentText := strings.TrimSpace(h.mainComment.Find(".commtext").Text())

		// Use first 50 characters of comment as title
		preview := commentText
		if len(commentText) > 50 {
			preview = commentText[:50] + "..."
		}

		return fmt.Sprintf("Comment by %s: %s", author, preview)
	}

	if h.mainPost.Length() == 0 {
		return ""
	}

	return strings.TrimSpace(h.mainPost.Find(".titleline").Text())
}

// getPostAuthor extracts the post author
// TypeScript original code:
//
//	private getPostAuthor(): string {
//		if (!this.mainPost) return '';
//
//		return this.mainPost.querySelector('.hnuser')?.textContent?.trim() || '';
//	}
func (h *HackerNewsExtractor) getPostAuthor() string {
	if h.mainPost.Length() == 0 {
		return ""
	}

	return strings.TrimSpace(h.mainPost.Find(".hnuser").Text())
}

// createDescription creates a description for the post
// TypeScript original code:
//
//	private createDescription(): string {
//		const title = this.getPostTitle();
//		const author = this.getPostAuthor();
//
//		if (this.isCommentPage) {
//			return `Comment by ${author} on Hacker News`;
//		}
//
//		return `${title} - by ${author} on Hacker News`;
//	}
func (h *HackerNewsExtractor) createDescription() string {
	title := h.getPostTitle()
	author := h.getPostAuthor()

	if h.isCommentPage {
		return fmt.Sprintf("Comment by %s on Hacker News", author)
	}

	return fmt.Sprintf("%s - by %s on Hacker News", title, author)
}

// getPostDate extracts the post date
// TypeScript original code:
//
//	private getPostDate(): string {
//		if (!this.mainPost) return '';
//
//		const timeElement = this.mainPost.querySelector('.age');
//		const timestamp = timeElement?.getAttribute('title') || '';
//
//		if (timestamp) {
//			return timestamp.split('T')[0];
//		}
//
//		return '';
//	}
func (h *HackerNewsExtractor) getPostDate() string {
	if h.mainPost.Length() == 0 {
		return ""
	}

	timeElement := h.mainPost.Find(".age")
	timestamp, _ := timeElement.Attr("title")

	if timestamp != "" {
		parts := strings.Split(timestamp, "T")
		if len(parts) > 0 {
			return parts[0]
		}
	}

	return ""
}
