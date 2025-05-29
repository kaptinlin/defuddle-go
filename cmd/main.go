package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kaptinlin/defuddle-go"
	"github.com/spf13/cobra"
)

var (
	version = "0.6.4"
)

// ParseOptions holds all CLI options for the parse command
type ParseOptions struct {
	Output   string
	Markdown bool
	MD       bool
	JSON     bool
	Debug    bool
	Property string
}

// JSONOutput represents the JSON output structure matching TypeScript version
type JSONOutput struct {
	Content       string      `json:"content"`
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	Domain        string      `json:"domain"`
	Favicon       string      `json:"favicon"`
	Image         string      `json:"image"`
	MetaTags      interface{} `json:"metaTags"`
	ParseTime     int64       `json:"parseTime"`
	Published     string      `json:"published"`
	Author        string      `json:"author"`
	Site          string      `json:"site"`
	SchemaOrgData interface{} `json:"schemaOrgData"`
	WordCount     int         `json:"wordCount"`
}

func main() {
	rootCmd := &cobra.Command{
		Use:     "defuddle",
		Short:   "Extract article content from web pages",
		Long:    "Command line interface for Defuddle - extract clean HTML, markdown and metadata from web pages.",
		Version: version,
	}

	parseCmd := &cobra.Command{
		Use:   "parse <source>",
		Short: "Parse HTML content from a file or URL",
		Long: `Parse HTML content from a file or URL and extract clean article content.

The source can be either:
  - A local HTML file path
  - A URL (http:// or https://)

Examples:
  defuddle parse article.html
  defuddle parse https://example.com/article --md
  defuddle parse article.html --json
  defuddle parse article.html --property title`,
		Args: cobra.ExactArgs(1),
		RunE: runParse,
	}

	var opts ParseOptions

	// Add flags matching TypeScript version exactly
	parseCmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path (default: stdout)")
	parseCmd.Flags().BoolVarP(&opts.Markdown, "markdown", "m", false, "Convert content to markdown format")
	parseCmd.Flags().BoolVar(&opts.MD, "md", false, "Alias for --markdown")
	parseCmd.Flags().BoolVarP(&opts.JSON, "json", "j", false, "Output as JSON with metadata and content")
	parseCmd.Flags().StringVarP(&opts.Property, "property", "p", "", "Extract a specific property (e.g., title, description, domain)")
	parseCmd.Flags().BoolVar(&opts.Debug, "debug", false, "Enable debug mode")

	// Store options in context for access in RunE
	parseCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		cmd.SetContext(context.WithValue(cmd.Context(), "options", &opts))
		return nil
	}

	rootCmd.AddCommand(parseCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runParse(cmd *cobra.Command, args []string) error {
	opts := cmd.Context().Value("options").(*ParseOptions)
	source := args[0]

	// Handle --md alias
	if opts.MD {
		opts.Markdown = true
	}

	// Read content from source
	var htmlContent string
	var sourceURL string
	var err error

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		htmlContent, err = fetchURL(source)
		sourceURL = source
	} else {
		htmlContent, err = readFile(source)
	}

	if err != nil {
		return fmt.Errorf("error loading content: %w", err)
	}

	// Configure defuddle options
	defuddleOpts := &defuddle.Options{
		Debug:            opts.Debug,
		Markdown:         opts.Markdown,
		SeparateMarkdown: opts.Markdown,
		URL:              sourceURL,
	}

	// Parse content
	defuddleInstance, err := defuddle.NewDefuddle(htmlContent, defuddleOpts)
	if err != nil {
		return fmt.Errorf("error creating defuddle instance: %w", err)
	}

	result, err := defuddleInstance.Parse(context.Background())
	if err != nil {
		return fmt.Errorf("error during parsing: %w", err)
	}

	// If in debug mode, don't show content output (matching TypeScript behavior)
	if opts.Debug {
		return nil
	}

	// Format output
	var output string

	if opts.Property != "" {
		// Extract specific property
		output, err = extractProperty(result, opts.Property)
		if err != nil {
			return err
		}
	} else if opts.JSON {
		// JSON output matching TypeScript structure
		jsonOutput := JSONOutput{
			Content:       result.Content,
			Title:         result.Title,
			Description:   result.Description,
			Domain:        result.Domain,
			Favicon:       result.Favicon,
			Image:         result.Image,
			MetaTags:      result.MetaTags,
			ParseTime:     result.ParseTime,
			Published:     result.Published,
			Author:        result.Author,
			Site:          result.Site,
			SchemaOrgData: result.SchemaOrgData,
			WordCount:     result.WordCount,
		}

		jsonBytes, err := json.MarshalIndent(jsonOutput, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling JSON: %w", err)
		}
		output = string(jsonBytes)
	} else {
		// Default: return content (HTML or Markdown)
		if opts.Markdown && result.ContentMarkdown != nil {
			output = *result.ContentMarkdown
		} else {
			output = result.Content
		}
	}

	// Handle output
	if opts.Output != "" {
		err := writeFile(opts.Output, output)
		if err != nil {
			return fmt.Errorf("error writing output file: %w", err)
		}
		fmt.Printf("Output written to %s\n", opts.Output)
	} else {
		fmt.Print(output)
	}

	return nil
}

// fetchURL fetches content from a URL
func fetchURL(url string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// readFile reads content from a local file
func readFile(filename string) (string, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// writeFile writes content to a file
func writeFile(filename, content string) error {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	return os.WriteFile(absPath, []byte(content), 0644)
}

// extractProperty extracts a specific property from the result
func extractProperty(result *defuddle.Result, property string) (string, error) {
	property = strings.ToLower(property)

	// Direct mapping of common properties
	switch property {
	case "title":
		return result.Title, nil
	case "description":
		return result.Description, nil
	case "domain":
		return result.Domain, nil
	case "favicon":
		return result.Favicon, nil
	case "image":
		return result.Image, nil
	case "published":
		return result.Published, nil
	case "author":
		return result.Author, nil
	case "site":
		return result.Site, nil
	case "content":
		return result.Content, nil
	case "parsetime":
		return fmt.Sprintf("%d", result.ParseTime), nil
	case "wordcount":
		return fmt.Sprintf("%d", result.WordCount), nil
	case "extractortype":
		if result.ExtractorType != nil {
			return *result.ExtractorType, nil
		}
		return "", nil
	case "contentmarkdown":
		if result.ContentMarkdown != nil {
			return *result.ContentMarkdown, nil
		}
		return "", nil
	}

	return "", fmt.Errorf("property \"%s\" not found in response", property)
}
