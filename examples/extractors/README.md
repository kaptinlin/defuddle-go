# Site-Specific Extractors

Demonstrates how Defuddle Go automatically selects extractors based on URL patterns.

## Run Example

```bash
cd examples/extractors
go run main.go
```

## What It Does

Shows automatic extractor selection for specific websites:
- **URL Pattern Matching** - Reddit URL triggers Reddit extractor
- **Specialized Processing** - Site-specific extraction logic
- **Content Structure** - Handles Reddit's `shreddit-post` and `shreddit-comment` elements

## Sample Output

```
=== Site-Specific Extractor Demo ===
URL: https://www.reddit.com/r/programming/comments/abc123/
Title: Go Programming Discussion
Site: reddit.com
Word Count: 35
Parse Time: 4 ms

=== Extracted Content ===
<article>
    <h1>Go Programming Discussion</h1>
    <div class="post-content">
        <p>I've been working with Go and really like its simplicity.</p>
        <p>The concurrency model with goroutines is excellent.</p>
    </div>
    <div class="comments">
        <div class="comment">
            <p>I agree! Go's approach to concurrency is very elegant.</p>
        </div>
    </div>
</article>

=== Extractor Info ===
Extractor Used: Reddit extractor for reddit.com content
```

Perfect for understanding how site-specific extraction works automatically. 