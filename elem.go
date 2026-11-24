package elem

import (
	"fmt"
	"slices"
	"strings"

	"github.com/chasefleming/elem-go/attrs"
)

// List of HTML5 void elements. Void elements, also known as self-closing or empty elements,
// are elements that don't have a closing tag because they can't contain any content.
// For example, the <img> tag cannot wrap text or other tags, it stands alone, so it doesn't have a closing tag.
var voidElements = map[string]struct{}{
	"area":    {},
	"base":    {},
	"br":      {},
	"col":     {},
	"command": {},
	"embed":   {},
	"hr":      {},
	"img":     {},
	"input":   {},
	"keygen":  {},
	"link":    {},
	"meta":    {},
	"param":   {},
	"source":  {},
	"track":   {},
	"wbr":     {},
}

// List of boolean attributes. Boolean attributes can't have literal values. The presence of an boolean
// attribute represents the "true" value. To represent the "false" value, the attribute has to be omitted.
// See https://html.spec.whatwg.org/multipage/indices.html#attributes-3 for reference
var booleanAttrs = map[string]struct{}{
	attrs.AllowFullscreen: {},
	attrs.Async:           {},
	attrs.Autofocus:       {},
	attrs.Autoplay:        {},
	attrs.Checked:         {},
	attrs.Controls:        {},
	attrs.Defer:           {},
	attrs.Disabled:        {},
	attrs.Ismap:           {},
	attrs.Loop:            {},
	attrs.Multiple:        {},
	attrs.Muted:           {},
	attrs.Nomodule:        {},
	attrs.Novalidate:      {},
	attrs.Open:            {},
	attrs.Playsinline:     {},
	attrs.Readonly:        {},
	attrs.Required:        {},
	attrs.Selected:        {},
}

type CSSGenerator interface {
	GenerateCSS() string // TODO: Change to CSS()
}

// InvalidNameMode controls how Element rendering handles invalid HTML tag or attribute names.
type InvalidNameMode int

const (
	// InvalidNamePlaceholder renders an HTML comment (<!-- empty tag --> or
	// <!-- invalid tag: ... --> with invalid characters replaced by '_') in place
	// of elements with invalid tag names, and replaces invalid attribute names
	// with valueless data-invalid-attr (for empty names) or
	// data-invalid-attr-... attributes (with characters outside [a-zA-Z0-9-_]
	// replaced by '_').
	// This is the default (zero value) so invalid names are never silently dropped.
	InvalidNamePlaceholder InvalidNameMode = iota
	// InvalidNamePanic panics when encountering an invalid tag or attribute name.
	InvalidNamePanic
	// InvalidNameAsIs renders invalid tag and attribute names verbatim without
	// validation, preserving pre-validation behavior for backwards compatibility.
	//
	// Deprecated: InvalidNameAsIs is unsafe (it permits HTML and attribute
	// injection if names are not strictly trusted), is not recommended for use,
	// and may be removed in a future version. Use InvalidNamePlaceholder or
	// InvalidNamePanic instead.
	InvalidNameAsIs
)

const invalidAttrPlaceholder = "data-invalid-attr"

type RenderOptions struct {
	// DisableHtmlPreamble disables the doctype preamble for the HTML tag if it exists in the rendering tree
	DisableHtmlPreamble bool
	StyleManager        CSSGenerator
	// InvalidTagMode controls how invalid HTML tag names are handled.
	// The zero value (InvalidNamePlaceholder) renders an HTML comment placeholder.
	InvalidTagMode InvalidNameMode
	// InvalidAttrMode controls how invalid HTML attribute names are handled.
	// The zero value (InvalidNamePlaceholder) emits valueless data-invalid-attr-... placeholder attributes.
	InvalidAttrMode InvalidNameMode
}

type Node interface {
	RenderTo(builder *strings.Builder, opts RenderOptions)
	Render() string
	RenderWithOptions(opts RenderOptions) string
}

// NoneNode represents a node that renders nothing.
type NoneNode struct{}

// RenderTo for NoneNode does nothing.
func (n NoneNode) RenderTo(builder *strings.Builder, opts RenderOptions) {
	// Intentionally left blank to render nothing
}

// Render for NoneNode returns an empty string.
func (n NoneNode) Render() string {
	return ""
}

// RenderWithOptions for NoneNode returns an empty string.
func (n NoneNode) RenderWithOptions(opts RenderOptions) string {
	return ""
}

type TextNode string

func (t TextNode) RenderTo(builder *strings.Builder, opts RenderOptions) {
	// Stream the escaped text directly into the builder to avoid
	// allocating an intermediate escaped string.
	nodeContentReplacer.WriteString(builder, string(t))
}

func (t TextNode) Render() string {
	return EscapeNodeContents(string(t))
}

func (t TextNode) RenderWithOptions(opts RenderOptions) string {
	return EscapeNodeContents(string(t))
}

type RawNode string

func (r RawNode) RenderTo(builder *strings.Builder, opts RenderOptions) {
	builder.WriteString(string(r))
}

func (r RawNode) Render() string {
	return string(r)
}

func (t RawNode) RenderWithOptions(opts RenderOptions) string {
	return string(t)
}

type ScriptNode struct {
	node Node
}

func NewScriptNode(node Node) ScriptNode {
	return ScriptNode{node}
}

func (t ScriptNode) RenderTo(builder *strings.Builder, opts RenderOptions) {
	scriptContents := EscapeScriptContents(t.node.RenderWithOptions(opts))
	builder.WriteString(scriptContents)
}

func (t ScriptNode) Render() string {
	return t.RenderWithOptions(RenderOptions{})
}

func (t ScriptNode) RenderWithOptions(opts RenderOptions) string {
	var builder strings.Builder
	t.RenderTo(&builder, opts)
	return builder.String()
}

type CdataNode string

func (t CdataNode) RenderTo(builder *strings.Builder, opts RenderOptions) {
	builder.WriteString("<![CDATA[")
	builder.WriteString(EscapeCdataContents(string(t)))
	builder.WriteString("]]>")
}

func (t CdataNode) Render() string {
	return t.RenderWithOptions(RenderOptions{})
}

func (t CdataNode) RenderWithOptions(opts RenderOptions) string {
	var builder strings.Builder
	t.RenderTo(&builder, opts)
	return builder.String()
}

type CommentNode string

func (c CommentNode) RenderTo(builder *strings.Builder, opts RenderOptions) {
	builder.WriteString("<!-- ")
	builder.WriteString(EscapeCommentContents(string(c)))
	builder.WriteString(" -->")
}

func (c CommentNode) Render() string {
	return c.RenderWithOptions(RenderOptions{})
}

func (c CommentNode) RenderWithOptions(opts RenderOptions) string {
	var builder strings.Builder
	c.RenderTo(&builder, opts)
	return builder.String()
}

type Element struct {
	Tag      string
	Attrs    attrs.Props
	Children []Node
}

// RenderTo writes the HTML representation of e into builder.
// By default (InvalidNamePlaceholder), invalid tag or attribute names are
// replaced with safe placeholders; set RenderOptions.InvalidTagMode or
// RenderOptions.InvalidAttrMode to InvalidNamePanic to panic instead.
func (e *Element) RenderTo(builder *strings.Builder, opts RenderOptions) {
	if e.Tag == "fragment" {
		for _, child := range e.Children {
			child.RenderTo(builder, opts)
		}
		return
	}

	if opts.InvalidTagMode != InvalidNameAsIs && !validTagName(e.Tag) {
		if opts.InvalidTagMode == InvalidNamePanic {
			panic(fmt.Sprintf("invalid tag name %q", e.Tag))
		}
		if e.Tag == "" {
			builder.WriteString("<!-- empty tag -->")
		} else {
			builder.WriteString("<!-- invalid tag: ")
			writeSanitizedName(builder, e.Tag)
			builder.WriteString(" -->")
		}
		return
	}

	// The HTML tag needs a doctype preamble in order to ensure
	// browsers don't render in legacy/quirks mode
	// https://developer.mozilla.org/en-US/docs/Glossary/Doctype
	if !opts.DisableHtmlPreamble && e.Tag == "html" {
		builder.WriteString("<!DOCTYPE html>")
	}

	// Start with opening tag
	builder.WriteString("<")
	builder.WriteString(e.Tag)

	// Append the attributes to the builder in sorted order for consistent
	// output. Elements with zero or one attribute skip the sort entirely.
	switch len(e.Attrs) {
	case 0:
	case 1:
		for k, v := range e.Attrs {
			if !renderAttrTo(k, v, builder, opts.InvalidAttrMode) {
				builder.WriteByte(' ')
				writeInvalidAttrPlaceholder(builder, k)
			}
		}
	default:
		// Use a stack-allocated array for typical attribute counts so
		// sorting the keys doesn't allocate.
		var keysArr [16]string
		var keys []string
		if len(e.Attrs) <= len(keysArr) {
			keys = keysArr[:0]
		} else {
			keys = make([]string, 0, len(e.Attrs))
		}
		for k := range e.Attrs {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		var invalidAttrs []string
		for _, k := range keys {
			if !renderAttrTo(k, e.Attrs[k], builder, opts.InvalidAttrMode) {
				invalidAttrs = append(invalidAttrs, invalidAttrName(k))
			}
		}
		if len(invalidAttrs) > 0 {
			if len(invalidAttrs) > 1 {
				slices.Sort(invalidAttrs)
			}
			var prev string
			for i, placeholder := range invalidAttrs {
				if i > 0 && placeholder == prev {
					continue
				}
				prev = placeholder
				if _, exists := e.Attrs[placeholder]; !exists {
					builder.WriteByte(' ')
					builder.WriteString(placeholder)
				}
			}
		}
	}

	// If it's a void element, close it and return
	if _, exists := voidElements[e.Tag]; exists {
		builder.WriteString(`>`)
		return
	}

	// Close opening tag
	builder.WriteString(`>`)

	// Build the content
	for _, child := range e.Children {
		child.RenderTo(builder, opts)
	}

	// Append closing tag
	builder.WriteString(`</`)
	builder.WriteString(e.Tag)
	builder.WriteString(`>`)
}

// renderAttrTo writes the string representation of the given attribute and value
// to builder, returning false if attrName was invalid and replaced by a placeholder.
func renderAttrTo(attrName, attrVal string, builder *strings.Builder, mode InvalidNameMode) bool {
	if mode != InvalidNameAsIs && !validAttrName(attrName) {
		if mode == InvalidNamePanic {
			panic(fmt.Sprintf("invalid attribute name %q", attrName))
		}
		return false
	}
	if _, exists := booleanAttrs[attrName]; exists {
		// boolean attribute presents its name only if the value is "true"
		if attrVal == "true" {
			builder.WriteString(` `)
			builder.WriteString(attrName)
		}
	} else {
		// A necessary check to avoid adding extra quotes around values that are already single-quoted
		// An example is '{"quantity": 5}'
		isSingleQuoted := len(attrVal) >= 2 && strings.HasPrefix(attrVal, "'") && strings.HasSuffix(attrVal, "'")

		builder.WriteString(` `)
		builder.WriteString(attrName)
		builder.WriteString(`=`)
		if isSingleQuoted {
			builder.WriteString(`'`)
			singleQuoteAttrReplacer.WriteString(builder, attrVal[1:len(attrVal)-1])
			builder.WriteString(`'`)
		} else {
			builder.WriteString(`"`)
			doubleQuoteAttrReplacer.WriteString(builder, attrVal)
			builder.WriteString(`"`)
		}
	}
	return true
}

func (e *Element) Render() string {
	return e.RenderWithOptions(RenderOptions{})
}

// estimateSize walks the tree and returns an approximate rendered length,
// used to preallocate the output buffer and avoid repeated growth copies.
func (e *Element) estimateSize() int {
	// Attributes are estimated at a flat average rather than iterated,
	// since ranging over the map costs more than an occasional resize.
	size := 2*len(e.Tag) + 5 + len(e.Attrs)*24
	for _, child := range e.Children {
		switch c := child.(type) {
		case *Element:
			size += c.estimateSize()
		case TextNode:
			size += len(c)
		case RawNode:
			size += len(c)
		}
	}
	return size
}

func (e *Element) RenderWithOptions(opts RenderOptions) string {
	var builder strings.Builder
	builder.Grow(e.estimateSize())
	e.RenderTo(&builder, opts)

	if opts.StyleManager != nil && (e.Tag == "fragment" || opts.InvalidTagMode == InvalidNameAsIs || validTagName(e.Tag)) {
		htmlContent := builder.String()
		cssContent := opts.StyleManager.GenerateCSS()

		// Define the <style> element with the generated CSS content
		styleElement := fmt.Sprintf("<style>%s</style>", cssContent)

		// Check if a <head> tag exists in the HTML content
		headStartIndex := strings.Index(htmlContent, "<head>")
		headEndIndex := strings.Index(htmlContent, "</head>")

		if headStartIndex != -1 && headEndIndex != -1 {
			// If <head> exists, inject the style content just before </head>
			beforeHead := htmlContent[:headEndIndex]
			afterHead := htmlContent[headEndIndex:]
			modifiedHTML := beforeHead + styleElement + afterHead
			return modifiedHTML
		} else {
			// If <head> does not exist, create it and inject the style content
			// Assuming <html> tag exists and injecting <head> immediately after <html>
			htmlTagEnd := strings.Index(htmlContent, ">") + 1
			if htmlTagEnd > 0 {
				beforeHTML := htmlContent[:htmlTagEnd]
				afterHTML := htmlContent[htmlTagEnd:]
				modifiedHTML := beforeHTML + "<head>" + styleElement + "</head>" + afterHTML
				return modifiedHTML
			}
		}
	}

	// Return the original HTML content if no modifications were made
	return builder.String()
}

func newElement(tag string, attrs attrs.Props, children ...Node) *Element {
	return &Element{
		Tag:      tag,
		Attrs:    attrs,
		Children: children,
	}
}
