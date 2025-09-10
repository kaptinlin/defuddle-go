// Package metadata provides functionality for extracting and processing document metadata.
// It extracts metadata from HTML documents including title, description, author, and Schema.org data.
package metadata

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// MetaTag represents a meta tag item from HTML
// JavaScript original code:
//
//	export interface MetaTagItem {
//	  name?: string | null;
//	  property?: string | null;
//	  content: string | null;
//	}
type MetaTag struct {
	Name     *string `json:"name,omitempty"`
	Property *string `json:"property,omitempty"`
	Content  *string `json:"content"`
}

// Metadata represents extracted metadata from a document
// JavaScript original code:
//
//	export interface DefuddleMetadata {
//	  title: string;
//	  description: string;
//	  domain: string;
//	  favicon: string;
//	  image: string;
//	  parseTime: number;
//	  published: string;
//	  author: string;
//	  site: string;
//	  schemaOrgData: any;
//	  wordCount: number;
//	}
type Metadata struct {
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	Domain        string      `json:"domain"`
	Favicon       string      `json:"favicon"`
	Image         string      `json:"image"`
	ParseTime     int64       `json:"parseTime"`
	Published     string      `json:"published"`
	Author        string      `json:"author"`
	Site          string      `json:"site"`
	SchemaOrgData interface{} `json:"schemaOrgData"`
	WordCount     int         `json:"wordCount"`
}

// Extract extracts metadata from a document
// JavaScript original code:
//
//	static extract(doc: Document, schemaOrgData: any, metaTags: MetaTagItem[]): DefuddleMetadata {
//	  let domain = '';
//	  let url = '';
//
//	  try {
//	    // Try to get URL from document location
//	    url = doc.location?.href || '';
//
//	    // If no URL from location, try other sources
//	    if (!url) {
//	      url = this.getMetaContent(metaTags, "property", "og:url") ||
//	        this.getMetaContent(metaTags, "property", "twitter:url") ||
//	        this.getSchemaProperty(schemaOrgData, 'url') ||
//	        this.getSchemaProperty(schemaOrgData, 'mainEntityOfPage.url') ||
//	        this.getSchemaProperty(schemaOrgData, 'mainEntity.url') ||
//	        this.getSchemaProperty(schemaOrgData, 'WebSite.url') ||
//	        doc.querySelector('link[rel="canonical"]')?.getAttribute('href') || '';
//	    }
//
//	    if (url) {
//	      try {
//	        domain = new URL(url).hostname.replace(/^www\./, '');
//	      } catch (e) {
//	        console.warn('Failed to parse URL:', e);
//	      }
//	    }
//	  } catch (e) {
//	    // If URL parsing fails, try to get from base tag
//	    const baseTag = doc.querySelector('base[href]');
//	    if (baseTag) {
//	      try {
//	        url = baseTag.getAttribute('href') || '';
//	        domain = new URL(url).hostname.replace(/^www\./, '');
//	      } catch (e) {
//	        console.warn('Failed to parse base URL:', e);
//	      }
//	    }
//	  }
//
//	  return {
//	    title: this.getTitle(doc, schemaOrgData, metaTags),
//	    description: this.getDescription(doc, schemaOrgData, metaTags),
//	    domain,
//	    favicon: this.getFavicon(doc, url, metaTags),
//	    image: this.getImage(doc, schemaOrgData, metaTags),
//	    published: this.getPublished(doc, schemaOrgData, metaTags),
//	    author: this.getAuthor(doc, schemaOrgData, metaTags),
//	    site: this.getSite(doc, schemaOrgData, metaTags),
//	    schemaOrgData,
//	    wordCount: 0,
//	    parseTime: 0
//	  };
//	}
func Extract(doc *goquery.Document, schemaOrgData interface{}, metaTags []MetaTag, baseURL string) *Metadata {
	domain := ""
	documentURL := baseURL

	// If no base URL provided, try to extract from meta tags and canonical links
	if documentURL == "" {
		documentURL = getMetaContent(metaTags, "property", "og:url")
		if documentURL == "" {
			documentURL = getMetaContent(metaTags, "property", "twitter:url")
		}
		if documentURL == "" {
			documentURL = getSchemaProperty(schemaOrgData, "url")
		}
		if documentURL == "" {
			documentURL = getSchemaProperty(schemaOrgData, "mainEntityOfPage.url")
		}
		if documentURL == "" {
			documentURL = getSchemaProperty(schemaOrgData, "mainEntity.url")
		}
		if documentURL == "" {
			documentURL = getSchemaProperty(schemaOrgData, "WebSite.url")
		}
		if documentURL == "" {
			canonical := doc.Find(`link[rel="canonical"]`).First()
			if canonical.Length() > 0 {
				documentURL, _ = canonical.Attr("href")
			}
		}
	}

	// Extract domain from URL
	if documentURL != "" {
		if parsedURL, err := url.Parse(documentURL); err == nil {
			domain = strings.TrimPrefix(parsedURL.Hostname(), "www.")
		}
	}

	// If still no URL, try base tag
	if documentURL == "" {
		baseTag := doc.Find("base[href]").First()
		if baseTag.Length() > 0 {
			if href, exists := baseTag.Attr("href"); exists {
				documentURL = href
				if parsedURL, err := url.Parse(documentURL); err == nil {
					domain = strings.TrimPrefix(parsedURL.Hostname(), "www.")
				}
			}
		}
	}

	return &Metadata{
		Title:         getTitle(doc, schemaOrgData, metaTags),
		Description:   getDescription(doc, schemaOrgData, metaTags),
		Domain:        domain,
		Favicon:       getFavicon(doc, documentURL, metaTags),
		Image:         getImage(doc, schemaOrgData, metaTags),
		Published:     getPublished(doc, schemaOrgData, metaTags),
		Author:        getAuthor(doc, schemaOrgData, metaTags),
		Site:          getSite(doc, schemaOrgData, metaTags),
		SchemaOrgData: schemaOrgData,
		WordCount:     0,
		ParseTime:     0,
	}
}

// getAuthor extracts author information
// JavaScript original code:
//
//	private static getAuthor(doc: Document, schemaOrgData: any, metaTags: MetaTagItem[]): string {
//	  let authorsString: string | undefined;
//
//	  // Meta tags - typically expect a single string, possibly comma-separated
//	  authorsString = this.getMetaContent(metaTags, "name", "sailthru.author") ||
//	    this.getMetaContent(metaTags, "property", "author") ||
//	    this.getMetaContent(metaTags, "name", "author") ||
//	    this.getMetaContent(metaTags, "name", "byl") ||
//	    this.getMetaContent(metaTags, "name", "authorList");
//	  if (authorsString) return authorsString;
//
//	  // 2. Schema.org data - deduplicate if it's a list
//	  let schemaAuthors = this.getSchemaProperty(schemaOrgData, 'author.name') ||
//	    this.getSchemaProperty(schemaOrgData, 'author.[].name');
//
//	  if (schemaAuthors) {
//	    const parts = schemaAuthors.split(',')
//	      .map(part => part.trim().replace(/,$/, '').trim())
//	      .filter(Boolean);
//	    if (parts.length > 0) {
//	      let uniqueSchemaAuthors = [...new Set(parts)];
//	      if (uniqueSchemaAuthors.length > 10) {
//	        uniqueSchemaAuthors = uniqueSchemaAuthors.slice(0, 10);
//	      }
//	      return uniqueSchemaAuthors.join(', ');
//	    }
//	  }
//
//	  // 3. DOM elements
//	  const collectedAuthorsFromDOM: string[] = [];
//	  const addDomAuthor = (value: string | null | undefined) => {
//	    if (!value) return;
//	    value.split(',').forEach(namePart => {
//	      const cleanedName = namePart.trim().replace(/,$/, '').trim();
//	      const lowerCleanedName = cleanedName.toLowerCase();
//	      if (cleanedName && lowerCleanedName !== 'author' && lowerCleanedName !== 'authors') {
//	        collectedAuthorsFromDOM.push(cleanedName);
//	      }
//	    });
//	  };
//
//	  const domAuthorSelectors = [
//	    '[itemprop="author"]',
//	    '.author',
//	    '[href*="author"]',
//	    '.authors a',
//	  ];
//
//	  domAuthorSelectors.forEach(selector => {
//	    doc.querySelectorAll(selector).forEach(el => {
//	      addDomAuthor(el.textContent);
//	    });
//	  });
//
//	  if (collectedAuthorsFromDOM.length > 0) {
//	    let uniqueAuthors = [...new Set(collectedAuthorsFromDOM.map(name => name.trim()).filter(Boolean))];
//	    if (uniqueAuthors.length > 0) {
//	      if (uniqueAuthors.length > 10) {
//	        uniqueAuthors = uniqueAuthors.slice(0, 10);
//	      }
//	      return uniqueAuthors.join(', ');
//	    }
//	  }
//
//	  // 4. Fallback meta tags and schema properties (less direct for author names)
//	  authorsString = this.getMetaContent(metaTags, "name", "copyright") ||
//	    this.getSchemaProperty(schemaOrgData, 'copyrightHolder.name') ||
//	    this.getMetaContent(metaTags, "property", "og:site_name") ||
//	    this.getSchemaProperty(schemaOrgData, 'publisher.name') ||
//	    this.getSchemaProperty(schemaOrgData, 'sourceOrganization.name') ||
//	    this.getSchemaProperty(schemaOrgData, 'isPartOf.name') ||
//	    this.getMetaContent(metaTags, "name", "twitter:creator") ||
//	    this.getMetaContent(metaTags, "name", "application-name");
//	  if (authorsString) return authorsString;
//
//	  return '';
//	}
func getAuthor(doc *goquery.Document, schemaOrgData interface{}, metaTags []MetaTag) string {
	// Meta tags - typically expect a single string, possibly comma-separated
	authorsString := getMetaContent(metaTags, "name", "sailthru.author")
	if authorsString == "" {
		authorsString = getMetaContent(metaTags, "property", "author")
	}
	if authorsString == "" {
		authorsString = getMetaContent(metaTags, "name", "author")
	}
	if authorsString == "" {
		authorsString = getMetaContent(metaTags, "name", "byl")
	}
	if authorsString == "" {
		authorsString = getMetaContent(metaTags, "name", "authorList")
	}
	if authorsString != "" {
		return authorsString
	}

	// Schema.org data - deduplicate if it's a list
	schemaAuthors := getSchemaProperty(schemaOrgData, "author.name")
	if schemaAuthors == "" {
		schemaAuthors = getSchemaProperty(schemaOrgData, "author.[].name")
	}

	if schemaAuthors != "" {
		parts := strings.Split(schemaAuthors, ",")
		var cleanParts []string
		for _, part := range parts {
			cleaned := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(part), ","))
			if cleaned != "" {
				cleanParts = append(cleanParts, cleaned)
			}
		}
		if len(cleanParts) > 0 {
			// Remove duplicates
			uniqueAuthors := removeDuplicates(cleanParts)
			if len(uniqueAuthors) > 10 {
				uniqueAuthors = uniqueAuthors[:10]
			}
			return strings.Join(uniqueAuthors, ", ")
		}
	}

	// DOM elements
	var collectedAuthorsFromDOM []string
	addDomAuthor := func(value string) {
		if value == "" {
			return
		}
		parts := strings.Split(value, ",")
		for _, namePart := range parts {
			cleanedName := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(namePart), ","))
			lowerCleanedName := strings.ToLower(cleanedName)
			if cleanedName != "" && lowerCleanedName != "author" && lowerCleanedName != "authors" {
				collectedAuthorsFromDOM = append(collectedAuthorsFromDOM, cleanedName)
			}
		}
	}

	domAuthorSelectors := []string{
		`[itemprop="author"]`,
		".author",
		`[href*="author"]`,
		".authors a",
	}

	for _, selector := range domAuthorSelectors {
		doc.Find(selector).Each(func(_ int, el *goquery.Selection) {
			addDomAuthor(strings.TrimSpace(el.Text()))
		})
	}

	if len(collectedAuthorsFromDOM) > 0 {
		var cleanAuthors []string
		for _, name := range collectedAuthorsFromDOM {
			trimmed := strings.TrimSpace(name)
			if trimmed != "" {
				cleanAuthors = append(cleanAuthors, trimmed)
			}
		}
		uniqueAuthors := removeDuplicates(cleanAuthors)
		if len(uniqueAuthors) > 0 {
			if len(uniqueAuthors) > 10 {
				uniqueAuthors = uniqueAuthors[:10]
			}
			return strings.Join(uniqueAuthors, ", ")
		}
	}

	// Fallback meta tags and schema properties
	authorsString = getMetaContent(metaTags, "name", "copyright")
	if authorsString == "" {
		authorsString = getSchemaProperty(schemaOrgData, "copyrightHolder.name")
	}
	if authorsString == "" {
		authorsString = getMetaContent(metaTags, "property", "og:site_name")
	}
	if authorsString == "" {
		authorsString = getSchemaProperty(schemaOrgData, "publisher.name")
	}
	if authorsString == "" {
		authorsString = getSchemaProperty(schemaOrgData, "sourceOrganization.name")
	}
	if authorsString == "" {
		authorsString = getSchemaProperty(schemaOrgData, "isPartOf.name")
	}
	if authorsString == "" {
		authorsString = getMetaContent(metaTags, "name", "twitter:creator")
	}
	if authorsString == "" {
		authorsString = getMetaContent(metaTags, "name", "application-name")
	}
	if authorsString != "" {
		return authorsString
	}

	return ""
}

// getSite extracts site name
// JavaScript original code:
//
//	private static getSite(doc: Document, schemaOrgData: any, metaTags: MetaTagItem[]): string {
//	  return (
//	    this.getSchemaProperty(schemaOrgData, 'publisher.name') ||
//	    this.getMetaContent(metaTags, "property", "og:site_name") ||
//	    this.getSchemaProperty(schemaOrgData, 'WebSite.name') ||
//	    this.getSchemaProperty(schemaOrgData, 'sourceOrganization.name') ||
//	    this.getMetaContent(metaTags, "name", "copyright") ||
//	    this.getSchemaProperty(schemaOrgData, 'copyrightHolder.name') ||
//	    this.getSchemaProperty(schemaOrgData, 'isPartOf.name') ||
//	    this.getMetaContent(metaTags, "name", "application-name") ||
//	    this.getAuthor(doc, schemaOrgData, metaTags) ||
//	    ''
//	  );
//	}
func getSite(doc *goquery.Document, schemaOrgData interface{}, metaTags []MetaTag) string {
	site := getSchemaProperty(schemaOrgData, "publisher.name")
	if site == "" {
		site = getMetaContent(metaTags, "property", "og:site_name")
	}
	if site == "" {
		site = getSchemaProperty(schemaOrgData, "WebSite.name")
	}
	if site == "" {
		site = getSchemaProperty(schemaOrgData, "sourceOrganization.name")
	}
	if site == "" {
		site = getMetaContent(metaTags, "name", "copyright")
	}
	if site == "" {
		site = getSchemaProperty(schemaOrgData, "copyrightHolder.name")
	}
	if site == "" {
		site = getSchemaProperty(schemaOrgData, "isPartOf.name")
	}
	if site == "" {
		site = getMetaContent(metaTags, "name", "application-name")
	}
	if site == "" {
		site = getAuthor(doc, schemaOrgData, metaTags)
	}
	return site
}

// getTitle extracts title
// JavaScript original code:
//
//	private static getTitle(doc: Document, schemaOrgData: any, metaTags: MetaTagItem[]): string {
//	  const rawTitle = (
//	    this.getMetaContent(metaTags, "property", "og:title") ||
//	    this.getMetaContent(metaTags, "name", "twitter:title") ||
//	    this.getSchemaProperty(schemaOrgData, 'headline') ||
//	    this.getMetaContent(metaTags, "name", "title") ||
//	    this.getMetaContent(metaTags, "name", "sailthru.title") ||
//	    doc.querySelector('title')?.textContent?.trim() ||
//	    ''
//	  );
//
//	  return this.cleanTitle(rawTitle, this.getSite(doc, schemaOrgData, metaTags));
//	}
func getTitle(doc *goquery.Document, schemaOrgData interface{}, metaTags []MetaTag) string {
	rawTitle := getMetaContent(metaTags, "property", "og:title")
	if rawTitle == "" {
		rawTitle = getMetaContent(metaTags, "name", "twitter:title")
	}
	if rawTitle == "" {
		rawTitle = getSchemaProperty(schemaOrgData, "headline")
	}
	if rawTitle == "" {
		rawTitle = getMetaContent(metaTags, "name", "title")
	}
	if rawTitle == "" {
		rawTitle = getMetaContent(metaTags, "name", "sailthru.title")
	}
	if rawTitle == "" {
		titleEl := doc.Find("title").First()
		if titleEl.Length() > 0 {
			rawTitle = strings.TrimSpace(titleEl.Text())
		}
	}

	return cleanTitle(rawTitle, getSite(doc, schemaOrgData, metaTags))
}

// cleanTitle removes site name from title
// JavaScript original code:
//
//	private static cleanTitle(title: string, siteName: string): string {
//	  if (!title || !siteName) return title;
//
//	  // Remove site name if it exists
//	  const siteNameEscaped = siteName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
//	  const patterns = [
//	    `\\s*[\\|\\-–—]\\s*${siteNameEscaped}\\s*$`, // Title | Site Name
//	    `^\\s*${siteNameEscaped}\\s*[\\|\\-–—]\\s*`, // Site Name | Title
//	  ];
//
//	  for (const pattern of patterns) {
//	    const regex = new RegExp(pattern, 'i');
//	    if (regex.test(title)) {
//	      title = title.replace(regex, '');
//	      break;
//	    }
//	  }
//
//	  return title.trim();
//	}
func cleanTitle(title, siteName string) string {
	if title == "" || siteName == "" {
		return title
	}

	// Remove site name if it exists
	siteNameEscaped := regexp.QuoteMeta(siteName)
	patterns := []string{
		`\s*[\|\-–—]\s*` + siteNameEscaped + `\s*$`, // Title | Site Name
		`^\s*` + siteNameEscaped + `\s*[\|\-–—]\s*`, // Site Name | Title
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(`(?i)` + pattern)
		if regex.MatchString(title) {
			title = regex.ReplaceAllString(title, "")
			break
		}
	}

	return strings.TrimSpace(title)
}

// getDescription extracts description
// JavaScript original code:
//
//	private static getDescription(doc: Document, schemaOrgData: any, metaTags: MetaTagItem[]): string {
//	  return (
//	    this.getMetaContent(metaTags, "name", "description") ||
//	    this.getMetaContent(metaTags, "property", "description") ||
//	    this.getMetaContent(metaTags, "property", "og:description") ||
//	    this.getSchemaProperty(schemaOrgData, 'description') ||
//	    this.getMetaContent(metaTags, "name", "twitter:description") ||
//	    this.getMetaContent(metaTags, "name", "sailthru.description") ||
//	    ''
//	  );
//	}
func getDescription(_ *goquery.Document, schemaOrgData interface{}, metaTags []MetaTag) string {
	description := getMetaContent(metaTags, "name", "description")
	if description == "" {
		description = getMetaContent(metaTags, "property", "description")
	}
	if description == "" {
		description = getMetaContent(metaTags, "property", "og:description")
	}
	if description == "" {
		description = getSchemaProperty(schemaOrgData, "description")
	}
	if description == "" {
		description = getMetaContent(metaTags, "name", "twitter:description")
	}
	if description == "" {
		description = getMetaContent(metaTags, "name", "sailthru.description")
	}
	return description
}

// getImage extracts image URL
// JavaScript original code:
//
//	private static getImage(doc: Document, schemaOrgData: any, metaTags: MetaTagItem[]): string {
//	  return (
//	    this.getMetaContent(metaTags, "property", "og:image") ||
//	    this.getMetaContent(metaTags, "name", "twitter:image") ||
//	    this.getSchemaProperty(schemaOrgData, 'image.url') ||
//	    this.getSchemaProperty(schemaOrgData, 'image') ||
//	    this.getMetaContent(metaTags, "name", "sailthru.image.full") ||
//	    this.getMetaContent(metaTags, "name", "sailthru.image.thumb") ||
//	    ''
//	  );
//	}
func getImage(_ *goquery.Document, schemaOrgData interface{}, metaTags []MetaTag) string {
	image := getMetaContent(metaTags, "property", "og:image")
	if image == "" {
		image = getMetaContent(metaTags, "name", "twitter:image")
	}
	if image == "" {
		image = getSchemaProperty(schemaOrgData, "image.url")
	}
	if image == "" {
		image = getSchemaProperty(schemaOrgData, "image")
	}
	if image == "" {
		image = getMetaContent(metaTags, "name", "sailthru.image.full")
	}
	if image == "" {
		image = getMetaContent(metaTags, "name", "sailthru.image.thumb")
	}
	return image
}

// getFavicon extracts favicon URL
// JavaScript original code:
//
//	private static getFavicon(doc: Document, baseUrl: string, metaTags: MetaTagItem[]): string {
//	  const favicon = doc.querySelector('link[rel*="icon"]')?.getAttribute('href') ||
//	    this.getMetaContent(metaTags, "name", "msapplication-TileImage") ||
//	    '/favicon.ico';
//
//	  if (favicon.startsWith('http')) {
//	    return favicon;
//	  }
//
//	  if (baseUrl) {
//	    try {
//	      return new URL(favicon, baseUrl).href;
//	    } catch (e) {
//	      return favicon;
//	    }
//	  }
//
//	  return favicon;
//	}
func getFavicon(doc *goquery.Document, baseURL string, metaTags []MetaTag) string {
	favicon := ""
	iconLink := doc.Find(`link[rel*="icon"]`).First()
	if iconLink.Length() > 0 {
		href, exists := iconLink.Attr("href")
		if exists {
			favicon = href
		}
	}

	if favicon == "" {
		favicon = getMetaContent(metaTags, "name", "msapplication-TileImage")
	}

	if favicon == "" {
		favicon = "/favicon.ico"
	}

	if strings.HasPrefix(favicon, "http") {
		return favicon
	}

	if baseURL != "" {
		if parsedBase, err := url.Parse(baseURL); err == nil {
			if resolvedURL, err := parsedBase.Parse(favicon); err == nil {
				return resolvedURL.String()
			}
		}
	}

	return favicon
}

// getPublished extracts publication date
// JavaScript original code:
//
//	private static getPublished(doc: Document, schemaOrgData: any, metaTags: MetaTagItem[]): string {
//	  return (
//	    this.getSchemaProperty(schemaOrgData, 'datePublished') ||
//	    this.getMetaContent(metaTags, "property", "article:published_time") ||
//	    this.getMetaContent(metaTags, "name", "sailthru.date") ||
//	    this.getMetaContent(metaTags, "name", "date") ||
//	    this.getTimeElement(doc) ||
//	    ''
//	  );
//	}
func getPublished(doc *goquery.Document, schemaOrgData interface{}, metaTags []MetaTag) string {
	published := getSchemaProperty(schemaOrgData, "datePublished")
	if published == "" {
		published = getMetaContent(metaTags, "property", "article:published_time")
	}
	if published == "" {
		published = getMetaContent(metaTags, "name", "sailthru.date")
	}
	if published == "" {
		published = getMetaContent(metaTags, "name", "date")
	}
	if published == "" {
		published = getTimeElement(doc)
	}
	return published
}

// getMetaContent finds meta tag content by attribute and value
// JavaScript original code:
//
//	private static getMetaContent(metaTags: MetaTagItem[], attr: string, value: string): string {
//	  const tag = metaTags.find(tag => tag[attr] === value);
//	  return tag?.content || '';
//	}
func getMetaContent(metaTags []MetaTag, attr, value string) string {
	for _, tag := range metaTags {
		var tagValue *string
		switch attr {
		case "name":
			tagValue = tag.Name
		case "property":
			tagValue = tag.Property
		}
		if tagValue != nil && *tagValue == value && tag.Content != nil {
			return *tag.Content
		}
	}
	return ""
}

// getTimeElement extracts time from time elements
// JavaScript original code:
//
//	private static getTimeElement(doc: Document): string {
//	  const timeEl = doc.querySelector('time[datetime]');
//	  return timeEl?.getAttribute('datetime') || '';
//	}
func getTimeElement(doc *goquery.Document) string {
	timeEl := doc.Find("time[datetime]").First()
	if timeEl.Length() > 0 {
		datetime, exists := timeEl.Attr("datetime")
		if exists {
			return datetime
		}
	}
	return ""
}

// getSchemaProperty extracts property from schema.org data
// JavaScript original code:
//
//	private static getSchemaProperty(schemaOrgData: any, property: string, defaultValue: string = ''): string {
//	  if (!schemaOrgData) return defaultValue;
//
//	  const searchSchema = (data: any, props: string[], fullPath: string, isExactMatch: boolean = true): string[] => {
//	    if (typeof data === 'string') {
//	      return props.length === 0 ? [data] : [];
//	    }
//
//	    if (!data || typeof data !== 'object') {
//	      return [];
//	    }
//
//	    if (Array.isArray(data)) {
//	      const currentProp = props[0];
//	      if (/^\\[\\d+\\]$/.test(currentProp)) {
//	        const index = parseInt(currentProp.slice(1, -1));
//	        if (data[index]) {
//	          return searchSchema(data[index], props.slice(1), fullPath, isExactMatch);
//	        }
//	        return [];
//	      }
//
//	      if (props.length === 0 && data.every(item => typeof item === 'string' || typeof item === 'number')) {
//	        return data.map(String);
//	      }
//
//	      return data.flatMap(item => searchSchema(item, props, fullPath, isExactMatch));
//	    }
//
//	    const [currentProp, ...remainingProps] = props;
//
//	    if (!currentProp) {
//	      if (typeof data === 'string') return [data];
//	      if (typeof data === 'object' && data.name) {
//	        return [data.name];
//	      }
//	      return [];
//	    }
//
//	    if (data.hasOwnProperty(currentProp)) {
//	      return searchSchema(data[currentProp], remainingProps,
//	        fullPath ? `${fullPath}.${currentProp}` : currentProp, true);
//	    }
//
//	    if (!isExactMatch) {
//	      const nestedResults: string[] = [];
//	      for (const key in data) {
//	        if (typeof data[key] === 'object') {
//	          const results = searchSchema(data[key], props,
//	            fullPath ? `${fullPath}.${key}` : key, false);
//	          nestedResults.push(...results);
//	        }
//	      }
//	      if (nestedResults.length > 0) {
//	        return nestedResults;
//	      }
//	    }
//
//	    return [];
//	  };
//
//	  try {
//	    let results = searchSchema(schemaOrgData, property.split('.'), '', true);
//	    if (results.length === 0) {
//	      results = searchSchema(schemaOrgData, property.split('.'), '', false);
//	    }
//	    const result = results.length > 0 ? results.filter(Boolean).join(', ') : defaultValue;
//	    return result;
//	  } catch (error) {
//	    console.error(`Error in getSchemaProperty for ${property}:`, error);
//	    return defaultValue;
//	  }
//	}
func getSchemaProperty(schemaOrgData interface{}, property string) string {
	if schemaOrgData == nil {
		return ""
	}

	var searchSchema func(data interface{}, props []string, isExactMatch bool) []string
	searchSchema = func(data interface{}, props []string, isExactMatch bool) []string {
		// Handle string data
		if str, ok := data.(string); ok {
			if len(props) == 0 {
				return []string{str}
			}
			return []string{}
		}

		// Handle non-object data
		if data == nil {
			return []string{}
		}

		// Handle arrays
		if arr, ok := data.([]interface{}); ok {
			if len(props) > 0 {
				currentProp := props[0]
				// Handle array index notation like [0]
				if matched, _ := regexp.MatchString(`^\[\d+\]$`, currentProp); matched {
					indexStr := currentProp[1 : len(currentProp)-1]
					if index, err := strconv.Atoi(indexStr); err == nil && index < len(arr) {
						return searchSchema(arr[index], props[1:], isExactMatch)
					}
					return []string{}
				}
			}

			// If no props left and all items are strings/numbers, return them
			if len(props) == 0 {
				var results []string
				for _, item := range arr {
					if str, ok := item.(string); ok {
						results = append(results, str)
					} else if num, ok := item.(float64); ok {
						results = append(results, strconv.FormatFloat(num, 'f', -1, 64))
					}
				}
				if len(results) == len(arr) {
					return results
				}
			}

			// Search in all array items
			var allResults []string
			for _, item := range arr {
				results := searchSchema(item, props, isExactMatch)
				allResults = append(allResults, results...)
			}
			return allResults
		}

		// Handle maps/objects
		if obj, ok := data.(map[string]interface{}); ok {
			if len(props) == 0 {
				if str, ok := obj["name"].(string); ok {
					return []string{str}
				}
				if str, ok := data.(string); ok {
					return []string{str}
				}
				return []string{}
			}

			currentProp := props[0]
			remainingProps := props[1:]

			// Check if property exists
			if value, exists := obj[currentProp]; exists {
				return searchSchema(value, remainingProps, true)
			}

			// If not exact match, search nested objects
			if !isExactMatch {
				var nestedResults []string
				for _, value := range obj {
					if _, ok := value.(map[string]interface{}); ok {
						results := searchSchema(value, props, false)
						nestedResults = append(nestedResults, results...)
					}
				}
				return nestedResults
			}
		}

		return []string{}
	}

	props := strings.Split(property, ".")
	results := searchSchema(schemaOrgData, props, true)
	if len(results) == 0 {
		results = searchSchema(schemaOrgData, props, false)
	}

	var filteredResults []string
	for _, result := range results {
		if result != "" {
			filteredResults = append(filteredResults, result)
		}
	}

	return strings.Join(filteredResults, ", ")
}

// removeDuplicates removes duplicate strings from slice while preserving order
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
