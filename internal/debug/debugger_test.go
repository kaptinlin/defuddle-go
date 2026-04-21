package debug

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebuggerDisabled(t *testing.T) {
	d := NewDebugger(false)
	called := false

	d.StartTimer("parse")
	d.EndTimer("parse")
	d.AddRemovedElement(".ads", "clutter", "div", "text", 1)
	d.AddProcessingStep("parse", "Parse content", 1, "details")
	d.SetStatistics(Statistics{OriginalElementCount: 10})
	d.SetExtractorUsed("example")
	d.LogStep("log", "Logged step", func() int {
		called = true
		return 2
	})

	assert.True(t, called)
	assert.Nil(t, d.GetInfo())
	assert.Equal(t, "Debug mode is disabled", d.GetSummary())
}

func TestDebuggerGetInfoAndSummary(t *testing.T) {
	d := NewDebugger(true)
	d.durations["parse"] = 5 * time.Millisecond
	d.AddProcessingStep("parse", "Parse content", 2, "Trimmed nodes")
	d.AddRemovedElement(".ads", "clutter", "div", "short text", 3)
	d.SetStatistics(Statistics{
		OriginalElementCount: 10,
		FinalElementCount:    7,
		RemovedElementCount:  3,
		WordCount:            42,
		CharacterCount:       256,
		ImageCount:           1,
		LinkCount:            4,
	})
	d.SetExtractorUsed("example")

	info := d.GetInfo()
	require.NotNil(t, info)
	require.Len(t, info.ProcessingSteps, 1)
	require.Len(t, info.RemovedElements, 1)

	assert.Equal(t, int64((5 * time.Millisecond).Nanoseconds()), info.Timings["parse"])
	assert.Equal(t, 5*time.Millisecond, info.ProcessingSteps[0].Duration)
	assert.Equal(t, "example", info.ExtractorUsed)

	summary := d.GetSummary()
	assert.Contains(t, summary, "=== Defuddle Debug Summary ===")
	assert.Contains(t, summary, "Extractor Used: example")
	assert.Contains(t, summary, "Original Elements: 10")
	assert.Contains(t, summary, "Final Elements: 7")
	assert.Contains(t, summary, "1. Parse content (5ms)")
	assert.Contains(t, summary, "Elements affected: 2")
	assert.Contains(t, summary, "Details: Trimmed nodes")
	assert.Contains(t, summary, "parse: 5ms")
	assert.Contains(t, summary, "Removed Elements (1 total):")
	assert.Contains(t, summary, "clutter: 3 elements")
}
