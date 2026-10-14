package bodycodegen

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestScopedCounterBoundaries18417(t *testing.T) {
	invalid := []struct{ name, body string }{
		{"escaped_read", `let total=input;for let i=0;i<3;i++ {total=total+i};return i`},
		{"escaped_write", `let total=input;for let i=0;i<3;i++ {total=total+i};i=9;return total`},
		{"body_mutation", `let total=input;for let i=0;i<3;i++ {i=i+2};return total`},
		{"body_shadow", `let total=input;for let i=0;i<3;i++ {let i=1};return total`},
		{"outer_bound", `let i=13;for let i=0;i<3;i++ {};return i`},
		{"parameter_bound", `for let input=0;input<3;input++ {};return input`},
		{"decrement", `let total=input;for let i=0;i<3;i-- {total=total+i};return total`},
		{"large_bound", `let total=input;for let i=0;i<1025;i++ {total=total+i};return total`},
		{"dynamic_bound", `let total=input;for let i=0;i<input;i++ {total=total+i};return total`},
		{"nested", `let total=input;for let i=0;i<3;i++ {for let j=0;j<2;j++ {total=total+j}};return total`},
		{"nonzero", `let total=input;for let i=1;i<3;i++ {total=total+i};return total`},
	}
	for n, c := range invalid {
		source := []byte(fmt.Sprintf("package bounded\nnamespace bounded18417_%d\nentity Integer id \"urn:gooo:type:integer\"\nactivity Sum(Integer)->Integer computes %s\n", n, strconv.Quote(c.body)))
		if _, err := Generate("boundary18417.gooo", source, "Sum"); err == nil {
			t.Fatal("invalid loop/source scope accepted", c.name)
		} else {
			t.Log(c.name, err)
		}
	}
	const body = `let note="for let i=0;i<3;i++";return note`
	normalized, _, err := NormalizeBodySource(body)
	if err != nil || !strings.Contains(normalized, `"for let i=0;i<3;i++"`) {
		t.Fatal("string literal changed", normalized, err)
	}
	t.Log("11 new invalid source programs rejected;finite1024 bound,nesting,scope,readonly and step restrictions retained;literal text untouched;native/model0")
}
