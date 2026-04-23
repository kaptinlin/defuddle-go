package defuddle

import (
	"context"
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
