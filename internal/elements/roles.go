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
	p.convertRoleElements(`[role="paragraph"]`, "p")
}

// convertListRoles converts role-based lists to semantic HTML lists
func (p *RoleProcessor) convertListRoles() {
	p.doc.Find(`[role="list"]`).Each(func(_ int, listElement *goquery.Selection) {
		newTag := "ul"
		if p.isOrderedList(listElement) {
			newTag = "ol"
		}

		listElement.Find(`[role="listitem"]`).Each(func(_ int, itemElement *goquery.Selection) {
			p.convertListItem(itemElement)
		})

		p.replaceElementTag(listElement, newTag)
	})
}

// isOrderedList determines if a role-based list should be an ordered list
func (p *RoleProcessor) isOrderedList(listElement *goquery.Selection) bool {
	hasNumbers := false
	listElement.Find(`[role="listitem"]`).EachWithBreak(func(_ int, itemElement *goquery.Selection) bool {
		labelText := strings.TrimSpace(itemElement.Find(".label").First().Text())
		if strings.Contains(labelText, ")") || strings.Contains(labelText, ".") {
			hasNumbers = true
			return false
		}
		return true
	})
	return hasNumbers
}

// convertListItem converts a role="listitem" to <li>
func (p *RoleProcessor) convertListItem(itemElement *goquery.Selection) {
	itemElement.Find(".label").Remove()
	itemElement.Find(`[role="paragraph"]`).Each(func(_ int, s *goquery.Selection) {
		p.replaceElementTag(s, "p")
	})
	p.replaceElementTag(itemElement, "li")
}

// convertButtonRoles converts elements with role="button" to <button> tags
func (p *RoleProcessor) convertButtonRoles() {
	p.convertRoleElements(`[role="button"]`, "button")
}

// convertLinkRoles converts elements with role="link" to <a> tags
func (p *RoleProcessor) convertLinkRoles() {
	p.convertRoleElements(`[role="link"]`, "a")
}

func (p *RoleProcessor) convertRoleElements(selector, tag string) {
	p.doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		p.replaceElementTag(s, tag)
	})
}

// replaceElementTag replaces an element's tag while preserving content and attributes
func (p *RoleProcessor) replaceElementTag(s *goquery.Selection, newTagName string) {
	if s.Length() == 0 {
		return
	}

	attrs := make(map[string]string)
	for _, attr := range s.Get(0).Attr {
		if attr.Key != "role" {
			attrs[attr.Key] = attr.Val
		}
	}

	innerHTML, _ := s.Html()

	attrStrings := make([]string, 0, len(attrs))
	for key, value := range attrs {
		attrStrings = append(attrStrings, fmt.Sprintf(`%s=%q`, key, value))
	}

	var attrString string
	if len(attrStrings) > 0 {
		attrString = " " + strings.Join(attrStrings, " ")
	}

	newHTML := fmt.Sprintf("<%s%s>%s</%s>", newTagName, attrString, innerHTML, newTagName)

	s.ReplaceWithHtml(newHTML)
}
