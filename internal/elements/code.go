// Package elements provides enhanced element processing functionality
// This module handles code block processing including syntax highlighting,
// language detection, and code formatting
package elements

import (
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Pre-compiled regex patterns for language detection and code normalization.
var (
	highlighterPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^language-(\w+)$`),
		regexp.MustCompile(`^lang-(\w+)$`),
		regexp.MustCompile(`^(\w+)-code$`),
		regexp.MustCompile(`^code-(\w+)$`),
		regexp.MustCompile(`^syntax-(\w+)$`),
		regexp.MustCompile(`^code-snippet__(\w+)$`),
		regexp.MustCompile(`^highlight-(\w+)$`),
		regexp.MustCompile(`^(\w+)-snippet$`),
		regexp.MustCompile(`(?:^|\s)(?:language|lang|brush|syntax)-(\w+)(?:\s|$)`),
	}

	codeThreeNewlinesRe = regexp.MustCompile(`\n{3,}`)
	codeLeadingNlRe     = regexp.MustCompile(`^\n+`)
	codeTrailingNlRe    = regexp.MustCompile(`\n+$`)

	codeLanguages = map[string]bool{
		"abap": true, "actionscript": true, "ada": true, "adoc": true, "agda": true, "antlr4": true,
		"applescript": true, "arduino": true, "armasm": true, "asciidoc": true, "aspnet": true, "atom": true,
		"bash": true, "batch": true, "c": true, "clojure": true, "cmake": true, "cobol": true,
		"coffeescript": true, "cpp": true, "c++": true, "crystal": true, "csharp": true, "cs": true,
		"dart": true, "django": true, "dockerfile": true, "dotnet": true, "elixir": true, "elm": true,
		"erlang": true, "fortran": true, "fsharp": true, "gdscript": true, "gitignore": true, "glsl": true,
		"golang": true, "go": true, "gradle": true, "graphql": true, "groovy": true, "haskell": true,
		"hs": true, "haxe": true, "hlsl": true, "html": true, "idris": true, "java": true,
		"javascript": true, "js": true, "jsx": true, "jsdoc": true, "json": true, "jsonp": true,
		"julia": true, "kotlin": true, "latex": true, "lisp": true, "elisp": true, "livescript": true,
		"lua": true, "makefile": true, "markdown": true, "md": true, "markup": true, "masm": true,
		"mathml": true, "matlab": true, "mongodb": true, "mysql": true, "nasm": true, "nginx": true,
		"nim": true, "nix": true, "objc": true, "ocaml": true, "pascal": true, "perl": true,
		"php": true, "postgresql": true, "powershell": true, "prolog": true, "puppet": true, "python": true,
		"regex": true, "rss": true, "ruby": true, "rb": true, "rust": true, "scala": true,
		"scheme": true, "shell": true, "sh": true, "solidity": true, "sparql": true, "sql": true,
		"ssml": true, "svg": true, "swift": true, "tcl": true, "terraform": true, "tex": true,
		"toml": true, "typescript": true, "ts": true, "tsx": true, "unrealscript": true, "verilog": true,
		"vhdl": true, "webassembly": true, "wasm": true, "xml": true, "yaml": true, "yml": true,
		"zig": true,
	}
)

/*
TypeScript source code (code.ts, 319 lines):

This module provides code block processing functionality including:
- Language detection from class names and content analysis
- Code formatting and normalization
- Syntax highlighting preparation
- Code block structure optimization

Key functions:
- processCodeBlocks(): Main processing function for all code blocks
- detectLanguage(): Language detection from various sources
- formatCodeBlock(): Code formatting and structure optimization
- normalizeCodeContent(): Content normalization and cleanup

Original TypeScript implementation:
const HIGHLIGHTER_PATTERNS = [
	/^language-(\w+)$/,          // language-javascript
	/^lang-(\w+)$/,              // lang-javascript
	/^(\w+)-code$/,              // javascript-code
	/^code-(\w+)$/,              // code-javascript
	/^syntax-(\w+)$/,            // syntax-javascript
	/^code-snippet__(\w+)$/,     // code-snippet__javascript
	/^highlight-(\w+)$/,         // highlight-javascript
	/^(\w+)-snippet$/,           // javascript-snippet
	/(?:^|\s)(?:language|lang|brush|syntax)-(\w+)(?:\s|$)/i
];

const CODE_LANGUAGES = new Set([
	'javascript', 'js', 'jsx', 'typescript', 'ts', 'tsx',
	'python', 'java', 'c', 'cpp', 'c++', 'csharp', 'cs',
	// ... extensive language list
]);

export const codeBlockRules = [
	{
		selector: [
			'pre',
			'div[class*="prismjs"]',
			'.syntaxhighlighter',
			'.highlight',
			'.highlight-source',
			'.wp-block-syntaxhighlighter-code',
			'.wp-block-code',
			'div[class*="language-"]'
		].join(', '),
		element: 'pre',
		transform: (el: Element, doc: Document): Element => {
			// Processing logic here
		}
	}
];
*/

// CodeBlockProcessor handles code block processing and enhancement
// TypeScript original code:
//
//	class CodeBlockProcessor {
//	  constructor(private document: Document) {}
//	}
type CodeBlockProcessor struct {
	doc *goquery.Document
}

// CodeBlockProcessingOptions contains options for code block processing
// TypeScript original code:
//
//	interface CodeBlockOptions {
//	  detectLanguage?: boolean;
//	  formatCode?: boolean;
//	  addLineNumbers?: boolean;
//	  enableSyntaxHighlight?: boolean;
//	  wrapInPre?: boolean;
//	}
type CodeBlockProcessingOptions struct {
	DetectLanguage        bool
	FormatCode            bool
	AddLineNumbers        bool
	EnableSyntaxHighlight bool
	WrapInPre             bool
}

// DefaultCodeBlockProcessingOptions returns default options for code block processing
// TypeScript original code:
//
//	const defaultOptions: CodeBlockOptions = {
//	  detectLanguage: true,
//	  formatCode: true,
//	  addLineNumbers: false,
//	  enableSyntaxHighlight: true,
//	  wrapInPre: true
//	};
func DefaultCodeBlockProcessingOptions() *CodeBlockProcessingOptions {
	return &CodeBlockProcessingOptions{
		DetectLanguage:        true,
		FormatCode:            true,
		AddLineNumbers:        false,
		EnableSyntaxHighlight: true,
		WrapInPre:             true,
	}
}

// NewCodeBlockProcessor creates a new code block processor
// TypeScript original code:
// constructor(private doc: Document) {}
func NewCodeBlockProcessor(doc *goquery.Document) *CodeBlockProcessor {
	return &CodeBlockProcessor{
		doc: doc,
	}
}

// ProcessCodeBlocks processes all code blocks in the document
// TypeScript original code:
// export const codeBlockRules = [
//
//	{
//	  selector: [
//	    'pre',
//	    'div[class*="prismjs"]',
//	    '.syntaxhighlighter',
//	    '.highlight',
//	    '.highlight-source',
//	    '.wp-block-syntaxhighlighter-code',
//	    '.wp-block-code',
//	    'div[class*="language-"]'
//	  ].join(', '),
//	  element: 'pre',
//	  transform: (el: Element, doc: Document): Element => {
//	    // Processing logic here
//	  }
//	}
//
// ];
func (p *CodeBlockProcessor) ProcessCodeBlocks(options *CodeBlockProcessingOptions) {
	if options == nil {
		options = DefaultCodeBlockProcessingOptions()
	}

	slog.Debug("processing code blocks", "detectLanguage", options.DetectLanguage, "formatCode", options.FormatCode)

	// Process code blocks with the same selector logic as TypeScript
	selector := []string{
		"pre",
		"div[class*=\"prismjs\"]",
		".syntaxhighlighter",
		".highlight",
		".highlight-source",
		".wp-block-syntaxhighlighter-code",
		".wp-block-code",
		"div[class*=\"language-\"]",
	}

	combinedSelector := strings.Join(selector, ", ")
	slog.Debug("using code block selector", "selector", combinedSelector)

	var processedCount int
	p.doc.Find(combinedSelector).Each(func(_ int, s *goquery.Selection) {
		p.processCodeBlock(s, options)
		processedCount++
	})

	slog.Info("code blocks processed", "count", processedCount)
}

// processCodeBlock processes a single code block
// TypeScript original code:
//
//	transform: (el: Element, doc: Document): Element => {
//	  const getCodeLanguage = (element: Element): string => {
//	    const dataLang = element.getAttribute('data-lang') || element.getAttribute('data-language');
//	    if (dataLang) {
//	      return dataLang.toLowerCase();
//	    }
//	    // Check class names for patterns and supported languages
//	    const classNames = Array.from(element.classList || []);
//	    // Pattern matching logic...
//	  };
//
//	  let language = '';
//	  let currentElement: Element | null = el;
//	  while (currentElement && !language) {
//	    language = getCodeLanguage(currentElement);
//	    currentElement = currentElement.parentElement;
//	  }
//	}
func (p *CodeBlockProcessor) processCodeBlock(s *goquery.Selection, options *CodeBlockProcessingOptions) {
	slog.Debug("processing individual code block")

	// Detect language using hierarchical approach like TypeScript
	var language string
	if options.DetectLanguage {
		language = p.detectLanguageHierarchical(s)
		if language != "" {
			slog.Debug("detected language", "language", language)
		}
	}

	// Extract content using structured text extraction (TypeScript equivalent)
	content := p.extractStructuredContent(s)
	content = p.normalizeCodeContent(content)

	// Format the code block
	if options.FormatCode {
		p.formatCodeBlock(s, language, content, options)
	}
}

// detectLanguageHierarchical detects language using hierarchical approach like TypeScript
// TypeScript original code:
// let language = ”;
// let currentElement: Element | null = el;
//
//	while (currentElement && !language) {
//	  language = getCodeLanguage(currentElement);
//	  // Also check for code elements within the current element
//	  const codeEl = currentElement.querySelector('code');
//	  if (!language && codeEl) {
//	    language = getCodeLanguage(codeEl);
//	  }
//	  currentElement = currentElement.parentElement;
//	}
func (p *CodeBlockProcessor) detectLanguageHierarchical(s *goquery.Selection) string {
	var language string
	current := s

	// Traverse hierarchy like TypeScript implementation
	for current.Length() > 0 && language == "" {
		language = p.getCodeLanguage(current)

		// Also check for code elements within current element
		if language == "" {
			codeEl := current.Find("code").First()
			if codeEl.Length() > 0 {
				language = p.getCodeLanguage(codeEl)
			}
		}

		current = current.Parent()
	}

	return language
}

// getCodeLanguage extracts language from element attributes and classes
// TypeScript original code:
//
//	const getCodeLanguage = (element: Element): string => {
//	  // Check data-lang attribute first
//	  const dataLang = element.getAttribute('data-lang') || element.getAttribute('data-language');
//	  if (dataLang) {
//	    return dataLang.toLowerCase();
//	  }
//
//	  // Check class names for patterns and supported languages
//	  const classNames = Array.from(element.classList || []);
//
//	  // Check for syntax highlighter specific format
//	  if (element.classList?.contains('syntaxhighlighter')) {
//	    const langClass = classNames.find(c => !['syntaxhighlighter', 'nogutter'].includes(c));
//	    if (langClass && CODE_LANGUAGES.has(langClass.toLowerCase())) {
//	      return langClass.toLowerCase();
//	    }
//	  }
//
//	  // Check patterns
//	  for (const className of classNames) {
//	    for (const pattern of HIGHLIGHTER_PATTERNS) {
//	      const match = className.toLowerCase().match(pattern);
//	      if (match && match[1] && CODE_LANGUAGES.has(match[1].toLowerCase())) {
//	        return match[1].toLowerCase();
//	      }
//	    }
//	  }
//
//	  // If all else fails, check for bare language names
//	  for (const className of classNames) {
//	    if (CODE_LANGUAGES.has(className.toLowerCase())) {
//	      return className.toLowerCase();
//	    }
//	  }
//
//	  return '';
//	};
func (p *CodeBlockProcessor) getCodeLanguage(s *goquery.Selection) string {
	// Check data-lang attribute first
	if dataLang, exists := s.Attr("data-lang"); exists && dataLang != "" {
		return strings.ToLower(dataLang)
	}
	if dataLanguage, exists := s.Attr("data-language"); exists && dataLanguage != "" {
		return strings.ToLower(dataLanguage)
	}

	// Get class names for pattern matching
	class, hasClass := s.Attr("class")
	if !hasClass {
		return ""
	}

	classNames := strings.Fields(class)

	// Check for syntax highlighter specific format
	if slices.Contains(classNames, "syntaxhighlighter") {
		for _, className := range classNames {
			if className != "syntaxhighlighter" && className != "nogutter" {
				langLower := strings.ToLower(className)
				if p.isCodeLanguage(langLower) {
					return langLower
				}
			}
		}
	}

	// Check highlighter patterns (same as TypeScript)
	for _, className := range classNames {
		classLower := strings.ToLower(className)
		for _, re := range highlighterPatterns {
			if matches := re.FindStringSubmatch(classLower); len(matches) > 1 {
				lang := matches[1]
				if p.isCodeLanguage(lang) {
					return lang
				}
			}
		}
	}

	// Check for bare language names
	for _, className := range classNames {
		classLower := strings.ToLower(className)
		if p.isCodeLanguage(classLower) {
			return classLower
		}
	}

	return ""
}

// extractStructuredContent extracts content using structured approach like TypeScript
// TypeScript original code:
//
//	const extractStructuredText = (element: Node): string => {
//	  if (isTextNode(element)) {
//	    return element.textContent || '';
//	  }
//
//	  let text = '';
//	  if (isElement(element)) {
//	    // Handle explicit line breaks
//	    if (element.tagName === 'BR') {
//	      return '\n';
//	    }
//
//	    // Handle common line-based code formats
//	    if (element.matches('div[class*="line"], span[class*="line"], .ec-line, [data-line-number], [data-line]')) {
//	      // Processing logic for line-based formats
//	    }
//
//	    element.childNodes.forEach(child => {
//	      text += extractStructuredText(child);
//	    });
//	  }
//	  return text;
//	};
func (p *CodeBlockProcessor) extractStructuredContent(s *goquery.Selection) string {
	// First try WordPress syntax highlighter extraction
	if s.HasClass("syntaxhighlighter") || s.HasClass("wp-block-syntaxhighlighter-code") {
		if content := p.extractWordPressContent(s); content != "" {
			return content
		}
	}

	// Use structured text extraction as fallback
	return p.extractStructuredText(s)
}

// extractWordPressContent extracts content from WordPress syntax highlighter
// TypeScript original code:
//
//	const extractWordPressContent = (element: Element): string => {
//	  // Handle WordPress syntax highlighter table format
//	  const codeContainer = element.querySelector('.syntaxhighlighter table .code .container');
//	  if (codeContainer) {
//	    return Array.from(codeContainer.children)
//	      .map(line => {
//	        const codeParts = Array.from(line.querySelectorAll('code'))
//	          .map(code => {
//	            let text = code.textContent || '';
//	            if (code.classList?.contains('spaces')) {
//	              text = ' '.repeat(text.length);
//	            }
//	            return text;
//	          })
//	          .join('');
//	        return codeParts || line.textContent || '';
//	      })
//	      .join('\n');
//	  }
//
//	  // Handle WordPress syntax highlighter non-table format
//	  const codeLines = element.querySelectorAll('.code .line');
//	  if (codeLines.length > 0) {
//	    return Array.from(codeLines)
//	      .map(line => {
//	        const codeParts = Array.from(line.querySelectorAll('code'))
//	          .map(code => code.textContent || '')
//	          .join('');
//	        return codeParts || line.textContent || '';
//	      })
//	      .join('\n');
//	  }
//
//	  return '';
//	};
func (p *CodeBlockProcessor) extractWordPressContent(s *goquery.Selection) string {
	var builder strings.Builder

	// Handle WordPress syntax highlighter table format
	codeContainer := s.Find(".syntaxhighlighter table .code .container")
	if codeContainer.Length() > 0 {
		codeContainer.Children().Each(func(i int, line *goquery.Selection) {
			if i > 0 {
				builder.WriteString("\n")
			}

			var lineBuilder strings.Builder
			line.Find("code").Each(func(_ int, code *goquery.Selection) {
				text := code.Text()
				if code.HasClass("spaces") {
					// Replace with spaces of same length
					lineBuilder.WriteString(strings.Repeat(" ", len(text)))
				} else {
					lineBuilder.WriteString(text)
				}
			})

			if lineContent := lineBuilder.String(); lineContent != "" {
				builder.WriteString(lineContent)
			} else {
				builder.WriteString(line.Text())
			}
		})
		return builder.String()
	}

	// Handle WordPress syntax highlighter non-table format
	codeLines := s.Find(".code .line")
	if codeLines.Length() > 0 {
		codeLines.Each(func(i int, line *goquery.Selection) {
			if i > 0 {
				builder.WriteString("\n")
			}

			var lineBuilder strings.Builder
			line.Find("code").Each(func(_ int, code *goquery.Selection) {
				lineBuilder.WriteString(code.Text())
			})

			if lineContent := lineBuilder.String(); lineContent != "" {
				builder.WriteString(lineContent)
			} else {
				builder.WriteString(line.Text())
			}
		})
		return builder.String()
	}

	return ""
}

// extractStructuredText recursively extracts text with structure preservation
// TypeScript original code:
//
//	const extractStructuredText = (element: Node): string => {
//	  if (isTextNode(element)) {
//	    return element.textContent || '';
//	  }
//
//	  let text = '';
//	  if (isElement(element)) {
//	    // Handle explicit line breaks
//	    if (element.tagName === 'BR') {
//	      return '\n';
//	    }
//
//	    // Handle common line-based code formats
//	    if (element.matches('div[class*="line"], span[class*="line"], .ec-line, [data-line-number], [data-line]')) {
//	      // Try to find the actual code content in common structures:
//	      // 1. A dedicated code container
//	      const codeContainer = element.querySelector('.code, .content, [class*="code-"], [class*="content-"]');
//	      if (codeContainer) {
//	        return (codeContainer.textContent || '') + '\n';
//	      }
//
//	      // 2. Line number is in a separate element
//	      const lineNumber = element.querySelector('.line-number, .gutter, [class*="line-number"], [class*="gutter"]');
//	      if (lineNumber) {
//	        const withoutLineNum = Array.from(element.childNodes)
//	          .filter(node => !lineNumber.contains(node))
//	          .map(node => extractStructuredText(node))
//	          .join('');
//	        return withoutLineNum + '\n';
//	      }
//
//	      // 3. Fallback to the entire line content
//	      return element.textContent + '\n';
//	    }
//
//	    element.childNodes.forEach(child => {
//	      text += extractStructuredText(child);
//	    });
//	  }
//	  return text;
//	};
func (p *CodeBlockProcessor) extractStructuredText(s *goquery.Selection) string {
	var builder strings.Builder

	s.Contents().Each(func(_ int, node *goquery.Selection) {
		// Handle text nodes
		if goquery.NodeName(node) == "#text" {
			builder.WriteString(node.Text())
			return
		}

		// Handle BR elements
		if node.Is("br") {
			builder.WriteString("\n")
			return
		}

		// Handle common line-based code formats
		lineSelectors := []string{
			"div[class*=\"line\"]",
			"span[class*=\"line\"]",
			".ec-line",
			"[data-line-number]",
			"[data-line]",
		}

		for _, lineSelector := range lineSelectors {
			if node.Is(lineSelector) {
				// Try to find dedicated code container
				codeContainer := node.Find(".code, .content, [class*=\"code-\"], [class*=\"content-\"]")
				if codeContainer.Length() > 0 {
					builder.WriteString(codeContainer.Text())
					builder.WriteString("\n")
					return
				}

				// Handle line numbers in separate element
				lineNumber := node.Find(".line-number, .gutter, [class*=\"line-number\"], [class*=\"gutter\"]")
				if lineNumber.Length() > 0 {
					// Extract content without line numbers
					var lineContent strings.Builder
					node.Contents().Each(func(_ int, child *goquery.Selection) {
						// Check if child is not contained in lineNumber elements
						childNode := child.Get(0)
						isLineNumber := false
						lineNumber.Each(func(_ int, ln *goquery.Selection) {
							if ln.Get(0) == childNode {
								isLineNumber = true
							}
						})
						if !isLineNumber {
							lineContent.WriteString(p.extractStructuredText(child))
						}
					})
					builder.WriteString(lineContent.String())
					builder.WriteString("\n")
					return
				}

				// Fallback to entire line content
				builder.WriteString(node.Text())
				builder.WriteString("\n")
				return
			}
		}

		// Recursively process child elements
		builder.WriteString(p.extractStructuredText(node))
	})

	return builder.String()
}

// normalizeCodeContent normalizes and cleans up code content
// TypeScript original code:
// codeContent = codeContent
//
//	.replace(/^\s+|\s+$/g, '')      // Trim start/end whitespace
//	.replace(/\t/g, '    ')         // Convert tabs to spaces
//	.replace(/\n{3,}/g, '\n\n')     // Normalize multiple newlines
//	.replace(/\u00a0/g, ' ')        // Replace non-breaking spaces
//	.replace(/^\n+/, '')            // Remove extra newlines at start
//	.replace(/\n+$/, '');           // Remove extra newlines at end
func (p *CodeBlockProcessor) normalizeCodeContent(content string) string {
	// Trim whitespace
	content = strings.TrimSpace(content)

	// Convert tabs to spaces
	content = strings.ReplaceAll(content, "\t", "    ")

	// Replace non-breaking spaces
	content = strings.ReplaceAll(content, "\u00a0", " ")

	// Normalize multiple newlines
	content = codeThreeNewlinesRe.ReplaceAllString(content, "\n\n")

	// Remove extra newlines at start and end
	content = codeLeadingNlRe.ReplaceAllString(content, "")
	content = codeTrailingNlRe.ReplaceAllString(content, "")

	return content
}

// formatCodeBlock formats a code block with language and options
// TypeScript original code:
// // Create new pre element
// const newPre = doc.createElement('pre');
//
// // Create code element
// const code = doc.createElement('code');
//
//	if (language) {
//	  code.setAttribute('data-lang', language);
//	  code.setAttribute('class', `language-${language}`);
//	}
//
// code.textContent = codeContent;
//
// newPre.appendChild(code);
// return newPre;
func (p *CodeBlockProcessor) formatCodeBlock(s *goquery.Selection, language, content string, _ *CodeBlockProcessingOptions) {
	// Create new pre and code structure using HTML strings (simpler approach)
	var preHTML strings.Builder
	preHTML.WriteString("<pre>")
	preHTML.WriteString("<code")

	if language != "" {
		fmt.Fprintf(&preHTML, ` data-lang="%s" class="language-%s"`, language, language)
	}

	preHTML.WriteString(">")
	// Escape HTML content
	escapedContent := strings.ReplaceAll(content, "&", "&amp;")
	escapedContent = strings.ReplaceAll(escapedContent, "<", "&lt;")
	escapedContent = strings.ReplaceAll(escapedContent, ">", "&gt;")
	preHTML.WriteString(escapedContent)
	preHTML.WriteString("</code>")
	preHTML.WriteString("</pre>")

	// Replace original element with new structure
	s.ReplaceWithHtml(preHTML.String())

	slog.Debug("formatted code block", "language", language, "contentLength", len(content))
}

// isCodeLanguage checks if a language is in the supported languages set
// TypeScript original code:
// const CODE_LANGUAGES = new Set([
//
//	'abap', 'actionscript', 'ada', 'adoc', 'agda', 'antlr4',
//	// ... extensive language list ...
//
// ]);
func (p *CodeBlockProcessor) isCodeLanguage(lang string) bool {
	return codeLanguages[lang]
}

// ProcessCodeBlocks processes all code blocks in the document (public interface)
// TypeScript original code:
//
//	export function processCodeBlocks(doc: Document, options?: CodeBlockOptions): void {
//	  const processor = new CodeBlockProcessor(doc);
//	  processor.processAllCodeBlocks(options || defaultOptions);
//	}
func ProcessCodeBlocks(doc *goquery.Document, options *CodeBlockProcessingOptions) {
	processor := NewCodeBlockProcessor(doc)
	processor.ProcessCodeBlocks(options)
}
