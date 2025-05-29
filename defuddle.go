package defuddle

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kaptinlin/defuddle-go/extractors"
	"github.com/kaptinlin/defuddle-go/internal/constants"
	"github.com/kaptinlin/defuddle-go/internal/debug"
	"github.com/kaptinlin/defuddle-go/internal/markdown"
	"github.com/kaptinlin/defuddle-go/internal/metadata"
	"github.com/kaptinlin/defuddle-go/internal/scoring"
	"github.com/kaptinlin/defuddle-go/internal/standardize"
	"github.com/kaptinlin/requests"
	"github.com/piprate/json-gold/ld"
)

// Defuddle represents a document parser instance
type Defuddle struct {
	doc      *goquery.Document
	options  *Options
	debug    bool
	debugger *debug.Debugger
}

// NewDefuddle creates a new Defuddle instance from HTML content
// JavaScript original code:
//
//	constructor(document: Document, options: DefuddleOptions = {}) {
//	  this.doc = document;
//	  this.options = options;
//	}
func NewDefuddle(html string, options *Options) (*Defuddle, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	debugEnabled := false
	if options != nil {
		debugEnabled = options.Debug
	}
	debugger := debug.NewDebugger(debugEnabled)

	return &Defuddle{
		doc:      doc,
		options:  options,
		debug:    debugEnabled,
		debugger: debugger,
	}, nil
}

// Parse extracts the main content from the document
// JavaScript original code:
//
//	parse(): DefuddleResponse {
//	  // Try first with default settings
//	  const result = this.parseInternal();
//
//	  // If result has very little content, try again without clutter removal
//	  if (result.wordCount < 200) {
//	    console.log('Initial parse returned very little content, trying again');
//	    const retryResult = this.parseInternal({
//	      removePartialSelectors: false
//	    });
//
//	    // Return the result with more content
//	    if (retryResult.wordCount > result.wordCount) {
//	      this._log('Retry produced more content');
//	      return retryResult;
//	    }
//	  }
//
//	  return result;
//	}
func (d *Defuddle) Parse(ctx context.Context) (*Result, error) {
	// Try first with default settings
	result, err := d.parseInternal(ctx, nil)
	if err != nil {
		return nil, err
	}

	// If result has very little content, try again without clutter removal
	if result.WordCount < 200 {
		if d.debug {
			slog.Debug("Initial parse returned very little content, trying again")
		}

		retryOptions := &Options{}
		if d.options != nil {
			*retryOptions = *d.options
		}
		retryOptions.RemovePartialSelectors = false

		retryResult, retryErr := d.parseInternal(ctx, retryOptions)
		if retryErr != nil {
			return result, retryErr
		}

		// Return the result with more content
		if retryResult.WordCount > result.WordCount {
			if d.debug {
				slog.Debug("Retry produced more content", "originalWordCount", result.WordCount, "retryWordCount", retryResult.WordCount)
			}
			return retryResult, nil
		}
	}

	return result, nil
}

// ParseFromURL fetches content from a URL and parses it
// JavaScript original code:
// // This corresponds to Node.js usage: Defuddle(htmlOrDom, url?, options?)
func ParseFromURL(ctx context.Context, url string, options *Options) (*Result, error) {
	if options == nil {
		options = &Options{}
	}

	// Set URL in options if not already set
	if options.URL == "" {
		options.URL = url
	}

	// Create HTTP client and make request
	client := requests.URL(url)
	resp, err := client.Get("").
		UserAgent("Mozilla/5.0 (compatible; Defuddle/1.0; +https://github.com/kaptinlin/defuddle-go)").
		Timeout(30 * time.Second).
		Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", url, err)
	}
	defer func() {
		if closeErr := resp.Close(); closeErr != nil {
			slog.Warn("Failed to close response", "error", closeErr)
		}
	}()

	html := resp.String()

	// Create Defuddle instance and parse
	defuddle, err := NewDefuddle(html, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create Defuddle instance: %w", err)
	}

	return defuddle.Parse(ctx)
}

// parseInternal performs the actual parsing work
// JavaScript original code:
//
//	private parseInternal(overrideOptions: Partial<DefuddleOptions> = {}): DefuddleResponse {
//	  const startTime = Date.now();
//	  const options = {
//	    removeExactSelectors: true,
//	    removePartialSelectors: true,
//	    ...this.options,
//	    ...overrideOptions
//	  };
//
//	  // Extract schema.org data
//	  const schemaOrgData = this._extractSchemaOrgData(this.doc);
//
//	  // Collect meta tags
//	  const pageMetaTags: MetaTagItem[] = [];
//	  this.doc.querySelectorAll('meta').forEach(meta => {
//	    const name = meta.getAttribute('name');
//	    const property = meta.getAttribute('property');
//	    let content = meta.getAttribute('content');
//	    if (content) { // Only include tags that have content
//	      pageMetaTags.push({ name, property, content: this._decodeHTMLEntities(content) });
//	    }
//	  });
//
//	  // Extract metadata
//	  const metadata = MetadataExtractor.extract(this.doc, schemaOrgData, pageMetaTags);
//
//	  try {
//	    // Use site-specific extractor first, if there is one
//	    const url = options.url || this.doc.URL;
//	    const extractor = ExtractorRegistry.findExtractor(this.doc, url, schemaOrgData);
//	    if (extractor && extractor.canExtract()) {
//	      const extracted = extractor.extract();
//	      const endTime = Date.now();
//	      // console.log('Using extractor:', extractor.constructor.name.replace('Extractor', ''));
//	      return {
//	        content: extracted.contentHtml,
//	        title: extracted.variables?.title || metadata.title,
//	        description: metadata.description,
//	        domain: metadata.domain,
//	        favicon: metadata.favicon,
//	        image: metadata.image,
//	        published: extracted.variables?.published || metadata.published,
//	        author: extracted.variables?.author || metadata.author,
//	        site: metadata.site,
//	        schemaOrgData: metadata.schemaOrgData,
//	        wordCount: this.countWords(extracted.contentHtml),
//	        parseTime: Math.round(endTime - startTime),
//	        extractorType: extractor.constructor.name.replace('Extractor', '').toLowerCase(),
//	        metaTags: pageMetaTags
//	      };
//	    }
//
//	    // Continue if there is no extractor...
//
//	    // Evaluate mobile styles and sizes on original document
//	    const mobileStyles = this._evaluateMediaQueries(this.doc);
//
//	    // Find small images in original document, excluding lazy-loaded ones
//	    const smallImages = this.findSmallImages(this.doc);
//
//	    // Clone document
//	    const clone = this.doc.cloneNode(true) as Document;
//
//	    // Apply mobile styles to clone
//	    this.applyMobileStyles(clone, mobileStyles);
//
//	    // Find main content
//	    const mainContent = this.findMainContent(clone);
//	    if (!mainContent) {
//	      const endTime = Date.now();
//	      return {
//	        content: this.doc.body.innerHTML,
//	        ...metadata,
//	        wordCount: this.countWords(this.doc.body.innerHTML),
//	        parseTime: Math.round(endTime - startTime),
//	        metaTags: pageMetaTags
//	      };
//	    }
//
//	    // Remove small images
//	    this.removeSmallImages(clone, smallImages);
//
//	    // Remove hidden elements using computed styles
//	    this.removeHiddenElements(clone);
//
//	    // Remove non-content blocks by scoring
//	    // Tries to find lists, navigation based on text content and link density
//	    ContentScorer.scoreAndRemove(clone, this.debug);
//
//	    // Remove clutter using selectors
//	    if (options.removeExactSelectors || options.removePartialSelectors) {
//	      this.removeBySelector(clone, options.removeExactSelectors, options.removePartialSelectors);
//	    }
//
//	    // Normalize the main content
//	    standardizeContent(mainContent, metadata, this.doc, this.debug);
//
//	    const content = mainContent.outerHTML;
//	    const endTime = Date.now();
//
//	    return {
//	      content,
//	      ...metadata,
//	      wordCount: this.countWords(content),
//	      parseTime: Math.round(endTime - startTime),
//	      metaTags: pageMetaTags
//	    };
//	  } catch (error) {
//	    console.error('Defuddle', 'Error processing document:', error);
//	    const endTime = Date.now();
//	    return {
//	      content: this.doc.body.innerHTML,
//	      ...metadata,
//	      wordCount: this.countWords(this.doc.body.innerHTML),
//	      parseTime: Math.round(endTime - startTime),
//	      metaTags: pageMetaTags
//	    };
//	  }
//	}
func (d *Defuddle) parseInternal(ctx context.Context, overrideOptions *Options) (*Result, error) {
	startTime := time.Now()

	// Merge options with defaults
	options := d.mergeOptions(overrideOptions)

	// Extract schema.org data
	schemaOrgData := d.extractSchemaOrgData()

	// Collect meta tags
	metaTags := d.collectMetaTags()

	// Get base URL for metadata extraction
	baseURL := options.URL

	// Extract metadata
	extractedMetadata := metadata.Extract(d.doc, schemaOrgData, metaTags, baseURL)

	// Initialize debug tracking
	if d.debugger.IsEnabled() {
		d.debugger.StartTimer("total_parsing")
		d.debugger.SetStatistics(debug.Statistics{
			OriginalElementCount: d.doc.Find("*").Length(),
		})
	}

	// Try site-specific extractor first, if there is one
	url := options.URL
	extractor := extractors.FindExtractor(d.doc, url, schemaOrgData)
	if extractor != nil && extractor.CanExtract() {
		d.debugger.SetExtractorUsed(extractor.GetName())
		extracted := extractor.Extract()
		parseTime := time.Since(startTime).Milliseconds()

		// Get site name from extractor variables or use metadata
		siteName := extractedMetadata.Site
		if extracted.Variables != nil {
			if site, exists := extracted.Variables["site"]; exists {
				siteName = site
			}
		}

		// Create extractor type name (remove "Extractor" suffix)
		extractorType := strings.ToLower(strings.TrimSuffix(extractor.GetName(), "Extractor"))

		result := &Result{
			Metadata: Metadata{
				Title:         extractedMetadata.Title,
				Description:   extractedMetadata.Description,
				Domain:        extractedMetadata.Domain,
				Favicon:       extractedMetadata.Favicon,
				Image:         extractedMetadata.Image,
				ParseTime:     parseTime,
				Published:     extractedMetadata.Published,
				Author:        extractedMetadata.Author,
				Site:          siteName,
				SchemaOrgData: schemaOrgData,
				WordCount:     d.countWords(extracted.ContentHTML),
			},
			Content:       extracted.ContentHTML,
			ExtractorType: &extractorType,
			MetaTags:      metaTags,
		}

		// Override metadata from extractor if available
		if extracted.Variables != nil {
			if title, exists := extracted.Variables["title"]; exists && title != "" {
				result.Title = title
			}
			if author, exists := extracted.Variables["author"]; exists && author != "" {
				result.Author = author
			}
			if published, exists := extracted.Variables["published"]; exists && published != "" {
				result.Published = published
			}
			if description, exists := extracted.Variables["description"]; exists && description != "" {
				result.Description = description
			}
			if image, exists := extracted.Variables["image"]; exists && image != "" {
				result.Image = image
			}
		}

		// Add debug info if enabled
		if d.debugger.IsEnabled() {
			d.debugger.EndTimer("total_parsing")
			d.debugger.AddProcessingStep("extractor", "Used site-specific extractor: "+extractor.GetName(), 1, "")
			result.DebugInfo = d.debugger.GetDebugInfo()
		}

		return result, nil
	}

	// Evaluate mobile styles and sizes on original document
	mobileStyles := d.evaluateMediaQueries()

	// Find small images in original document, excluding lazy-loaded ones
	smallImages := d.findSmallImages(d.doc)

	// Work with the original document for processing
	// Note: goquery doesn't have true document cloning, so we work with the original
	workingDoc := d.doc

	// Apply mobile styles to document
	d.applyMobileStyles(workingDoc, mobileStyles)

	// Find main content
	mainContent := d.findMainContent(workingDoc)
	if mainContent == nil {
		// Fallback to body content
		content, _ := d.doc.Find("body").Html()
		wordCount := d.countWords(content)
		parseTime := time.Since(startTime).Milliseconds()

		result := &Result{
			Metadata: Metadata{
				Title:         extractedMetadata.Title,
				Description:   extractedMetadata.Description,
				Domain:        extractedMetadata.Domain,
				Favicon:       extractedMetadata.Favicon,
				Image:         extractedMetadata.Image,
				ParseTime:     parseTime,
				Published:     extractedMetadata.Published,
				Author:        extractedMetadata.Author,
				Site:          extractedMetadata.Site,
				SchemaOrgData: schemaOrgData,
				WordCount:     wordCount,
			},
			Content:  content,
			MetaTags: metaTags,
		}

		// Add debug info if enabled (fallback case)
		if d.debugger.IsEnabled() {
			d.debugger.EndTimer("total_parsing")
			d.debugger.AddProcessingStep("fallback", "Used fallback body content extraction", 1, "No main content found")
			result.DebugInfo = d.debugger.GetDebugInfo()
		}

		return result, nil
	}

	// Remove small images
	d.removeSmallImages(workingDoc, smallImages)

	// Remove hidden elements using computed styles
	d.removeHiddenElements(workingDoc)

	// Remove non-content blocks by scoring
	scoring.ScoreAndRemove(workingDoc, d.debug)

	// Remove clutter using selectors
	if options.RemoveExactSelectors || options.RemovePartialSelectors {
		d.removeBySelector(workingDoc, options.RemoveExactSelectors, options.RemovePartialSelectors)
	}

	// Normalize the main content
	standardize.StandardizeContent(mainContent, extractedMetadata, workingDoc, d.debug)

	content, _ := mainContent.Html()
	wordCount := d.countWords(content)
	parseTime := time.Since(startTime).Milliseconds()

	// Convert to Markdown if requested
	var contentMarkdown *string
	if options.Markdown || options.SeparateMarkdown {
		if markdownContent, err := d.convertHTMLToMarkdown(content); err == nil {
			contentMarkdown = &markdownContent
		} else if d.debug {
			slog.Debug("Failed to convert to Markdown", "error", err)
		}
	}

	result := &Result{
		Metadata: Metadata{
			Title:         extractedMetadata.Title,
			Description:   extractedMetadata.Description,
			Domain:        extractedMetadata.Domain,
			Favicon:       extractedMetadata.Favicon,
			Image:         extractedMetadata.Image,
			ParseTime:     parseTime,
			Published:     extractedMetadata.Published,
			Author:        extractedMetadata.Author,
			Site:          extractedMetadata.Site,
			SchemaOrgData: schemaOrgData,
			WordCount:     wordCount,
		},
		Content:         content,
		ContentMarkdown: contentMarkdown,
		MetaTags:        metaTags,
	}

	// Add debug info if enabled
	if d.debugger.IsEnabled() {
		d.debugger.EndTimer("total_parsing")
		d.debugger.AddProcessingStep("standard_parsing", "Used standard content extraction algorithm", 1, "")

		// Update final statistics
		finalStats := debug.Statistics{
			OriginalElementCount: d.doc.Find("*").Length(),
			FinalElementCount:    workingDoc.Find("*").Length(),
			WordCount:            wordCount,
			CharacterCount:       len(content),
			ImageCount:           workingDoc.Find("img").Length(),
			LinkCount:            workingDoc.Find("a").Length(),
		}
		finalStats.RemovedElementCount = finalStats.OriginalElementCount - finalStats.FinalElementCount
		d.debugger.SetStatistics(finalStats)

		result.DebugInfo = d.debugger.GetDebugInfo()
	}

	return result, nil
}

// findMainContent finds the main content element
// JavaScript original code:
//
//	private findMainContent(doc: Document): Element | null {
//	  // Try entry point elements first
//	  for (const selector of ENTRY_POINT_ELEMENTS) {
//	    const element = doc.querySelector(selector);
//	    if (element) {
//	      return element;
//	    }
//	  }
//
//	  // Try table-based content
//	  const tableContent = this.findTableBasedContent(doc);
//	  if (tableContent) {
//	    return tableContent;
//	  }
//
//	  // Try content scoring
//	  const scoredContent = this.findContentByScoring(doc);
//	  if (scoredContent) {
//	    return scoredContent;
//	  }
//
//	  return null;
//	}
func (d *Defuddle) findMainContent(doc *goquery.Document) *goquery.Selection {
	// Try entry point elements first
	entryPoints := constants.GetEntryPointElements()
	for _, selector := range entryPoints {
		element := doc.Find(selector).First()
		if element.Length() > 0 {
			if d.debug {
				slog.Debug("Found main content using entry point", "selector", selector)
			}
			return element
		}
	}

	// Try table-based content
	tableContent := d.findTableBasedContent(doc)
	if tableContent != nil {
		if d.debug {
			slog.Debug("Found main content using table-based detection")
		}
		return tableContent
	}

	// Try content scoring
	scoredContent := d.findContentByScoring(doc)
	if scoredContent != nil {
		if d.debug {
			slog.Debug("Found main content using scoring")
		}
		return scoredContent
	}

	return nil
}

// findTableBasedContent finds content in table-based layouts
// JavaScript original code:
//
//	private findTableBasedContent(doc: Document): Element | null {
//	  const tables = doc.querySelectorAll('table');
//	  let bestTable: Element | null = null;
//	  let bestScore = 0;
//
//	  tables.forEach(table => {
//	    const cells = table.querySelectorAll('td');
//	    cells.forEach(cell => {
//	      const score = ContentScorer.scoreElement(cell);
//	      if (score > bestScore) {
//	        bestScore = score;
//	        bestTable = cell;
//	      }
//	    });
//	  });
//
//	  return bestScore > 50 ? bestTable : null;
//	}
func (d *Defuddle) findTableBasedContent(doc *goquery.Document) *goquery.Selection {
	var bestElement *goquery.Selection
	bestScore := 0.0

	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("td").Each(func(j int, cell *goquery.Selection) {
			score := scoring.ScoreElement(cell)
			if score > bestScore {
				bestScore = score
				bestElement = cell
			}
		})
	})

	if bestScore > 50 {
		return bestElement
	}
	return nil
}

// findContentByScoring finds content using scoring algorithm
// JavaScript original code:
//
//	private findContentByScoring(doc: Document): Element | null {
//	  const candidates = doc.querySelectorAll('div, section, article, main');
//	  const elements = Array.from(candidates);
//	  return ContentScorer.findBestElement(elements, 50);
//	}
func (d *Defuddle) findContentByScoring(doc *goquery.Document) *goquery.Selection {
	var candidates []*goquery.Selection
	doc.Find("div, section, article, main").Each(func(i int, s *goquery.Selection) {
		candidates = append(candidates, s)
	})

	return scoring.FindBestElement(candidates, 50)
}

// removeBySelector removes elements by exact and partial selectors
// JavaScript original code:
//
//	private removeBySelector(doc: Document, removeExact: boolean = true, removePartial: boolean = true) {
//	  if (removeExact) {
//	    EXACT_SELECTORS.forEach(selector => {
//	      doc.querySelectorAll(selector).forEach(el => el.remove());
//	    });
//	  }
//
//	  if (removePartial) {
//	    const testAttributes = TEST_ATTRIBUTES;
//	    const partialSelectors = PARTIAL_SELECTORS;
//
//	    doc.querySelectorAll('*').forEach(element => {
//	      testAttributes.forEach(attr => {
//	        const value = element.getAttribute(attr);
//	        if (value) {
//	          const lowerValue = value.toLowerCase();
//	          partialSelectors.forEach(pattern => {
//	            if (lowerValue.includes(pattern.toLowerCase())) {
//	              element.remove();
//	            }
//	          });
//	        }
//	      });
//	    });
//	  }
//	}
func (d *Defuddle) removeBySelector(doc *goquery.Document, removeExact, removePartial bool) {
	if removeExact {
		exactSelectors := constants.GetExactSelectors()
		for _, selector := range exactSelectors {
			doc.Find(selector).Remove()
		}
	}

	if removePartial {
		testAttributes := constants.GetTestAttributes()
		partialSelectors := constants.GetPartialSelectors()

		doc.Find("*").Each(func(i int, element *goquery.Selection) {
			for _, attr := range testAttributes {
				value, exists := element.Attr(attr)
				if exists && value != "" {
					lowerValue := strings.ToLower(value)
					for _, pattern := range partialSelectors {
						if strings.Contains(lowerValue, strings.ToLower(pattern)) {
							element.Remove()
							return
						}
					}
				}
			}
		})
	}
}

// mergeOptions merges override options with instance options and defaults
// JavaScript original code:
//
//	const options = {
//	  removeExactSelectors: true,
//	  removePartialSelectors: true,
//	  ...this.options,
//	  ...overrideOptions
//	};
func (d *Defuddle) mergeOptions(overrideOptions *Options) *Options {
	// Start with defaults (exactly like TypeScript version)
	options := &Options{
		RemoveExactSelectors:   true,
		RemovePartialSelectors: true,
	}

	// Apply instance options if they exist (...this.options)
	if d.options != nil {
		// Copy all values from instance options, including false values
		options.Debug = d.options.Debug
		if d.options.URL != "" {
			options.URL = d.options.URL
		}
		options.Markdown = d.options.Markdown
		options.SeparateMarkdown = d.options.SeparateMarkdown

		// For boolean options that can override defaults, always apply them
		options.RemoveExactSelectors = d.options.RemoveExactSelectors
		options.RemovePartialSelectors = d.options.RemovePartialSelectors
		options.ProcessCode = d.options.ProcessCode
		options.ProcessImages = d.options.ProcessImages
		options.ProcessHeadings = d.options.ProcessHeadings
		options.ProcessMath = d.options.ProcessMath
		options.ProcessFootnotes = d.options.ProcessFootnotes
		options.ProcessRoles = d.options.ProcessRoles

		// Copy pointer fields
		if d.options.CodeOptions != nil {
			options.CodeOptions = d.options.CodeOptions
		}
		if d.options.ImageOptions != nil {
			options.ImageOptions = d.options.ImageOptions
		}
		if d.options.HeadingOptions != nil {
			options.HeadingOptions = d.options.HeadingOptions
		}
		if d.options.MathOptions != nil {
			options.MathOptions = d.options.MathOptions
		}
		if d.options.FootnoteOptions != nil {
			options.FootnoteOptions = d.options.FootnoteOptions
		}
		if d.options.RoleOptions != nil {
			options.RoleOptions = d.options.RoleOptions
		}
	}

	// Apply override options if they exist (...overrideOptions)
	if overrideOptions != nil {
		// Copy all values from override options, including false values
		options.Debug = overrideOptions.Debug
		if overrideOptions.URL != "" {
			options.URL = overrideOptions.URL
		}
		options.Markdown = overrideOptions.Markdown
		options.SeparateMarkdown = overrideOptions.SeparateMarkdown

		// Override boolean options (these will override any previous values)
		options.RemoveExactSelectors = overrideOptions.RemoveExactSelectors
		options.RemovePartialSelectors = overrideOptions.RemovePartialSelectors
		options.ProcessCode = overrideOptions.ProcessCode
		options.ProcessImages = overrideOptions.ProcessImages
		options.ProcessHeadings = overrideOptions.ProcessHeadings
		options.ProcessMath = overrideOptions.ProcessMath
		options.ProcessFootnotes = overrideOptions.ProcessFootnotes
		options.ProcessRoles = overrideOptions.ProcessRoles

		// Copy pointer fields
		if overrideOptions.CodeOptions != nil {
			options.CodeOptions = overrideOptions.CodeOptions
		}
		if overrideOptions.ImageOptions != nil {
			options.ImageOptions = overrideOptions.ImageOptions
		}
		if overrideOptions.HeadingOptions != nil {
			options.HeadingOptions = overrideOptions.HeadingOptions
		}
		if overrideOptions.MathOptions != nil {
			options.MathOptions = overrideOptions.MathOptions
		}
		if overrideOptions.FootnoteOptions != nil {
			options.FootnoteOptions = overrideOptions.FootnoteOptions
		}
		if overrideOptions.RoleOptions != nil {
			options.RoleOptions = overrideOptions.RoleOptions
		}
	}

	return options
}

// countWords counts words in HTML content
// JavaScript original code:
//
//	private countWords(content: string): number {
//	  // Create a temporary div to parse HTML content
//	  const tempDiv = this.doc.createElement('div');
//	  tempDiv.innerHTML = content;
//
//	  // Get text content, removing extra whitespace
//	  const text = tempDiv.textContent || '';
//	  const words = text
//	    .trim()
//	    .replace(/\s+/g, ' ') // Replace multiple spaces with single space
//	    .split(' ')
//	    .filter(word => word.length > 0); // Filter out empty strings
//
//	  return words.length;
//	}
func (d *Defuddle) countWords(content string) int {
	// Parse HTML content to extract text
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		// Fallback: count words in raw content
		text := strings.TrimSpace(content)
		if text == "" {
			return 0
		}
		words := strings.Fields(text)
		return len(words)
	}

	// Get text content, removing extra whitespace
	text := strings.TrimSpace(doc.Text())
	if text == "" {
		return 0
	}

	// Replace multiple spaces with single space and split
	text = strings.Join(strings.Fields(text), " ")
	words := strings.Split(text, " ")

	// Filter out empty strings
	count := 0
	for _, word := range words {
		if len(strings.TrimSpace(word)) > 0 {
			count++
		}
	}

	return count
}

// extractSchemaOrgData extracts and processes schema.org structured data using JSON-LD processor
// JavaScript original code:
//
//	private _extractSchemaOrgData(document: Document) {
//	  const schemaItems = [];
//	  const scripts = document.querySelectorAll('script[type="application/ld+json"]');
//
//	  scripts.forEach(script => {
//	    try {
//	      const jsonData = JSON.parse(script.textContent);
//	      if (jsonData['@graph']) {
//	        schemaItems.push(...jsonData['@graph']);
//	      } else {
//	        schemaItems.push(jsonData);
//	      }
//	    } catch (e) {
//	      console.warn('Failed to parse schema.org data:', e);
//	    }
//	  });
//
//	  return schemaItems;
//	}
func (d *Defuddle) extractSchemaOrgData() interface{} {
	processor := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.ProcessingMode = ld.JsonLd_1_1

	var allSchemaItems []interface{}

	if d.debugger.IsEnabled() {
		d.debugger.StartTimer("schema_extraction")
	}

	d.doc.Find(`script[type="application/ld+json"]`).Each(func(i int, script *goquery.Selection) {
		jsonContent := strings.TrimSpace(script.Text())
		if jsonContent == "" {
			return
		}

		// Clean and validate JSON-LD content
		cleanedContent := d.cleanJSONLDContent(jsonContent)
		if cleanedContent == "" {
			if d.debug {
				slog.Debug("Empty JSON-LD content after cleaning", "index", i)
			}
			return
		}

		// Parse and process JSON-LD using json-gold
		processedData, err := d.processSchemaOrgData(processor, options, cleanedContent)
		if err != nil {
			if d.debug {
				slog.Debug("Failed to process schema.org JSON-LD",
					"error", err,
					"index", i,
					"content_preview", cleanedContent[:min(len(cleanedContent), 100)])
			}
			return
		}

		// Extract items from processed data
		items := d.extractSchemaItems(processedData)
		allSchemaItems = append(allSchemaItems, items...)
	})

	if d.debugger.IsEnabled() {
		d.debugger.EndTimer("schema_extraction")
		d.debugger.AddProcessingStep("schema_org_extraction",
			fmt.Sprintf("Extracted %d schema.org items", len(allSchemaItems)),
			len(allSchemaItems), "")
	}

	if d.debug {
		slog.Debug("Schema.org data extraction completed",
			"total_items", len(allSchemaItems),
			"unique_types", d.countSchemaTypes(allSchemaItems))
	}

	return allSchemaItems
}

// cleanJSONLDContent cleans and normalizes JSON-LD content
// JavaScript original code:
//
//	// Remove comments, CDATA, and other non-JSON content
func (d *Defuddle) cleanJSONLDContent(content string) string {
	// Remove HTML comments
	commentRegex := regexp.MustCompile(`<!--[\s\S]*?-->`)
	content = commentRegex.ReplaceAllString(content, "")

	// Remove JavaScript-style comments
	jsCommentRegex := regexp.MustCompile(`/\*[\s\S]*?\*/|^\s*//.*$`)
	content = jsCommentRegex.ReplaceAllString(content, "")

	// Handle CDATA sections
	cdataRegex := regexp.MustCompile(`^\s*<!\[CDATA\[([\s\S]*?)\]\]>\s*$`)
	if matches := cdataRegex.FindStringSubmatch(content); len(matches) > 1 {
		content = matches[1]
	}

	// Remove comment markers that might be left
	commentMarkerRegex := regexp.MustCompile(`^\s*(\*/|/\*)\s*|\s*(\*/|/\*)\s*$`)
	content = commentMarkerRegex.ReplaceAllString(content, "")

	// Remove leading/trailing whitespace
	content = strings.TrimSpace(content)

	// Basic JSON validation - check if it starts and ends correctly
	isValidJSON := (strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}")) ||
		(strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]"))

	if content != "" && !isValidJSON {
		if d.debug {
			slog.Debug("Invalid JSON-LD format detected", "content_preview", content[:min(len(content), 50)])
		}
		return ""
	}

	return content
}

// processSchemaOrgData processes JSON-LD data using json-gold processor
// JavaScript original code:
//
//	// Standard JSON-LD processing with context expansion and validation
func (d *Defuddle) processSchemaOrgData(processor *ld.JsonLdProcessor, options *ld.JsonLdOptions, jsonContent string) (interface{}, error) {
	// Parse raw JSON first
	var rawData interface{}
	if err := json.Unmarshal([]byte(jsonContent), &rawData); err != nil {
		return nil, fmt.Errorf("invalid JSON syntax: %w", err)
	}

	// Expand JSON-LD to resolve contexts and normalize structure
	expanded, err := processor.Expand(rawData, options)
	if err != nil {
		return nil, fmt.Errorf("JSON-LD expansion failed: %w", err)
	}

	// If expansion succeeded, try to compact with schema.org context for cleaner output
	if len(expanded) > 0 {
		schemaContext := map[string]interface{}{
			"@context": "https://schema.org/",
		}

		compacted, err := processor.Compact(expanded, schemaContext, options)
		if err != nil {
			// If compaction fails, use expanded data
			if d.debug {
				slog.Debug("Schema.org compaction failed, using expanded data", "error", err)
			}
			return expanded, nil
		}

		return compacted, nil
	}

	return expanded, nil
}

// extractSchemaItems extracts individual schema items from processed JSON-LD data
// JavaScript original code:
//
//	// Handle both single items and @graph arrays
func (d *Defuddle) extractSchemaItems(data interface{}) []interface{} {
	var items []interface{}

	switch typedData := data.(type) {
	case map[string]interface{}:
		// Check for @graph property (common in schema.org JSON-LD)
		if graph, exists := typedData["@graph"]; exists {
			if graphArray, ok := graph.([]interface{}); ok {
				items = append(items, graphArray...)
			} else {
				items = append(items, graph)
			}
		} else {
			// Single item
			items = append(items, typedData)
		}

	case []interface{}:
		// Array of items (from JSON-LD expansion)
		items = append(items, typedData...)

	default:
		// Single item of unknown type
		items = append(items, data)
	}

	// Filter and validate schema items
	var validItems []interface{}
	for _, item := range items {
		if d.isValidSchemaItem(item) {
			validItems = append(validItems, item)
		}
	}

	return validItems
}

// isValidSchemaItem validates if an item is a valid schema.org item
// JavaScript original code:
//
//	// Check for @type or other schema.org indicators
func (d *Defuddle) isValidSchemaItem(item interface{}) bool {
	itemMap, ok := item.(map[string]interface{})
	if !ok {
		return false
	}

	// Check for @type or type property (required for schema.org items)
	var itemType interface{}
	var exists bool
	if itemType, exists = itemMap["@type"]; !exists {
		itemType, exists = itemMap["type"]
	}

	if exists {
		switch typedValue := itemType.(type) {
		case string:
			return typedValue != ""
		case []interface{}:
			return len(typedValue) > 0
		}
	}

	// Check for schema.org URL in @id
	if itemId, exists := itemMap["@id"]; exists {
		if idStr, ok := itemId.(string); ok {
			return strings.Contains(idStr, "schema.org") ||
				strings.Contains(idStr, "http") // Any URL-like identifier
		}
	}

	// Check if it has common schema.org properties
	commonProps := []string{"name", "description", "url", "image", "author", "publisher"}
	propCount := 0
	for _, prop := range commonProps {
		if _, exists := itemMap[prop]; exists {
			propCount++
		}
	}

	// Consider valid if it has multiple common properties
	return propCount >= 2
}

// countSchemaTypes counts unique schema types for debugging
// JavaScript original code:
//
//	// Helper for debugging and logging
func (d *Defuddle) countSchemaTypes(items []interface{}) int {
	typeSet := make(map[string]bool)

	for _, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			// Check both @type and type (after JSON-LD processing)
			var itemType interface{}
			var exists bool
			if itemType, exists = itemMap["@type"]; !exists {
				itemType, exists = itemMap["type"]
			}

			if exists {
				switch typedValue := itemType.(type) {
				case string:
					typeSet[typedValue] = true
				case []interface{}:
					for _, t := range typedValue {
						if typeStr, ok := t.(string); ok {
							typeSet[typeStr] = true
						}
					}
				}
			}
		}
	}

	return len(typeSet)
}

// collectMetaTags collects meta tags from the document
func (d *Defuddle) collectMetaTags() []MetaTag {
	var metaTags []MetaTag

	d.doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, nameExists := s.Attr("name")
		property, propertyExists := s.Attr("property")
		content, contentExists := s.Attr("content")

		if contentExists && content != "" {
			metaTag := MetaTag{
				Content: &content,
			}
			if nameExists {
				metaTag.Name = &name
			}
			if propertyExists {
				metaTag.Property = &property
			}
			metaTags = append(metaTags, metaTag)
		}
	})

	return metaTags
}

// evaluateMediaQueries evaluates mobile styles from CSS media queries
// JavaScript original code:
//
//	private _evaluateMediaQueries(doc: Document): StyleChange[] {
//		const mobileStyles: StyleChange[] = [];
//		const maxWidthRegex = /max-width[^:]*:\s*(\d+)/;
//
//		try {
//			// Get all styles, including inline styles
//			const sheets = Array.from(doc.styleSheets).filter(sheet => {
//				try {
//					// Access rules once to check validity
//					sheet.cssRules;
//					return true;
//				} catch (e) {
//					// Expected error for cross-origin stylesheets or Node.js environment
//					if (e instanceof DOMException && e.name === 'SecurityError') {
//						return false;
//					}
//					return false;
//				}
//			});
//
//			// Process all sheets in a single pass
//			const mediaRules = sheets.flatMap(sheet => {
//				try {
//					// Check if we're in a browser environment where CSSMediaRule is available
//					if (typeof CSSMediaRule === 'undefined') {
//						return [];
//					}
//
//					return Array.from(sheet.cssRules)
//						.filter((rule): rule is CSSMediaRule =>
//							rule instanceof CSSMediaRule &&
//							rule.conditionText.includes('max-width')
//						);
//				} catch (e) {
//					if (this.debug) {
//						console.warn('Defuddle: Failed to process stylesheet:', e);
//					}
//					return [];
//				}
//			});
//
//			// Process all media rules in a single pass
//			mediaRules.forEach(rule => {
//				const match = rule.conditionText.match(maxWidthRegex);
//				if (match) {
//					const maxWidth = parseInt(match[1]);
//
//					if (MOBILE_WIDTH <= maxWidth) {
//						// Batch process all style rules
//						const styleRules = Array.from(rule.cssRules)
//							.filter((r): r is CSSStyleRule => r instanceof CSSStyleRule);
//
//						styleRules.forEach(cssRule => {
//							try {
//								mobileStyles.push({
//									selector: cssRule.selectorText,
//									styles: cssRule.style.cssText
//								});
//							} catch (e) {
//								if (this.debug) {
//									console.warn('Defuddle: Failed to process CSS rule:', e);
//								}
//							}
//						});
//					}
//				}
//			});
//		} catch (e) {
//			console.error('Defuddle: Error evaluating media queries:', e);
//		}
//
//		return mobileStyles;
//	}
func (d *Defuddle) evaluateMediaQueries() []StyleChange {
	// Note: In Go/server environment, we don't have access to CSS stylesheets
	// This is a placeholder for future implementation if needed
	// Most content extraction doesn't require CSS evaluation
	return []StyleChange{}
}

// StyleChange represents a CSS style change for mobile
type StyleChange struct {
	Selector string
	Styles   string
}

// applyMobileStyles applies mobile styles to the document
// JavaScript original code:
//
//	private applyMobileStyles(doc: Document, mobileStyles: StyleChange[]) {
//		let appliedCount = 0;
//
//		mobileStyles.forEach(({selector, styles}) => {
//			try {
//				const elements = doc.querySelectorAll(selector);
//				elements.forEach(element => {
//					element.setAttribute('style',
//						(element.getAttribute('style') || '') + styles
//					);
//					appliedCount++;
//				});
//			} catch (e) {
//				console.error('Defuddle', 'Error applying styles for selector:', selector, e);
//			}
//		});
//	}
func (d *Defuddle) applyMobileStyles(doc *goquery.Document, mobileStyles []StyleChange) {
	appliedCount := 0

	for _, change := range mobileStyles {
		doc.Find(change.Selector).Each(func(i int, element *goquery.Selection) {
			existingStyle, _ := element.Attr("style")
			newStyle := existingStyle + change.Styles
			element.SetAttr("style", newStyle)
			appliedCount++
		})
	}

	if d.debug {
		slog.Debug("Applied mobile styles", "count", appliedCount)
	}
}

// removeHiddenElements removes elements that are hidden via CSS
// JavaScript original code:
//
//	private removeHiddenElements(doc: Document) {
//		let count = 0;
//		const elementsToRemove = new Set<Element>();
//
//		// Get all elements and check their styles
//		const allElements = Array.from(doc.getElementsByTagName('*'));
//
//		// Process styles in batches to minimize layout thrashing
//		const BATCH_SIZE = 100;
//		for (let i = 0; i < allElements.length; i += BATCH_SIZE) {
//			const batch = allElements.slice(i, i + BATCH_SIZE);
//
//			// Read phase - gather all computedStyles
//			const styles = batch.map(element => {
//				try {
//					return element.ownerDocument.defaultView?.getComputedStyle(element);
//				} catch (e) {
//					// If we can't get computed style, check inline styles
//					const style = element.getAttribute('style');
//					if (!style) return null;
//
//					// Create a temporary style element to parse inline styles
//					const tempStyle = doc.createElement('style');
//					tempStyle.textContent = `* { ${style} }`;
//					doc.head.appendChild(tempStyle);
//					const computedStyle = element.ownerDocument.defaultView?.getComputedStyle(element);
//					doc.head.removeChild(tempStyle);
//					return computedStyle;
//				}
//			});
//
//			// Write phase - mark elements for removal
//			batch.forEach((element, index) => {
//				const computedStyle = styles[index];
//				if (computedStyle && (
//					computedStyle.display === 'none' ||
//					computedStyle.visibility === 'hidden' ||
//					computedStyle.opacity === '0'
//				)) {
//					elementsToRemove.add(element);
//					count++;
//				}
//			});
//		}
//
//		// Batch remove all hidden elements
//		this._log('Removed hidden elements:', count);
//	}
func (d *Defuddle) removeHiddenElements(doc *goquery.Document) {
	count := 0

	// Check inline styles for hidden elements
	doc.Find("*").Each(func(i int, element *goquery.Selection) {
		style, exists := element.Attr("style")
		if !exists {
			return
		}

		lowerStyle := strings.ToLower(style)
		if strings.Contains(lowerStyle, "display:none") ||
			strings.Contains(lowerStyle, "display: none") ||
			strings.Contains(lowerStyle, "visibility:hidden") ||
			strings.Contains(lowerStyle, "visibility: hidden") ||
			strings.Contains(lowerStyle, "opacity:0") ||
			strings.Contains(lowerStyle, "opacity: 0") {
			element.Remove()
			count++
		}
	})

	if d.debug {
		slog.Debug("Removed hidden elements", "count", count)
	}
}

// findSmallImages finds small images that should be removed
// JavaScript original code:
//
//	private findSmallImages(doc: Document): Set<string> {
//		const MIN_DIMENSION = 33;
//		const smallImages = new Set<string>();
//		const transformRegex = /scale\(([\d.]+)\)/;
//		const startTime = Date.now();
//		let processedCount = 0;
//
//		// 1. Read phase - Gather all elements in a single pass
//		const elements = [
//			...Array.from(doc.getElementsByTagName('img')),
//			...Array.from(doc.getElementsByTagName('svg'))
//		];
//
//		if (elements.length === 0) {
//			return smallImages;
//		}
//
//		// 2. Batch process - Collect all measurements in one go
//		const measurements = elements.map(element => ({
//			element,
//			// Static attributes (no reflow)
//			naturalWidth: element.tagName.toLowerCase() === 'img' ?
//				parseInt(element.getAttribute('width') || '0') || 0 : 0,
//			naturalHeight: element.tagName.toLowerCase() === 'img' ?
//				parseInt(element.getAttribute('height') || '0') || 0 : 0,
//			attrWidth: parseInt(element.getAttribute('width') || '0'),
//			attrHeight: parseInt(element.getAttribute('height') || '0')
//		}));
//
//		// 3. Batch compute styles - Process in chunks to avoid long tasks
//		const BATCH_SIZE = 50;
//		for (let i = 0; i < measurements.length; i += BATCH_SIZE) {
//			const batch = measurements.slice(i, i + BATCH_SIZE);
//
//			try {
//				// Read phase - compute all styles at once
//				const styles = batch.map(({ element }) => {
//					try {
//						return element.ownerDocument.defaultView?.getComputedStyle(element);
//					} catch (e) {
//						return null;
//					}
//				});
//
//				// Get bounding rectangles if available
//				const rects = batch.map(({ element }) => {
//					try {
//						return element.getBoundingClientRect();
//					} catch (e) {
//						return null;
//					}
//				});
//
//				// Process phase - no DOM operations
//				batch.forEach((measurement, index) => {
//					try {
//						const style = styles[index];
//						const rect = rects[index];
//
//						if (!style) return;
//
//						// Get transform scale in the same batch
//						const transform = style.transform;
//						const scale = transform ?
//							parseFloat(transform.match(transformRegex)?.[1] || '1') : 1;
//
//						// Calculate effective dimensions
//						const widths = [
//							measurement.naturalWidth,
//							measurement.attrWidth,
//							parseInt(style.width) || 0,
//							rect ? rect.width * scale : 0
//						].filter(dim => typeof dim === 'number' && dim > 0);
//
//						const heights = [
//							measurement.naturalHeight,
//							measurement.attrHeight,
//							parseInt(style.height) || 0,
//							rect ? rect.height * scale : 0
//						].filter(dim => typeof dim === 'number' && dim > 0);
//
//						// Decision phase - no DOM operations
//						if (widths.length > 0 && heights.length > 0) {
//							const effectiveWidth = Math.min(...widths);
//							const effectiveHeight = Math.min(...heights);
//
//							if (effectiveWidth < MIN_DIMENSION || effectiveHeight < MIN_DIMENSION) {
//								const identifier = this.getElementIdentifier(measurement.element);
//								if (identifier) {
//									smallImages.add(identifier);
//									processedCount++;
//								}
//							}
//						}
//					} catch (e) {
//						if (this.debug) {
//							console.warn('Defuddle: Failed to process element dimensions:', e);
//						}
//					}
//				});
//			} catch (e) {
//				if (this.debug) {
//					console.warn('Defuddle: Failed to process batch:', e);
//				}
//			}
//		}
//
//		const endTime = Date.now();
//		this._log('Found small elements:', {
//			count: processedCount,
//			processingTime: `${(endTime - startTime).toFixed(2)}ms`
//		});
//
//		return smallImages;
//	}
func (d *Defuddle) findSmallImages(doc *goquery.Document) map[string]bool {
	const minDimension = 33
	smallImages := make(map[string]bool)
	processedCount := 0

	// Process img and svg elements
	doc.Find("img, svg").Each(func(i int, element *goquery.Selection) {
		tagName := goquery.NodeName(element)

		// Get dimensions from attributes
		widthStr, _ := element.Attr("width")
		heightStr, _ := element.Attr("height")

		width := 0
		height := 0

		if widthStr != "" {
			if w, err := strconv.Atoi(widthStr); err == nil {
				width = w
			}
		}

		if heightStr != "" {
			if h, err := strconv.Atoi(heightStr); err == nil {
				height = h
			}
		}

		// Check if dimensions are small
		if (width > 0 && width < minDimension) || (height > 0 && height < minDimension) {
			identifier := d.getElementIdentifier(element, tagName)
			if identifier != "" {
				smallImages[identifier] = true
				processedCount++
			}
		}
	})

	if d.debug {
		slog.Debug("Found small images", "count", processedCount)
	}

	return smallImages
}

// removeSmallImages removes small images from the document
// JavaScript original code:
//
//	private removeSmallImages(doc: Document, smallImages: Set<string>) {
//		let removedCount = 0;
//
//		['img', 'svg'].forEach(tag => {
//			const elements = doc.getElementsByTagName(tag);
//			Array.from(elements).forEach(element => {
//				const identifier = this.getElementIdentifier(element);
//				if (identifier && smallImages.has(identifier)) {
//					element.remove();
//					removedCount++;
//				}
//			});
//		});
//
//		this._log('Removed small elements:', removedCount);
//	}
func (d *Defuddle) removeSmallImages(doc *goquery.Document, smallImages map[string]bool) {
	removedCount := 0

	doc.Find("img, svg").Each(func(i int, element *goquery.Selection) {
		tagName := goquery.NodeName(element)
		identifier := d.getElementIdentifier(element, tagName)
		if identifier != "" && smallImages[identifier] {
			element.Remove()
			removedCount++
		}
	})

	if d.debug {
		slog.Debug("Removed small images", "count", removedCount)
	}
}

// getElementIdentifier creates a unique identifier for an element
// JavaScript original code:
//
//	private getElementIdentifier(element: Element): string | null {
//		// Try to create a unique identifier using various attributes
//		if (element.tagName.toLowerCase() === 'img') {
//			// For lazy-loaded images, use data-src as identifier if available
//			const dataSrc = element.getAttribute('data-src');
//			if (dataSrc) return `src:${dataSrc}`;
//
//			const src = element.getAttribute('src') || '';
//			const srcset = element.getAttribute('srcset') || '';
//			const dataSrcset = element.getAttribute('data-srcset');
//
//			if (src) return `src:${src}`;
//			if (srcset) return `srcset:${srcset}`;
//			if (dataSrcset) return `srcset:${dataSrcset}`;
//		}
//
//		const id = element.id || '';
//		const className = element.className || '';
//		const viewBox = element.tagName.toLowerCase() === 'svg' ? element.getAttribute('viewBox') || '' : '';
//
//		if (id) return `id:${id}`;
//		if (viewBox) return `viewBox:${viewBox}`;
//		if (className) return `class:${className}`;
//
//		return null;
//	}
func (d *Defuddle) getElementIdentifier(element *goquery.Selection, tagName string) string {
	if tagName == "img" {
		// For lazy-loaded images, use data-src as identifier if available
		if dataSrc, exists := element.Attr("data-src"); exists && dataSrc != "" {
			return "src:" + dataSrc
		}

		if src, exists := element.Attr("src"); exists && src != "" {
			return "src:" + src
		}

		if srcset, exists := element.Attr("srcset"); exists && srcset != "" {
			return "srcset:" + srcset
		}

		if dataSrcset, exists := element.Attr("data-srcset"); exists && dataSrcset != "" {
			return "srcset:" + dataSrcset
		}
	}

	if id, exists := element.Attr("id"); exists && id != "" {
		return "id:" + id
	}

	if tagName == "svg" {
		if viewBox, exists := element.Attr("viewBox"); exists && viewBox != "" {
			return "viewBox:" + viewBox
		}
	}

	if className, exists := element.Attr("class"); exists && className != "" {
		return "class:" + className
	}

	return ""
}

// convertHTMLToMarkdown converts HTML content to Markdown
func (d *Defuddle) convertHTMLToMarkdown(htmlContent string) (string, error) {
	return markdown.ConvertHTML(htmlContent)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
