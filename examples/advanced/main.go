package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kaptinlin/defuddle-go"
)

func main() {
	// HTML with elements that need advanced processing
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Advanced Processing Demo</title>
	</head>
	<body>
		<article>
			<h1>Advanced Element Processing</h1>
			
			<div role="paragraph">This div should become a paragraph.</div>
			
			<h2>Code Example</h2>
			<pre><code class="language-go">
func main() {
    fmt.Println("Hello, World!")
}
			</code></pre>
			
			<h2>ARIA Role List</h2>
			<div role="list">
				<div role="listitem">Item 1</div>
				<div role="listitem">Item 2</div>
			</div>
			
			<p>Math formula: <span class="math">E = mc^2</span></p>
		</article>
	</body>
	</html>
	`

	// Enable advanced element processing
	options := &defuddle.Options{
		ProcessCode:     true,
		ProcessMath:     true,
		ProcessRoles:    true,
		ProcessHeadings: true,
		Debug:           true,
	}

	defuddleInstance, err := defuddle.NewDefuddle(html, options)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	result, err := defuddleInstance.Parse(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("=== Advanced Element Processing ===")
	fmt.Printf("Title: %s\n", result.Title)
	fmt.Printf("Word Count: %d\n", result.WordCount)
	fmt.Printf("Parse Time: %d ms\n", result.ParseTime)

	fmt.Println("\n=== Processed Content ===")
	fmt.Println(result.Content)

	// Show debug info for processing steps
	if result.DebugInfo != nil {
		fmt.Println("\n=== Processing Steps ===")
		for i, step := range result.DebugInfo.ProcessingSteps {
			fmt.Printf("%d. %s: %s\n", i+1, step.Step, step.Description)
		}
	}
}
