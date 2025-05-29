# Custom Extractor Development

Demonstrates creating and registering a custom extractor in Defuddle Go.

## Run Example

```bash
cd examples/custom_extractor
go run custom_extractor.go
```

## What It Does

Shows how to build custom extractors for specific sites:
- **Custom Extractor Implementation** - Implement BaseExtractor interface
- **Pattern Registration** - Register for specific URL patterns
- **Specialized Processing** - Custom extraction logic for .post-content
- **Automatic Selection** - Extractor selected based on URL pattern

## Sample Output

```
=== Custom Extractor Demo ===
URL: https://blog.example.com/post/123
Title: Custom Extractor Demo
Site: Custom Blog
Word Count: 18

=== Extracted Content ===
<article>
    <h1>Custom Extractor Demo</h1>
    <div class="post-content">
        <p>This content will be extracted by our custom blog extractor.</p>
        <p>The extractor looks for specific CSS classes like .post-content.</p>
    </div>
</article>

✅ Custom extractor successfully used!
```

## Key Components

```go
// Implement BaseExtractor interface
type CustomBlogExtractor struct {
    *extractors.ExtractorBase
}

// Register for specific URL patterns
extractors.Register(extractors.ExtractorMapping{
    Patterns:  []interface{}{"blog.example.com"},
    Extractor: NewCustomBlogExtractor,
})
```

Perfect for building site-specific extraction logic. 