// Package debug provides debugging functionality for the defuddle content extraction system.
// It tracks removed elements, processing steps, timing information, and parsing statistics.
package debug

import (
	"fmt"
	"strings"
	"time"
)

// Info contains detailed debugging information about the parsing process
type Info struct {
	RemovedElements []RemovedElement `json:"removedElements"`
	ProcessingSteps []ProcessingStep `json:"processingSteps"`
	Timings         map[string]int64 `json:"timings"` // Duration in nanoseconds
	Statistics      Statistics       `json:"statistics"`
	ExtractorUsed   string           `json:"extractorUsed,omitempty"`
}

// RemovedElement represents an element that was removed during processing
type RemovedElement struct {
	Selector    string `json:"selector"`
	Reason      string `json:"reason"`
	Count       int    `json:"count"`
	ElementType string `json:"elementType"`
	TextContent string `json:"textContent,omitempty"`
}

// ProcessingStep represents a step in the content processing pipeline
type ProcessingStep struct {
	Step             string        `json:"step"`
	Description      string        `json:"description"`
	Duration         time.Duration `json:"duration"`
	ElementsAffected int           `json:"elementsAffected"`
	Details          string        `json:"details,omitempty"`
}

// Statistics contains parsing statistics
type Statistics struct {
	OriginalElementCount int `json:"originalElementCount"`
	FinalElementCount    int `json:"finalElementCount"`
	RemovedElementCount  int `json:"removedElementCount"`
	WordCount            int `json:"wordCount"`
	CharacterCount       int `json:"characterCount"`
	ImageCount           int `json:"imageCount"`
	LinkCount            int `json:"linkCount"`
}

// Debugger provides debugging functionality for the parsing process
type Debugger struct {
	enabled         bool
	removedElements []RemovedElement
	processingSteps []ProcessingStep
	timings         map[string]time.Time
	durations       map[string]time.Duration
	statistics      Statistics
	extractorUsed   string
}

// NewDebugger creates a new debugger instance
func NewDebugger(enabled bool) *Debugger {
	return &Debugger{
		enabled:         enabled,
		removedElements: make([]RemovedElement, 0),
		processingSteps: make([]ProcessingStep, 0),
		timings:         make(map[string]time.Time),
		durations:       make(map[string]time.Duration),
	}
}

// IsEnabled returns whether debugging is enabled
func (d *Debugger) IsEnabled() bool {
	return d.enabled
}

// StartTimer starts a timer for the given operation
func (d *Debugger) StartTimer(operation string) {
	if !d.enabled {
		return
	}
	d.timings[operation] = time.Now()
}

// EndTimer ends a timer for the given operation
func (d *Debugger) EndTimer(operation string) {
	if !d.enabled {
		return
	}
	if startTime, exists := d.timings[operation]; exists {
		d.durations[operation] = time.Since(startTime)
		delete(d.timings, operation)
	}
}

// AddRemovedElement records an element that was removed
func (d *Debugger) AddRemovedElement(selector, reason, elementType, textContent string, count int) {
	if !d.enabled {
		return
	}

	// Truncate text content for readability
	if len(textContent) > 100 {
		textContent = textContent[:100] + "..."
	}

	d.removedElements = append(d.removedElements, RemovedElement{
		Selector:    selector,
		Reason:      reason,
		Count:       count,
		ElementType: elementType,
		TextContent: strings.TrimSpace(textContent),
	})
}

// AddProcessingStep records a processing step
func (d *Debugger) AddProcessingStep(step, description string, elementsAffected int, details string) {
	if !d.enabled {
		return
	}

	duration := d.durations[step]

	d.processingSteps = append(d.processingSteps, ProcessingStep{
		Step:             step,
		Description:      description,
		Duration:         duration,
		ElementsAffected: elementsAffected,
		Details:          details,
	})
}

// SetStatistics sets the parsing statistics
func (d *Debugger) SetStatistics(stats Statistics) {
	if !d.enabled {
		return
	}
	d.statistics = stats
}

// SetExtractorUsed sets the name of the extractor that was used
func (d *Debugger) SetExtractorUsed(extractor string) {
	if !d.enabled {
		return
	}
	d.extractorUsed = extractor
}

// GetInfo returns the collected debug information
func (d *Debugger) GetInfo() *Info {
	if !d.enabled {
		return nil
	}

	timings := make(map[string]int64, len(d.durations))
	for operation, duration := range d.durations {
		timings[operation] = duration.Nanoseconds()
	}

	return &Info{
		RemovedElements: d.removedElements,
		ProcessingSteps: d.processingSteps,
		Timings:         timings,
		Statistics:      d.statistics,
		ExtractorUsed:   d.extractorUsed,
	}
}

// GetSummary returns a human-readable summary of the debug information
func (d *Debugger) GetSummary() string {
	if !d.enabled {
		return "Debug mode is disabled"
	}

	var summary strings.Builder
	summary.WriteString("=== Defuddle Debug Summary ===\n\n")

	if d.extractorUsed != "" {
		fmt.Fprintf(&summary, "Extractor Used: %s\n\n", d.extractorUsed)
	}

	summary.WriteString("Statistics:\n")
	fmt.Fprintf(&summary, "  Original Elements: %d\n", d.statistics.OriginalElementCount)
	fmt.Fprintf(&summary, "  Final Elements: %d\n", d.statistics.FinalElementCount)
	fmt.Fprintf(&summary, "  Removed Elements: %d\n", d.statistics.RemovedElementCount)
	fmt.Fprintf(&summary, "  Word Count: %d\n", d.statistics.WordCount)
	fmt.Fprintf(&summary, "  Character Count: %d\n", d.statistics.CharacterCount)
	fmt.Fprintf(&summary, "  Images: %d\n", d.statistics.ImageCount)
	fmt.Fprintf(&summary, "  Links: %d\n\n", d.statistics.LinkCount)

	summary.WriteString("Processing Steps:\n")
	for i, step := range d.processingSteps {
		fmt.Fprintf(&summary, "  %d. %s (%v)\n", i+1, step.Description, step.Duration)
		if step.ElementsAffected > 0 {
			fmt.Fprintf(&summary, "     Elements affected: %d\n", step.ElementsAffected)
		}
		if step.Details != "" {
			fmt.Fprintf(&summary, "     Details: %s\n", step.Details)
		}
	}

	if len(d.durations) > 0 {
		summary.WriteString("\nTiming Information:\n")
		for operation, duration := range d.durations {
			fmt.Fprintf(&summary, "  %s: %v\n", operation, duration)
		}
	}

	if len(d.removedElements) > 0 {
		fmt.Fprintf(&summary, "\nRemoved Elements (%d total):\n", len(d.removedElements))

		reasonCounts := make(map[string]int)
		for _, elem := range d.removedElements {
			reasonCounts[elem.Reason] += elem.Count
		}

		for reason, count := range reasonCounts {
			fmt.Fprintf(&summary, "  %s: %d elements\n", reason, count)
		}
	}

	return summary.String()
}

// LogStep is a convenience method to log a processing step with timing
func (d *Debugger) LogStep(step, description string, fn func() int) {
	if !d.enabled {
		fn()
		return
	}

	d.StartTimer(step)
	elementsAffected := fn()
	d.EndTimer(step)
	d.AddProcessingStep(step, description, elementsAffected, "")
}
