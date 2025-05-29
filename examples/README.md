# Defuddle Go Examples

Simple examples demonstrating core features of the Defuddle Go content extraction library.

## Available Examples

### 📁 [basic/](./basic/)
**Basic Content Extraction**
- Simple HTML parsing and content extraction
- Metadata extraction and processing statistics
- Perfect starting point for beginners

### 📁 [advanced/](./advanced/)
**Element Processing**
- ARIA role conversion and code block processing
- Math formula handling and heading standardization
- Debug information and processing steps

### 📁 [markdown/](./markdown/)
**HTML to Markdown Conversion**
- Convert HTML content to clean Markdown format
- Text formatting, code blocks, and lists
- Format comparison and compression analysis

### 📁 [extractors/](./extractors/)
**Site-Specific Extractors**
- Automatic extractor selection by URL pattern
- Reddit content extraction example
- Specialized processing for different sites

### 📁 [custom_extractor/](./custom_extractor/)
**Custom Extractor Development**
- Create custom extractors for specific sites
- Pattern registration and BaseExtractor interface
- Site-specific extraction logic implementation

## Quick Start

```bash
# Run any example
go run examples/basic/main.go
go run examples/advanced/main.go
go run examples/extractors/main.go
go run examples/markdown/main.go
go run examples/custom_extractor/main.go
```

## Common Configurations

### Basic Extraction
```go
options := &defuddle.Options{
    Debug: true,
}
```

### Advanced Processing
```go
options := &defuddle.Options{
    ProcessCode:      true,
    ProcessMath:      true,
    ProcessRoles:     true,
    ProcessHeadings:  true,
    Debug:            true,
}
```

### Markdown Conversion
```go
options := &defuddle.Options{
    Markdown: true,
}
```
