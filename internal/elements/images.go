// Package elements provides enhanced element processing functionality
// This module handles image processing including optimization, lazy loading,
// responsive processing, and Alt text generation
package elements

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Pre-compiled regex patterns for image filename processing.
var (
	fileExtRe       = regexp.MustCompile(`\.[^.]+$`)
	separatorsRe    = regexp.MustCompile(`[-_]`)
	camelCaseRe     = regexp.MustCompile(`([a-z])([A-Z])`)
	imgWhitespaceRe = regexp.MustCompile(`\s+`)
)

/*
TypeScript source code (images.ts, 977 lines):

This module provides comprehensive image processing functionality including:
- Image optimization and responsive processing
- Lazy loading implementation
- Alt text generation and enhancement
- Image size detection and optimization
- Broken image handling
- Figure and caption processing

Key functions:
- processImages(): Main processing function for all images
- optimizeImage(): Image optimization and responsive handling
- generateAltText(): Alt text generation from context
- processLazyLoading(): Lazy loading implementation
- handleResponsiveImages(): Responsive image processing
*/

// ImageProcessor handles image processing and enhancement
// TypeScript original code:
// const b64DataUrlRegex = /^data:image\/([^;]+);base64,/;
// const srcsetPattern = /\.(jpg|jpeg|png|webp)\s+\d/;
// const srcPattern = /^\s*\S+\.(jpg|jpeg|png|webp)\S*\s*$/;
// const imageUrlPattern = /\.(jpg|jpeg|png|webp|gif|avif)(\?.*)?$/i;
// const widthPattern = /\s(\d+)w/;
// const dprPattern = /dpr=(\d+(?:\.\d+)?)/;
// const urlPattern = /^([^\s]+)/;
// const filenamePattern = /^[\w\-\.\/\\]+\.(jpg|jpeg|png|gif|webp|svg)$/i;
// const datePattern = /^\d{4}-\d{2}-\d{2}$/;
//
// export const imageRules = [
//
//	{
//	  selector: 'picture',
//	  element: 'picture',
//	  transform: (el: Element, doc: Document): Element => { ... }
//	},
//	// ... other image processing rules
//
// ];
type ImageProcessor struct {
	doc *goquery.Document
}

// ImageProcessingOptions contains options for image processing
// TypeScript original code:
//
//	interface ImageProcessingOptions {
//	  enableLazyLoading?: boolean;
//	  enableResponsive?: boolean;
//	  generateAltText?: boolean;
//	  optimizeImages?: boolean;
//	  removeSmallImages?: boolean;
//	  minImageWidth?: number;
//	  minImageHeight?: number;
//	  maxImageWidth?: number;
//	  maxImageHeight?: number;
//	}
type ImageProcessingOptions struct {
	EnableLazyLoading bool
	EnableResponsive  bool
	GenerateAltText   bool
	OptimizeImages    bool
	RemoveSmallImages bool
	MinImageWidth     int
	MinImageHeight    int
	MaxImageWidth     int
	MaxImageHeight    int
}

// DefaultImageProcessingOptions returns default options for image processing
// TypeScript original code:
//
//	const defaultImageOptions: ImageProcessingOptions = {
//	  enableLazyLoading: true,
//	  enableResponsive: true,
//	  generateAltText: true,
//	  optimizeImages: true,
//	  removeSmallImages: true,
//	  minImageWidth: 50,
//	  minImageHeight: 50,
//	  maxImageWidth: 1200,
//	  maxImageHeight: 800
//	};
func DefaultImageProcessingOptions() *ImageProcessingOptions {
	return &ImageProcessingOptions{
		EnableLazyLoading: true,
		EnableResponsive:  true,
		GenerateAltText:   true,
		OptimizeImages:    true,
		RemoveSmallImages: true,
		MinImageWidth:     50,
		MinImageHeight:    50,
		MaxImageWidth:     1200,
		MaxImageHeight:    800,
	}
}

// NewImageProcessor creates a new image processor
// TypeScript original code:
//
//	class ImageProcessor {
//	  constructor(private document: Document) {}
//	}
func NewImageProcessor(doc *goquery.Document) *ImageProcessor {
	return &ImageProcessor{
		doc: doc,
	}
}

// ProcessImages processes all images in the document
// TypeScript original code:
//
//	processImages(options?: ImageProcessingOptions): void {
//	  const imgs = this.document.querySelectorAll('img');
//	  imgs.forEach(img => this.processImage(img, options));
//
//	  const figures = this.document.querySelectorAll('figure');
//	  figures.forEach(figure => this.processFigure(figure, options));
//
//	  const pictures = this.document.querySelectorAll('picture');
//	  pictures.forEach(picture => this.processPicture(picture, options));
//
//	  if (options?.removeSmallImages) {
//	    this.removeSmallImages(options);
//	  }
//	}
func (p *ImageProcessor) ProcessImages(options *ImageProcessingOptions) {
	if options == nil {
		options = DefaultImageProcessingOptions()
	}

	// Process all img elements
	p.doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		p.processImage(s, options)
	})

	// Process figure elements
	p.doc.Find("figure").Each(func(_ int, s *goquery.Selection) {
		p.processFigure(s, options)
	})

	// Process picture elements
	p.doc.Find("picture").Each(func(_ int, s *goquery.Selection) {
		p.processPicture(s, options)
	})

	// Remove small or decorative images if enabled
	if options.RemoveSmallImages {
		p.removeSmallImages(options)
	}
}

// processImage processes a single image element
// TypeScript original code:
//
//	processImage(img: Element, options?: ImageProcessingOptions): void {
//	  const src = img.getAttribute('src');
//	  if (!src) {
//	    const dataSrc = img.getAttribute('data-src');
//	    if (dataSrc) {
//	      img.setAttribute('src', dataSrc);
//	    } else {
//	      return;
//	    }
//	  }
//
//	  if (this.isDecorativeImage(img, src) && options?.removeSmallImages) {
//	    img.remove();
//	    return;
//	  }
//
//	  if (options?.optimizeImages) {
//	    this.optimizeImageAttributes(img, src);
//	  }
//
//	  if (options?.generateAltText) {
//	    this.enhanceAltText(img);
//	  }
//
//	  if (options?.enableLazyLoading) {
//	    this.addLazyLoading(img);
//	  }
//
//	  if (options?.enableResponsive) {
//	    this.makeResponsive(img, options);
//	  }
//	}
func (p *ImageProcessor) processImage(s *goquery.Selection, options *ImageProcessingOptions) {
	// Get image source
	src, exists := s.Attr("src")
	if !exists {
		// Check for data-src (lazy loading)
		if dataSrc, hasDataSrc := s.Attr("data-src"); hasDataSrc {
			src = dataSrc
		} else {
			return
		}
	}

	// Skip if it's a small decorative image
	if p.isDecorativeImage(s, src) && options.RemoveSmallImages {
		s.Remove()
		return
	}

	// Optimize image attributes
	if options.OptimizeImages {
		p.optimizeImageAttributes(s, src)
	}

	// Generate or enhance alt text
	if options.GenerateAltText {
		p.enhanceAltText(s)
	}

	// Add lazy loading
	if options.EnableLazyLoading {
		p.addLazyLoading(s)
	}

	// Make responsive
	if options.EnableResponsive {
		p.makeResponsive(s, options)
	}

	// Add loading optimization
	p.addLoadingOptimization(s)
}

// processFigure processes figure elements with images
// TypeScript original code:
//
//	processFigure(figure: Element, options?: ImageProcessingOptions): void {
//	  const img = figure.querySelector('img');
//	  if (!img) return;
//
//	  this.processImage(img, options);
//
//	  const caption = figure.querySelector('figcaption');
//	  if (caption) {
//	    this.processFigcaption(img, caption);
//	  } else if (options?.generateAltText) {
//	    this.generateFigcaption(figure, img);
//	  }
//
//	  this.addFigureClasses(figure);
//	}
func (p *ImageProcessor) processFigure(s *goquery.Selection, options *ImageProcessingOptions) {
	// Find image within figure
	img := s.Find("img").First()
	if img.Length() == 0 {
		return
	}

	// Process the image
	p.processImage(img, options)

	// Process figcaption
	caption := s.Find("figcaption").First()
	if caption.Length() > 0 {
		p.processFigcaption(img, caption)
	} else if options.GenerateAltText {
		// Try to generate caption from alt text or context
		p.generateFigcaption(s, img)
	}

	// Add figure styling classes
	p.addFigureClasses(s)
}

// processPicture processes picture elements
// TypeScript original code:
//
//	processPicture(picture: Element, options?: ImageProcessingOptions): void {
//	  const sources = picture.querySelectorAll('source');
//	  sources.forEach(source => this.processSource(source, options));
//
//	  const img = picture.querySelector('img');
//	  if (img) {
//	    this.processImage(img, options);
//	  }
//	}
func (p *ImageProcessor) processPicture(s *goquery.Selection, options *ImageProcessingOptions) {
	// Process all source elements
	s.Find("source").Each(func(_ int, source *goquery.Selection) {
		p.processSource(source, options)
	})

	// Process the fallback img element
	img := s.Find("img").First()
	if img.Length() > 0 {
		p.processImage(img, options)
	}
}

// isDecorativeImage determines if an image is decorative/small
// TypeScript original code:
//
//	isDecorativeImage(img: Element, src: string): boolean {
//	  const width = parseInt(img.getAttribute('width') || '0');
//	  const height = parseInt(img.getAttribute('height') || '0');
//
//	  if (width < 50 || height < 50) {
//	    return true;
//	  }
//
//	  const classes = img.className.toLowerCase();
//	  const decorativeClasses = ['icon', 'avatar', 'emoji', 'bullet', 'decoration'];
//	  if (decorativeClasses.some(cls => classes.includes(cls))) {
//	    return true;
//	  }
//
//	  return this.isTrackingPixel(src);
//	}
func (p *ImageProcessor) isDecorativeImage(s *goquery.Selection, src string) bool {
	// Check explicit dimensions
	if width, hasWidth := s.Attr("width"); hasWidth {
		if w, err := strconv.Atoi(width); err == nil && w < 50 {
			return true
		}
	}
	if height, hasHeight := s.Attr("height"); hasHeight {
		if h, err := strconv.Atoi(height); err == nil && h < 50 {
			return true
		}
	}

	// Check CSS classes that indicate decorative images
	if class, hasClass := s.Attr("class"); hasClass {
		decorativeClasses := []string{"icon", "avatar", "emoji", "bullet", "decoration", "logo-small"}
		classLower := strings.ToLower(class)
		for _, decorativeClass := range decorativeClasses {
			if strings.Contains(classLower, decorativeClass) {
				return true
			}
		}
	}

	// Check if it's a tracking pixel
	return p.isTrackingPixel(src)
}

// optimizeImageAttributes optimizes image attributes
// TypeScript original code:
//
//	optimizeImageAttributes(img: Element, src: string): void {
//	  // Handle lazy-loaded images
//	  const dataSrc = img.getAttribute('data-src');
//	  if (dataSrc && !img.getAttribute('src')) {
//	    img.setAttribute('src', dataSrc);
//	  }
//
//	  const dataSrcset = img.getAttribute('data-srcset');
//	  if (dataSrcset && !img.getAttribute('srcset')) {
//	    img.setAttribute('srcset', dataSrcset);
//	  }
//
//	  // Remove lazy loading attributes
//	  img.removeAttribute('data-src');
//	  img.removeAttribute('data-srcset');
//	  img.classList.remove('lazy', 'lazyload');
//
//	  // Optimize URL parameters
//	  if (src.includes('?')) {
//	    const optimizedSrc = this.optimizeImageUrl(src);
//	    if (optimizedSrc !== src) {
//	      img.setAttribute('src', optimizedSrc);
//	    }
//	  }
//	}
func (p *ImageProcessor) optimizeImageAttributes(s *goquery.Selection, src string) {
	// Handle lazy-loaded images by moving data-src to src
	if dataSrc, hasDataSrc := s.Attr("data-src"); hasDataSrc && src == "" {
		s.SetAttr("src", dataSrc)
		src = dataSrc
	}

	// Handle srcset lazy loading
	if dataSrcset, hasDataSrcset := s.Attr("data-srcset"); hasDataSrcset {
		if _, hasSrcset := s.Attr("srcset"); !hasSrcset {
			s.SetAttr("srcset", dataSrcset)
		}
	}

	// Remove lazy loading attributes
	s.RemoveAttr("data-src")
	s.RemoveAttr("data-srcset")
	s.RemoveAttr("data-lazy")
	s.RemoveClass("lazy")
	s.RemoveClass("lazyload")

	// Validate and clean URL
	if src != "" && !p.isRelativeURL(src) {
		if parsedURL, err := url.Parse(src); err == nil {
			// Clean up URL if needed
			s.SetAttr("src", parsedURL.String())
		}
	}
}

// enhanceAltText generates or enhances alt text for images
// TypeScript original code:
//
//	enhanceAltText(img: Element): void {
//	  let altText = img.getAttribute('alt') || '';
//
//	  if (!altText || this.isGenericAltText(altText)) {
//	    const generatedAlt = this.generateAltText(img);
//	    if (generatedAlt) {
//	      img.setAttribute('alt', generatedAlt);
//	    }
//	  } else if (this.isGenericAltText(altText)) {
//	    const enhancedAlt = this.enhanceGenericAltText(img, altText);
//	    if (enhancedAlt !== altText) {
//	      img.setAttribute('alt', enhancedAlt);
//	    }
//	  }
//	}
func (p *ImageProcessor) enhanceAltText(s *goquery.Selection) {
	currentAlt := s.AttrOr("alt", "")

	// If no alt text or generic alt text, try to generate better alt text
	if currentAlt == "" || p.isGenericAltText(currentAlt) {
		generatedAlt := p.generateAltText(s)
		if generatedAlt != "" {
			s.SetAttr("alt", generatedAlt)
		}
	}
}

// generateAltText generates alt text from various sources
// TypeScript original code:
//
//	generateAltText(img: Element): string {
//	  // Try title attribute
//	  const title = img.getAttribute('title');
//	  if (title && title.length > 3) {
//	    return title;
//	  }
//
//	  // Try figure caption
//	  const figure = img.closest('figure');
//	  if (figure) {
//	    const caption = figure.querySelector('figcaption');
//	    if (caption && caption.textContent) {
//	      return caption.textContent.trim();
//	    }
//	  }
//
//	  // Try contextual text
//	  const contextualAlt = this.getContextualAltText(img);
//	  if (contextualAlt) {
//	    return contextualAlt;
//	  }
//
//	  // Try filename
//	  const src = img.getAttribute('src');
//	  if (src) {
//	    return this.getAltFromFilename(src);
//	  }
//
//	  return '';
//	}
func (p *ImageProcessor) generateAltText(s *goquery.Selection) string {
	// Try title attribute first
	if title, hasTitle := s.Attr("title"); hasTitle && len(title) > 3 {
		return strings.TrimSpace(title)
	}

	// Try figure caption
	figure := s.Closest("figure")
	if figure.Length() > 0 {
		caption := figure.Find("figcaption").First()
		if caption.Length() > 0 {
			captionText := strings.TrimSpace(caption.Text())
			if captionText != "" {
				return captionText
			}
		}
	}

	// Try contextual text from surrounding elements
	contextualAlt := p.getContextualAltText(s)
	if contextualAlt != "" {
		return contextualAlt
	}

	// Try to extract from filename
	if src, hasSrc := s.Attr("src"); hasSrc {
		return p.getAltFromFilename(src)
	}

	return ""
}

// getContextualAltText extracts alt text from context
// TypeScript original code:
//
//	getContextualAltText(img: Element): string {
//	  // Look for nearby headings
//	  const nearbyHeading = this.findNearbyHeading(img);
//	  if (nearbyHeading) {
//	    return nearbyHeading;
//	  }
//
//	  // Look for link text if image is inside a link
//	  const link = img.closest('a');
//	  if (link && link.textContent) {
//	    const linkText = link.textContent.trim();
//	    if (linkText && linkText !== img.getAttribute('alt')) {
//	      return linkText;
//	    }
//	  }
//
//	  // Look for aria-label
//	  const ariaLabel = img.getAttribute('aria-label');
//	  if (ariaLabel) {
//	    return ariaLabel;
//	  }
//
//	  // Look in parent elements for descriptive text
//	  let parent = img.parentElement;
//	  while (parent && parent !== document.body) {
//	    const text = this.extractMeaningfulText(parent);
//	    if (text && text.length > 10) {
//	      return text.substring(0, 100);
//	    }
//	    parent = parent.parentElement;
//	  }
//
//	  return '';
//	}
func (p *ImageProcessor) getContextualAltText(s *goquery.Selection) string {
	// Look for nearby headings
	if heading := p.findNearbyHeading(s); heading != "" {
		return heading
	}

	// Check if image is inside a link and use link text
	link := s.Closest("a")
	if link.Length() > 0 {
		linkText := strings.TrimSpace(link.Text())
		if linkText != "" && linkText != s.AttrOr("alt", "") {
			return linkText
		}
	}

	// Check for aria-label
	if ariaLabel, hasAriaLabel := s.Attr("aria-label"); hasAriaLabel {
		return ariaLabel
	}

	// Look in parent elements for descriptive text
	parent := s.Parent()
	for parent.Length() > 0 && !parent.Is("body") {
		text := strings.TrimSpace(parent.Text())
		if len(text) > 10 && len(text) < 100 {
			// Avoid using text that's too long or includes many child elements
			children := parent.Children()
			if children.Length() <= 2 {
				return text
			}
		}
		parent = parent.Parent()
	}

	return ""
}

// getAltFromFilename extracts meaningful text from filename
// TypeScript original code:
//
//	getAltFromFilename(src: string): string {
//	  try {
//	    const url = new URL(src, window.location.href);
//	    const pathname = url.pathname;
//	    const filename = pathname.split('/').pop() || '';
//
//	    if (!filename || this.isGenericFilename(filename)) {
//	      return '';
//	    }
//
//	    // Remove file extension
//	    const nameWithoutExt = filename.replace(/\.[^.]+$/, '');
//
//	    // Replace common separators with spaces
//	    let readable = nameWithoutExt
//	      .replace(/[-_]/g, ' ')
//	      .replace(/([a-z])([A-Z])/g, '$1 $2') // camelCase to spaces
//	      .replace(/\s+/g, ' ')
//	      .trim();
//
//	    // Capitalize first letter
//	    if (readable) {
//	      readable = readable.charAt(0).toUpperCase() + readable.slice(1);
//	    }
//
//	    return readable;
//	  } catch (e) {
//	    return '';
//	  }
//	}
func (p *ImageProcessor) getAltFromFilename(src string) string {
	parsedURL, err := url.Parse(src)
	if err != nil {
		return ""
	}

	// Extract filename from path
	pathParts := strings.Split(parsedURL.Path, "/")
	filename := pathParts[len(pathParts)-1]

	if filename == "" || p.isGenericFilename(filename) {
		return ""
	}

	// Remove file extension
	nameWithoutExt := fileExtRe.ReplaceAllString(filename, "")

	// Convert to readable format
	// Replace common separators with spaces
	readable := separatorsRe.ReplaceAllString(nameWithoutExt, " ")

	// Handle camelCase
	readable = camelCaseRe.ReplaceAllString(readable, "$1 $2")

	// Clean up multiple spaces
	readable = imgWhitespaceRe.ReplaceAllString(readable, " ")
	readable = strings.TrimSpace(readable)

	// Capitalize first letter
	if readable != "" {
		readable = strings.ToUpper(readable[:1]) + readable[1:]
	}

	return readable
}

// isGenericAltText checks if alt text is generic or placeholder text
// TypeScript original code:
//
//	isGenericAltText(alt: string): boolean {
//	  const genericTerms = ['image', 'picture', 'photo', 'screenshot', 'icon', 'logo', 'banner'];
//	  const altLower = alt.toLowerCase().trim();
//
//	  if (altLower.length < 3) {
//	    return true;
//	  }
//
//	  return genericTerms.some(term => altLower === term || altLower.includes(term));
//	}
func (p *ImageProcessor) isGenericAltText(alt string) bool {
	genericTerms := []string{"image", "picture", "photo", "screenshot", "icon", "logo", "banner", "graphic"}
	altLower := strings.ToLower(strings.TrimSpace(alt))

	// Very short alt text is considered generic
	if len(altLower) < 3 {
		return true
	}

	// Check if alt text exactly matches or contains generic terms
	for _, term := range genericTerms {
		if altLower == term || strings.Contains(altLower, term) {
			return true
		}
	}

	return false
}

// addLazyLoading adds lazy loading attributes to images
// TypeScript original code:
//
//	addLazyLoading(img: Element): void {
//	  if (!img.hasAttribute('loading')) {
//	    img.setAttribute('loading', 'lazy');
//	  }
//
//	  // Add intersection observer polyfill data
//	  img.setAttribute('data-lazy', 'true');
//
//	  // Preserve original src in data attribute for fallback
//	  const src = img.getAttribute('src');
//	  if (src && !img.hasAttribute('data-original-src')) {
//	    img.setAttribute('data-original-src', src);
//	  }
//	}
func (p *ImageProcessor) addLazyLoading(s *goquery.Selection) {
	// Add native lazy loading
	if _, hasLoading := s.Attr("loading"); !hasLoading {
		s.SetAttr("loading", "lazy")
	}

	// Don't add lazy loading to above-the-fold images
	if p.isAboveFold(s) {
		s.SetAttr("loading", "eager")
		return
	}

	// Add data attribute for custom lazy loading implementation
	s.SetAttr("data-lazy", "true")
}

// makeResponsive makes images responsive
// TypeScript original code:
//
//	makeResponsive(img: Element, options: ImageProcessingOptions): void {
//	  if (!img.style.maxWidth) {
//	    img.style.maxWidth = '100%';
//	  }
//	  if (!img.style.height) {
//	    img.style.height = 'auto';
//	  }
//
//	  // Add responsive class
//	  img.classList.add('responsive-image');
//
//	  // Generate srcset if not present
//	  if (!img.hasAttribute('srcset')) {
//	    const srcset = this.generateSrcset(img, options);
//	    if (srcset) {
//	      img.setAttribute('srcset', srcset);
//	    }
//	  }
//
//	  // Add sizes attribute
//	  if (!img.hasAttribute('sizes')) {
//	    img.setAttribute('sizes', '(max-width: 768px) 100vw, 50vw');
//	  }
//	}
func (p *ImageProcessor) makeResponsive(s *goquery.Selection, _ *ImageProcessingOptions) {
	// Add responsive styling via class
	s.AddClass("responsive-image")

	// Set max-width and height styles if not present
	style := s.AttrOr("style", "")
	if !strings.Contains(style, "max-width") {
		if style == "" {
			style = "max-width: 100%;"
		} else {
			style += " max-width: 100%;"
		}
	}
	if !strings.Contains(style, "height") {
		style += " height: auto;"
	}
	s.SetAttr("style", style)

	// Add sizes attribute for responsive behavior
	if _, hasSizes := s.Attr("sizes"); !hasSizes {
		s.SetAttr("sizes", "(max-width: 768px) 100vw, 50vw")
	}
}

// addLoadingOptimization adds loading optimization attributes
// TypeScript original code:
//
//	addLoadingOptimization(img: Element): void {
//	  // Add decode optimization
//	  if (!img.hasAttribute('decoding')) {
//	    img.setAttribute('decoding', 'async');
//	  }
//
//	  // Add fetchpriority for important images
//	  if (this.isImportantImage(img) && !img.hasAttribute('fetchpriority')) {
//	    img.setAttribute('fetchpriority', 'high');
//	  }
//	}
func (p *ImageProcessor) addLoadingOptimization(s *goquery.Selection) {
	// Add async decoding
	if _, hasDecoding := s.Attr("decoding"); !hasDecoding {
		s.SetAttr("decoding", "async")
	}

	// Add high priority for important images
	if p.isImportantImage(s) {
		if _, hasFetchPriority := s.Attr("fetchpriority"); !hasFetchPriority {
			s.SetAttr("fetchpriority", "high")
		}
	}
}

// processFigcaption processes figure captions
// TypeScript original code:
//
//	processFigcaption(img: Element, caption: Element): void {
//	  const captionText = caption.textContent?.trim();
//	  if (!captionText) {
//	    caption.remove();
//	    return;
//	  }
//
//	  // Enhance caption formatting
//	  if (captionText.length > 200) {
//	    caption.classList.add('long-caption');
//	  }
//
//	  // Link caption to image for accessibility
//	  const imgId = img.getAttribute('id') || `img-${Date.now()}`;
//	  img.setAttribute('id', imgId);
//	  caption.setAttribute('aria-describedby', imgId);
//	}
func (p *ImageProcessor) processFigcaption(img, caption *goquery.Selection) {
	captionText := strings.TrimSpace(caption.Text())
	if captionText == "" {
		caption.Remove()
		return
	}

	// Add class for long captions
	if len(captionText) > 200 {
		caption.AddClass("long-caption")
	}

	// Link caption to image for accessibility
	imgID := img.AttrOr("id", "")
	if imgID == "" {
		imgID = fmt.Sprintf("img-%d", p.generateImageID())
		img.SetAttr("id", imgID)
	}
	caption.SetAttr("aria-describedby", imgID)
}

// generateFigcaption generates a caption for figures without one
// TypeScript original code:
//
//	generateFigcaption(figure: Element, img: Element): void {
//	  const alt = img.getAttribute('alt');
//	  if (!alt || alt.length < 10) return;
//
//	  const figcaption = document.createElement('figcaption');
//	  figcaption.textContent = alt;
//	  figure.appendChild(figcaption);
//	}
func (p *ImageProcessor) generateFigcaption(figure, img *goquery.Selection) {
	alt := img.AttrOr("alt", "")
	if alt == "" || len(alt) < 10 {
		return
	}

	// Only generate if there's meaningful alt text
	if !p.isGenericAltText(alt) {
		figcaption := fmt.Sprintf("<figcaption>%s</figcaption>", alt)
		figure.AppendHtml(figcaption)
	}
}

// addFigureClasses adds appropriate classes to figure elements
// TypeScript original code:
//
//	addFigureClasses(figure: Element): void {
//	  figure.classList.add('image-figure');
//
//	  const img = figure.querySelector('img');
//	  if (img) {
//	    const width = parseInt(img.getAttribute('width') || '0');
//	    if (width > 800) {
//	      figure.classList.add('large-image');
//	    } else if (width < 300) {
//	      figure.classList.add('small-image');
//	    }
//	  }
//	}
func (p *ImageProcessor) addFigureClasses(s *goquery.Selection) {
	s.AddClass("image-figure")

	// Add size-based classes
	img := s.Find("img").First()
	if img.Length() > 0 {
		if width, hasWidth := img.Attr("width"); hasWidth {
			if w, err := strconv.Atoi(width); err == nil {
				if w > 800 {
					s.AddClass("large-image")
				} else if w < 300 {
					s.AddClass("small-image")
				}
			}
		}
	}
}

// processSource processes source elements in picture tags
// TypeScript original code:
//
//	processSource(source: Element, options: ImageProcessingOptions): void {
//	  const srcset = source.getAttribute('srcset');
//	  if (!srcset) return;
//
//	  // Optimize srcset URLs
//	  const optimizedSrcset = this.optimizeSrcset(srcset);
//	  if (optimizedSrcset !== srcset) {
//	    source.setAttribute('srcset', optimizedSrcset);
//	  }
//	}
func (p *ImageProcessor) processSource(s *goquery.Selection, _ *ImageProcessingOptions) {
	srcset, hasSrcset := s.Attr("srcset")
	if !hasSrcset || srcset == "" {
		return
	}

	// Basic srcset validation and cleanup could go here
	// For now, just ensure the attribute is properly formatted
	srcset = strings.TrimSpace(srcset)
	s.SetAttr("srcset", srcset)
}

// removeSmallImages removes small or decorative images
// TypeScript original code:
//
//	removeSmallImages(options: ImageProcessingOptions): void {
//	  const imgs = this.document.querySelectorAll('img');
//	  imgs.forEach(img => {
//	    if (this.shouldRemoveSmallImage(img, options)) {
//	      img.remove();
//	    }
//	  });
//	}
func (p *ImageProcessor) removeSmallImages(options *ImageProcessingOptions) {
	p.doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		if p.shouldRemoveSmallImage(s, options) {
			s.Remove()
		}
	})
}

// shouldRemoveSmallImage determines if a small image should be removed
// TypeScript original code:
//
//	shouldRemoveSmallImage(img: Element, options: ImageProcessingOptions): boolean {
//	  const width = parseInt(img.getAttribute('width') || '0');
//	  const height = parseInt(img.getAttribute('height') || '0');
//
//	  if (width > 0 && width < options.minImageWidth) return true;
//	  if (height > 0 && height < options.minImageHeight) return true;
//
//	  // Don't remove important images
//	  if (this.isImportantImage(img)) return false;
//
//	  // Remove tracking pixels and decorative images
//	  const src = img.getAttribute('src') || '';
//	  return this.isTrackingPixel(src) || this.isDecorativeImage(img, src);
//	}
func (p *ImageProcessor) shouldRemoveSmallImage(s *goquery.Selection, options *ImageProcessingOptions) bool {
	// Check dimensions
	if width, hasWidth := s.Attr("width"); hasWidth {
		if w, err := strconv.Atoi(width); err == nil && w > 0 && w < options.MinImageWidth {
			return true
		}
	}
	if height, hasHeight := s.Attr("height"); hasHeight {
		if h, err := strconv.Atoi(height); err == nil && h > 0 && h < options.MinImageHeight {
			return true
		}
	}

	// Don't remove important images
	if p.isImportantImage(s) {
		return false
	}

	// Remove tracking pixels and decorative images
	src := s.AttrOr("src", "")
	return p.isTrackingPixel(src) || p.isDecorativeImage(s, src)
}

// isRelativeURL checks if a URL is relative
// TypeScript original code:
//
//	isRelativeURL(src: string): boolean {
//	  return !src.startsWith('http://') && !src.startsWith('https://') && !src.startsWith('//');
//	}
func (p *ImageProcessor) isRelativeURL(src string) bool {
	return !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") && !strings.HasPrefix(src, "//")
}

// isImportantImage determines if an image is important (shouldn't be removed)
// TypeScript original code:
//
//	isImportantImage(img: Element): boolean {
//	  // Check if it's the main content image
//	  const figure = img.closest('figure');
//	  if (figure && figure.classList.contains('featured')) {
//	    return true;
//	  }
//
//	  // Check if it's above the fold
//	  if (this.isAboveFold(img)) {
//	    return true;
//	  }
//
//	  // Check if it has meaningful alt text
//	  const alt = img.getAttribute('alt') || '';
//	  if (alt.length > 20 && !this.isGenericAltText(alt)) {
//	    return true;
//	  }
//
//	  // Check if it's part of the main content
//	  const article = img.closest('article, main, .content');
//	  return article.Length > 0;
//	}
func (p *ImageProcessor) isImportantImage(s *goquery.Selection) bool {
	// Check if it's in a featured figure
	figure := s.Closest("figure")
	if figure.Length() > 0 && figure.HasClass("featured") {
		return true
	}

	// Check if it's above the fold
	if p.isAboveFold(s) {
		return true
	}

	// Check if it has meaningful alt text
	alt := s.AttrOr("alt", "")
	if len(alt) > 20 && !p.isGenericAltText(alt) {
		return true
	}

	// Check if it's in main content area
	mainContent := s.Closest("article, main, .content, .post")
	return mainContent.Length() > 0
}

// isAboveFold determines if an image is likely above the fold
// TypeScript original code:
//
//	isAboveFold(img: Element): boolean {
//	  // Simple heuristic: first few images in the document are likely above fold
//	  const allImages = Array.from(document.querySelectorAll('img'));
//	  const index = allImages.indexOf(img);
//	  return index < 3;
//	}
func (p *ImageProcessor) isAboveFold(s *goquery.Selection) bool {
	// Simple heuristic: check if it's one of the first few images
	var imageIndex int
	found := false

	p.doc.Find("img").Each(func(i int, img *goquery.Selection) {
		if img.Get(0) == s.Get(0) {
			imageIndex = i
			found = true
		}
	})

	return found && imageIndex < 3
}

// findNearbyHeading finds a nearby heading that could describe the image
// TypeScript original code:
//
//	findNearbyHeading(img: Element): string {
//	  // Look for previous headings
//	  let current = img.previousElementSibling;
//	  while (current) {
//	    if (/^h[1-6]$/i.test(current.tagName)) {
//	      return current.textContent?.trim() || '';
//	    }
//	    current = current.previousElementSibling;
//	  }
//
//	  // Look in parent elements
//	  let parent = img.parentElement;
//	  while (parent && parent !== document.body) {
//	    const heading = parent.querySelector('h1, h2, h3, h4, h5, h6');
//	    if (heading) {
//	      return heading.textContent?.trim() || '';
//	    }
//	    parent = parent.parentElement;
//	  }
//
//	  return '';
//	}
func (p *ImageProcessor) findNearbyHeading(s *goquery.Selection) string {
	// Look for previous headings in the same container
	parent := s.Parent()
	var headingText string

	parent.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, heading *goquery.Selection) {
		text := strings.TrimSpace(heading.Text())
		if text != "" && len(text) < 100 {
			headingText = text
		}
	})

	if headingText != "" {
		return headingText
	}

	// Look in ancestor elements
	ancestor := parent.Parent()
	for ancestor.Length() > 0 && !ancestor.Is("body") {
		heading := ancestor.Find("h1, h2, h3, h4, h5, h6").First()
		if heading.Length() > 0 {
			text := strings.TrimSpace(heading.Text())
			if text != "" && len(text) < 100 {
				return text
			}
		}
		ancestor = ancestor.Parent()
	}

	return ""
}

// isGenericFilename checks if a filename is generic
// TypeScript original code:
//
//	isGenericFilename(filename: string): boolean {
//	  const genericPatterns = [
//	    /^image\d*\.(jpg|jpeg|png|gif|webp)$/i,
//	    /^img\d*\.(jpg|jpeg|png|gif|webp)$/i,
//	    /^picture\d*\.(jpg|jpeg|png|gif|webp)$/i,
//	    /^photo\d*\.(jpg|jpeg|png|gif|webp)$/i,
//	    /^screenshot\d*\.(jpg|jpeg|png|gif|webp)$/i,
//	    /^\d+\.(jpg|jpeg|png|gif|webp)$/i,
//	    /^untitled\d*\.(jpg|jpeg|png|gif|webp)$/i
//	  ];
//
//	  return genericPatterns.some(pattern => pattern.test(filename));
//	}
func (p *ImageProcessor) isGenericFilename(filename string) bool {
	genericPatterns := []string{
		`^image\d*\.(jpg|jpeg|png|gif|webp)$`,
		`^img\d*\.(jpg|jpeg|png|gif|webp)$`,
		`^picture\d*\.(jpg|jpeg|png|gif|webp)$`,
		`^photo\d*\.(jpg|jpeg|png|gif|webp)$`,
		`^screenshot\d*\.(jpg|jpeg|png|gif|webp)$`,
		`^\d+\.(jpg|jpeg|png|gif|webp)$`,
		`^untitled\d*\.(jpg|jpeg|png|gif|webp)$`,
	}

	filenameLower := strings.ToLower(filename)
	for _, pattern := range genericPatterns {
		if matched, _ := regexp.MatchString(pattern, filenameLower); matched {
			return true
		}
	}

	return false
}

// isTrackingPixel determines if an image is a tracking pixel
// TypeScript original code:
//
//	isTrackingPixel(src: string): boolean {
//	  if (!src) return false;
//
//	  // Check common tracking pixel patterns
//	  const trackingPatterns = [
//	    /pixel\.gif/i,
//	    /1x1\.gif/i,
//	    /tracking\.gif/i,
//	    /analytics/i,
//	    /metrics/i,
//	    /beacon/i
//	  ];
//
//	  return trackingPatterns.some(pattern => pattern.test(src));
//	}
func (p *ImageProcessor) isTrackingPixel(src string) bool {
	if src == "" {
		return false
	}

	trackingPatterns := []string{
		`pixel\.gif`,
		`1x1\.gif`,
		`tracking\.gif`,
		`analytics`,
		`metrics`,
		`beacon`,
	}

	srcLower := strings.ToLower(src)
	for _, pattern := range trackingPatterns {
		if matched, _ := regexp.MatchString(pattern, srcLower); matched {
			return true
		}
	}

	return false
}

// generateImageID generates a unique ID for images
// TypeScript original code:
//
//	generateImageId(): string {
//	  return `img-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
//	}
func (p *ImageProcessor) generateImageID() int {
	// Simple counter-based ID generation
	var counter int
	p.doc.Find("img[id]").Each(func(_ int, _ *goquery.Selection) {
		counter++
	})
	return counter + 1
}

// ProcessImages processes all images in the document (public interface)
// TypeScript original code:
//
//	export function processImages(doc: Document, options?: ImageProcessingOptions): void {
//	  const processor = new ImageProcessor(doc);
//	  processor.processImages(options);
//	}
func ProcessImages(doc *goquery.Document, options *ImageProcessingOptions) {
	processor := NewImageProcessor(doc)
	processor.ProcessImages(options)
}
