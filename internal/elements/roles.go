package elements

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// RoleProcessor handles conversion of ARIA roles to semantic HTML elements
type RoleProcessor struct {
	doc *goquery.Document
}

// RoleProcessingOptions configures role processing behavior
type RoleProcessingOptions struct {
	ConvertParagraphs bool
	ConvertLists      bool
	ConvertButtons    bool
	ConvertLinks      bool
}

// DefaultRoleProcessingOptions returns default options for role processing
func DefaultRoleProcessingOptions() *RoleProcessingOptions {
	return &RoleProcessingOptions{
		ConvertParagraphs: true,
		ConvertLists:      true,
		ConvertButtons:    true,
		ConvertLinks:      true,
	}
}

// NewRoleProcessor creates a new role processor
func NewRoleProcessor(doc *goquery.Document) *RoleProcessor {
	return &RoleProcessor{
		doc: doc,
	}
}

// ProcessRoles processes all role-based elements in the document
func (p *RoleProcessor) ProcessRoles(options *RoleProcessingOptions) {
	if options == nil {
		options = DefaultRoleProcessingOptions()
	}

	if options.ConvertParagraphs {
		p.convertParagraphRoles()
	}

	if options.ConvertLists {
		p.convertListRoles()
	}

	if options.ConvertButtons {
		p.convertButtonRoles()
	}

	if options.ConvertLinks {
		p.convertLinkRoles()
	}
}

// convertParagraphRoles converts elements with role="paragraph" to <p> tags
func (p *RoleProcessor) convertParagraphRoles() {
	p.doc.Find(`[role="paragraph"]`).Each(func(i int, s *goquery.Selection) {
		p.replaceElementTag(s, "p")
	})
}

// convertListRoles converts role-based lists to semantic HTML lists
func (p *RoleProcessor) convertListRoles() {
	// Convert role="list" to <ol> or <ul>
	p.doc.Find(`[role="list"]`).Each(func(i int, listElement *goquery.Selection) {
		// Check if it's an ordered list by looking for numbered items
		isOrdered := p.isOrderedList(listElement)

		var newTag string
		if isOrdered {
			newTag = "ol"
		} else {
			newTag = "ul"
		}

		// Convert list items first
		listElement.Find(`[role="listitem"]`).Each(func(j int, itemElement *goquery.Selection) {
			p.convertListItem(itemElement)
		})

		// Convert the list container
		p.replaceElementTag(listElement, newTag)
	})
}

// isOrderedList determines if a role-based list should be an ordered list
func (p *RoleProcessor) isOrderedList(listElement *goquery.Selection) bool {
	// Look for numbered labels in list items
	hasNumbers := false
	listElement.Find(`[role="listitem"]`).Each(func(i int, itemElement *goquery.Selection) {
		labelElement := itemElement.Find(".label").First()
		if labelElement.Length() > 0 {
			labelText := strings.TrimSpace(labelElement.Text())
			// Check for patterns like "1)", "2.", "1.", etc.
			if strings.Contains(labelText, ")") || strings.Contains(labelText, ".") {
				hasNumbers = true
			}
		}
	})
	return hasNumbers
}

// convertListItem converts a role="listitem" to <li>
func (p *RoleProcessor) convertListItem(itemElement *goquery.Selection) {
	// Remove label elements (like "1)", "2)", etc.)
	itemElement.Find(".label").Remove()

	// Convert content divs to paragraphs if they have role="paragraph"
	itemElement.Find(`[role="paragraph"]`).Each(func(i int, s *goquery.Selection) {
		p.replaceElementTag(s, "p")
	})

	// Convert the list item itself
	p.replaceElementTag(itemElement, "li")
}

// convertButtonRoles converts elements with role="button" to <button> tags
func (p *RoleProcessor) convertButtonRoles() {
	p.doc.Find(`[role="button"]`).Each(func(i int, s *goquery.Selection) {
		p.replaceElementTag(s, "button")
	})
}

// convertLinkRoles converts elements with role="link" to <a> tags
func (p *RoleProcessor) convertLinkRoles() {
	p.doc.Find(`[role="link"]`).Each(func(i int, s *goquery.Selection) {
		p.replaceElementTag(s, "a")
	})
}

// replaceElementTag replaces an element's tag while preserving content and attributes
func (p *RoleProcessor) replaceElementTag(s *goquery.Selection, newTagName string) {
	if s.Length() == 0 {
		return
	}

	// Get current attributes (excluding role)
	attrs := make(map[string]string)
	for _, attr := range s.Get(0).Attr {
		if attr.Key != "role" { // Remove role attribute
			attrs[attr.Key] = attr.Val
		}
	}

	// Get inner HTML
	innerHTML, _ := s.Html()

	// Build attribute string
	attrStrings := make([]string, 0, len(attrs)) // Pre-allocate with capacity
	for key, value := range attrs {
		attrStrings = append(attrStrings, fmt.Sprintf(`%s="%s"`, key, value))
	}

	var attrString string
	if len(attrStrings) > 0 {
		attrString = " " + strings.Join(attrStrings, " ")
	}

	newHTML := fmt.Sprintf("<%s%s>%s</%s>", newTagName, attrString, innerHTML, newTagName)

	// Replace the element
	s.ReplaceWithHtml(newHTML)
}
