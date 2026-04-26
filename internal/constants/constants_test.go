package constants

import "testing"

func TestAccessorsExposeExpectedConstants(t *testing.T) {
	t.Parallel()

	if GetMobileWidth() != 600 {
		t.Fatalf("GetMobileWidth() = %d, want 600", GetMobileWidth())
	}
	if len(GetEntryPointElements()) == 0 {
		t.Fatal("GetEntryPointElements() returned no selectors")
	}
	if len(GetExactSelectors()) == 0 {
		t.Fatal("GetExactSelectors() returned no selectors")
	}
	if len(GetTestAttributes()) == 0 {
		t.Fatal("GetTestAttributes() returned no attributes")
	}
	if len(GetPartialSelectors()) == 0 {
		t.Fatal("GetPartialSelectors() returned no selectors")
	}
	if len(GetFootnoteInlineReferences()) == 0 {
		t.Fatal("GetFootnoteInlineReferences() returned no selectors")
	}
	if len(GetFootnoteListSelectors()) == 0 {
		t.Fatal("GetFootnoteListSelectors() returned no selectors")
	}
	if len(GetBlockElements()) == 0 {
		t.Fatal("GetBlockElements() returned no elements")
	}
	if len(GetInlineElements()) == 0 {
		t.Fatal("GetInlineElements() returned no elements")
	}
	if len(GetAllowedEmptyElements()) == 0 {
		t.Fatal("GetAllowedEmptyElements() returned no elements")
	}
}

func TestElementAndAttributeClassifiers(t *testing.T) {
	t.Parallel()

	if !IsPreserveElement("table") {
		t.Fatal("IsPreserveElement(table) = false, want true")
	}
	if IsPreserveElement("div") {
		t.Fatal("IsPreserveElement(div) = true, want false")
	}
	if !IsInlineElement("span") {
		t.Fatal("IsInlineElement(span) = false, want true")
	}
	if IsInlineElement("section") {
		t.Fatal("IsInlineElement(section) = true, want false")
	}
	if !IsAllowedEmptyElement("img") {
		t.Fatal("IsAllowedEmptyElement(img) = false, want true")
	}
	if IsAllowedEmptyElement("p") {
		t.Fatal("IsAllowedEmptyElement(p) = true, want false")
	}
	if !IsAllowedAttribute("href") {
		t.Fatal("IsAllowedAttribute(href) = false, want true")
	}
	if IsAllowedAttribute("onclick") {
		t.Fatal("IsAllowedAttribute(onclick) = true, want false")
	}
	if !IsAllowedAttributeDebug("class") {
		t.Fatal("IsAllowedAttributeDebug(class) = false, want true")
	}
	if IsAllowedAttributeDebug("style") {
		t.Fatal("IsAllowedAttributeDebug(style) = true, want false")
	}
}
