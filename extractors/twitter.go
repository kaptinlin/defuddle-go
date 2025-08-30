package extractors

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// TwitterExtractor handles Twitter/X content extraction
// TypeScript original code:
// import { BaseExtractor } from './_base';
// import { ExtractorResult } from '../types/extractors';
//
//	export class TwitterExtractor extends BaseExtractor {
//		private mainTweet: Element | null = null;
//		private threadTweets: Element[] = [];
//
//		constructor(document: Document, url: string) {
//			super(document, url);
//
//			// Get all tweets from the timeline
//			const timeline = document.querySelector('[aria-label="Timeline: Conversation"]');
//			if (!timeline) {
//				// Try to find a single tweet if not in timeline view
//				const singleTweet = document.querySelector('article[data-testid="tweet"]');
//				if (singleTweet) {
//					this.mainTweet = singleTweet;
//				}
//				return;
//			}
//
//			// Get all tweets before any section with "Discover more" or similar headings
//			const allTweets = Array.from(timeline.querySelectorAll('article[data-testid="tweet"]'));
//			const firstSection = timeline.querySelector('section, h2')?.parentElement;
//
//			if (firstSection) {
//				// Filter out tweets that appear after the first section
//				allTweets.forEach((tweet, index) => {
//					if (firstSection.compareDocumentPosition(tweet) & Node.DOCUMENT_POSITION_FOLLOWING) {
//						allTweets.splice(index);
//						return false;
//					}
//				});
//			}
//
//			// Set main tweet and thread tweets
//			this.mainTweet = allTweets[0] || null;
//			this.threadTweets = allTweets.slice(1);
//		}
//	}
type TwitterExtractor struct {
	*ExtractorBase
	mainTweet    *goquery.Selection
	threadTweets []*goquery.Selection
}

// UserInfo represents Twitter user information
type UserInfo struct {
	FullName  string
	Handle    string
	Date      string
	Permalink string
}

// NewTwitterExtractor creates a new Twitter extractor
// TypeScript original code:
//
//	constructor(document: Document, url: string) {
//		super(document, url);
//
//		// Get all tweets from the timeline
//		const timeline = document.querySelector('[aria-label="Timeline: Conversation"]');
//		if (!timeline) {
//			// Try to find a single tweet if not in timeline view
//			const singleTweet = document.querySelector('article[data-testid="tweet"]');
//			if (singleTweet) {
//				this.mainTweet = singleTweet;
//			}
//			return;
//		}
//
//		// Get all tweets before any section with "Discover more" or similar headings
//		const allTweets = Array.from(timeline.querySelectorAll('article[data-testid="tweet"]'));
//
//		// Set main tweet and thread tweets
//		if (allTweets.length > 0) {
//			this.mainTweet = allTweets[0];
//			this.threadTweets = allTweets.slice(1);
//		}
//	}
func NewTwitterExtractor(document *goquery.Document, url string, schemaOrgData interface{}) *TwitterExtractor {
	extractor := &TwitterExtractor{
		ExtractorBase: NewExtractorBase(document, url, schemaOrgData),
		threadTweets:  make([]*goquery.Selection, 0),
	}

	// Primary method: Get all tweets from the timeline
	timeline := document.Find(`[aria-label="Timeline: Conversation"]`).First()
	if timeline.Length() == 0 {
		// Fallback: Try alternative timeline selectors
		timelineSelectors := []string{
			`[aria-label*="timeline"]`,
			`[aria-label*="Timeline"]`,
			`main[role="main"]`,
			`section[role="region"]`,
		}

		for _, selector := range timelineSelectors {
			timeline = document.Find(selector).First()
			if timeline.Length() > 0 {
				break
			}
		}
	}

	var allTweets []*goquery.Selection

	if timeline.Length() > 0 {
		// Try to find tweets within the timeline
		timeline.Find(`article[data-testid="tweet"]`).Each(func(i int, s *goquery.Selection) {
			allTweets = append(allTweets, s)
		})
	}

	// Fallback: Try to find tweets anywhere in the document if timeline method fails
	if len(allTweets) == 0 {
		// Try alternative tweet selectors
		tweetSelectors := []string{
			`article[data-testid="tweet"]`,
			`[data-testid="tweet"]`,
			`.tweet`,
			`article[role="article"]`,
			`div[data-tweet-id]`,
		}

		for _, selector := range tweetSelectors {
			document.Find(selector).Each(func(i int, s *goquery.Selection) {
				allTweets = append(allTweets, s)
			})
			if len(allTweets) > 0 {
				break
			}
		}
	}

	// Set main tweet and thread tweets
	if len(allTweets) > 0 {
		extractor.mainTweet = allTweets[0]
		extractor.threadTweets = allTweets[1:]
	}

	return extractor
}

// CanExtract checks if the extractor can extract content
// TypeScript original code:
//
//	canExtract(): boolean {
//		return !!this.mainTweet;
//	}
func (t *TwitterExtractor) CanExtract() bool {
	return t.mainTweet != nil && t.mainTweet.Length() > 0
}

// GetName returns the name of the extractor
func (t *TwitterExtractor) GetName() string {
	return "TwitterExtractor"
}

// Extract extracts the Twitter content
// TypeScript original code:
//
//	extract(): ExtractorResult {
//		const mainContent = this.extractTweet(this.mainTweet);
//
//		const threadContents = this.threadTweets
//			.map(tweet => this.extractTweet(tweet))
//			.filter(content => content);
//		const threadContent = threadContents.join('\n<hr>\n');
//
//		let contentHtml = '<div class="tweet-thread">';
//		contentHtml += '<div class="main-tweet">' + mainContent + '</div>';
//
//		if (threadContent) {
//			contentHtml += '<hr><div class="thread-tweets">' + threadContent + '</div>';
//		}
//
//		contentHtml += '</div>';
//
//		const tweetId = this.getTweetId();
//		const tweetAuthor = this.getTweetAuthor();
//		const description = this.createDescription(this.mainTweet);
//
//		return {
//			content: contentHtml,
//			contentHtml: contentHtml,
//			extractedContent: {
//				tweetId: tweetId,
//				tweetAuthor: tweetAuthor
//			},
//			variables: {
//				title: `Thread by ${tweetAuthor}`,
//				author: tweetAuthor,
//				site: 'X (Twitter)',
//				description: description
//			}
//		};
//	}
func (t *TwitterExtractor) Extract() *ExtractorResult {
	mainContent := t.extractTweet(t.mainTweet)

	var threadContents []string
	for _, tweet := range t.threadTweets {
		content := t.extractTweet(tweet)
		if content != "" {
			threadContents = append(threadContents, content)
		}
	}
	threadContent := strings.Join(threadContents, "\n<hr>\n")

	var contentHTML strings.Builder
	contentHTML.WriteString(`<div class="tweet-thread">`)
	contentHTML.WriteString(`<div class="main-tweet">`)
	contentHTML.WriteString(mainContent)
	contentHTML.WriteString(`</div>`)

	if threadContent != "" {
		contentHTML.WriteString(`<hr><div class="thread-tweets">`)
		contentHTML.WriteString(threadContent)
		contentHTML.WriteString(`</div>`)
	}

	contentHTML.WriteString(`</div>`)

	tweetID := t.getTweetID()
	tweetAuthor := t.getTweetAuthor()
	description := t.createDescription(t.mainTweet)

	return &ExtractorResult{
		Content:     contentHTML.String(),
		ContentHTML: contentHTML.String(),
		ExtractedContent: map[string]interface{}{
			"tweetId":     tweetID,
			"tweetAuthor": tweetAuthor,
		},
		Variables: map[string]string{
			"title":       fmt.Sprintf("Thread by %s", tweetAuthor),
			"author":      tweetAuthor,
			"site":        "X (Twitter)",
			"description": description,
		},
	}
}

// formatTweetText formats tweet text content
// TypeScript original code:
//
//	private formatTweetText(text: string): string {
//		if (!text) return '';
//
//		// Parse HTML content to clean it
//		const parser = new DOMParser();
//		const doc = parser.parseFromString(text, 'text/html');
//
//		// Convert links to plain text with @ handles
//		doc.querySelectorAll('a').forEach(link => {
//			const handle = link.textContent?.trim() || '';
//			link.replaceWith(handle);
//		});
//
//		// Remove unnecessary spans and divs but keep their content
//		doc.querySelectorAll('span, div').forEach(element => {
//			const content = element.textContent || '';
//			element.replaceWith(content);
//		});
//
//		// Get cleaned text and split into paragraphs
//		const cleanText = doc.body.innerHTML;
//		const paragraphs = cleanText.split('\n').filter(p => p.trim());
//
//		return paragraphs.map(p => `<p>${p.trim()}</p>`).join('\n');
//	}
func (t *TwitterExtractor) formatTweetText(text string) string {
	if text == "" {
		return ""
	}

	// Add safety check for base document to mirror TypeScript fix
	if t.document == nil {
		return text
	}

	// Parse HTML content to clean it
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(text))
	if err != nil {
		return text
	}

	// Convert links to plain text with @ handles
	doc.Find("a").Each(func(i int, link *goquery.Selection) {
		handle := strings.TrimSpace(link.Text())
		link.ReplaceWithHtml(handle)
	})

	// Remove unnecessary spans and divs but keep their content
	doc.Find("span, div").Each(func(i int, element *goquery.Selection) {
		content := element.Text()
		element.ReplaceWithHtml(content)
	})

	// Get cleaned text and split into paragraphs
	cleanText, _ := doc.Html()
	paragraphs := strings.Split(cleanText, "\n")

	var formattedParagraphs []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p != "" {
			formattedParagraphs = append(formattedParagraphs, fmt.Sprintf("<p>%s</p>", p))
		}
	}

	return strings.Join(formattedParagraphs, "\n")
}

// extractTweet extracts content from a single tweet
// TypeScript original code:
//
//	private extractTweet(tweet: Element | null): string {
//		if (!tweet) return '';
//
//		// Get tweet text
//		const tweetText = tweet.querySelector('[data-testid="tweetText"]');
//		const tweetHtml = tweetText ? tweetText.innerHTML : '';
//		const formattedText = this.formatTweetText(tweetHtml);
//
//		// Get images
//		const images = this.extractImages(tweet);
//
//		// Get user info and date
//		const userInfo = this.extractUserInfo(tweet);
//
//		// Extract quoted tweet if present
//		const quotedTweet = tweet.querySelector('[aria-labelledby*="id__"]');
//		let quotedContent = '';
//		if (quotedTweet) {
//			const quotedUserName = quotedTweet.querySelector('[data-testid="User-Name"]');
//			if (quotedUserName) {
//				const quotedTweetContainer = quotedUserName.closest('[aria-labelledby*="id__"]');
//				if (quotedTweetContainer) {
//					quotedContent = this.extractTweet(quotedTweetContainer);
//				}
//			}
//		}
//
//		let result = '<div class="tweet">';
//		result += '<div class="tweet-header">';
//		result += `<span class="tweet-author"><strong>${userInfo.fullName}</strong> <span class="tweet-handle">${userInfo.handle}</span></span>`;
//
//		if (userInfo.date) {
//			result += ` <a href="${userInfo.permalink}" class="tweet-date">${userInfo.date}</a>`;
//		}
//
//		result += '</div>';
//
//		if (formattedText) {
//			result += `<div class="tweet-text">${formattedText}</div>`;
//		}
//
//		if (images.length > 0) {
//			result += '<div class="tweet-media">';
//			result += images.join('\n');
//			result += '</div>';
//		}
//
//		if (quotedContent) {
//			result += `<blockquote class="quoted-tweet">${quotedContent}</blockquote>`;
//		}
//
//		result += '</div>';
//		return result.trim();
//	}
func (t *TwitterExtractor) extractTweet(tweet *goquery.Selection) string {
	if tweet == nil || tweet.Length() == 0 {
		return ""
	}

	// Get tweet text
	tweetText := tweet.Find(`[data-testid="tweetText"]`).First()
	tweetHTML, _ := tweetText.Html()
	formattedText := t.formatTweetText(tweetHTML)

	// Get images
	images := t.extractImages(tweet)

	// Get user info and date
	userInfo := t.extractUserInfo(tweet)

	// Extract quoted tweet if present
	quotedTweet := tweet.Find(`[aria-labelledby*="id__"]`).First()
	var quotedContent string
	if quotedTweet.Length() > 0 {
		quotedUserName := quotedTweet.Find(`[data-testid="User-Name"]`).First()
		if quotedUserName.Length() > 0 {
			// Find the closest parent with aria-labelledby
			quotedTweetContainer := quotedUserName.Closest(`[aria-labelledby*="id__"]`)
			if quotedTweetContainer.Length() > 0 {
				quotedContent = t.extractTweet(quotedTweetContainer)
			}
		}
	}

	var result strings.Builder
	result.WriteString(`<div class="tweet">`)
	result.WriteString(`<div class="tweet-header">`)
	result.WriteString(fmt.Sprintf(`<span class="tweet-author"><strong>%s</strong> <span class="tweet-handle">%s</span></span>`,
		userInfo.FullName, userInfo.Handle))

	if userInfo.Date != "" {
		result.WriteString(fmt.Sprintf(` <a href="%s" class="tweet-date">%s</a>`, userInfo.Permalink, userInfo.Date))
	}

	result.WriteString(`</div>`)

	if formattedText != "" {
		result.WriteString(fmt.Sprintf(`<div class="tweet-text">%s</div>`, formattedText))
	}

	if len(images) > 0 {
		result.WriteString(`<div class="tweet-media">`)
		for _, img := range images {
			result.WriteString(img)
			result.WriteString("\n")
		}
		result.WriteString(`</div>`)
	}

	if quotedContent != "" {
		result.WriteString(fmt.Sprintf(`<blockquote class="quoted-tweet">%s</blockquote>`, quotedContent))
	}

	result.WriteString(`</div>`)
	return strings.TrimSpace(result.String())
}

// extractUserInfo extracts user information from a tweet
// TypeScript original code:
//
//	private extractUserInfo(tweet: Element): UserInfo {
//		const userInfo: UserInfo = {
//			fullName: '',
//			handle: '',
//			date: '',
//			permalink: ''
//		};
//
//		const nameElement = tweet.querySelector('[data-testid="User-Name"]');
//		if (!nameElement) return userInfo;
//
//		// Try to get name and handle from links first (main tweet structure)
//		const links = nameElement.querySelectorAll('a');
//		if (links.length >= 2) {
//			userInfo.fullName = links[0].textContent?.trim() || '';
//			userInfo.handle = links[1].textContent?.trim() || '';
//		}
//
//		// If links don't have the info, try to get from spans (quoted tweet structure)
//		if (!userInfo.fullName || !userInfo.handle) {
//			const fullNameSpan = nameElement.querySelector('span[style*="color: rgb(15, 20, 25)"] span');
//			if (fullNameSpan) {
//				userInfo.fullName = fullNameSpan.textContent?.trim() || '';
//			}
//
//			const handleSpan = nameElement.querySelector('span[style*="color: rgb(83, 100, 113)"]');
//			if (handleSpan) {
//				userInfo.handle = handleSpan.textContent?.trim() || '';
//			}
//		}
//
//		// Get timestamp information
//		const timestamp = tweet.querySelector('time');
//		if (timestamp) {
//			const datetime = timestamp.getAttribute('datetime');
//			if (datetime && datetime.length >= 10) {
//				userInfo.date = datetime.substring(0, 10); // YYYY-MM-DD format
//			}
//
//			// Get permalink from parent link
//			const timestampLink = timestamp.closest('a');
//			if (timestampLink) {
//				userInfo.permalink = timestampLink.getAttribute('href') || '';
//			}
//		}
//
//		return userInfo;
//	}
func (t *TwitterExtractor) extractUserInfo(tweet *goquery.Selection) UserInfo {
	userInfo := UserInfo{
		FullName:  "",
		Handle:    "",
		Date:      "",
		Permalink: "",
	}

	nameElement := tweet.Find(`[data-testid="User-Name"]`).First()
	if nameElement.Length() == 0 {
		return userInfo
	}

	// Try to get name and handle from links first (main tweet structure)
	links := nameElement.Find("a")
	if links.Length() >= 2 {
		userInfo.FullName = strings.TrimSpace(links.Eq(0).Text())
		userInfo.Handle = strings.TrimSpace(links.Eq(1).Text())
	}

	// If links don't have the info, try to get from spans (quoted tweet structure)
	if userInfo.FullName == "" || userInfo.Handle == "" {
		fullNameSpan := nameElement.Find(`span[style*="color: rgb(15, 20, 25)"] span`).First()
		if fullNameSpan.Length() > 0 {
			userInfo.FullName = strings.TrimSpace(fullNameSpan.Text())
		}

		handleSpan := nameElement.Find(`span[style*="color: rgb(83, 100, 113)"]`).First()
		if handleSpan.Length() > 0 {
			userInfo.Handle = strings.TrimSpace(handleSpan.Text())
		}
	}

	// Get timestamp information
	timestamp := tweet.Find("time").First()
	if timestamp.Length() > 0 {
		if datetime, exists := timestamp.Attr("datetime"); exists {
			// Parse datetime and extract date part
			if len(datetime) >= 10 {
				userInfo.Date = datetime[:10] // YYYY-MM-DD format
			}
		}

		// Get permalink from parent link
		timestampLink := timestamp.Closest("a")
		if timestampLink.Length() > 0 {
			if href, exists := timestampLink.Attr("href"); exists {
				userInfo.Permalink = href
			}
		}
	}

	return userInfo
}

// extractImages extracts images from a tweet
// TypeScript original code:
//
//	private extractImages(tweet: Element): string[] {
//		const images: string[] = [];
//
//		// Look for images in different containers
//		const imageSelectors = [
//			'[data-testid="tweetPhoto"]',
//			'[data-testid="tweet-image"]',
//			'img[src*="media"]'
//		];
//
//		// Skip images that are inside quoted tweets
//		const quotedTweet = tweet.querySelector('[aria-labelledby*="id__"]');
//		let quotedTweetContainer: Element | null = null;
//		if (quotedTweet) {
//			const quotedUserName = quotedTweet.querySelector('[data-testid="User-Name"]');
//			if (quotedUserName) {
//				quotedTweetContainer = quotedUserName.closest('[aria-labelledby*="id__"]');
//			}
//		}
//
//		for (const selector of imageSelectors) {
//			tweet.querySelectorAll(selector).forEach(img => {
//				// Skip if the image is inside a quoted tweet
//				if (quotedTweetContainer && quotedTweetContainer.contains(img)) {
//					return;
//				}
//
//				if (img.tagName === 'IMG') {
//					const src = img.getAttribute('src');
//					if (src) {
//						// Improve image quality
//						const highQualitySrc = src.replace(/&name=\w+$/, '&name=large');
//
//						const alt = img.getAttribute('alt') || '';
//						const cleanAlt = alt.trim().replace(/\s+/g, ' ');
//
//						images.push(`<img src="${highQualitySrc}" alt="${cleanAlt}" />`);
//					}
//				}
//			});
//		}
//
//		return images;
//	}
func (t *TwitterExtractor) extractImages(tweet *goquery.Selection) []string {
	var images []string

	// Look for images in different containers
	imageSelectors := []string{
		`[data-testid="tweetPhoto"]`,
		`[data-testid="tweet-image"]`,
		`img[src*="media"]`,
	}

	// Skip images that are inside quoted tweets
	quotedTweet := tweet.Find(`[aria-labelledby*="id__"]`).First()
	var quotedTweetContainer *goquery.Selection
	if quotedTweet.Length() > 0 {
		quotedUserName := quotedTweet.Find(`[data-testid="User-Name"]`).First()
		if quotedUserName.Length() > 0 {
			quotedTweetContainer = quotedUserName.Closest(`[aria-labelledby*="id__"]`)
		}
	}

	for _, selector := range imageSelectors {
		tweet.Find(selector).Each(func(i int, img *goquery.Selection) {
			// Skip if the image is inside a quoted tweet
			if quotedTweetContainer != nil && quotedTweetContainer.Length() > 0 {
				// Check if img is contained within quotedTweetContainer
				isInQuoted := false
				quotedTweetContainer.Find("*").Each(func(j int, el *goquery.Selection) {
					if el.Get(0) == img.Get(0) {
						isInQuoted = true
						return
					}
				})
				if isInQuoted {
					return
				}
			}

			if goquery.NodeName(img) == "img" {
				if src, exists := img.Attr("src"); exists {
					// Improve image quality
					highQualitySrc := regexp.MustCompile(`&name=\w+$`).ReplaceAllString(src, "&name=large")

					alt := img.AttrOr("alt", "")
					cleanAlt := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(alt, " "))

					images = append(images, fmt.Sprintf(`<img src="%s" alt="%s" />`, highQualitySrc, cleanAlt))
				}
			}
		})
	}

	return images
}

// getTweetID extracts the tweet ID from the URL
// TypeScript original code:
//
//	private getTweetId(): string {
//		const match = this.url.match(/status\/(\d+)/);
//		return match ? match[1] : '';
//	}
func (t *TwitterExtractor) getTweetID() string {
	re := regexp.MustCompile(`status/(\d+)`)
	matches := re.FindStringSubmatch(t.url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// getTweetAuthor extracts the author handle from the main tweet
// TypeScript original code:
//
//	private getTweetAuthor(): string {
//		if (!this.mainTweet) return '';
//
//		const nameElement = this.mainTweet.querySelector('[data-testid="User-Name"]');
//		if (!nameElement) return '';
//
//		const links = nameElement.querySelectorAll('a');
//		if (links.length >= 2) {
//			let handle = links[1].textContent?.trim() || '';
//			if (!handle.startsWith('@')) {
//				handle = '@' + handle;
//			}
//			return handle;
//		}
//
//		return '';
//	}
func (t *TwitterExtractor) getTweetAuthor() string {
	if t.mainTweet == nil {
		return ""
	}

	nameElement := t.mainTweet.Find(`[data-testid="User-Name"]`).First()
	if nameElement.Length() == 0 {
		return ""
	}

	links := nameElement.Find("a")
	if links.Length() >= 2 {
		handle := strings.TrimSpace(links.Eq(1).Text())
		if !strings.HasPrefix(handle, "@") {
			handle = "@" + handle
		}
		return handle
	}

	return ""
}

// createDescription creates a description from the main tweet
// TypeScript original code:
//
//	private createDescription(tweet: Element | null): string {
//		if (!tweet) return '';
//
//		const tweetText = tweet.querySelector('[data-testid="tweetText"]');
//		if (!tweetText) return '';
//
//		let text = tweetText.textContent?.trim() || '';
//		if (text.length > 140) {
//			text = text.substring(0, 140);
//		}
//
//		// Replace multiple spaces with single space
//		return text.replace(/\s+/g, ' ');
//	}
func (t *TwitterExtractor) createDescription(tweet *goquery.Selection) string {
	if tweet == nil || tweet.Length() == 0 {
		return ""
	}

	tweetText := tweet.Find(`[data-testid="tweetText"]`).First()
	if tweetText.Length() == 0 {
		return ""
	}

	text := strings.TrimSpace(tweetText.Text())
	if len(text) > 140 {
		text = text[:140]
	}

	// Replace multiple spaces with single space
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return text
}
