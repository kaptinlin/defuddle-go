package defuddle

import (
	"context"
	"testing"

	"github.com/kaptinlin/defuddle-go/extractors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractors(t *testing.T) {
	t.Run("GitHub extractor registration and detection", func(t *testing.T) {
		// Initialize extractors
		extractors.InitializeBuiltins()

		// Test GitHub URL detection
		githubHTML := `<html>
			<head>
				<meta name="expected-hostname" content="github.com">
				<meta name="github-keyboard-shortcuts" content="">
				<title>Test Issue · kepano/defuddle</title>
			</head>
			<body>
				<div data-testid="issue-metadata-sticky">Issue metadata</div>
				<div data-testid="issue-title">Test Issue</div>
				<div data-testid="issue-viewer-issue-container">
					<div data-testid="issue-body-viewer">
						<div class="markdown-body">
							<p>This is a test issue body.</p>
						</div>
					</div>
				</div>
			</body>
		</html>`

		defuddleInstance, err := NewDefuddle(githubHTML, &Options{
			URL: "https://github.com/kepano/defuddle/issues/123",
		})
		require.NoError(t, err)

		result, err := defuddleInstance.Parse(context.Background())
		require.NoError(t, err)

		t.Logf("GitHub extraction result: %+v", result)

		// Check if GitHub extractor was used
		require.NotNil(t, result.ExtractorType)
		assert.Equal(t, "github", *result.ExtractorType)

		// Check content extraction
		assert.Contains(t, result.Content, "This is a test issue body")
	})

	t.Run("YouTube extractor with empty videoId", func(t *testing.T) {
		// Test YouTube URL that might have empty videoId
		youtubeHTML := `<html>
			<head>
				<title>YouTube</title>
				<script type="application/ld+json">
				{
					"@type": "VideoObject",
					"name": "Test Video",
					"description": "Test video description",
					"author": "Test Author",
					"uploadDate": "2024-01-01T00:00:00Z"
				}
				</script>
			</head>
			<body>
				<h1>Test Video</h1>
				<p>Test video description</p>
			</body>
		</html>`

		defuddleInstance, err := NewDefuddle(youtubeHTML, &Options{
			URL: "https://youtube.com/watch?v=", // Empty video ID
		})
		require.NoError(t, err)

		result, err := defuddleInstance.Parse(context.Background())
		require.NoError(t, err)

		t.Logf("YouTube extraction result: %+v", result)

		// Should handle empty videoId gracefully
		if result.ExtractorType != nil && *result.ExtractorType == "youtube" {
			// If YouTube extractor was used, check content doesn't have empty iframe
			assert.NotContains(t, result.Content, `src="https://www.youtube.com/embed/"`, "Found empty iframe src, should be handled gracefully")
		}
	})

	t.Run("Twitter extractor safety", func(t *testing.T) {
		twitterHTML := `<html>
			<head><title>Twitter</title></head>
			<body>
				<article data-testid="tweet">
					<div data-testid="tweetText">
						<span>This is a test tweet</span>
					</div>
				</article>
			</body>
		</html>`

		defuddleInstance, err := NewDefuddle(twitterHTML, &Options{
			URL: "https://twitter.com/user/status/123",
		})
		require.NoError(t, err)

		result, err := defuddleInstance.Parse(context.Background())
		require.NoError(t, err)

		t.Logf("Twitter extraction result: %+v", result)

		// Should not crash with document undefined issues
		if result.ExtractorType != nil && *result.ExtractorType == "twitter" {
			assert.Contains(t, result.Content, "test tweet", "Expected tweet content to be extracted")
		}
	})
}
