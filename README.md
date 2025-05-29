# Defuddle Go

[![Release](https://img.shields.io/github/v/release/kaptinlin/defuddle-go)](https://github.com/kaptinlin/defuddle-go/releases)
[![Test](https://github.com/kaptinlin/defuddle-go/workflows/Test/badge.svg)](https://github.com/kaptinlin/defuddle-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kaptinlin/defuddle-go)](https://goreportcard.com/report/github.com/kaptinlin/defuddle-go)
[![GoDoc](https://godoc.org/github.com/kaptinlin/defuddle-go?status.svg)](https://godoc.org/github.com/kaptinlin/defuddle-go)

A Go implementation of the [Defuddle](https://github.com/kepano/defuddle) TypeScript library for intelligent web content extraction. Defuddle Go extracts clean, readable content from HTML documents using advanced algorithms to remove clutter while preserving meaningful content.

**Available as both a Go library and a command-line tool.**

## Features

- 🧠 **Intelligent Content Extraction**: Advanced algorithms to identify and extract main content
- 🎯 **Site-Specific Extractors**: Built-in support for popular platforms (ChatGPT, Grok, Hacker News, Reddit, etc.)
- 🧹 **Clutter Removal**: Automatically removes ads, navigation, sidebars, and other non-content elements
- 📱 **Mobile-First**: Applies mobile styles for better content detection
- 🔍 **Metadata Extraction**: Extracts titles, descriptions, authors, images, and more
- 🏷️ **Schema.org Support**: Parses structured data using JSON-LD processing
- 📝 **Markdown Conversion**: High-quality HTML to Markdown conversion
- 🔧 **Element Processing**: Advanced processing for code blocks, images, math formulas, and more
- 🐛 **Debug Mode**: Detailed processing information for troubleshooting
- ⚡ **High Performance**: Optimized for Go with efficient DOM processing
- 🖥️ **CLI Tool**: Powerful command-line interface for extracting content

## Installation

### CLI Tool

#### Download Pre-built Binaries
Download the latest binary for your platform from the [releases page](https://github.com/kaptinlin/defuddle-go/releases).

#### Install with Go
```bash
go install github.com/kaptinlin/defuddle-go/cmd@latest
```

#### Install from Source
```bash
git clone https://github.com/kaptinlin/defuddle-go.git
cd defuddle-go
make build-cli
sudo make install-cli
```

### Go Library

```bash
go get github.com/kaptinlin/defuddle-go
```

## CLI Usage

The `defuddle` command-line tool provides a simple interface for extracting content from web pages and HTML files.

### Basic Usage

```bash
# Extract content from a URL
defuddle parse https://example.com/article

# Extract from local HTML file
defuddle parse article.html

# Convert to Markdown
defuddle parse https://example.com/article --markdown

# Get JSON output with metadata
defuddle parse https://example.com/article --json

# Extract specific properties
defuddle parse https://example.com/article --property title
defuddle parse https://example.com/article --property author
defuddle parse https://example.com/article --property description

# Save output to file
defuddle parse https://example.com/article --markdown --output article.md
```

### CLI Options

| Option | Short | Description |
|--------|-------|-------------|
| `--output` | `-o` | Output file path (default: stdout) |
| `--markdown` | `-m` | Convert content to markdown format |
| `--md` | | Alias for --markdown |
| `--json` | `-j` | Output as JSON with metadata and content |
| `--property` | `-p` | Extract a specific property |
| `--debug` | | Enable debug mode |
| `--help` | `-h` | Show help message |
| `--version` | `-v` | Show version information |

### CLI Examples

```bash
# Extract Reddit post title
defuddle parse https://www.reddit.com/r/golang/comments/xyz/... --property title

# Get full JSON metadata
defuddle parse https://news.ycombinator.com/item?id=123456 --json

# Convert article to Markdown and save
defuddle parse https://blog.example.com/post --markdown --output post.md

# Debug parsing process
defuddle parse https://example.com/article --debug
```

## Library Usage

## Library Quick Start

### Basic Content Extraction

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
                <p>It contains multiple paragraphs of text.</p>
            </article>
        </main>
        <aside>Sidebar content</aside>
        <footer>Footer content</footer>
    </body>
    </html>
    `

    // Create Defuddle instance
    defuddleInstance, err := defuddle.NewDefuddle(html, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Parse the content
    result, err := defuddleInstance.Parse(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // Output results
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Author: %s\n", result.Author)
    fmt.Printf("Description: %s\n", result.Description)
    fmt.Printf("Word Count: %d\n", result.WordCount)
    fmt.Printf("Parse Time: %dms\n", result.ParseTime)
    fmt.Printf("Content: %s\n", result.Content)
}
```

### Parsing from URL

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kaptinlin/defuddle-go"
)

func main() {
    options := &defuddle.Options{
        Debug: true,
        URL:   "https://example.com/article",
    }

    result, err := defuddle.ParseFromURL(context.Background(), "https://example.com/article", options)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Content length: %d\n", len(result.Content))
}
```

### Advanced Usage with All Options

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kaptinlin/defuddle-go"
)

func main() {
    html := `<html>...</html>` // Your HTML content here

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
        RemoveExactSelectors:   true,  // Remove exact clutter selectors
        RemovePartialSelectors: true,  // Remove partial clutter selectors
    }

    defuddleInstance, err := defuddle.NewDefuddle(html, options)
    if err != nil {
        log.Fatal(err)
    }

    result, err := defuddleInstance.Parse(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Title: %s\n", result.Title)
    if result.ContentMarkdown != nil {
        fmt.Printf("Markdown content: %s\n", *result.ContentMarkdown)
    }

    if result.DebugInfo != nil {
        fmt.Printf("Processing steps: %d\n", len(result.DebugInfo.ProcessingSteps))
        fmt.Printf("Original elements: %d\n", result.DebugInfo.Statistics.OriginalElementCount)
        fmt.Printf("Final elements: %d\n", result.DebugInfo.Statistics.FinalElementCount)
    }
}
```

## API Reference

### Result Structure

The `Result` object contains the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `Title` | string | Article title |
| `Author` | string | Article author |
| `Description` | string | Article description or summary |
| `Domain` | string | Website domain |
| `Favicon` | string | Website favicon URL |
| `Image` | string | Main image URL |
| `Published` | string | Publication date |
| `Site` | string | Website name |
| `Content` | string | Cleaned HTML content |
| `ContentMarkdown` | *string | Markdown version (if enabled) |
| `WordCount` | int | Word count in extracted content |
| `ParseTime` | int64 | Parse time in milliseconds |
| `SchemaOrgData` | interface{} | Schema.org structured data |
| `MetaTags` | []MetaTag | Document meta tags |
| `ExtractorType` | *string | Extractor type used |
| `DebugInfo` | *DebugInfo | Debug information (if enabled) |

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Debug` | bool | false | Enable debug logging |
| `URL` | string | "" | Source URL for the content |
| `Markdown` | bool | false | Convert content to Markdown |
| `SeparateMarkdown` | bool | false | Keep both HTML and Markdown |
| `RemoveExactSelectors` | bool | true | Remove exact clutter matches |
| `RemovePartialSelectors` | bool | true | Remove partial clutter matches |
| `ProcessCode` | bool | false | Process code blocks |
| `ProcessImages` | bool | false | Process and optimize images |
| `ProcessHeadings` | bool | false | Standardize heading structure |
| `ProcessMath` | bool | false | Process mathematical formulas |
| `ProcessFootnotes` | bool | false | Extract and format footnotes |
| `ProcessRoles` | bool | false | Convert ARIA roles to semantic HTML |

### Core Functions

#### `NewDefuddle(html string, options *Options) (*Defuddle, error)`
Creates a new Defuddle instance from HTML content.

#### `ParseFromURL(ctx context.Context, url string, options *Options) (*Result, error)`
Fetches content from a URL and parses it directly.

#### `Parse(ctx context.Context) (*Result, error)`
Parses the HTML content and returns extracted results.

## Content Processing

### Processing Pipeline

Defuddle Go processes content through these stages:

1. **Schema.org Extraction** - Extracts structured data using JSON-LD
2. **Site-Specific Detection** - Uses specialized extractors when available
3. **Main Content Detection** - Identifies primary content areas
4. **Clutter Removal** - Removes navigation, ads, and decorative elements
5. **Content Standardization** - Normalizes HTML structure
6. **Element Processing** - Processes code, math, images, and footnotes
7. **Markdown Conversion** - Converts to Markdown if requested

### HTML Standardization

#### Headings
- Duplicate H1/H2 headings matching the title are removed
- Heading hierarchy is normalized
- Navigation links within headings are removed

#### Code Blocks
Code blocks are standardized with preserved language information:

```html
<pre><code data-lang="javascript" class="language-javascript">
console.log("Hello, World!");
</code></pre>
```

#### Footnotes
Footnotes are converted to a standard format with proper linking:

```html
Text with footnote<sup id="fnref:1"><a href="#fn:1">1</a></sup>.

<div id="footnotes">
  <ol>
    <li class="footnote" id="fn:1">
      <p>Footnote content <a href="#fnref:1" class="footnote-backref">↩</a></p>
    </li>
  </ol>
</div>
```

## Site-Specific Extractors

Built-in extractors automatically activate for supported platforms:

- **ChatGPT** - Extracts conversation content and metadata
- **Grok** - Extracts AI conversation content  
- **Hacker News** - Extracts posts and comments with proper threading

Custom extractors can be implemented using the `BaseExtractor` interface.

## Examples

The [`examples/`](./examples/) directory contains ready-to-run examples:

- **[Basic](./examples/basic/)** - Simple content extraction
- **[Advanced](./examples/advanced/)** - Full feature demonstration
- **[Markdown](./examples/markdown/)** - HTML to Markdown conversion
- **[Extractors](./examples/extractors/)** - Site-specific extraction
- **[Custom Extractor](./examples/custom_extractor/)** - Building custom extractors

Run examples with:
```bash
cd examples/basic && go run main.go
cd examples/advanced && go run main.go
cd examples/markdown && go run main.go
cd examples/extractors && go run main.go
cd examples/custom_extractor && go run custom_extractor.go
```

## Performance

Typical performance characteristics:

- **Processing Speed**: 5-15ms for standard web pages
- **Memory Usage**: Optimized with object pooling and efficient DOM processing  
- **Concurrent Safe**: Can process multiple documents simultaneously

## Dependencies

- [goquery](https://github.com/PuerkitoBio/goquery) - DOM manipulation and traversal
- [requests](https://github.com/kaptinlin/requests) - HTTP client for URL fetching
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - HTML to Markdown conversion
- [json-gold](https://github.com/piprate/json-gold) - JSON-LD processing

## Contributing

Contributions are welcome. Please open an issue or submit a pull request.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original [Defuddle TypeScript library](https://github.com/kepano/defuddle) by Steph Ango (@kepano)
- Inspired by Mozilla's Readability algorithm 
