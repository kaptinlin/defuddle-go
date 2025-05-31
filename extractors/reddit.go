package extractors

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// RedditExtractor handles Reddit post and comment content extraction
// TypeScript original code:
// import { BaseExtractor } from './_base';
// import { ExtractorResult } from '../types/extractors';
//
//	export class RedditExtractor extends BaseExtractor {
//		private shredditPost: Element | null;
//
//		constructor(document: Document, url: string) {
//			super(document, url);
//			this.shredditPost = document.querySelector('shreddit-post');
//		}
//	}
type RedditExtractor struct {
	*ExtractorBase
	shredditPost *goquery.Selection
}

// NewRedditExtractor creates a new Reddit extractor
// TypeScript original code:
//
//	constructor(document: Document, url: string) {
//		super(document, url);
//		this.shredditPost = document.querySelector('shreddit-post');
//	}
func NewRedditExtractor(document *goquery.Document, url string, schemaOrgData interface{}) *RedditExtractor {
	shredditPost := document.Find("shreddit-post").First()

	slog.Debug("Reddit extractor initialized",
		"hasShredditPost", shredditPost.Length() > 0,
		"url", url)

	return &RedditExtractor{
		ExtractorBase: NewExtractorBase(document, url, schemaOrgData),
		shredditPost:  shredditPost,
	}
}

// CanExtract checks if the extractor can extract content
// TypeScript original code:
//
//	canExtract(): boolean {
//		return !!this.shredditPost;
//	}
func (r *RedditExtractor) CanExtract() bool {
	// Primary check: shreddit-post elements
	if r.shredditPost.Length() > 0 {
		slog.Debug("Reddit extractor can extract check", "canExtract", true, "method", "shreddit-post")
		return true
	}

	// Fallback check: alternative selectors for Reddit content
	fallbackSelectors := []string{
		"[data-testid='post-content']",
		".usertext-body",
		".md",
		"div[data-click-id='text']",
		"div[data-click-id='body']",
		"div[id^='thing_t3_']", // Reddit post format
		".thing.link",          // Old Reddit format
	}

	for _, selector := range fallbackSelectors {
		if r.document.Find(selector).Length() > 0 {
			slog.Debug("Reddit extractor can extract check", "canExtract", true, "method", "fallback", "selector", selector)
			return true
		}
	}

	slog.Debug("Reddit extractor can extract check", "canExtract", false)
	return false
}

// GetName returns the name of the extractor
func (r *RedditExtractor) GetName() string {
	return "RedditExtractor"
}

// Extract extracts the Reddit post and comments
// TypeScript original code:
//
//	extract(): ExtractorResult {
//		const postContent = this.getPostContent();
//		const comments = this.extractComments();
//
//		const contentHtml = this.createContentHtml(postContent, comments);
//		const postTitle = this.document.querySelector('h1')?.textContent?.trim() || '';
//		const subreddit = this.getSubreddit();
//		const postAuthor = this.getPostAuthor();
//		const description = this.createDescription(postContent);
//
//		return {
//			content: contentHtml,
//			contentHtml: contentHtml,
//			extractedContent: {
//				postId: this.getPostId(),
//				subreddit,
//				 postAuthor,
//			},
//			variables: {
//				title: postTitle,
//				author: postAuthor,
//				site: `r/${subreddit}`,
//				description,
//			}
//		};
//	}
func (r *RedditExtractor) Extract() *ExtractorResult {
	slog.Debug("Reddit extractor starting extraction", "url", r.url)

	postContent := r.getPostContent()
	comments := r.extractComments()

	contentHTML := r.createContentHTML(postContent, comments)
	postTitle := r.getPostTitle()
	subreddit := r.getSubreddit()
	postAuthor := r.getPostAuthor()
	description := r.createDescription(postContent)
	postID := r.getPostID()

	slog.Debug("Reddit extraction completed",
		"postTitle", postTitle,
		"postAuthor", postAuthor,
		"subreddit", subreddit,
		"postId", postID,
		"hasComments", comments != "")

	return &ExtractorResult{
		Content:     contentHTML,
		ContentHTML: contentHTML,
		ExtractedContent: map[string]interface{}{
			"postId":     postID,
			"subreddit":  subreddit,
			"postAuthor": postAuthor,
		},
		Variables: map[string]string{
			"title":       postTitle,
			"author":      postAuthor,
			"site":        fmt.Sprintf("r/%s", subreddit),
			"description": description,
		},
	}
}

// getPostContent extracts the main post content
// TypeScript original code:
//
//	private getPostContent(): string {
//		const textBody = this.shredditPost?.querySelector('[slot="text-body"]')?.innerHTML || '';
//		const mediaBody = this.shredditPost?.querySelector('#post-image')?.outerHTML || '';
//
//		return textBody + mediaBody;
//	}
func (r *RedditExtractor) getPostContent() string {
	var content strings.Builder

	// Primary method: Look for shreddit-post elements
	if r.shredditPost.Length() > 0 {
		slog.Debug("Reddit extractor: using shreddit-post element")

		// Get text body content
		textBody := r.shredditPost.Find(`[slot="text-body"]`).First()
		if textBody.Length() > 0 {
			textBodyHTML, _ := textBody.Html()
			content.WriteString(textBodyHTML)
		}

		// Get media body content
		mediaBody := r.shredditPost.Find("#post-image").First()
		if mediaBody.Length() > 0 {
			mediaBodyHTML, _ := mediaBody.Html()
			// Use innerHTML equivalent since TypeScript uses outerHTML
			content.WriteString(fmt.Sprintf(`<div id="post-image">%s</div>`, mediaBodyHTML))
		}
	} else {
		// Fallback method: Look for alternative selectors
		slog.Debug("Reddit extractor: using fallback selectors")

		// Try to find post content using alternative selectors
		alternativeSelectors := []string{
			"div[data-testid='post-content']",
			".usertext-body",
			".md",
			"div[data-click-id='text']",
			"div[data-click-id='body']",
		}

		for _, selector := range alternativeSelectors {
			postContent := r.document.Find(selector).First()
			if postContent.Length() > 0 {
				if html, err := postContent.Html(); err == nil && html != "" {
					content.WriteString(html)
					slog.Debug("Reddit extractor: found content with selector", "selector", selector)
					break
				}
			}
		}

		// Try to find images separately
		imageSelectors := []string{
			"img[src*='i.redd.it']",
			"img[src*='preview.redd.it']",
			"img[src*='external-preview.redd.it']",
		}

		for _, selector := range imageSelectors {
			images := r.document.Find(selector)
			if images.Length() > 0 {
				images.Each(func(i int, img *goquery.Selection) {
					if outerHTML, err := img.Clone().Wrap("<div>").Parent().Html(); err == nil {
						content.WriteString(outerHTML)
					}
				})
				break
			}
		}
	}

	result := content.String()
	slog.Debug("Reddit extractor: extracted post content",
		"hasShredditPost", r.shredditPost.Length() > 0,
		"contentLength", len(result))

	return result
}

// createContentHTML creates the formatted HTML content
// TypeScript original code:
//
//	private createContentHtml(postContent: string, comments: string): string {
//		return `
//			<div class="reddit-post">
//				<div class="post-content">
//					${postContent}
//				</div>
//			</div>
//			${comments ? `
//				<hr>
//				<h2>Comments</h2>
//				<div class="reddit-comments">
//					${comments}
//				</div>
//			` : ''}
//		`.trim();
//	}
func (r *RedditExtractor) createContentHTML(postContent, comments string) string {
	var content strings.Builder

	content.WriteString(`<div class="reddit-post">`)
	content.WriteString(`<div class="post-content">`)
	content.WriteString(postContent)
	content.WriteString(`</div>`)
	content.WriteString(`</div>`)

	if comments != "" {
		content.WriteString(`<hr>`)
		content.WriteString(`<h2>Comments</h2>`)
		content.WriteString(`<div class="reddit-comments">`)
		content.WriteString(comments)
		content.WriteString(`</div>`)
	}

	return strings.TrimSpace(content.String())
}

// extractComments extracts comments from the page
// TypeScript original code:
//
//	private extractComments(): string {
//		const comments = Array.from(this.document.querySelectorAll('shreddit-comment'));
//		return this.processComments(comments);
//	}
func (r *RedditExtractor) extractComments() string {
	var comments []*goquery.Selection

	// Primary method: Look for shreddit-comment elements
	r.document.Find("shreddit-comment").Each(func(i int, s *goquery.Selection) {
		comments = append(comments, s)
	})

	// Fallback method: Look for alternative comment selectors
	if len(comments) == 0 {
		slog.Debug("Reddit extractor: using fallback comment selectors")

		alternativeSelectors := []string{
			"div[data-testid='comment']",
			".comment",
			".comment-area .comment",
			"div[data-click-id='text']",
			"div[data-click-id='body']",
			"div[id^='thing_t3_']", // Reddit post format
			".thing.link",          // Old Reddit format
		}

		for _, selector := range alternativeSelectors {
			r.document.Find(selector).Each(func(i int, s *goquery.Selection) {
				comments = append(comments, s)
			})
			if len(comments) > 0 {
				slog.Debug("Reddit extractor: found comments with selector", "selector", selector, "count", len(comments))
				break
			}
		}
	}

	slog.Debug("Reddit extractor: found comments", "commentCount", len(comments))

	if len(comments) == 0 {
		return ""
	}

	return r.processComments(comments)
}

// processComments processes the comments with proper nesting
// TypeScript original code:
//
//	private processComments(comments: Element[]): string {
//		let html = '';
//		let currentDepth = -1;
//		let blockquoteStack: number[] = []; // Keep track of open blockquotes at each depth
//
//		for (const comment of comments) {
//			const depth = parseInt(comment.getAttribute('depth') || '0');
//			const author = comment.getAttribute('author') || '';
//			const score = comment.getAttribute('score') || '0';
//			const permalink = comment.getAttribute('permalink') || '';
//			const content = comment.querySelector('[slot="comment"]')?.innerHTML || '';
//
//			// Get timestamp from faceplate-timeago element
//			const timeElement = comment.querySelector('faceplate-timeago');
//			const timestamp = timeElement?.getAttribute('ts') || '';
//			const date = timestamp ? new Date(timestamp).toISOString().split('T')[0] : '';
//
//			// For top-level comments, close all previous blockquotes and start fresh
//			if (depth === 0) {
//				// Close all open blockquotes
//				while (blockquoteStack.length > 0) {
//					html += '</blockquote>';
//					blockquoteStack.pop();
//				}
//				html += '<blockquote>';
//				blockquoteStack = [0];
//				currentDepth = 0;
//			}
//			// For nested comments
//			else {
//				// If we're moving back up the tree
//				if (depth < currentDepth) {
//					// Close blockquotes until we reach the current depth
//					while (blockquoteStack.length > 0 && blockquoteStack[blockquoteStack.length - 1] >= depth) {
//						html += '</blockquote>';
//						blockquoteStack.pop();
//					}
//				}
//				// If we're going deeper
//				else if (depth > currentDepth) {
//					html += '<blockquote>';
//					blockquoteStack.push(depth);
//				}
//				// If we're at the same depth, no need to close or open blockquotes
//			}
//
//			html += `<div class="comment">
//	<div class="comment-metadata">
//		<span class="comment-author"><strong>${author}</strong></span> •
//		<a href="https://reddit.com${permalink}" class="comment-link">${score} points</a> •
//		<span class="comment-date">${date}</span>
//	</div>
//	<div class="comment-content">${content}</div>
//
// </div>`;
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
func (r *RedditExtractor) processComments(comments []*goquery.Selection) string {
	var html strings.Builder
	currentDepth := -1
	var blockquoteStack []int // Keep track of open blockquotes at each depth

	slog.Debug("Reddit extractor: processing comments", "totalComments", len(comments))

	for _, comment := range comments {
		depthStr, _ := comment.Attr("depth")
		depth, _ := strconv.Atoi(depthStr)

		author, _ := comment.Attr("author")
		score, _ := comment.Attr("score")
		permalink, _ := comment.Attr("permalink")

		contentElement := comment.Find(`[slot="comment"]`).First()
		content, _ := contentElement.Html()

		// Get timestamp from faceplate-timeago element
		timeElement := comment.Find("faceplate-timeago").First()
		timestamp, _ := timeElement.Attr("ts")

		var date string
		if timestamp != "" {
			// Parse timestamp and convert to date
			if ts, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
				date = time.Unix(ts, 0).Format("2006-01-02")
			}
		}

		// For top-level comments, close all previous blockquotes and start fresh
		if depth == 0 {
			// Close all open blockquotes
			for len(blockquoteStack) > 0 {
				html.WriteString("</blockquote>")
				blockquoteStack = blockquoteStack[:len(blockquoteStack)-1]
			}
			html.WriteString("<blockquote>")
			blockquoteStack = []int{0}
		} else {
			// For nested comments
			// If we're moving back up the tree
			if depth < currentDepth {
				// Close blockquotes until we reach the current depth
				for len(blockquoteStack) > 0 && blockquoteStack[len(blockquoteStack)-1] >= depth {
					html.WriteString("</blockquote>")
					blockquoteStack = blockquoteStack[:len(blockquoteStack)-1]
				}
			} else if depth > currentDepth {
				// If we're going deeper
				html.WriteString("<blockquote>")
				blockquoteStack = append(blockquoteStack, depth)
			}
			// If we're at the same depth, no need to close or open blockquotes
		}

		html.WriteString(`<div class="comment">`)
		html.WriteString(`<div class="comment-metadata">`)
		html.WriteString(fmt.Sprintf(`<span class="comment-author"><strong>%s</strong></span> •`, author))
		html.WriteString(fmt.Sprintf(` <a href="https://reddit.com%s" class="comment-link">%s points</a> •`, permalink, score))
		html.WriteString(fmt.Sprintf(` <span class="comment-date">%s</span>`, date))
		html.WriteString(`</div>`)
		html.WriteString(fmt.Sprintf(`<div class="comment-content">%s</div>`, content))
		html.WriteString(`</div>`)

		currentDepth = depth
	}

	// Close any remaining blockquotes
	for len(blockquoteStack) > 0 {
		html.WriteString("</blockquote>")
		blockquoteStack = blockquoteStack[:len(blockquoteStack)-1]
	}

	slog.Debug("Reddit extractor: comments processed", "processedCount", len(comments))
	return html.String()
}

// getPostID extracts the post ID from URL
// TypeScript original code:
//
//	private getPostId(): string {
//		const match = this.url.match(/comments\/([a-zA-Z0-9]+)/);
//		return match?.[1] || '';
//	}
func (r *RedditExtractor) getPostID() string {
	re := regexp.MustCompile(`comments/([a-zA-Z0-9]+)`)
	matches := re.FindStringSubmatch(r.url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// getSubreddit extracts the subreddit name from URL
// TypeScript original code:
//
//	private getSubreddit(): string {
//		const match = this.url.match(/\/r\/([^/]+)/);
//		return match?.[1] || '';
//	}
func (r *RedditExtractor) getSubreddit() string {
	re := regexp.MustCompile(`/r/([^/]+)`)
	matches := re.FindStringSubmatch(r.url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// getPostAuthor extracts the post author
// TypeScript original code:
//
//	private getPostAuthor(): string {
//		return this.shredditPost?.getAttribute('author') || '';
//	}
func (r *RedditExtractor) getPostAuthor() string {
	if r.shredditPost.Length() > 0 {
		author, _ := r.shredditPost.Attr("author")
		return author
	}
	return ""
}

// getPostTitle extracts the post title
// TypeScript original code:
//
//	const postTitle = this.document.querySelector('h1')?.textContent?.trim() || '';
func (r *RedditExtractor) getPostTitle() string {
	// First try to get title from h1 element
	h1Title := strings.TrimSpace(r.document.Find("h1").First().Text())
	if h1Title != "" {
		return h1Title
	}

	// Fallback to page title
	pageTitle := strings.TrimSpace(r.document.Find("title").Text())
	if pageTitle != "" && pageTitle != "Reddit - The heart of the internet" {
		return pageTitle
	}

	return ""
}

// createDescription creates a description from post content
// TypeScript original code:
//
//	private createDescription(postContent: string): string {
//		if (!postContent) return '';
//
//		const tempDiv = document.createElement('div');
//		tempDiv.innerHTML = postContent;
//		return tempDiv.textContent?.trim()
//			.slice(0, 140)
//			.replace(/\s+/g, ' ') || '';
//	}
func (r *RedditExtractor) createDescription(postContent string) string {
	if postContent == "" {
		return ""
	}

	// Create a temporary document to extract text content
	tempDoc, err := goquery.NewDocumentFromReader(strings.NewReader(postContent))
	if err != nil {
		slog.Warn("Reddit extractor: failed to parse post content for description", "error", err)
		return ""
	}

	textContent := strings.TrimSpace(tempDoc.Text())

	// Replace multiple whitespace with single space
	re := regexp.MustCompile(`\s+`)
	textContent = re.ReplaceAllString(textContent, " ")

	// Limit to 140 characters
	if len(textContent) > 140 {
		return textContent[:140]
	}

	return textContent
}
