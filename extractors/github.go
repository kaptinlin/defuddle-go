package extractors

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Pre-compiled regex patterns for GitHub extraction.
var (
	githubUserRe       = regexp.MustCompile(`github\.com/([^/?#]+)`)
	githubRepoRe       = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)`)
	githubTitleRepoRe  = regexp.MustCompile(`([^/\s]+)/([^/\s]+)`)
	githubIssueRe      = regexp.MustCompile(`/issues/(\d+)`)
	githubWhitespaceRe = regexp.MustCompile(`\s+`)
)

// GitHubExtractor handles GitHub content extraction
// TypeScript original code:
//
//	export class GitHubExtractor extends BaseExtractor {
//		canExtract(): boolean {
//			const githubIndicators = [
//				'meta[name="expected-hostname"][content="github.com"]',
//				'meta[name="octolytics-url"]',
//				'meta[name="github-keyboard-shortcuts"]',
//				'.js-header-wrapper',
//				'#js-repo-pjax-container',
//			];
//
//			const githubPageIndicators = {
//				issue: [
//					'[data-testid="issue-metadata-sticky"]',
//					'[data-testid="issue-title"]',
//				],
//			}
//
//			return githubIndicators.some(selector => this.document.querySelector(selector) !== null)
//				&& Object.values(githubPageIndicators).some(selectors => selectors.some(selector => this.document.querySelector(selector) !== null));
//		}
//
//		extract(): ExtractorResult {
//			return this.extractIssue();
//		}
type GitHubExtractor struct {
	*ExtractorBase
}

// NewGitHubExtractor creates a new GitHub extractor
func NewGitHubExtractor(document *goquery.Document, url string, schemaOrgData any) *GitHubExtractor {
	extractor := &GitHubExtractor{
		ExtractorBase: NewExtractorBase(document, url, schemaOrgData),
	}

	slog.Debug("GitHub extractor initialized", "url", url)
	return extractor
}

// CanExtract checks if the extractor can extract content
// TypeScript original code:
//
//	canExtract(): boolean {
//		const githubIndicators = [
//			'meta[name="expected-hostname"][content="github.com"]',
//			'meta[name="octolytics-url"]',
//			'meta[name="github-keyboard-shortcuts"]',
//			'.js-header-wrapper',
//			'#js-repo-pjax-container',
//		];
//
//		const githubPageIndicators = {
//			issue: [
//				'[data-testid="issue-metadata-sticky"]',
//				'[data-testid="issue-title"]',
//			],
//		}
//
//		return githubIndicators.some(selector => this.document.querySelector(selector) !== null)
//			&& Object.values(githubPageIndicators).some(selectors => selectors.some(selector => this.document.querySelector(selector) !== null));
//	}
func (g *GitHubExtractor) CanExtract() bool {
	githubIndicators := []string{
		`meta[name="expected-hostname"][content="github.com"]`,
		`meta[name="octolytics-url"]`,
		`meta[name="github-keyboard-shortcuts"]`,
		`.js-header-wrapper`,
		`#js-repo-pjax-container`,
	}

	githubPageIndicators := []string{
		`[data-testid="issue-metadata-sticky"]`,
		`[data-testid="issue-title"]`,
	}

	// Check for GitHub indicators
	hasGitHubIndicator := false
	for _, selector := range githubIndicators {
		if g.document.Find(selector).Length() > 0 {
			hasGitHubIndicator = true
			break
		}
	}

	// Check for page-specific indicators
	hasPageIndicator := false
	for _, selector := range githubPageIndicators {
		if g.document.Find(selector).Length() > 0 {
			hasPageIndicator = true
			break
		}
	}

	canExtract := hasGitHubIndicator && hasPageIndicator
	slog.Debug("GitHub extractor can extract check", "canExtract", canExtract, "url", g.url)
	return canExtract
}

// GetName returns the name of the extractor
func (g *GitHubExtractor) Name() string {
	return "GitHubExtractor"
}

// Extract extracts the GitHub content
// TypeScript original code:
//
//	extract(): ExtractorResult {
//		return this.extractIssue();
//	}
func (g *GitHubExtractor) Extract() *ExtractorResult {
	slog.Debug("GitHub extractor starting extraction", "url", g.url)
	return g.extractIssue()
}

// extractIssue extracts GitHub issue content with comprehensive structure
// TypeScript original code: Full implementation with issue body, comments, and metadata
func (g *GitHubExtractor) extractIssue() *ExtractorResult {
	slog.Debug("GitHub extractor extracting issue")

	repoInfo := g.extractRepoInfo()
	issueNumber := g.extractIssueNumber()

	var content strings.Builder

	// Extract the main issue body first
	issueContainer := g.document.Find(`[data-testid="issue-viewer-issue-container"]`).First()
	if issueContainer.Length() > 0 {
		issueAuthor := g.extractAuthor(issueContainer, []string{
			`a[data-testid="issue-body-header-author"]`,
			`.IssueBodyHeaderAuthor-module__authorLoginLink--_S7aT`,
			`.ActivityHeader-module__AuthorLink--iofTU`,
			`a[href*="/users/"][data-hovercard-url*="/users/"]`,
			`a[aria-label*="profile"]`,
		})

		issueTimeElement := issueContainer.Find("relative-time").First()
		issueTimestamp := ""
		if issueTimeElement.Length() > 0 {
			if datetime, exists := issueTimeElement.Attr("datetime"); exists {
				issueTimestamp = datetime
			}
		}

		issueBodyElement := issueContainer.Find(`[data-testid="issue-body-viewer"] .markdown-body`).First()
		if issueBodyElement.Length() > 0 {
			bodyContent := g.cleanBodyContent(issueBodyElement)

			// Add the main issue
			content.WriteString(fmt.Sprintf(`<div class="issue-author"><strong>%s</strong>`, issueAuthor))
			if issueTimestamp != "" {
				if date, err := time.Parse(time.RFC3339, issueTimestamp); err == nil {
					content.WriteString(fmt.Sprintf(` opened this issue on %s`, date.Format("January 2, 2006")))
				}
			}
			content.WriteString("</div>\n\n")
			content.WriteString(fmt.Sprintf(`<div class="issue-body">%s</div>\n\n`, bodyContent))
		}
	}

	// Extract comments
	commentElements := g.document.Find(`[data-wrapper-timeline-id]`)
	processedComments := make(map[string]bool)

	commentElements.Each(func(_ int, commentElement *goquery.Selection) {
		commentContainer := commentElement.Find(".react-issue-comment").First()
		if commentContainer.Length() == 0 {
			return
		}

		commentID, exists := commentElement.Attr("data-wrapper-timeline-id")
		if !exists || commentID == "" || processedComments[commentID] {
			return
		}
		processedComments[commentID] = true

		author := g.extractAuthor(commentContainer, []string{
			`.ActivityHeader-module__AuthorLink--iofTU`,
			`a[data-testid="avatar-link"]`,
			`a[href^="/"][data-hovercard-url*="/users/"]`,
		})

		timeElement := commentContainer.Find("relative-time").First()
		timestamp := ""
		if timeElement.Length() > 0 {
			if datetime, exists := timeElement.Attr("datetime"); exists {
				timestamp = datetime
			}
		}

		bodyElement := commentContainer.Find(".markdown-body").First()
		if bodyElement.Length() > 0 {
			bodyContent := g.cleanBodyContent(bodyElement)
			if bodyContent != "" {
				content.WriteString(`<div class="comment">\n`)
				content.WriteString(fmt.Sprintf(`<div class="comment-header"><strong>%s</strong>`, author))
				if timestamp != "" {
					if date, err := time.Parse(time.RFC3339, timestamp); err == nil {
						content.WriteString(fmt.Sprintf(` commented on %s`, date.Format("January 2, 2006")))
					}
				}
				content.WriteString(`</div>\n`)
				content.WriteString(fmt.Sprintf(`<div class="comment-body">%s</div>\n`, bodyContent))
				content.WriteString(`</div>\n\n`)
			}
		}
	})

	contentHTML := content.String()
	description := g.createDescription(contentHTML)
	title := g.document.Find("title").Text()

	slog.Debug("GitHub issue extraction completed",
		"title", title,
		"issueNumber", issueNumber,
		"repo", fmt.Sprintf("%s/%s", repoInfo["owner"], repoInfo["repo"]),
		"contentLength", len(contentHTML))

	return &ExtractorResult{
		Content:     contentHTML,
		ContentHTML: contentHTML,
		ExtractedContent: map[string]any{
			"type":        "issue",
			"issueNumber": issueNumber,
			"repository":  repoInfo["repo"],
			"owner":       repoInfo["owner"],
		},
		Variables: map[string]string{
			"title":       title,
			"author":      "",
			"site":        fmt.Sprintf("GitHub - %s/%s", repoInfo["owner"], repoInfo["repo"]),
			"description": description,
		},
	}
}

// extractAuthor extracts author from a container using multiple selectors
// TypeScript original code:
//
//	private extractAuthor(container: Element, selectors: string[]): string {
//		for (const selector of selectors) {
//			const authorLink = container.querySelector(selector);
//			if (authorLink) {
//				const href = authorLink.getAttribute('href');
//				if (href) {
//					if (href.startsWith('/')) {
//						return href.substring(1);
//					} else if (href.includes('github.com/')) {
//						const match = href.match(/github\.com\/([^\/\?#]+)/);
//						if (match && match[1]) {
//							return match[1];
//						}
//					}
//				}
//			}
//		}
//		return 'Unknown';
//	}
func (g *GitHubExtractor) extractAuthor(container *goquery.Selection, selectors []string) string {
	for _, selector := range selectors {
		authorLink := container.Find(selector).First()
		if authorLink.Length() > 0 {
			if href, exists := authorLink.Attr("href"); exists {
				if strings.HasPrefix(href, "/") {
					return href[1:]
				} else if strings.Contains(href, "github.com/") {
					matches := githubUserRe.FindStringSubmatch(href)
					if len(matches) > 1 && matches[1] != "" {
						return matches[1]
					}
				}
			}
		}
	}
	return "Unknown"
}

// cleanBodyContent cleans markdown body content by removing buttons and clipboard elements
// TypeScript original code:
//
//	private cleanBodyContent(bodyElement: Element): string {
//		const cleanBody = bodyElement.cloneNode(true) as Element;
//		cleanBody.querySelectorAll('button, [data-testid*="button"], [data-testid*="menu"]').forEach(el => el.remove());
//		cleanBody.querySelectorAll('.js-clipboard-copy, .zeroclipboard-container').forEach(el => el.remove());
//		return cleanBody.innerHTML.trim();
//	}
func (g *GitHubExtractor) cleanBodyContent(bodyElement *goquery.Selection) string {
	// Clone the selection to avoid modifying the original
	htmlContent, err := bodyElement.Html()
	if err != nil {
		return ""
	}

	// Create a new document from the HTML content
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent // Return original if parsing fails
	}

	// Remove buttons and menu elements
	doc.Find(`button, [data-testid*="button"], [data-testid*="menu"]`).Remove()

	// Remove clipboard elements
	doc.Find(".js-clipboard-copy, .zeroclipboard-container").Remove()

	// Get the cleaned HTML
	cleanedHTML, err := doc.Html()
	if err != nil {
		return htmlContent // Return original if extraction fails
	}

	return strings.TrimSpace(cleanedHTML)
}

// extractRepoInfo extracts repository owner and name
// TypeScript original code:
//
//	private extractRepoInfo(): { owner: string; repo: string } {
//		// Try URL first (most reliable)
//		const urlMatch = this.url.match(/github\.com\/([^\/]+)\/([^\/]+)/);
//		if (urlMatch) {
//			return { owner: urlMatch[1], repo: urlMatch[2] };
//		}
//
//		// Fallback to HTML extraction
//		const titleMatch = this.document.title.match(/([^\/\s]+)\/([^\/\s]+)/);
//		return titleMatch ? { owner: titleMatch[1], repo: titleMatch[2] } : { owner: '', repo: '' };
//	}
func (g *GitHubExtractor) extractRepoInfo() map[string]string {
	// Try URL first (most reliable)
	matches := githubRepoRe.FindStringSubmatch(g.url)
	if len(matches) >= 3 {
		return map[string]string{
			"owner": matches[1],
			"repo":  matches[2],
		}
	}

	// Fallback to HTML extraction
	title := g.document.Find("title").Text()
	titleMatches := githubTitleRepoRe.FindStringSubmatch(title)
	if len(titleMatches) >= 3 {
		return map[string]string{
			"owner": titleMatches[1],
			"repo":  titleMatches[2],
		}
	}

	return map[string]string{
		"owner": "",
		"repo":  "",
	}
}

// extractIssueNumber extracts the issue number from URL
// TypeScript original code:
//
//	private extractIssueNumber(): string {
//		const match = this.url.match(/\/issues\/(\d+)/);
//		return match ? match[1] : '';
//	}
func (g *GitHubExtractor) extractIssueNumber() string {
	matches := githubIssueRe.FindStringSubmatch(g.url)
	if len(matches) > 1 {
		issueNumber := matches[1]
		slog.Debug("GitHub extractor: extracted issue number", "issueNumber", issueNumber)
		return issueNumber
	}

	slog.Debug("GitHub extractor: no issue number found in URL", "url", g.url)
	return ""
}

// createDescription creates a description from HTML content
// TypeScript original code:
//
//	private createDescription(content: string): string {
//		if (!content) return '';
//
//		const tempDiv = this.document.createElement('div');
//		tempDiv.innerHTML = content;
//		return tempDiv.textContent?.trim()
//			.slice(0, 140)
//			.replace(/\s+/g, ' ') || '';
//	}
func (g *GitHubExtractor) createDescription(content string) string {
	if content == "" {
		return ""
	}

	// Parse HTML and extract text content
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		slog.Warn("GitHub extractor: failed to parse HTML for description", "error", err)
		return ""
	}

	text := strings.TrimSpace(doc.Text())

	// Truncate to 140 characters to match TypeScript implementation
	if len(text) > 140 {
		text = text[:140]
	}

	// Replace multiple spaces with single space
	text = githubWhitespaceRe.ReplaceAllString(text, " ")

	slog.Debug("GitHub extractor: created description", "descriptionLength", len(text))
	return text
}
