package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kaptinlin/defuddle-go"
)

func main() {
	// Simple HTML content for extraction
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>My Blog Post</title>
		<meta name="description" content="A simple blog post example">
	</head>
	<body>
		<nav>Navigation</nav>
		<article>
			<h1>Welcome to My Blog</h1>
			<p>This is the main content of my blog post.</p>
			<p>Here's another paragraph with important information.</p>
		</article>
		<footer>Footer content</footer>
	</body>
	</html>
	`

	// Basic extraction with minimal options
	options := &defuddle.Options{
		Debug: true,
	}

	defuddleInstance, err := defuddle.NewDefuddle(html, options)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	result, err := defuddleInstance.Parse(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Display basic results
	fmt.Println("=== Basic Content Extraction ===")
	fmt.Printf("Title: %s\n", result.Title)
	fmt.Printf("Description: %s\n", result.Description)
	fmt.Printf("Word Count: %d\n", result.WordCount)
	fmt.Printf("Parse Time: %d ms\n", result.ParseTime)

	fmt.Println("\n=== Extracted Content ===")
	fmt.Println(result.Content)
}
