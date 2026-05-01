// Package main provides the defuddle CLI application.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"

	"github.com/spf13/cobra"

	"github.com/kaptinlin/defuddle-go"
	"github.com/kaptinlin/defuddle-go/extractors"
)

const version = "0.1.3"

// ErrInvalidHeaderFormat is returned when a header flag is not in Key: Value form.
var ErrInvalidHeaderFormat = fmt.Errorf("invalid header format (expected 'Key: Value')")

// ErrDirectoryTraversal is returned when a file path contains a parent-directory segment.
var ErrDirectoryTraversal = fmt.Errorf("invalid file path: directory traversal detected")

// ErrPropertyNotFound is returned when a requested output property is missing.
var ErrPropertyNotFound = fmt.Errorf("property not found in response")

var rootCmd = &cobra.Command{
	Use:     "defuddle",
	Short:   "Extract and structure content from web pages",
	Version: version,
	Long: `defuddle is a CLI tool for extracting and structuring content from web pages.
It can parse HTML, extract metadata, and convert content to various formats.`,
}

var parseCmd = &cobra.Command{
	Use:   "parse <source>",
	Short: "Parse and extract content from a URL or HTML file",
	Long: `Parse content from a URL or local HTML file and extract structured information.
You can output the content in different formats and extract specific properties.`,
	Args: cobra.ExactArgs(1),
	RunE: parseContent,
}

// ParseOptions configures the parse command.
type ParseOptions struct {
	Source    string
	JSON      bool
	Markdown  bool
	Property  string
	Output    string
	UserAgent string
	Headers   []string
	Timeout   time.Duration
	Debug     bool
	Proxy     string
}

func init() {
	extractors.InitializeBuiltins()

	parseCmd.Flags().BoolP("json", "j", false, "Output as JSON with metadata and content")
	parseCmd.Flags().BoolP("markdown", "m", false, "Convert content to markdown format")
	parseCmd.Flags().Bool("md", false, "Alias for --markdown")
	parseCmd.Flags().StringP("property", "p", "", "Extract a specific property (e.g., title, description, domain)")
	parseCmd.Flags().StringP("output", "o", "", "Output file path (default: stdout)")
	parseCmd.Flags().String("user-agent", "", "Custom user agent string")
	parseCmd.Flags().StringArrayP("header", "H", []string{}, "Custom headers in format 'Key: Value'")
	parseCmd.Flags().Duration("timeout", 30*time.Second, "Request timeout")
	parseCmd.Flags().Bool("debug", false, "Enable debug mode")
	parseCmd.Flags().String("proxy", "", "Proxy URL (e.g., http://localhost:8080, socks5://localhost:1080)")

	rootCmd.AddCommand(parseCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseContent(cmd *cobra.Command, args []string) error {
	source := args[0]

	jsonOutput, _ := cmd.Flags().GetBool("json")
	markdown, _ := cmd.Flags().GetBool("markdown")
	mdAlias, _ := cmd.Flags().GetBool("md")
	property, _ := cmd.Flags().GetString("property")
	output, _ := cmd.Flags().GetString("output")
	userAgent, _ := cmd.Flags().GetString("user-agent")
	headers, _ := cmd.Flags().GetStringArray("header")
	timeout, _ := cmd.Flags().GetDuration("timeout")
	debug, _ := cmd.Flags().GetBool("debug")
	proxy, _ := cmd.Flags().GetString("proxy")

	if mdAlias {
		markdown = true
	}

	opts := &ParseOptions{
		Source:    source,
		JSON:      jsonOutput,
		Markdown:  markdown,
		Property:  property,
		Output:    output,
		UserAgent: userAgent,
		Headers:   headers,
		Timeout:   timeout,
		Debug:     debug,
		Proxy:     proxy,
	}

	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	return executeParseContent(opts)
}

func executeParseContent(opts *ParseOptions) error {
	headerMap := make(map[string]string)
	for _, header := range opts.Headers {
		key, value, err := parseHeader(header)
		if err != nil {
			return err
		}
		headerMap[key] = value
	}

	defuddleOpts := &defuddle.Options{
		Debug:            opts.Debug,
		URL:              opts.Source,
		Markdown:         opts.Markdown,
		SeparateMarkdown: opts.Markdown,
	}

	var result *defuddle.Result
	var err error

	if strings.HasPrefix(opts.Source, "http://") || strings.HasPrefix(opts.Source, "https://") {
		ctx := context.Background()
		if opts.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
			defer cancel()
		}
		result, err = defuddle.ParseFromURL(ctx, opts.Source, defuddleOpts)
	} else {
		htmlContent, fileErr := readFile(opts.Source)
		if fileErr != nil {
			return fmt.Errorf("error reading file: %w", fileErr)
		}

		defuddleInstance, createErr := defuddle.NewDefuddle(htmlContent, defuddleOpts)
		if createErr != nil {
			return fmt.Errorf("error creating defuddle instance: %w", createErr)
		}

		ctx := context.Background()
		if opts.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
			defer cancel()
		}
		result, err = defuddleInstance.Parse(ctx)
	}

	if err != nil {
		return fmt.Errorf("error loading content: %w", err)
	}

	if opts.Debug {
		return nil
	}

	if opts.Property != "" {
		value := getProperty(result, opts.Property)
		if value == "" {
			return fmt.Errorf("%w: \"%s\"", ErrPropertyNotFound, opts.Property)
		}
		return writeOutput(opts.Output, value)
	}

	var content string
	switch {
	case opts.JSON:
		jsonData, err := json.Marshal(result, jsontext.Multiline(true))
		if err != nil {
			return fmt.Errorf("error marshaling JSON: %w", err)
		}
		content = string(jsonData)
	case opts.Markdown:
		content = markdownContent(result, opts)
	default:
		content = result.Content
	}

	return writeOutput(opts.Output, content)
}

func markdownContent(result *defuddle.Result, opts *ParseOptions) string {
	if result.ContentMarkdown != nil {
		return *result.ContentMarkdown
	}

	markdownOpts := &defuddle.Options{
		Debug:            false,
		URL:              opts.Source,
		Markdown:         true,
		SeparateMarkdown: true,
	}

	htmlContent := fmt.Sprintf("<html><body>%s</body></html>", result.Content)
	defuddleInstance, err := defuddle.NewDefuddle(htmlContent, markdownOpts)
	if err != nil {
		return result.Content
	}

	ctx := context.Background()
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	markdownResult, err := defuddleInstance.Parse(ctx)
	if err != nil || markdownResult.ContentMarkdown == nil {
		return result.Content
	}

	return *markdownResult.ContentMarkdown
}

func parseHeader(header string) (string, string, error) {
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("%w: %s", ErrInvalidHeaderFormat, header)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func readFile(filename string) (string, error) {
	if err := validateFilePath(filename); err != nil {
		return "", err
	}
	content, err := os.ReadFile(filename) // #nosec G304 - path validated above
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}
	return string(content), nil
}

func validateFilePath(filename string) error {
	if strings.Contains(filename, "..") {
		return ErrDirectoryTraversal
	}
	return nil
}

func writeOutput(filename, content string) error {
	if filename == "" {
		fmt.Print(content)
		return nil
	}

	err := os.WriteFile(filename, []byte(content), 0600) // More secure file permissions
	if err != nil {
		return err
	}

	fmt.Printf("Output written to %s\n", filename)
	return nil
}

func getProperty(result *defuddle.Result, property string) string {
	// Convert to lowercase for case-insensitive matching (matching TypeScript behavior)
	prop := strings.ToLower(property)

	switch prop {
	case "content":
		return result.Content
	case "title":
		return result.Title
	case "description":
		return result.Description
	case "domain":
		return result.Domain
	case "favicon":
		return result.Favicon
	case "image":
		return result.Image
	case "author":
		return result.Author
	case "site":
		return result.Site
	case "published":
		return result.Published
	case "wordcount":
		return strconv.Itoa(result.WordCount)
	case "parsetime":
		return strconv.FormatInt(result.ParseTime, 10)
	case "metatags":
		if result.MetaTags != nil {
			jsonBytes, err := json.Marshal(result.MetaTags)
			if err != nil {
				return ""
			}
			return string(jsonBytes)
		}
		return ""
	case "schemaorgdata":
		if result.SchemaOrgData != nil {
			jsonBytes, err := json.Marshal(result.SchemaOrgData)
			if err != nil {
				return ""
			}
			return string(jsonBytes)
		}
		return "null"
	case "extractortype":
		if result.ExtractorType != nil {
			return *result.ExtractorType
		}
		return ""
	case "contentmarkdown":
		if result.ContentMarkdown != nil {
			return *result.ContentMarkdown
		}
		return ""
	default:
		return ""
	}
}
