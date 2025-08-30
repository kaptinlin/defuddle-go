package defuddle

import (
	"github.com/kaptinlin/defuddle-go/internal/debug"
	"github.com/kaptinlin/defuddle-go/internal/elements"
	"github.com/kaptinlin/defuddle-go/internal/metadata"
)

// MetaTag represents a meta tag item from HTML
// This is an alias to the internal metadata.MetaTag type
type MetaTag = metadata.MetaTag

// Options represents configuration options for Defuddle parsing
// JavaScript original code:
//
//	export interface DefuddleOptions {
//	  debug?: boolean;
//	  url?: string;
//	  markdown?: boolean;
//	  separateMarkdown?: boolean;
//	  removeExactSelectors?: boolean;
//	  removePartialSelectors?: boolean;
//	}
type Options struct {
	// Enable debug logging
	Debug bool `json:"debug,omitempty"`

	// URL of the page being parsed
	URL string `json:"url,omitempty"`

	// Convert output to Markdown
	Markdown bool `json:"markdown,omitempty"`

	// Include Markdown in the response
	SeparateMarkdown bool `json:"separateMarkdown,omitempty"`

	// Whether to remove elements matching exact selectors like ads, social buttons, etc.
	// Defaults to true.
	RemoveExactSelectors bool `json:"removeExactSelectors,omitempty"`

	// Whether to remove elements matching partial selectors like ads, social buttons, etc.
	// Defaults to true.
	RemovePartialSelectors bool `json:"removePartialSelectors,omitempty"`

	// Remove images from the extracted content
	// Defaults to false.
	RemoveImages bool `json:"removeImages,omitempty"`

	// Element processing options
	ProcessCode      bool                                 `json:"processCode,omitempty"`
	ProcessImages    bool                                 `json:"processImages,omitempty"`
	ProcessHeadings  bool                                 `json:"processHeadings,omitempty"`
	ProcessMath      bool                                 `json:"processMath,omitempty"`
	ProcessFootnotes bool                                 `json:"processFootnotes,omitempty"`
	ProcessRoles     bool                                 `json:"processRoles,omitempty"`
	CodeOptions      *elements.CodeBlockProcessingOptions `json:"codeOptions,omitempty"`
	ImageOptions     *elements.ImageProcessingOptions     `json:"imageOptions,omitempty"`
	HeadingOptions   *elements.HeadingProcessingOptions   `json:"headingOptions,omitempty"`
	MathOptions      *elements.MathProcessingOptions      `json:"mathOptions,omitempty"`
	FootnoteOptions  *elements.FootnoteProcessingOptions  `json:"footnoteOptions,omitempty"`
	RoleOptions      *elements.RoleProcessingOptions      `json:"roleOptions,omitempty"`
}

// Metadata represents extracted metadata from a document
// This is an alias to the internal metadata.Metadata type
type Metadata = metadata.Metadata

// Result represents the complete response from Defuddle parsing
// JavaScript original code:
//
//	export interface DefuddleResponse extends DefuddleMetadata {
//	  content: string;
//	  contentMarkdown?: string;
//	  extractorType?: string;
//	  metaTags?: MetaTagItem[];
//	}
type Result struct {
	Metadata
	Content         string           `json:"content"`
	ContentMarkdown *string          `json:"contentMarkdown,omitempty"`
	ExtractorType   *string          `json:"extractorType,omitempty"`
	MetaTags        []MetaTag        `json:"metaTags,omitempty"`
	DebugInfo       *debug.DebugInfo `json:"debugInfo,omitempty"`
}

// ExtractorVariables represents variables extracted by site-specific extractors
// JavaScript original code:
//
//	export interface ExtractorVariables {
//	  [key: string]: string;
//	}
type ExtractorVariables map[string]string

// ExtractedContent represents content extracted by site-specific extractors
// JavaScript original code:
//
//	export interface ExtractedContent {
//	  title?: string;
//	  author?: string;
//	  published?: string;
//	  content?: string;
//	  contentHtml?: string;
//	  variables?: ExtractorVariables;
//	}
type ExtractedContent struct {
	Title       *string             `json:"title,omitempty"`
	Author      *string             `json:"author,omitempty"`
	Published   *string             `json:"published,omitempty"`
	Content     *string             `json:"content,omitempty"`
	ContentHTML *string             `json:"contentHtml,omitempty"`
	Variables   *ExtractorVariables `json:"variables,omitempty"`
}
