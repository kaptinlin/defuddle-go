package pool

import (
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// NodeSlicePool provides pooled slices for goquery selections
// This helps reduce memory allocations during DOM traversal
var NodeSlicePool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate with reasonable capacity
		return make([]*goquery.Selection, 0, 100)
	},
}

// StringBuilderPool provides pooled string builders
// This helps reduce memory allocations during string concatenation
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// StringSlicePool provides pooled string slices
// This helps reduce memory allocations during string processing
var StringSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 50)
	},
}

// GetNodeSlice gets a node slice from the pool
func GetNodeSlice() []*goquery.Selection {
	return NodeSlicePool.Get().([]*goquery.Selection)
}

// PutNodeSlice returns a node slice to the pool
func PutNodeSlice(slice []*goquery.Selection) {
	// Clear the slice but keep the underlying array
	slice = slice[:0]
	NodeSlicePool.Put(slice) //nolint:staticcheck // SA6002: slice is intentionally passed by value for pool reuse
}

// GetStringBuilder gets a string builder from the pool
func GetStringBuilder() *strings.Builder {
	sb := StringBuilderPool.Get().(*strings.Builder)
	sb.Reset() // Ensure it's clean
	return sb
}

// PutStringBuilder returns a string builder to the pool
func PutStringBuilder(sb *strings.Builder) {
	// Don't reset here, let GetStringBuilder handle it
	StringBuilderPool.Put(sb)
}

// GetStringSlice gets a string slice from the pool
func GetStringSlice() []string {
	return StringSlicePool.Get().([]string)
}

// PutStringSlice returns a string slice to the pool
func PutStringSlice(slice []string) {
	// Clear the slice but keep the underlying array
	slice = slice[:0]
	StringSlicePool.Put(slice) //nolint:staticcheck // SA6002: slice is intentionally passed by value for pool reuse
}
