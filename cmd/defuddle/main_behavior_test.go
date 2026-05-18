package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spf13/cobra"

	"github.com/kaptinlin/defuddle-go"
)

func TestParseHeaderTrimsKeyAndValue(t *testing.T) {
	t.Parallel()

	key, value, err := parseHeader(" Authorization : Bearer token:with-colon ")
	require.NoError(t, err)

	assert.Equal(t, "Authorization", key)
	assert.Equal(t, "Bearer token:with-colon", value)
}

func TestParseHeaderRejectsMissingSeparator(t *testing.T) {
	t.Parallel()

	_, _, err := parseHeader("Authorization Bearer token")

	require.ErrorIs(t, err, ErrInvalidHeaderFormat)
}

func TestValidateHeadersAcceptsMultipleHeaders(t *testing.T) {
	t.Parallel()

	err := validateHeaders([]string{
		"Authorization: Bearer token",
		"X-Trace: request:with:colons",
	})

	require.NoError(t, err)
}

func TestValidateHeadersRejectsInvalidHeader(t *testing.T) {
	t.Parallel()

	err := validateHeaders([]string{"Authorization: Bearer token", "Invalid"})

	require.ErrorIs(t, err, ErrInvalidHeaderFormat)
}

func TestReadFileValidatesAndReadsContent(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "article.html")
	require.NoError(t, os.WriteFile(path, []byte("<article>Readable</article>"), 0o600))

	content, err := readFile(path)
	require.NoError(t, err)

	assert.Equal(t, "<article>Readable</article>", content)
}

func TestReadFileWrapsFilesystemErrors(t *testing.T) {
	t.Parallel()

	_, err := readFile(filepath.Join(t.TempDir(), "missing.html"))

	require.Error(t, err)
	assert.NotNil(t, errors.Unwrap(err))
}

func TestValidateFilePathAcceptsSafePath(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateFilePath("articles/example.html"))
}

func TestWriteOutputWritesFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "result.txt")
	require.NoError(t, writeOutput(path, "Readable content"))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "Readable content", string(content))
}

func TestWriteOutputPrintsToStdout(t *testing.T) {
	// This test swaps os.Stdout, so it must not run in parallel.
	stdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = stdout
	})

	require.NoError(t, writeOutput("", "Readable content"))
	require.NoError(t, writer.Close())

	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	require.NoError(t, err)
	assert.Equal(t, "Readable content", buf.String())
}

func TestParseContextHonorsPositiveAndNonPositiveTimeouts(t *testing.T) {
	t.Parallel()

	timedCtx, timedCancel := parseContext(time.Second)
	defer timedCancel()
	if _, ok := timedCtx.Deadline(); !ok {
		t.Fatal("parseContext() with timeout did not set deadline")
	}

	plainCtx, plainCancel := parseContext(0)
	defer plainCancel()
	if _, ok := plainCtx.Deadline(); ok {
		t.Fatal("parseContext() without timeout set deadline")
	}
}

func TestJSONPropertyReturnsEmptyStringForUnmarshalableValues(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", jsonProperty(func() {}))
}

func TestGetPropertyReturnsScalarMetadata(t *testing.T) {
	t.Parallel()

	markdown := "# Markdown"
	extractorType := "github"
	result := &defuddle.Result{
		Content:         "HTML content",
		ContentMarkdown: &markdown,
		ExtractorType:   &extractorType,
	}
	result.Title = "Title"
	result.Description = "Description"
	result.Domain = "example.com"
	result.Favicon = "/favicon.ico"
	result.Image = "/image.jpg"
	result.Author = "Author"
	result.Site = "Site"
	result.Published = "2026-05-07"
	result.WordCount = 42
	result.ParseTime = 17

	tests := []struct {
		name     string
		property string
		want     string
	}{
		{name: "content", property: "content", want: "HTML content"},
		{name: "title case insensitive", property: "TITLE", want: "Title"},
		{name: "description", property: "description", want: "Description"},
		{name: "domain", property: "domain", want: "example.com"},
		{name: "favicon", property: "favicon", want: "/favicon.ico"},
		{name: "image", property: "image", want: "/image.jpg"},
		{name: "author", property: "author", want: "Author"},
		{name: "site", property: "site", want: "Site"},
		{name: "published", property: "published", want: "2026-05-07"},
		{name: "word count", property: "wordcount", want: "42"},
		{name: "parse time", property: "parsetime", want: "17"},
		{name: "extractor type", property: "extractortype", want: "github"},
		{name: "content markdown", property: "contentmarkdown", want: "# Markdown"},
		{name: "missing", property: "missing", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, getProperty(result, tc.property))
		})
	}
}

func TestGetPropertyReturnsJSONMetadata(t *testing.T) {
	t.Parallel()

	name := "description"
	content := "Summary"
	result := &defuddle.Result{}
	result.MetaTags = []defuddle.MetaTag{{Name: &name, Content: &content}}
	result.SchemaOrgData = map[string]any{"@type": "Article"}

	assert.Contains(t, getProperty(result, "metatags"), "Summary")
	assert.Contains(t, getProperty(result, "schemaorgdata"), "Article")
}

func TestGetPropertySchemaOrgDataNilReturnsNull(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "null", getProperty(&defuddle.Result{}, "schemaorgdata"))
}

func TestExecuteParseContentReadsFileAndWritesRequestedFormat(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	input := filepath.Join(dir, "article.html")
	output := filepath.Join(dir, "result.md")
	require.NoError(t, os.WriteFile(input, []byte(`<html><head><title>CLI Article</title></head><body><article><h1>CLI Article</h1><p>Readable CLI body content.</p></article></body></html>`), 0o600))

	err := executeParseContent(&ParseOptions{
		Source:   input,
		Markdown: true,
		Output:   output,
		Timeout:  5 * time.Second,
	})
	require.NoError(t, err)

	content, err := os.ReadFile(output)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Readable CLI body content")
	assert.NotContains(t, string(content), "<article")
}

func TestExecuteParseContentReturnsRequestedProperty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	input := filepath.Join(dir, "article.html")
	output := filepath.Join(dir, "title.txt")
	require.NoError(t, os.WriteFile(input, []byte(`<html><head><title>Property Title</title></head><body><article><h1>Property Title</h1><p>Readable property body content.</p></article></body></html>`), 0o600))

	err := executeParseContent(&ParseOptions{
		Source:   input,
		Property: "title",
		Output:   output,
		Timeout:  5 * time.Second,
	})
	require.NoError(t, err)

	content, err := os.ReadFile(output)
	require.NoError(t, err)
	assert.Equal(t, "Property Title", string(content))
}

func TestExecuteParseContentReportsInvalidHeader(t *testing.T) {
	t.Parallel()

	err := executeParseContent(&ParseOptions{
		Source:  "article.html",
		Headers: []string{"Invalid"},
	})

	require.ErrorIs(t, err, ErrInvalidHeaderFormat)
}

func TestExecuteParseContentReportsMissingProperty(t *testing.T) {
	t.Parallel()

	input := filepath.Join(t.TempDir(), "article.html")
	require.NoError(t, os.WriteFile(input, []byte(`<html><head><title>Article</title></head><body><article><p>Readable body content.</p></article></body></html>`), 0o600))

	err := executeParseContent(&ParseOptions{
		Source:   input,
		Property: "unknown",
		Timeout:  5 * time.Second,
	})

	require.ErrorIs(t, err, ErrPropertyNotFound)
}

func TestParseContentHonorsMarkdownAlias(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("markdown", false, "")
	cmd.Flags().Bool("md", true, "")
	cmd.Flags().String("property", "", "")
	cmd.Flags().String("output", filepath.Join(t.TempDir(), "out.md"), "")
	cmd.Flags().String("user-agent", "", "")
	cmd.Flags().StringArray("header", nil, "")
	cmd.Flags().Duration("timeout", 5*time.Second, "")
	cmd.Flags().Bool("debug", false, "")
	cmd.Flags().String("proxy", "", "")

	err := parseContent(cmd, []string{filepath.Join(t.TempDir(), "missing.html")})

	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "error reading file"))
}
