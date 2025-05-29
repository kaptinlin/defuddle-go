package extractors

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// RedditExtractor handles Reddit post and comment content extraction
type RedditExtractor struct {
	*ExtractorBase
	post     *goquery.Selection
	comments *goquery.Selection
}

// NewRedditExtractor creates a new Reddit extractor
// TypeScript original code:
//
//	constructor(document: Document, url: string, schemaOrgData?: any) {
//		super(document, url, schemaOrgData);
//		this.post = document.querySelector('shreddit-post');
//		this.comments = document.querySelectorAll('shreddit-comment');
//	}
func NewRedditExtractor(document *goquery.Document, url string, schemaOrgData interface{}) *RedditExtractor {
	return &RedditExtractor{
		ExtractorBase: NewExtractorBase(document, url, schemaOrgData),
		post:          document.Find("shreddit-post").First(),
		comments:      document.Find("shreddit-comment"),
	}
}

// CanExtract checks if the extractor can extract content
// TypeScript original code:
//
//	canExtract(): boolean {
//		return !!this.post;
//	}
func (r *RedditExtractor) CanExtract() bool {
	return r.post.Length() > 0
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
	postContent := r.extractPost()
	comments := r.extractComments()

	contentHTML := r.createContentHTML(postContent, comments)
	postTitle := r.getPostTitle()
	subreddit := r.getSubreddit()
	postAuthor := r.getPostAuthor()
	description := r.createDescription(postContent)
	postID := r.getPostID()

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

// extractPost extracts the main post content
// TypeScript original code:
//
//	private extractPost(): string {
//		if (!this.post) return '';
//
//		let content = '';
//		const author = this.post.getAttribute('author') || 'Unknown';
//		const postTitle = this.post.getAttribute('post-title') || '';
//
//		if (postTitle) {
//			content += `<h1>${postTitle}</h1>\n\n`;
//		}
//
//		content += `<p><strong>Posted by u/${author}</strong></p>\n\n`;
//
//		const postContentElement = this.post.querySelector('[slot="text-body"]');
//		if (postContentElement) {
//			const postBody = postContentElement.innerHTML;
//			if (postBody && postBody.trim()) {
//				content += postBody.trim() + '\n\n';
//			}
//		}
//
//		return content;
//	}
func (r *RedditExtractor) extractPost() string {
	if r.post.Length() == 0 {
		return ""
	}

	var content strings.Builder
	author := r.GetAttribute(r.post, "author")
	if author == "" {
		author = "Unknown"
	}
	postTitle := r.GetAttribute(r.post, "post-title")

	if postTitle != "" {
		content.WriteString(fmt.Sprintf("<h1>%s</h1>\n\n", postTitle))
	}

	content.WriteString(fmt.Sprintf("<p><strong>Posted by u/%s</strong></p>\n\n", author))

	postContentElement := r.post.Find(`[slot="text-body"]`).First()
	if postContentElement.Length() > 0 {
		postBody := r.GetHTMLContent(postContentElement)
		if strings.TrimSpace(postBody) != "" {
			content.WriteString(postBody)
			content.WriteString("\n\n")
		}
	}

	return content.String()
}

// extractComments extracts and nests the comments
// TypeScript original code:
//
//	private extractComments(): string {
//		if (!this.comments || this.comments.length === 0) return '';
//
//		const processComment = (comment: Element, depth: number = 0): string => {
//			const author = comment.getAttribute('author') || 'Unknown';
//			const commentBodyElement = comment.querySelector('[slot="comment"]');
//			const commentBody = commentBodyElement ? commentBodyElement.innerHTML : '';
//
//			if (!commentBody.trim()) return '';
//
//			let commentHtml = '';
//			if (depth === 0) {
//				commentHtml += `<div class="comment comment-level-${depth}">\n`;
//				commentHtml += `  <p><strong>u/${author}</strong></p>\n`;
//				commentHtml += `  <div class="comment-body">${commentBody.trim()}</div>\n`;
//				commentHtml += `</div>\n\n`;
//			} else {
//				commentHtml += `<blockquote>\n`;
//				commentHtml += `  <p><strong>u/${author}</strong></p>\n`;
//				commentHtml += `  <div class="comment-body">${commentBody.trim()}</div>\n`;
//				commentHtml += `</blockquote>\n\n`;
//			}
//
//			return commentHtml;
//		};
//
//		let commentsHtml = '';
//		this.comments.forEach((comment) => {
//			const depth = parseInt(comment.getAttribute('depth') || '0', 10);
//			commentsHtml += processComment(comment, depth);
//		});
//
//		return commentsHtml;
//	}
func (r *RedditExtractor) extractComments() string {
	if r.comments.Length() == 0 {
		return ""
	}

	processComment := func(comment *goquery.Selection, depth int) string {
		author := r.GetAttribute(comment, "author")
		if author == "" {
			author = "Unknown"
		}

		commentBodyElement := comment.Find(`[slot="comment"]`).First()
		commentBody := r.GetHTMLContent(commentBodyElement)

		if strings.TrimSpace(commentBody) == "" {
			return ""
		}

		var commentHTML strings.Builder
		if depth == 0 {
			commentHTML.WriteString(fmt.Sprintf(`<div class="comment comment-level-%d">`+"\n", depth))
			commentHTML.WriteString(fmt.Sprintf("  <p><strong>u/%s</strong></p>\n", author))
			commentHTML.WriteString(fmt.Sprintf("  <div class=\"comment-body\">%s</div>\n", strings.TrimSpace(commentBody)))
			commentHTML.WriteString("</div>\n\n")
		} else {
			commentHTML.WriteString("<blockquote>\n")
			commentHTML.WriteString(fmt.Sprintf("  <p><strong>u/%s</strong></p>\n", author))
			commentHTML.WriteString(fmt.Sprintf("  <div class=\"comment-body\">%s</div>\n", strings.TrimSpace(commentBody)))
			commentHTML.WriteString("</blockquote>\n\n")
		}

		return commentHTML.String()
	}

	var commentsHTML strings.Builder
	r.comments.Each(func(i int, comment *goquery.Selection) {
		depthStr := r.GetAttribute(comment, "depth")
		depth := 0
		if depthStr != "" {
			if d, err := parseDepthFromString(depthStr); err == nil {
				depth = d
			}
		}
		commentsHTML.WriteString(processComment(comment, depth))
	})

	return commentsHTML.String()
}

// parseDepthFromString safely parses depth from string
func parseDepthFromString(s string) (int, error) {
	// Simple integer parsing for depth
	depth := 0
	for _, char := range s {
		if char >= '0' && char <= '9' {
			depth = depth*10 + int(char-'0')
		} else {
			break
		}
	}
	return depth, nil
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

	// Try to get title from post-title attribute
	if r.post.Length() > 0 {
		postTitle := r.GetAttribute(r.post, "post-title")
		if postTitle != "" {
			return postTitle
		}
	}

	// Fallback to page title
	pageTitle := strings.TrimSpace(r.document.Find("title").Text())
	if pageTitle != "" && pageTitle != "Reddit - The heart of the internet" {
		return pageTitle
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
	// Extract subreddit from URL pattern like /r/subredditname/
	if strings.Contains(r.url, "/r/") {
		parts := strings.Split(r.url, "/r/")
		if len(parts) > 1 {
			subredditPart := parts[1]
			// Find the end of subreddit name (next slash)
			if slashIndex := strings.Index(subredditPart, "/"); slashIndex != -1 {
				return subredditPart[:slashIndex]
			}
			return subredditPart
		}
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
	if r.post.Length() > 0 {
		author := r.GetAttribute(r.post, "author")
		if author != "" {
			return author
		}
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
		return ""
	}

	textContent := strings.TrimSpace(tempDoc.Text())
	textContent = strings.ReplaceAll(textContent, "\n", " ")
	textContent = strings.ReplaceAll(textContent, "\t", " ")

	// Replace multiple spaces with single space
	for strings.Contains(textContent, "  ") {
		textContent = strings.ReplaceAll(textContent, "  ", " ")
	}

	// Limit to 140 characters
	if len(textContent) > 140 {
		return textContent[:140] + "..."
	}

	return textContent
}

// getPostID extracts the post ID from URL
// TypeScript original code:
//
//	private getPostId(): string {
//		const match = this.url.match(/comments\/([a-zA-Z0-9]+)/);
//		return match?.[1] || '';
//	}
func (r *RedditExtractor) getPostID() string {
	// Extract post ID from URL pattern like /comments/postid/
	if strings.Contains(r.url, "/comments/") {
		parts := strings.Split(r.url, "/comments/")
		if len(parts) > 1 {
			postIDPart := parts[1]
			// Find the end of post ID (next slash)
			if slashIndex := strings.Index(postIDPart, "/"); slashIndex != -1 {
				return postIDPart[:slashIndex]
			}
			return postIDPart
		}
	}
	return ""
}
