package bodycodegen

import (
	"strings"
	"testing"
)

const lengthContextPrelude20111 = `package lengthcontext
namespace lengthcontext20111
entity Text id "urn:lengthcontext20111:text"
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
entity Summary id "urn:lengthcontext20111:summary" fields {
 field size id "urn:lengthcontext20111:summary/size" type integer required one
 field present id "urn:lengthcontext20111:summary/present" type boolean required one
}
`

func TestLengthIntegerContexts20111(t *testing.T) {
	tests := []struct{ name, signature, body, extra, want string }{
		{"return", "(Text)->Integer", "return len(input)+20111", "", "int64(len(input) + 20111)"},
		{"mixed sum", "(Text,Integer)->Integer", "return len(input0)+input1", "", "int64(len(input0)) + input1"},
		{"derived local", "(Text,Integer)->Integer", "let n=len(input0);let m=n+2;let total=m+input1;return total", "", "int64(m) + input1"},
		{"comparison", "(Text,Integer)->Boolean", "return input1<len(input0)", "", "input1 < int64(len(input0))"},
		{"reverse comparison", "(Text,Integer)->Boolean", "return len(input0)>=input1", "", "int64(len(input0)) >= input1"},
		{"record constructor", "(Text)->Summary", "return Summary{size:len(input),present:true}", "", "int64(len(input))"},
		{"record assignment", "(Text)->Summary", "let out=Summary{size:0,present:true};out.size=len(input);return out", "", "int64(len(input))"},
		{"local assignment", "(Text)->Integer", "let n=20111;n=len(input);return n", "", "n = int64(len(input))"},
		{"pure helper", "(Text)->Integer", "return Add(len(input))", "activity Add(Integer)->Integer computes `return input+20111`\n", "Add(int64(len(input)))"},
		{"loop guard", "(Text)->Integer", "let total=20111;for let i=0;i<5;i++ {if i<len(input) {total=total+i}};return total", "", "i < int64(len(input))"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := []byte(lengthContextPrelude20111 + test.extra + "activity Run" + test.signature + " computes `" + test.body + "`\n")
			result, err := Generate("길이20111.gooo", source, "Run")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(result.Source, test.want) || !result.Report.RouteEquivalence.Equivalent ||
				result.Report.CompletenessPercent != 100 || !result.Report.DeterministicReplay {
				t.Fatal("Integer context, replay, source accounting", result.Source, result.Report)
			}
			for _, field := range result.Report.BodyFields {
				if strings.Contains(field.Value.Expression, "int64(len") {
					t.Fatal("implicit projection conversion claimed as authored field expression", field)
				}
			}
		})
	}
	t.Log("ten new Gooo length contexts;Integer compare/return/assignment/constructor/helper and scoped-loop guard;canonical units100%;model/native0")
}

func TestLengthIntegerRejectsUnrelatedTypes20111(t *testing.T) {
	for _, body := range []string{
		"return len(input)==true", "return len(true)", "return Summary{size:len(input),present:len(input)}",
	} {
		output := "Integer"
		if strings.Contains(body, "Summary{") {
			output = "Summary"
		}
		source := []byte(lengthContextPrelude20111 + "activity Bad(Text)->" + output + " computes `" + body + "`\n")
		if _, err := Generate("invalid20111.gooo", source, "Bad"); err == nil {
			t.Fatal("length widening accepted a wrong Boolean/Text contract", body)
		}
	}
}
