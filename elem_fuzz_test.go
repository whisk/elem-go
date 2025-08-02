//go:build go1.18
// +build go1.18

package elem

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/chasefleming/elem-go/attrs"
)

func FuzzElementRenderTo(f *testing.F) {
	f.Add("div", "class", "foo", "bar", "baz") // Seed corpus
	f.Fuzz(func(t *testing.T, tag, attrKey, attrVal, childText, commentText string) {
		// Build attributes
		props := attrs.Props{}
		if attrKey != "" {
			props[attrKey] = attrVal
		}
		// Build children
		children := []Node{TextNode(childText), Comment(commentText)}
		el := newElement(tag, props, children...)
		var sb strings.Builder
		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RenderTo panicked: %v", r)
			}
		}()
		el.RenderTo(&sb, RenderOptions{})

		// Try to decode the resulting HTML as XML
		r := strings.NewReader(sb.String())
		dec := xml.NewDecoder(r)
		for {
			_, err := dec.Token()
			if err != nil {
				if err.Error() != "EOF" {
					t.Logf("xml.Decoder error: %v", err)
				}
				break
			}
		}
	})
}
