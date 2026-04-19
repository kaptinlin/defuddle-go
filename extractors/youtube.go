package extractors

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// YouTubeExtractor handles YouTube content extraction
// TypeScript original code:
// import { BaseExtractor } from './_base';
// import { ExtractorResult } from '../types/extractors';
//
//	export class YoutubeExtractor extends BaseExtractor {
//		private videoElement: HTMLVideoElement | null;
//		protected override schemaOrgData: any;
//
//		constructor(document: Document, url: string, schemaOrgData?: any) {
//			super(document, url, schemaOrgData);
//			this.videoElement = document.querySelector('video');
//			this.schemaOrgData = schemaOrgData;
//		}
//
//		canExtract(): boolean {
//			return true;
//		}
//
//		extract(): ExtractorResult {
//			const videoData = this.getVideoData();
//			const description = videoData.description || '';
//			const formattedDescription = this.formatDescription(description);
//			const contentHtml = `<iframe width="560" height="315" src="https://www.youtube.com/embed/${this.getVideoId()}?si=_m0qv33lAuJFoGNh" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe><br>${formattedDescription}`;
//
//			return {
//				content: contentHtml,
//				contentHtml: contentHtml,
//				extractedContent: {
//					videoId: this.getVideoId(),
//					author: videoData.author || '',
//				},
//				variables: {
//					title: videoData.name || '',
//					author: videoData.author || '',
//					site: 'YouTube',
//					image: Array.isArray(videoData.thumbnailUrl) ? videoData.thumbnailUrl[0] || '' : '',
//					published: videoData.uploadDate,
//					description: description.slice(0, 200).trim(),
//				}
//			};
//		}
//	}
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
//		this.schemaOrgData = schemaOrgData;
//	}
func NewYouTubeExtractor(document *goquery.Document, url string, schemaOrgData any) *YouTubeExtractor {
	extractor := &YouTubeExtractor{
		ExtractorBase: NewExtractorBase(document, url, schemaOrgData),
	}

	// Find video element
	extractor.videoElement = document.Find("video").First()

	slog.Debug("YouTube extractor initialized",
		"hasVideoElement", extractor.videoElement.Length() > 0,
		"url", url,
		"hasSchemaOrgData", schemaOrgData != nil)

	return extractor
}

// CanExtract checks if the extractor can extract content
// TypeScript original code:
//
//	canExtract(): boolean {
//		return true; // YouTube extractor can always extract
//	}
func (y *YouTubeExtractor) CanExtract() bool {
	canExtract := true // YouTube extractor can always extract
	slog.Debug("YouTube extractor can extract check", "canExtract", canExtract)
	return canExtract
}

// Name returns the name of the extractor
func (y *YouTubeExtractor) Name() string {
	return "YouTubeExtractor"
}

// Extract extracts the YouTube content
// TypeScript original code:
//
//	extract(): ExtractorResult {
//		const videoData = this.getVideoData();
//		const description = videoData.description || '';
//		const formattedDescription = this.formatDescription(description);
//		const contentHtml = `<iframe width="560" height="315" src="https://www.youtube.com/embed/${this.getVideoId()}?si=_m0qv33lAuJFoGNh" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe><br>${formattedDescription}`;
//
//		return {
//			content: contentHtml,
//			contentHtml: contentHtml,
//			extractedContent: {
//				videoId: this.getVideoId(),
//				author: videoData.author || '',
//			},
//			variables: {
//				title: videoData.name || '',
//				author: videoData.author || '',
//				site: 'YouTube',
//				image: Array.isArray(videoData.thumbnailUrl) ? videoData.thumbnailUrl[0] || '' : '',
//				published: videoData.uploadDate,
//				description: description.slice(0, 200).trim(),
//			}
//		};
//	}
func (y *YouTubeExtractor) Extract() *ExtractorResult {
	slog.Debug("YouTube extractor starting extraction", "url", y.url)

	videoData := y.getVideoData()
	description := y.getDescription(videoData)
	formattedDescription := y.formatDescription(description)
	videoID := y.getVideoID()

	// Create iframe content - only if videoID is not empty
	var contentHTML string
	if videoID != "" {
		contentHTML = fmt.Sprintf(
			`<iframe width="560" height="315" src="https://www.youtube.com/embed/%s" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe><br>%s`,
			videoID,
			formattedDescription,
		)
	} else {
		// Fallback content when videoID is empty
		contentHTML = formattedDescription
	}

	title := y.getTitle(videoData)
	author := y.getAuthor(videoData)
	thumbnail := y.getThumbnail(videoData)
	published := y.getPublished(videoData)
	truncatedDescription := y.truncateDescription(description)

	slog.Debug("YouTube extraction completed",
		"videoId", videoID,
		"title", title,
		"author", author,
		"published", published,
		"descriptionLength", len(description))

	return &ExtractorResult{
		Content:     contentHTML,
		ContentHTML: contentHTML,
		ExtractedContent: map[string]any{
			"videoId": videoID,
			"author":  author,
		},
		Variables: map[string]string{
			"title":       title,
			"author":      author,
			"site":        "YouTube",
			"image":       thumbnail,
			"published":   published,
			"description": truncatedDescription,
		},
	}
}

// getVideoData extracts video data from schema.org structured data
// TypeScript original code:
//
//	private getVideoData(): any {
//		if (!this.schemaOrgData) return {};
//
//		const videoData = Array.isArray(this.schemaOrgData)
//			? this.schemaOrgData.find(item => item['@type'] === 'VideoObject')
//			: this.schemaOrgData['@type'] === 'VideoObject' ? this.schemaOrgData : null;
//
//		return videoData || {};
//	}
func (y *YouTubeExtractor) getVideoData() map[string]any {
	if y.schemaOrgData == nil {
		slog.Debug("YouTube extractor: no schema.org data available")
		return make(map[string]any)
	}

	// Handle both single object and array of objects
	switch data := y.schemaOrgData.(type) {
	case []any:
		// Find VideoObject in array
		for _, item := range data {
			if itemMap, ok := item.(map[string]any); ok {
				if itemType, exists := itemMap["@type"]; exists && itemType == "VideoObject" {
					slog.Debug("YouTube extractor: found VideoObject in array", "hasVideoData", true)
					return itemMap
				}
			}
		}
		slog.Debug("YouTube extractor: no VideoObject found in schema.org array")
	case map[string]any:
		// Check if it's a VideoObject
		if itemType, exists := data["@type"]; exists && itemType == "VideoObject" {
			slog.Debug("YouTube extractor: found VideoObject", "hasVideoData", true)
			return data
		}
		if itemType, exists := data["@type"]; exists {
			slog.Debug("YouTube extractor: schema.org data is not VideoObject", "type", itemType)
		}
	default:
		slog.Debug("YouTube extractor: unexpected schema.org data type", "type", fmt.Sprintf("%T", data))
	}

	return make(map[string]any)
}

// getVideoID extracts the video ID from the URL
// TypeScript original code:
//
//	private getVideoId(): string {
//		const urlParams = new URLSearchParams(new URL(this.url).search);
//		return urlParams.get('v') || '';
//	}
func (y *YouTubeExtractor) getVideoID() string {
	parsedURL, err := url.Parse(y.url)
	if err != nil {
		slog.Warn("YouTube extractor: failed to parse URL", "url", y.url, "error", err)
		return ""
	}

	// For youtube.com/watch?v=...
	if strings.Contains(parsedURL.Host, "youtube.com") {
		videoID := parsedURL.Query().Get("v")
		slog.Debug("YouTube extractor: extracted video ID from youtube.com", "videoId", videoID)
		return videoID
	}

	// For youtu.be/...
	if strings.Contains(parsedURL.Host, "youtu.be") {
		path := strings.TrimPrefix(parsedURL.Path, "/")
		slog.Debug("YouTube extractor: extracted video ID from youtu.be", "videoId", path)
		return path
	}

	slog.Debug("YouTube extractor: no video ID found in URL", "host", parsedURL.Host)
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
func (y *YouTubeExtractor) getTitle(videoData map[string]any) string {
	if name, exists := videoData["name"]; exists {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			slog.Debug("YouTube extractor: using title from schema.org", "title", nameStr)
			return nameStr
		}
	}

	// Fallback to document title
	title := y.document.Find("title").Text()
	// Remove " - YouTube" suffix if present
	title = strings.TrimSuffix(title, " - YouTube")

	slog.Debug("YouTube extractor: using title from document", "title", title)
	return title
}

// getAuthor gets the video author/channel name
// TypeScript original code:
//
//	private getAuthor(videoData: any): string {
//		return videoData.author || '';
//	}
func (y *YouTubeExtractor) getAuthor(videoData map[string]any) string {
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
func (y *YouTubeExtractor) getDescription(videoData map[string]any) string {
	if description, exists := videoData["description"]; exists {
		if descStr, ok := description.(string); ok && descStr != "" {
			slog.Debug("YouTube extractor: using description from schema.org", "descriptionLength", len(descStr))
			return descStr
		}
	}

	// Fallback to description element in DOM
	descElement := y.document.Find("#description").First()
	if descElement.Length() > 0 {
		description := descElement.Text()
		slog.Debug("YouTube extractor: using description from DOM", "descriptionLength", len(description))
		return description
	}

	slog.Debug("YouTube extractor: no description found")
	return ""
}

// getPublished gets the published date
// TypeScript original code:
//
//	private getPublished(videoData: any): string {
//		return videoData.uploadDate || '';
//	}
func (y *YouTubeExtractor) getPublished(videoData map[string]any) string {
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
func (y *YouTubeExtractor) getThumbnail(videoData map[string]any) string {
	if thumbnailURL, exists := videoData["thumbnailUrl"]; exists {
		switch thumb := thumbnailURL.(type) {
		case []any:
			if len(thumb) > 0 {
				if thumbStr, ok := thumb[0].(string); ok {
					slog.Debug("YouTube extractor: using thumbnail from schema.org array", "thumbnailUrl", thumbStr)
					return thumbStr
				}
			}
		case string:
			if thumb != "" {
				slog.Debug("YouTube extractor: using thumbnail from schema.org", "thumbnailUrl", thumb)
				return thumb
			}
		}
	}

	// Generate thumbnail URL from video ID if not found
	videoID := y.getVideoID()
	if videoID != "" {
		generatedThumbnail := fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID)
		slog.Debug("YouTube extractor: generated thumbnail URL", "thumbnailUrl", generatedThumbnail)
		return generatedThumbnail
	}

	slog.Debug("YouTube extractor: no thumbnail available")
	return ""
}

// formatDescription formats the video description
// TypeScript original code:
//
//	private formatDescription(description: string): string {
//		return `<p>${description.replace(/\n/g, '<br>')}</p>`;
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
