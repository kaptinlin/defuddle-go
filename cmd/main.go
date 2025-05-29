package main

import (
	"context"
	"encoding/json"
	"errors"
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
	version = "0.1.2"
)

// Custom error types for linting compliance
var (
	ErrHTTPRequest      = errors.New("HTTP request failed")
	ErrPropertyNotFound = errors.New("property not found in response")
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	OptionsContextKey ContextKey = "options"
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
		cmd.SetContext(context.WithValue(cmd.Context(), OptionsContextKey, &opts))
		return nil
	}

	rootCmd.AddCommand(parseCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runParse(cmd *cobra.Command, args []string) error {
	opts := cmd.Context().Value(OptionsContextKey).(*ParseOptions)
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
		htmlContent, err = fetchURL(cmd.Context(), source)
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

	// Format output using switch statement
	var output string
	switch {
	case opts.Property != "":
		// Extract specific property
		output, err = extractProperty(result, opts.Property)
		if err != nil {
			return err
		}
	case opts.JSON:
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
	default:
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

// fetchURL fetches content from a URL with context
func fetchURL(ctx context.Context, url string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrHTTPRequest, err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrHTTPRequest, err.Error())
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: HTTP %d: %s", ErrHTTPRequest, resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrHTTPRequest, err.Error())
	}

	return string(body), nil
}

// readFile reads content from a local file
func readFile(filename string) (string, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(absPath) // #nosec G304 - file path is from user input
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// writeFile writes content to a file with secure permissions
func writeFile(filename, content string) error {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	return os.WriteFile(absPath, []byte(content), 0600) // Use secure permissions
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

	return "", fmt.Errorf("%w: \"%s\"", ErrPropertyNotFound, property)
}
