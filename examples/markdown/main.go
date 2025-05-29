package main

import (
	"context"
	"fmt"
	"log"

	defuddle "github.com/kaptinlin/defuddle-go"
)

func main() {
	// Complex HTML content containing various element types
	html := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<title>Markdown Conversion Example - Technical Blog Post</title>
		<meta name="description" content="Demonstrates Defuddle's HTML to Markdown conversion features">
		<meta name="author" content="Technical Author">
		<meta property="og:title" content="Markdown Conversion Example">
		<meta property="og:description" content="Complete HTML to Markdown conversion demonstration">
		<meta property="og:image" content="https://example.com/featured.jpg">
	</head>
	<body>
		<article>
			<h1>Markdown Conversion Example</h1>
			<p class="subtitle">Demonstrating Defuddle's powerful HTML to Markdown conversion capabilities</p>
			
			<h2>Basic Text Formatting</h2>
			<p>This is a paragraph containing <strong>bold text</strong> and <em>italic text</em>.</p>
			<p>It can also include <code>inline code</code> and <a href="https://example.com">links</a>.</p>
			
			<h2>List Examples</h2>
			<h3>Unordered List</h3>
			<ul>
				<li>First list item</li>
				<li>Second list item with <strong>bold</strong> text</li>
				<li>Third list item with a <a href="https://example.com">link</a></li>
			</ul>
			
			<h3>Ordered List</h3>
			<ol>
				<li>Step one: Preparation</li>
				<li>Step two: Execute operation</li>
				<li>Step three: Verify results</li>
			</ol>
			
			<h2>Code Block Examples</h2>
			<p>Here's a JavaScript function:</p>
			<pre><code class="language-javascript">
// Calculate Fibonacci sequence
function fibonacci(n) {
    if (n <= 1) {
        return n;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

// Usage example
console.log("Fibonacci 10th term:", fibonacci(10));
			</code></pre>
			
			<p>Python code example:</p>
			<pre><code class="language-python">
def quick_sort(arr):
    """Quick sort algorithm implementation"""
    if len(arr) <= 1:
        return arr
    
    pivot = arr[len(arr) // 2]
    left = [x for x in arr if x < pivot]
    middle = [x for x in arr if x == pivot]
    right = [x for x in arr if x > pivot]
    
    return quick_sort(left) + middle + quick_sort(right)

# Test
numbers = [3, 6, 8, 10, 1, 2, 1]
print("Sorted result:", quick_sort(numbers))
			</code></pre>
			
			<h2>Table Example</h2>
			<table>
				<thead>
					<tr>
						<th>Feature</th>
						<th>HTML</th>
						<th>Markdown</th>
						<th>Status</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>Headings</td>
						<td>&lt;h1&gt;-&lt;h6&gt;</td>
						<td># - ######</td>
						<td>✅ Supported</td>
					</tr>
					<tr>
						<td>Bold</td>
						<td>&lt;strong&gt;</td>
						<td>**text**</td>
						<td>✅ Supported</td>
					</tr>
					<tr>
						<td>Italic</td>
						<td>&lt;em&gt;</td>
						<td>*text*</td>
						<td>✅ Supported</td>
					</tr>
					<tr>
						<td>Code Block</td>
						<td>&lt;pre&gt;&lt;code&gt;</td>
						<td>` + "```" + `language</td>
						<td>✅ Supported</td>
					</tr>
				</tbody>
			</table>
			
			<h2>Images and Media</h2>
			<p>Here are some image examples:</p>
			<img src="https://example.com/hero-image.jpg" alt="Main featured image" width="800" height="400">
			<p><em>Image caption: This is a featured image demonstrating image conversion in Markdown.</em></p>
			
			<img src="https://example.com/icon.png" alt="Small icon" width="32" height="32">
			
			<h2>Quotes and Footnotes</h2>
			<blockquote>
				<p>This is a blockquote used to display important quoted content. Blockquotes in Markdown are represented with the &gt; symbol.</p>
				<p>— Famous Author</p>
			</blockquote>
			
			<p>This text contains footnote references<sup><a href="#fn1">1</a></sup>, and another footnote<sup><a href="#fn2">2</a></sup>.</p>
			
			<h2>Mathematical Formulas</h2>
			<p>Inline math formula: <span class="math">E = mc^2</span></p>
			<p>Block math formula:</p>
			<div class="math-block">
				$$\int_{-\infty}^{\infty} e^{-x^2} dx = \sqrt{\pi}$$
			</div>
			
			<h2>Special Elements</h2>
			<p>Here are some special HTML element conversions:</p>
			
			<div role="paragraph">This div has a paragraph role and should be converted to a paragraph.</div>
			
			<div role="list">
				<div role="listitem">Role list item 1</div>
				<div role="listitem">Role list item 2</div>
				<div role="listitem">Role list item 3</div>
			</div>
			
			<hr>
			
			<h2>Summary</h2>
			<p>Defuddle's Markdown conversion features support:</p>
			<ul>
				<li>✅ Complete text format conversion</li>
				<li>✅ Code block syntax highlighting preservation</li>
				<li>✅ Table structure conversion</li>
				<li>✅ Image and link processing</li>
				<li>✅ Lists and blockquotes</li>
				<li>✅ Mathematical formula escaping</li>
				<li>✅ Semantic element processing</li>
			</ul>
			
			<div id="footnotes">
				<h3>Footnotes</h3>
				<p id="fn1">1. This is the first footnote providing additional explanatory information.</p>
				<p id="fn2">2. This is the second footnote demonstrating multiple footnote handling.</p>
			</div>
		</article>
		
		<!-- Sidebar content should be removed -->
		<aside class="sidebar">
			<div class="advertisement">
				<h3>Advertisement Content</h3>
				<p>This content should be removed by the content scoring algorithm.</p>
			</div>
		</aside>
	</body>
	</html>
	`

	fmt.Println("=== DEFUDDLE MARKDOWN CONVERSION EXAMPLE ===")
	fmt.Println()

	// Example 1: Basic Markdown conversion
	fmt.Println("📝 Example 1: Basic Markdown Conversion")
	fmt.Println("----------------------------------------")

	options1 := &defuddle.Options{
		Markdown: true,
		Debug:    false,
	}

	result1, err := parseAndConvert(html, options1)
	if err != nil {
		log.Fatalf("Example 1 failed: %v", err)
	}

	fmt.Printf("Title: %s\n", result1.Title)
	fmt.Printf("Word Count: %d\n", result1.WordCount)
	fmt.Printf("Parse Time: %d ms\n", result1.ParseTime)
	fmt.Println()

	if result1.ContentMarkdown != nil {
		fmt.Println("Markdown Content:")
		fmt.Println("```markdown")
		fmt.Println(*result1.ContentMarkdown)
		fmt.Println("```")
	}

	// Example 2: Advanced element processing + Markdown conversion
	fmt.Println("🔧 Example 2: Advanced Element Processing + Markdown Conversion")
	fmt.Println("----------------------------------------------------------------")

	options2 := &defuddle.Options{
		Markdown:         true,
		ProcessCode:      true,
		ProcessImages:    true,
		ProcessHeadings:  true,
		ProcessMath:      true,
		ProcessFootnotes: true,
		ProcessRoles:     true,
		Debug:            true,
	}

	result2, err := parseAndConvert(html, options2)
	if err != nil {
		log.Fatalf("Example 2 failed: %v", err)
	}

	fmt.Printf("Title: %s\n", result2.Title)
	fmt.Printf("Word Count: %d\n", result2.WordCount)
	fmt.Printf("Parse Time: %d ms\n", result2.ParseTime)

	if result2.DebugInfo != nil {
		fmt.Printf("Processing Steps: %d\n", len(result2.DebugInfo.ProcessingSteps))
	}

	if result2.ContentMarkdown != nil {
		fmt.Println("Enhanced Processed Markdown Content:")
		fmt.Println("```markdown")
		fmt.Println(*result2.ContentMarkdown)
		fmt.Println("```")
	}

	// Example 3: Compare HTML and Markdown lengths
	fmt.Println("📊 Example 3: Format Comparison")
	fmt.Println("--------------------------------")

	htmlLength := len(result2.Content)
	markdownLength := 0
	if result2.ContentMarkdown != nil {
		markdownLength = len(*result2.ContentMarkdown)
	}

	fmt.Printf("HTML Content Length: %d characters\n", htmlLength)
	fmt.Printf("Markdown Content Length: %d characters\n", markdownLength)

	if markdownLength > 0 {
		ratio := float64(markdownLength) / float64(htmlLength) * 100
		fmt.Printf("Markdown Relative Size: %.1f%%\n", ratio)
		fmt.Printf("Compression Ratio: %.1fx\n", float64(htmlLength)/float64(markdownLength))
	}

	fmt.Println("🎉 Markdown conversion example completed!")
}

// parseAndConvert parses HTML and converts to Markdown
func parseAndConvert(html string, options *defuddle.Options) (*defuddle.Result, error) {
	d, err := defuddle.NewDefuddle(html, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create Defuddle instance: %w", err)
	}

	result, err := d.Parse(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	return result, nil
}
