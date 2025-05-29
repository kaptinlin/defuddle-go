# Markdown Conversion

Demonstrates HTML to Markdown conversion capabilities of Defuddle Go.

## Run Example

```bash
cd examples/markdown
go run main.go
```

## Features

- **Text Formatting** - Bold, italic, inline code preservation
- **Headings** - H1-H6 heading levels
- **Lists** - Ordered and unordered lists
- **Code Blocks** - Syntax highlighting preservation
- **Tables** - Complete table structure conversion
- **Images & Links** - Proper Markdown syntax
- **Math Formulas** - LaTeX formula escaping

## Key Configuration

```go
options := &defuddle.Options{
    Markdown:         true,
    ProcessCode:      true,
    ProcessImages:    true,
    ProcessHeadings:  true,
    ProcessMath:      true,
    ProcessFootnotes: true,
    ProcessRoles:     true,
}
```

## Sample Output

The example shows 3 conversion scenarios:

1. **Basic Conversion** - Simple HTML to Markdown
2. **Advanced Processing** - All processors + Markdown
3. **Format Comparison** - HTML vs Markdown size comparison

## Performance Benefits

- **Fast Conversion** - 3ms processing time
- **Content Compression** - ~48% size reduction vs HTML
- **Quality Preservation** - All semantic information maintained

## Use Cases

- Content management systems
- Documentation generation
- Content migration between formats
- Static site generation