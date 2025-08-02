package elem

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"pgregory.net/rapid"
)

// MockStyleManager simulates the StyleManager for testing purposes.
type MockStyleManager struct{}

// GenerateCSS returns a fixed CSS string for testing.
func (m *MockStyleManager) GenerateCSS() string {
	return "body { background-color: #fff; }"
}

func TestRenderWithOptionsInjectsCSSIntoHead(t *testing.T) {
	// Setup a simple element that represents an HTML document structure
	e := &Element{
		Tag: "html",
		Children: []Node{
			&Element{Tag: "head"},
			&Element{Tag: "body"},
		},
	}

	// Use the MockStyleManager
	mockStyleManager := &MockStyleManager{}

	// Assuming RenderOptions expects a StyleManager interface, pass the mock
	opts := RenderOptions{
		StyleManager: mockStyleManager, // This should be adjusted to how your options are structured
	}
	htmlOutput := e.RenderWithOptions(opts)

	// Construct the expected HTML string with the CSS injected
	expectedHTML := "<!DOCTYPE html><html><head><style>body { background-color: #fff; }</style></head><body></body></html>"

	// Use testify's assert.Equal to check if the HTML output matches the expected HTML
	assert.Equal(t, expectedHTML, htmlOutput, "The generated HTML should include the CSS in the <head> section")
}

func TestRenderRapid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		attrsGen := rapid.MapOfN(rapid.String(), rapid.String(), 0, 4)
		bodyGen := rapid.StringOf(rapid.Rune())
		for i := 0; i < 10; i++ {

			str := Div(attrsGen.Draw(t, "a"), Comment(bodyGen.Draw(t, "t"))).Render()
			fmt.Println(str)

			doc, err := html.Parse(strings.NewReader(str))
			fmt.Println(doc)
			if err != nil {
				t.Logf("invalid html: %s", err)
			}
		}
	})
}
