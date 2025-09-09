// Package main demonstrates extractors usage.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kaptinlin/defuddle-go"
)

func main() {
	// Reddit content with specific structure
	redditHTML := `
	<html>
	<head>
		<title>Go Programming Discussion</title>
	</head>
	<body>
		<h1>Go Programming Discussion</h1>
		<shreddit-post author="gopher123">
			<div slot="text-body">
				<p>I've been working with Go and really like its simplicity.</p>
				<p>The concurrency model with goroutines is excellent.</p>
			</div>
		</shreddit-post>
		
		<shreddit-comment author="developer456" score="15">
			<div slot="comment">
				<p>I agree! Go's approach to concurrency is very elegant.</p>
			</div>
		</shreddit-comment>
	</body>
	</html>
	`

	// URL triggers Reddit extractor
	options := &defuddle.Options{
		URL:   "https://www.reddit.com/r/programming/comments/abc123/",
		Debug: true,
	}

	defuddleInstance, err := defuddle.NewDefuddle(redditHTML, options)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	result, err := defuddleInstance.Parse(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("=== Site-Specific Extractor Demo ===")
	fmt.Printf("URL: %s\n", options.URL)
	fmt.Printf("Title: %s\n", result.Title)
	fmt.Printf("Site: %s\n", result.Site)
	fmt.Printf("Word Count: %d\n", result.WordCount)
	fmt.Printf("Parse Time: %d ms\n", result.ParseTime)

	fmt.Println("\n=== Extracted Content ===")
	fmt.Println(result.Content)

	// Show which extractor was used (if debug info available)
	if result.DebugInfo != nil {
		fmt.Println("\n=== Extractor Info ===")
		for _, step := range result.DebugInfo.ProcessingSteps {
			if step.Step == "extractor_selection" {
				fmt.Printf("Extractor Used: %s\n", step.Description)
			}
		}
	}
}
