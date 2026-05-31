package defuddle

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kaptinlin/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><head><title>Fetched Title</title></head><body><article><h1>Fetched Title</h1><p>Fetched body content from test server.</p></article></body></html>`))
	}))
	defer server.Close()

	options := &Options{}
	result, err := ParseFromURL(context.Background(), server.URL, options)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, server.URL, options.URL)
	assert.Equal(t, "Fetched Title", result.Title)
	assert.Contains(t, result.Content, "Fetched body content from test server")
}

func TestParseFromURLUsesDefaultRequestsClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Mozilla/5.0 (compatible; Defuddle/1.0; +https://github.com/kaptinlin/defuddle-go)", r.Header.Get("User-Agent"))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><head><title>Default Client Title</title></head><body><article><p>Default client content.</p></article></body></html>`))
	}))
	defer server.Close()

	result, err := ParseFromURL(context.Background(), server.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "Default Client Title", result.Title)
	assert.Contains(t, result.Content, "Default client content")
}

func TestParseFromURLUsesCustomRequestsClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-agent", r.Header.Get("User-Agent"))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><head><title>Custom Client Title</title></head><body><article><p>Custom client content.</p></article></body></html>`))
	}))
	defer server.Close()

	options := &Options{
		Client: requests.New(requests.WithUserAgent("test-agent")),
	}

	result, err := ParseFromURL(context.Background(), server.URL, options)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, server.URL, options.URL)
	assert.Equal(t, "Custom Client Title", result.Title)
	assert.Contains(t, result.Content, "Custom client content")
}

func TestParseFromURLReturnsErrorForHTTPErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`<html><body><article><p>Temporary outage page.</p></article></body></html>`))
	}))
	defer server.Close()

	result, err := ParseFromURL(context.Background(), server.URL, nil)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "503 Service Unavailable")
}

func TestParseFromURLDecodesDeclaredCharset(t *testing.T) {
	t.Parallel()

	body := []byte("<html><head><title>Caf\xe9 Story</title><meta name=\"description\" content=\"Cr\xe8me summary\"></head><body><article><h1>Caf\xe9 Story</h1><p>Cr\xe8me br\xfbl\xe9e article body.</p></article></body></html>")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=iso-8859-1")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	result, err := ParseFromURL(context.Background(), server.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "Café Story", result.Title)
	assert.Equal(t, "Crème summary", result.Description)
	assert.Contains(t, result.Content, "Crème brûlée article body")
}

func TestParseFromURLNilOptionsUsesDefaultSelectorCleanup(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><head><title>Remote Defaults</title></head><body>
			<nav>Remote navigation clutter</nav>
			<main><article><h1>Remote Defaults</h1><p>Readable remote article body for default cleanup.</p></article></main>
			<footer>Remote footer clutter</footer>
		</body></html>`))
	}))
	defer server.Close()

	result, err := ParseFromURL(context.Background(), server.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Contains(t, result.Content, "Readable remote article body")
	assert.NotContains(t, result.Content, "Remote navigation clutter")
	assert.NotContains(t, result.Content, "Remote footer clutter")
}

func TestParseFromURLPreservesProvidedOptionsURLForMetadata(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><head><title>Logical URL</title><link rel="icon" href="/icon.svg"></head><body><article><h1>Logical URL</h1><p>Readable logical URL article body.</p></article></body></html>`))
	}))
	defer server.Close()

	options := &Options{URL: "https://www.example.com/articles/story"}
	result, err := ParseFromURL(context.Background(), server.URL, options)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "https://www.example.com/articles/story", options.URL)
	assert.Equal(t, "example.com", result.Domain)
	assert.Equal(t, "https://www.example.com/icon.svg", result.Favicon)
}

func TestParseFromURLHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be reached with a canceled context")
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := ParseFromURL(ctx, server.URL, nil)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, context.Canceled), "error = %v", err)
}
