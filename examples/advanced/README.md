# Advanced Element Processing

Demonstrates element processing capabilities in Defuddle Go.

## Run Example

```bash
cd examples/advanced
go run main.go
```

## What It Does

Shows how Defuddle processes different HTML elements:
- **ARIA Role Conversion** - `div[role="paragraph"]` → `<p>`
- **Code Block Processing** - Preserves syntax highlighting
- **Math Formula Processing** - Handles LaTeX formulas
- **Heading Standardization** - Normalizes heading structure

## Sample Output

```
=== Advanced Element Processing ===
Title: Advanced Processing Demo
Word Count: 18
Parse Time: 5 ms

=== Processed Content ===
<article>
    <h1>Advanced Element Processing</h1>
    <p>This div should become a paragraph.</p>
    <h2>Code Example</h2>
    <pre><code class="language-go">
func main() {
    fmt.Println("Hello, World!")
}
    </code></pre>
    <h2>ARIA Role List</h2>
    <ul>
        <li>Item 1</li>
        <li>Item 2</li>
    </ul>
    <p>Math formula: E = mc^2</p>
</article>

=== Processing Steps ===
1. role_processing: Converted 3 ARIA roles to semantic HTML
2. code_processing: Processed 1 code blocks
3. math_processing: Processed 1 mathematical formulas
4. heading_processing: Standardized 3 headings
```

Perfect for understanding advanced content processing features. 