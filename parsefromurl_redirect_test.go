package defuddle

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFromURLUsesRedirectTargetForRelativeMetadata(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			http.Redirect(w, r, "/articles/story/", http.StatusFound)
		case "/articles/story/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<html><head><title>Redirected Article</title><link rel="icon" href="icon.svg"></head><body><article><h1>Redirected Article</h1><p>Readable redirected article body.</p></article></body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	options := &Options{}
	result, err := ParseFromURL(context.Background(), server.URL+"/start", options)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, server.URL+"/articles/story/", options.URL)
	assert.Equal(t, server.URL+"/articles/story/icon.svg", result.Favicon)
}

func TestParseFromURLHTTPStatusErrorUsesRedirectTarget(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			http.Redirect(w, r, "/unavailable", http.StatusFound)
		case "/unavailable":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`<html><body><p>Temporary outage page.</p></body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	result, err := ParseFromURL(context.Background(), server.URL+"/start", nil)

	require.ErrorIs(t, err, ErrHTTPStatus)
	assert.Nil(t, result)

	var statusErr *HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, server.URL+"/unavailable", statusErr.URL)
}
