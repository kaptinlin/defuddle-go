// Package extractors provides site-specific content extraction functionality.
package extractors

import (
	"github.com/PuerkitoBio/goquery"
)

// ExtractorResult represents the result of content extraction
// Corresponding to TypeScript interface ExtractorResult
type ExtractorResult struct {
	Content          string            `json:"content"`
	ContentHTML      string            `json:"contentHtml"`
	ExtractedContent map[string]any    `json:"extractedContent,omitempty"`
	Variables        map[string]string `json:"variables,omitempty"`
}

// BaseExtractor defines the interface for site-specific extractors
// TypeScript original code:
//
//	export abstract class BaseExtractor {
//		protected document: Document;
//		protected url: string;
//		protected schemaOrgData?: any;
//
//		constructor(document: Document, url: string, schemaOrgData?: any) {
//			this.document = document;
//			this.url = url;
//			this.schemaOrgData = schemaOrgData;
//		}
//
//		abstract canExtract(): boolean;
//		abstract extract(): ExtractorResult;
//		abstract getName(): string;
//	}
type BaseExtractor interface {
	CanExtract() bool
	Extract() *ExtractorResult
	GetName() string
}

// ExtractorBase provides common functionality for extractors
// Implementation of the protected properties in TypeScript BaseExtractor
type ExtractorBase struct {
	document      *goquery.Document
	url           string
	schemaOrgData any
}

// NewExtractorBase creates a new base extractor
// TypeScript original code:
//
//	constructor(document: Document, url: string, schemaOrgData?: any) {
//		this.document = document;
//		this.url = url;
//		this.schemaOrgData = schemaOrgData;
//	}
func NewExtractorBase(document *goquery.Document, url string, schemaOrgData any) *ExtractorBase {
	return &ExtractorBase{
		document:      document,
		url:           url,
		schemaOrgData: schemaOrgData,
	}
}

// GetDocument returns the document
func (e *ExtractorBase) GetDocument() *goquery.Document {
	return e.document
}

// GetURL returns the URL
func (e *ExtractorBase) GetURL() string {
	return e.url
}

// GetSchemaOrgData returns the schema.org data
func (e *ExtractorBase) GetSchemaOrgData() any {
	return e.schemaOrgData
}

// GetTextContent safely extracts text content from a selection
func (e *ExtractorBase) GetTextContent(sel *goquery.Selection) string {
	if sel.Length() == 0 {
		return ""
	}
	return sel.Text()
}

// GetHTMLContent safely extracts HTML content from a selection
func (e *ExtractorBase) GetHTMLContent(sel *goquery.Selection) string {
	if sel.Length() == 0 {
		return ""
	}
	html, _ := sel.Html()
	return html
}

// GetAttribute safely gets an attribute value
func (e *ExtractorBase) GetAttribute(sel *goquery.Selection, attr string) string {
	if sel.Length() == 0 {
		return ""
	}
	value, _ := sel.Attr(attr)
	return value
}
