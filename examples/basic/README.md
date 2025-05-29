# Basic Content Extraction

Simplest content extraction example demonstrating core Defuddle Go functionality.

## Run Example

```bash
cd examples/basic
go run main.go
```

## What It Does

Extracts clean content from a simple HTML document:
- Title and meta information
- Main article content
- Word count and processing statistics

## Sample Output

```
=== Basic Content Extraction ===
Title: My Blog Post
Description: A simple blog post example
Word Count: 15
Parse Time: 2 ms

=== Extracted Content ===
<article>
    <h1>Welcome to My Blog</h1>
    <p>This is the main content of my blog post.</p>
    <p>Here's another paragraph with important information.</p>
</article>
```

Perfect starting point for understanding Defuddle Go basics. 