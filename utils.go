package elem

import "strings"

// commentContentsReplacer handles escaping of disallowed sequences in HTML comments
var commentContentsReplacer = strings.NewReplacer(
	"<!--", "&lt;!--",
	"-->", "--&gt;",
	"--!>", "--!&gt;",
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
	var nodes []Node
	for _, item := range items {
		nodes = append(nodes, fn(item))
	}
	return nodes
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
