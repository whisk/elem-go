package elem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIf(t *testing.T) {
	trueElement := Div(nil, Text("True Condition"))
	falseElement := Div(nil, Text("False Condition"))

	resultTrue := If(true, trueElement, falseElement)
	assert.Equal(t, trueElement.Render(), resultTrue.Render())

	resultFalse := If(false, trueElement, falseElement)
	assert.Equal(t, falseElement.Render(), resultFalse.Render())
}

func TestIfUsingNone(t *testing.T) {
	trueElement := Div(nil, Text("True Condition"))

	resultWithNone := If[Node](true, trueElement, None())
	assert.Equal(t, trueElement.Render(), resultWithNone.Render(), "If should render the true element when the condition is true")

	resultWithNoneFalse := If[Node](false, trueElement, None())
	assert.Equal(t, "", resultWithNoneFalse.Render(), "If should render nothing (empty string) when the condition is false and None is used")
}

func TestTransformEach(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}

	elements := TransformEach(items, func(item string) Node {
		return Li(nil, Text(item))
	})

	assert.Equal(t, len(items), len(elements))

	for i, el := range elements {
		expected := `<li>` + items[i] + `</li>`
		assert.Equal(t, expected, el.Render())
	}
}

func TestTransformEachEmpty(t *testing.T) {
	toLi := func(item string) Node { return Li(nil, Text(item)) }

	assert.Nil(t, TransformEach([]string{}, toLi), "TransformEach should return nil for an empty slice")
	assert.Nil(t, TransformEach(nil, toLi), "TransformEach should return nil for a nil slice")
}

func TestValidAttrName(t *testing.T) {
	valid := []string{
		"id",
		"class",
		"data-user-id",
		"data-user_id",
		"aria-label",
		"hx-on:click",
		"hx-on--after-request",
		"@click",
		":class",
		"xml:lang",
		"data-名前",
		"data-\u00a0",
		"data-\ufdcf",
		"data-\ufdf0",
		"data-\ufffd",
	}
	for _, name := range valid {
		assert.True(t, validAttrName(name), "expected %q to be valid", name)
	}

	invalid := []string{
		"",
		" ",
		"id onclick=alert(1)",
		"id\tonclick=alert(1)",
		"id\nonclick=alert(1)",
		"id\ronclick=alert(1)",
		"id\fonclick=alert(1)",
		"id\x00",
		"id\x1f",
		"id\x7f",
		"id\u0080",
		"id\u009f",
		"id\ufdd0",
		"id\ufdef",
		"id\ufffe",
		"id\uffff",
		"id\U0001fffe",
		"id\U0010ffff",
		"id\x80",
		"id\xff",
		"id\xc0\xaf",
		"id\xed\xa0\x80",
		`id"`,
		`id'`,
		`id/`,
		`id<`,
		`id=`,
		`id>`,
		"data-名前 foo",
	}
	for _, name := range invalid {
		assert.False(t, validAttrName(name), "expected %q to be invalid", name)
		placeholder := invalidAttrName(name)
		assert.True(t, validAttrName(placeholder), "expected placeholder %q for %q to be valid", placeholder, name)
	}
}

func TestInvalidAttrName(t *testing.T) {
	cases := []struct {
		in       string
		expected string
	}{
		{in: "", expected: "data-invalid-attr"},
		{in: " ", expected: "data-invalid-attr-_"},
		{in: "bad attr", expected: "data-invalid-attr-bad_attr"},
		{in: "onclick=alert(1) x", expected: "data-invalid-attr-onclick_alert_1__x"},
		{in: `data-x"><script>`, expected: "data-invalid-attr-data-x___script_"},
		{in: "data-名前 foo", expected: "data-invalid-attr-data-___foo"},
		{in: "data-\xff", expected: "data-invalid-attr-data-_"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.expected, invalidAttrName(tc.in))
	}
}

func TestValidTagName(t *testing.T) {
	valid := []string{
		"div",
		"h1",
		"DIV",
		"my-element",
		"my-el_1",
		"div_name",
		"x-card-2",
	}
	for _, tag := range valid {
		assert.True(t, validTagName(tag), "expected %q to be valid", tag)
	}

	invalid := []string{
		"",
		"1div",
		"-div",
		"_div",
		"div onclick=alert(1)",
		"div><b>hi</b><div",
		"div/class",
		"div\"a",
		"div'a",
		"div\n",
		"div<",
		"div=",
		"div:name",
		"div.name",
	}
	for _, tag := range invalid {
		assert.False(t, validTagName(tag), "expected %q to be invalid", tag)
	}
}
