package elem

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// nodeContentReplacer handles escaping of HTML special characters in text nodes
// note that single and double quotes (', ") do not need escaping in HTML5 text nodes
var nodeContentReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	// technically, > does not need escaping in HTML5, but we do it for consistency
	">", "&gt;",
)

// commentContentsReplacer handles escaping of disallowed sequences in HTML comments
var commentContentsReplacer = strings.NewReplacer(
	"<!--", "&lt;!--",
	"-->", "--&gt;",
	"--!>", "--!&gt;",
)

// doubleQuoteAttrReplacer handles escaping of HTML special characters in double-quoted attribute values.
// Note that single quotes (') do not need escaping inside double-quoted attributes in HTML5.
var doubleQuoteAttrReplacer = strings.NewReplacer(
	"&", "&amp;",
	`"`, "&quot;",
	"<", "&lt;",
	">", "&gt;",
)

// singleQuoteAttrReplacer handles escaping of HTML special characters in single-quoted attribute values.
// Note that double quotes (") do not need escaping inside single-quoted attributes in HTML5.
var singleQuoteAttrReplacer = strings.NewReplacer(
	"&", "&amp;",
	`'`, "&#39;",
	"<", "&lt;",
	">", "&gt;",
)

// If conditionally renders one of the provided elements based on the condition
func If[T any](condition bool, ifTrue, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// TransformEach maps a slice of items to a slice of Elements using the provided function
func TransformEach[T any](items []T, fn func(T) Node) []Node {
	// Return nil for empty input to match the historical behavior of the
	// unallocated slice; only preallocate when there are items to hold.
	if len(items) == 0 {
		return nil
	}
	nodes := make([]Node, 0, len(items))
	for _, item := range items {
		nodes = append(nodes, fn(item))
	}
	return nodes
}

// EscapeNodeContents escapes HTML5 special characters in a string to ensure safe rendering as a text node
func EscapeNodeContents(s string) string {
	return nodeContentReplacer.Replace(s)
}

// EscapeCdataContents escapes the contents of a CDATA section to ensure safe rendering
func EscapeCdataContents(s string) string {
	// text in cdata must not contain the string "]]>"
	if strings.Contains(s, "]]>") {
		s = strings.ReplaceAll(s, "]]>", "]]&gt;")
	}
	return s
}

// EscapeCommentContents escapes the contents of a comment node to ensure safe rendering according to https://html.spec.whatwg.org/multipage/syntax.html#comments
func EscapeCommentContents(s string) string {
	s = commentContentsReplacer.Replace(s)

	// comments cannot start with specific character sequences
	if strings.HasPrefix(s, "->") {
		s = "-&gt;" + s[2:]
	} else if strings.HasPrefix(s, ">") {
		s = "&gt;" + s[1:]
	}

	// comments cannot end with specific character sequences
	if strings.HasSuffix(s, "<!-") {
		s = s[:len(s)-3] + "&lt;!-"
	}

	return s
}

// invalidASCIIAttrNameChar marks ASCII bytes disallowed in HTML5 attribute names:
// controls (0x00-0x1F, 0x7F), space (0x20), '"', '\”, '/', '<', '=', '>'.
var invalidASCIIAttrNameChar = [128]bool{
	0x00: true, 0x01: true, 0x02: true, 0x03: true, 0x04: true, 0x05: true, 0x06: true, 0x07: true,
	0x08: true, 0x09: true, 0x0A: true, 0x0B: true, 0x0C: true, 0x0D: true, 0x0E: true, 0x0F: true,
	0x10: true, 0x11: true, 0x12: true, 0x13: true, 0x14: true, 0x15: true, 0x16: true, 0x17: true,
	0x18: true, 0x19: true, 0x1A: true, 0x1B: true, 0x1C: true, 0x1D: true, 0x1E: true, 0x1F: true,
	' ':  true,
	'"':  true,
	'\'': true,
	'/':  true,
	'<':  true,
	'=':  true,
	'>':  true,
	0x7F: true,
}

// validAttrName reports whether name is a valid HTML5 attribute name according to
// https://html.spec.whatwg.org/multipage/syntax.html#attributes-2 (also rejecting '<'
// and invalid UTF-8 sequences).
func validAttrName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for i := 0; i < len(name); i++ {
		b := name[i]
		if b < 0x80 {
			if invalidASCIIAttrNameChar[b] {
				return false
			}
			continue
		}
		if !utf8.ValidString(name[i:]) {
			return false
		}
		for _, r := range name[i:] {
			if r < 0x80 {
				if invalidASCIIAttrNameChar[r] {
					return false
				}
				continue
			}
			// Reject C1 controls (U+0080..U+009F) and Unicode noncharacters
			// (U+FDD0..U+FDEF, and U+nFFFE/U+nFFFF on every plane).
			if r <= 0x9F || (r >= 0xFDD0 && r <= 0xFDEF) || (r&0xFFFE) == 0xFFFE {
				return false
			}
		}
		return true
	}
	return true
}

// validTagName reports whether tag is a valid HTML tag name:
// an ASCII letter followed by zero or more ASCII letters, digits, hyphens, or underscores.
// Underscores are allowed because the WHATWG custom element name grammar (PCENChar) permits them.
func validTagName(tag string) bool {
	if len(tag) == 0 {
		return false
	}
	c := tag[0]
	if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
		return false
	}
	for i := 1; i < len(tag); i++ {
		c = tag[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// writeSanitizedName writes name to builder with every rune outside
// [a-zA-Z0-9-_] replaced by '_'.
func writeSanitizedName(builder *strings.Builder, name string) {
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			builder.WriteByte(byte(r))
		} else {
			builder.WriteByte('_')
		}
	}
}

// writeInvalidAttrPlaceholder writes the safe placeholder attribute name for an
// invalid attribute name into builder: invalidAttrPlaceholder for an empty name,
// or invalidAttrPlaceholder + "-" + sanitized name otherwise.
func writeInvalidAttrPlaceholder(builder *strings.Builder, attrName string) {
	builder.WriteString(invalidAttrPlaceholder)
	if attrName != "" {
		builder.WriteByte('-')
		writeSanitizedName(builder, attrName)
	}
}

// invalidAttrName returns the safe placeholder attribute name for an invalid
// attribute name.
func invalidAttrName(attrName string) string {
	if attrName == "" {
		return invalidAttrPlaceholder
	}
	var b strings.Builder
	b.Grow(len(invalidAttrPlaceholder) + 1 + len(attrName))
	writeInvalidAttrPlaceholder(&b, attrName)
	return b.String()
}

// scriptContentsRegexp matches <script and </script case-insensitively
var scriptContentsRegexp = regexp.MustCompile(`(?i)<(/?)(script)`)

// EscapeScriptContents escapes the contents of a script node according to https://html.spec.whatwg.org/multipage/scripting.html#restrictions-for-contents-of-script-elements
func EscapeScriptContents(s string) string {
	s = strings.ReplaceAll(s, `<!--`, `\x3C!--`)
	s = scriptContentsRegexp.ReplaceAllString(s, `\x3C$1$2`)
	return s
}
