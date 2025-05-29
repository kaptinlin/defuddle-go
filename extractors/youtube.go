package extractors

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// YouTubeExtractor handles YouTube content extraction
// Corresponding to TypeScript class YouTubeExtractor extends BaseExtractor
type YouTubeExtractor struct {
	*ExtractorBase
	videoElement *goquery.Selection
}

// NewYouTubeExtractor creates a new YouTube extractor
// TypeScript original code:
//
//	constructor(document: Document, url: string, schemaOrgData?: any) {
//		super(document, url, schemaOrgData);
//		this.videoElement = document.querySelector('video');
//	}
func NewYouTubeExtractor(document *goquery.Document, url string, schemaOrgData interface{}) *YouTubeExtractor {
	extractor := &YouTubeExtractor{
		ExtractorBase: NewExtractorBase(document, url, schemaOrgData),
	}

	// Find video element
	extractor.videoElement = document.Find("video").First()

	return extractor
}

// CanExtract checks if the extractor can extract content
// TypeScript original code:
//
//	canExtract(): boolean {
//		return true; // YouTube extractor can always extract
//	}
func (y *YouTubeExtractor) CanExtract() bool {
	return true // YouTube extractor can always extract
}

// GetName returns the name of the extractor
func (y *YouTubeExtractor) GetName() string {
	return "YouTubeExtractor"
}

// Extract extracts the YouTube content
// TypeScript original code:
//
//	extract(): ExtractorResult {
//		const videoData = this.getVideoData();
//		const description = this.getDescription(videoData);
//		const formattedDescription = this.formatDescription(description);
//		const videoId = this.getVideoId();
//
//		const contentHtml = `<iframe width="560" height="315" src="https://www.youtube.com/embed/${videoId}?si=_m0qv33lAuJFoGNh" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe><br>${formattedDescription}`;
//
//		return {
//			content: contentHtml,
//			contentHtml: contentHtml,
//			extractedContent: {
//				videoId: videoId,
//				author: this.getAuthor(videoData)
//			},
//			variables: {
//				title: this.getTitle(videoData),
//				author: this.getAuthor(videoData),
//				site: 'YouTube',
//				image: this.getThumbnail(videoData),
//				published: this.getPublished(videoData),
//				description: this.truncateDescription(description)
//			}
//		};
//	}
func (y *YouTubeExtractor) Extract() *ExtractorResult {
	videoData := y.getVideoData()
	description := y.getDescription(videoData)
	formattedDescription := y.formatDescription(description)
	videoID := y.getVideoID()

	// Create iframe content
	contentHTML := fmt.Sprintf(
		`<iframe width="560" height="315" src="https://www.youtube.com/embed/%s?si=_m0qv33lAuJFoGNh" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe><br>%s`,
		videoID,
		formattedDescription,
	)

	return &ExtractorResult{
		Content:     contentHTML,
		ContentHTML: contentHTML,
		ExtractedContent: map[string]interface{}{
			"videoId": videoID,
			"author":  y.getAuthor(videoData),
		},
		Variables: map[string]string{
			"title":       y.getTitle(videoData),
			"author":      y.getAuthor(videoData),
			"site":        "YouTube",
			"image":       y.getThumbnail(videoData),
			"published":   y.getPublished(videoData),
			"description": y.truncateDescription(description),
		},
	}
}

// getVideoData extracts video data from schema.org structured data
// TypeScript original code:
//
//	private getVideoData(): any {
//		if (!this.schemaOrgData) return {};
//
//		// Handle both single object and array of objects
//		if (Array.isArray(this.schemaOrgData)) {
//			return this.schemaOrgData.find(item => item['@type'] === 'VideoObject') || {};
//		}
//
//		return this.schemaOrgData['@type'] === 'VideoObject' ? this.schemaOrgData : {};
//	}
func (y *YouTubeExtractor) getVideoData() map[string]interface{} {
	if y.schemaOrgData == nil {
		return make(map[string]interface{})
	}

	// Handle both single object and array of objects
	switch data := y.schemaOrgData.(type) {
	case []interface{}:
		// Find VideoObject in array
		for _, item := range data {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemType, exists := itemMap["@type"]; exists && itemType == "VideoObject" {
					return itemMap
				}
			}
		}
	case map[string]interface{}:
		// Check if it's a VideoObject
		if itemType, exists := data["@type"]; exists && itemType == "VideoObject" {
			return data
		}
	}

	return make(map[string]interface{})
}

// getVideoID extracts the video ID from the URL
// TypeScript original code:
//
//	private getVideoId(): string {
//		const url = new URL(this.url);
//
//		// For youtube.com/watch?v=...
//		if (url.hostname.includes('youtube.com')) {
//			return url.searchParams.get('v') || '';
//		}
//
//		// For youtu.be/...
//		if (url.hostname.includes('youtu.be')) {
//			return url.pathname.slice(1);
//		}
//
//		return '';
//	}
func (y *YouTubeExtractor) getVideoID() string {
	parsedURL, err := url.Parse(y.url)
	if err != nil {
		return ""
	}

	// For youtube.com/watch?v=...
	if strings.Contains(parsedURL.Host, "youtube.com") {
		return parsedURL.Query().Get("v")
	}

	// For youtu.be/...
	if strings.Contains(parsedURL.Host, "youtu.be") {
		path := strings.TrimPrefix(parsedURL.Path, "/")
		return path
	}

	return ""
}

// getTitle gets the video title
// TypeScript original code:
//
//	private getTitle(videoData: any): string {
//		if (videoData.name) {
//			return videoData.name;
//		}
//
//		// Fallback to document title
//		let title = this.document.title;
//		// Remove " - YouTube" suffix if present
//		return title.replace(/ - YouTube$/, '');
//	}
func (y *YouTubeExtractor) getTitle(videoData map[string]interface{}) string {
	if name, exists := videoData["name"]; exists {
		if nameStr, ok := name.(string); ok {
			return nameStr
		}
	}

	// Fallback to document title
	title := y.document.Find("title").Text()
	// Remove " - YouTube" suffix if present
	title = strings.TrimSuffix(title, " - YouTube")
	return title
}

// getAuthor gets the video author/channel name
// TypeScript original code:
//
//	private getAuthor(videoData: any): string {
//		return videoData.author || '';
//	}
func (y *YouTubeExtractor) getAuthor(videoData map[string]interface{}) string {
	if author, exists := videoData["author"]; exists {
		if authorStr, ok := author.(string); ok {
			return authorStr
		}
	}
	return ""
}

// getDescription gets the video description
// TypeScript original code:
//
//	private getDescription(videoData: any): string {
//		if (videoData.description) {
//			return videoData.description;
//		}
//
//		// Fallback to description element in DOM
//		const descElement = this.document.querySelector('#description');
//		return descElement ? descElement.textContent || '' : '';
//	}
func (y *YouTubeExtractor) getDescription(videoData map[string]interface{}) string {
	if description, exists := videoData["description"]; exists {
		if descStr, ok := description.(string); ok {
			return descStr
		}
	}

	// Fallback to description element in DOM
	descElement := y.document.Find("#description").First()
	if descElement.Length() > 0 {
		return descElement.Text()
	}

	return ""
}

// getPublished gets the published date
// TypeScript original code:
//
//	private getPublished(videoData: any): string {
//		return videoData.uploadDate || '';
//	}
func (y *YouTubeExtractor) getPublished(videoData map[string]interface{}) string {
	if uploadDate, exists := videoData["uploadDate"]; exists {
		if dateStr, ok := uploadDate.(string); ok {
			return dateStr
		}
	}
	return ""
}

// getThumbnail gets the video thumbnail URL
// TypeScript original code:
//
//	private getThumbnail(videoData: any): string {
//		if (videoData.thumbnailUrl) {
//			return Array.isArray(videoData.thumbnailUrl) ? videoData.thumbnailUrl[0] : videoData.thumbnailUrl;
//		}
//
//		// Generate thumbnail URL from video ID if not found
//		const videoId = this.getVideoId();
//		return videoId ? `https://img.youtube.com/vi/${videoId}/maxresdefault.jpg` : '';
//	}
func (y *YouTubeExtractor) getThumbnail(videoData map[string]interface{}) string {
	if thumbnailUrl, exists := videoData["thumbnailUrl"]; exists {
		switch thumb := thumbnailUrl.(type) {
		case []interface{}:
			if len(thumb) > 0 {
				if thumbStr, ok := thumb[0].(string); ok {
					return thumbStr
				}
			}
		case string:
			return thumb
		}
	}

	// Generate thumbnail URL from video ID if not found
	videoID := y.getVideoID()
	if videoID != "" {
		return fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID)
	}

	return ""
}

// formatDescription formats the video description
// TypeScript original code:
//
//	private formatDescription(description: string): string {
//		if (!description) return '';
//
//		// Replace newlines with <br> tags
//		const formatted = description.replace(/\n/g, '<br>');
//		return `<p>${formatted}</p>`;
//	}
func (y *YouTubeExtractor) formatDescription(description string) string {
	if description == "" {
		return ""
	}

	// Replace newlines with <br> tags
	formatted := strings.ReplaceAll(description, "\n", "<br>")
	return fmt.Sprintf("<p>%s</p>", formatted)
}

// truncateDescription truncates description for metadata
// TypeScript original code:
//
//	private truncateDescription(description: string): string {
//		if (description.length <= 200) {
//			return description.trim();
//		}
//
//		// Find a good breaking point (end of sentence or word)
//		let truncated = description.substring(0, 200);
//		const lastSpace = truncated.lastIndexOf(' ');
//		if (lastSpace > 150) { // Only use word boundary if it's not too far back
//			truncated = truncated.substring(0, lastSpace);
//		}
//
//		return truncated.trim();
//	}
func (y *YouTubeExtractor) truncateDescription(description string) string {
	if len(description) <= 200 {
		return strings.TrimSpace(description)
	}

	// Find a good breaking point (end of sentence or word)
	truncated := description[:200]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 150 { // Only use word boundary if it's not too far back
		truncated = truncated[:lastSpace]
	}

	return strings.TrimSpace(truncated)
}
