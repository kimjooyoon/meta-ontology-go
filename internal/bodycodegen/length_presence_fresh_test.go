package bodycodegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestLengthPresence21983(t *testing.T) {
	for _, test := range []struct {
		source string
		want   bool
	}{
		{`package p;func Plain(input int64)int64{return input+21983}`, false},
		{`package p;func Text()string{return "len(input)"}`, false},
		{"package p\n// len(input)\nfunc Comment()int64{return 21983}", false},
		{`package p;func Count(input string)int{return len(input)}`, true},
		{`package p;func len(input string)int64{return 21983};func Named(input string)int64{return len(input)}`, true},
	} {
		file, err := parser.ParseFile(token.NewFileSet(), "presence21983.go", test.source, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		if found := bodyMayUseLength(file); found != test.want {
			t.Fatal("syntactic length eligibility", found, test.want, test.source)
		}
	}
	const prefix = "package presence\nnamespace presence21983\nentity Integer id \"urn:gooo:type:integer\"\nentity Text id \"urn:presence21983:text\"\n"
	for _, activity := range []string{
		"activity Run(Integer)->Integer computes `let n=input+21983;return n`\n",
		"activity Run(Text,Integer)->Integer computes `return input1+len(input0)`\n",
	} {
		result, err := Generate("presence21983.gooo", []byte(prefix+activity), "Run")
		if err != nil || result.Report.CompletenessPercent != 100 || !result.Report.RouteEquivalence.Equivalent {
			t.Fatal("fresh scalar/length paths", err, result.Report)
		}
		if strings.Contains(activity, "len(") && !strings.Contains(result.Source, "int64(len(input0))") {
			t.Fatal("length path was skipped", result.Source)
		}
	}
	t.Log("five new AST inputs distinguish calls from literals/comments;fresh scalar and mixed-length source paths preserve equivalence;no native/model or historical cohort execution")
}
