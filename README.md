# Defuddle Go

A Go port of the [Defuddle](https://github.com/kepano/defuddle) TypeScript library for intelligent web content extraction. Defuddle Go provides clean, readable content extraction from HTML documents with advanced algorithms for removing clutter and preserving meaningful content.

## Features

Defuddle Go aims to output clean and consistent HTML documents with enhanced Go-specific optimizations:

- 🧠 **Intelligent Content Extraction**: Advanced algorithms to identify and extract main content
- 🎯 **Site-Specific Extractors**: Built-in support for popular platforms (Twitter, YouTube, Reddit, etc.)
- 🧹 **Clutter Removal**: Automatically removes ads, navigation, sidebars, and other non-content elements
- 📱 **Mobile-First**: Applies mobile styles for better content detection
- 🔍 **Metadata Extraction**: Extracts titles, descriptions, authors, images, and more
- 🏷️ **Schema.org Support**: Parses structured data using JSON-LD processing
- 📝 **Markdown Conversion**: High-quality HTML to Markdown conversion
- 🔧 **Element Processing**: Advanced processing for code blocks, images, math formulas, and more
- 🐛 **Debug Mode**: Detailed processing information for troubleshooting
- ⚡ **High Performance**: Optimized for Go 1.21+ with efficient DOM processing

## Installation

```bash
go get github.com/kaptinlin/defuddle-go
```

## Usage

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kaptinlin/defuddle-go"
)

func main() {
    html := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Sample Article</title>
        <meta name="description" content="This is a sample article">
        <meta name="author" content="John Doe">
    </head>
    <body>
        <header>Navigation</header>
        <main>
            <article>
                <h1>Sample Article</h1>
                <p>This is the main content of the article.</p>
            </article>
        </main>
        <aside>Sidebar content</aside>
    </body>
    </html>
    `

    // Create Defuddle instance and parse
    defuddleInstance, err := defuddle.NewDefuddle(html, nil)
    if err != nil {
        log.Fatal(err)
    }

    result, err := defuddleInstance.Parse(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // Access the content and metadata
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Author: %s\n", result.Author)
    fmt.Printf("Word Count: %d\n", result.WordCount)
    fmt.Printf("Content: %s\n", result.Content)
}
```

### Parsing from URL

```go
result, err := defuddle.ParseFromURL(context.Background(), "https://example.com/article", &defuddle.Options{
    URL: "https://example.com/article",
})
if err != nil {
    log.Fatal(err)
}
```

### With Options

```go
options := &defuddle.Options{
    Debug:                  true,  // Enable debug mode
    Markdown:               true,  // Convert content to Markdown
    SeparateMarkdown:       true,  // Keep both HTML and Markdown
    URL:                    "https://example.com/article",
    ProcessCode:            true,  // Process code blocks
    ProcessImages:          true,  // Filter and optimize images
    ProcessHeadings:        true,  // Standardize headings
    ProcessMath:            true,  // Handle mathematical formulas
    ProcessFootnotes:       true,  // Extract footnotes
    ProcessRoles:           true,  // Convert ARIA roles to semantic HTML
}

result, err := defuddleInstance.Parse(context.Background())
```

## Response

Defuddle Go returns a `Result` object with the following properties:

| Property | Type | Description |
|----------|------|-------------|
| `Title` | string | Title of the article |
| `Author` | string | Author of the article |
| `Description` | string | Description or summary of the article |
| `Domain` | string | Domain name of the website |
| `Favicon` | string | URL of the website's favicon |
| `Image` | string | URL of the article's main image |
| `Published` | string | Publication date of the article |
| `Site` | string | Name of the website |
| `Content` | string | Cleaned HTML content |
| `ContentMarkdown` | *string | Markdown version of content (if enabled) |
| `WordCount` | int | Total number of words in the extracted content |
| `ParseTime` | int64 | Time taken to parse the page in milliseconds |
| `SchemaOrgData` | interface{} | Raw schema.org data extracted from the page |
| `MetaTags` | []MetaTag | Meta tags from the document |
| `ExtractorType` | *string | Type of extractor used (if any) |
| `DebugInfo` | *DebugInfo | Debug information (if debug mode enabled) |

## Options

Configure Defuddle's behavior with these options:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Debug` | bool | false | Enable debug logging and preserve structure |
| `URL` | string | "" | URL of the page being parsed |
| `Markdown` | bool | false | Convert content to Markdown |
| `SeparateMarkdown` | bool | false | Keep both HTML and Markdown versions |
| `RemoveExactSelectors` | bool | true | Remove elements matching exact clutter selectors |
| `RemovePartialSelectors` | bool | true | Remove elements matching partial clutter selectors |
| `ProcessCode` | bool | false | Process and standardize code blocks |
| `ProcessImages` | bool | false | Filter small images and optimize content |
| `ProcessHeadings` | bool | false | Standardize heading hierarchy |
| `ProcessMath` | bool | false | Process mathematical formulas |
| `ProcessFootnotes` | bool | false | Extract and standardize footnotes |
| `ProcessRoles` | bool | false | Convert ARIA roles to semantic HTML |

### Debug Mode

Enable debug mode for detailed processing information:

```go
options := &defuddle.Options{
    Debug: true,
}

result, err := defuddleInstance.Parse(context.Background())
if result.DebugInfo != nil {
    fmt.Printf("Processing steps: %d\n", len(result.DebugInfo.ProcessingSteps))
    fmt.Printf("Original elements: %d\n", result.DebugInfo.Statistics.OriginalElementCount)
    fmt.Printf("Final elements: %d\n", result.DebugInfo.Statistics.FinalElementCount)
}
```

Debug mode provides:
- Detailed console logging about the parsing process
- Processing step information
- Element count statistics
- Performance metrics

## Content Processing

Defuddle Go performs comprehensive content processing to ensure clean, consistent output:

### Content Processing Pipeline

1. **Schema.org Extraction**: Processes structured data using JSON-LD
2. **Mobile Styles**: Applies mobile CSS for better content detection
3. **Main Content Detection**: Identifies the primary content area
4. **Site-Specific Extraction**: Uses specialized extractors when available
5. **Image Processing**: Removes small/decorative images, preserves meaningful content
6. **Hidden Element Removal**: Removes elements hidden by CSS
7. **Content Scoring**: Scores and removes low-value content blocks
8. **Selector-Based Removal**: Removes known clutter elements
9. **Element Processing**: Processes code blocks, math formulas, footnotes, etc.
10. **Content Standardization**: Normalizes HTML structure
11. **Markdown Conversion**: Converts to Markdown format if enabled

### HTML Standardization

Defuddle Go standardizes HTML elements for consistent output:

#### Headings
- The first H1 or H2 heading is removed if it matches the title
- H1s are converted to H2s
- Anchor links in headings are removed

#### Code Blocks
Code blocks are standardized with language information preserved:

```html
<pre>
  <code data-lang="js" class="language-js">
    // code content
  </code>
</pre>
```

#### Footnotes
Inline references and footnotes are converted to a standard format:

```html
Text with footnote<sup id="fnref:1"><a href="#fn:1">1</a></sup>.

<div id="footnotes">
  <ol>
    <li class="footnote" id="fn:1">
      <p>Footnote content.&nbsp;<a href="#fnref:1" class="footnote-backref">↩</a></p>
    </li>
  </ol>
</div>
```

#### Math Elements
Mathematical content is processed and preserved with proper formatting.

## Site-Specific Extractors

Defuddle Go includes built-in extractors for popular platforms that automatically activate when parsing content from supported sites:

- **Twitter/X**: Extracts tweet content and metadata
- **YouTube**: Extracts video information and embeds  
- **Reddit**: Extracts post content and comments
- **Hacker News**: Extracts story and comment content

Extractors provide enhanced extraction quality for specific site structures and layouts.

## Performance

Defuddle Go is optimized for high performance:

- **Efficient DOM Processing**: Uses goquery for fast HTML parsing
- **Minimal Memory Allocation**: Optimized memory usage patterns
- **Concurrent Processing**: Leverages Go's concurrency where applicable
- **Fast Processing**: Typical processing times of 5-15ms for most documents

## Examples

See the [`examples/`](./examples/) directory for comprehensive usage examples:

- **[Basic Example](./examples/basic/)**: Simple content extraction
- **[Advanced Example](./examples/advanced/)**: All processors and Markdown conversion
- **[Markdown Example](./examples/markdown/)**: HTML to Markdown conversion focus
- **[Extractors Example](./examples/extractors/)**: Site-specific extractors
- **[Custom Extractor Example](./examples/custom_extractor/)**: Building custom extractors

## API Reference

### Functions

#### `NewDefuddle(html string, options *Options) (*Defuddle, error)`
Creates a new Defuddle instance from HTML content.

#### `ParseFromURL(ctx context.Context, url string, options *Options) (*Result, error)`
Fetches and parses content from a URL.

#### `Parse(ctx context.Context) (*Result, error)`
Parses the HTML content and extracts clean, readable content.

### Types

Complete type definitions are available in the source code. Key types include `Options`, `Result`, `Metadata`, and `MetaTag`.

## Dependencies

- [goquery](https://github.com/PuerkitoBio/goquery) - jQuery-like DOM manipulation
- [requests](https://github.com/kaptinlin/requests) - HTTP client for URL fetching
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - HTML to Markdown conversion
- [json-gold](https://github.com/piprate/json-gold) - JSON-LD processing for schema.org data
- Standard library packages for logging and utilities

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Original [Defuddle TypeScript library](https://github.com/kepano/defuddle) by Steph Ango (@kepano)
- Inspired by Mozilla's Readability algorithm
- Built with the excellent goquery library 