package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

const countedTextPrelude = `package search
namespace search
entity Text id "gooo://std/text"
entity Integer id "gooo://std/integer"
entity Result id "gooo://search/result" fields {
    field offset id "gooo://search/result/offset" type integer required one
}
`

const countedTextSearch = `activity IndexByte(Text, Integer) -> Integer id "gooo://search/index-byte" computes ` + "`" + `
if input1 < 0 || input1 > 255 { return -1 }
let offset = 0
for offset = 0; offset < 1024; offset = offset + 1 {
    if offset >= int64(len(input0)) { return -1 }
    if int64(input0[offset]) == input1 { return offset }
}
return -1
` + "`" + `
activity Run(Text, Integer) -> Result computes ` + "`" + `return Result{offset: IndexByte(input0, input1)}` + "`\n"

func TestCountedTextSearchMatchesGoByteOffsets(t *testing.T) {
	source := []byte(countedTextPrelude + countedTextSearch)
	generated, err := Generate("index-byte.gooo", source, "Run")
	if err != nil {
		t.Fatal(err)
	}
	if !generated.Report.DeterministicReplay || !generated.Report.RouteEquivalence.Equivalent {
		t.Fatal("source or replay identity lost", generated.Report)
	}
	for _, text := range []string{"", "gooo", "가🙂\x00aa", "a🙂가", strings.Repeat("a", 1023) + "z"} {
		cases := make([]assemblyspec.ValueCase, 0, 258)
		for needle := -1; needle <= 256; needle++ {
			want := -1
			if needle >= 0 && needle <= 255 {
				want = strings.IndexByte(text, byte(needle))
			}
			inputs, _ := json.Marshal([]any{text, needle})
			cases = append(cases, assemblyspec.ValueCase{Inputs: string(inputs), Expected: fmt.Sprintf(`{"offset":%d}`, want)})
		}
		results, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Run", generated.Report.RecordTypes, cases)
		if err != nil {
			t.Fatal(err)
		}
		for _, result := range results {
			if !result.Passed {
				t.Fatal("Go byte search differs", result)
			}
		}
	}
}

func TestTextByteBoundsAndUnsignedArithmetic(t *testing.T) {
	for expression, want := range map[string]int{
		"int64(input[0])": 195, "int64(input[1])": 191,
		"int64(input[0] + input[1])": 130, "int64(-input[0])": 61,
	} {
		source := []byte(countedTextPrelude + "activity Run(Text) -> Result computes `return Result{offset: " + expression + "}`\n")
		generated, err := Generate("byte.gooo", source, "Run")
		if err != nil {
			t.Fatal(err)
		}
		cases := []assemblyspec.ValueCase{{Inputs: `["ÿ"]`, Expected: fmt.Sprintf(`{"offset":%d}`, want)}}
		results, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Run", generated.Report.RecordTypes, cases)
		if err != nil || !results[0].Passed {
			t.Fatal(expression, results, err)
		}
	}
	for _, expression := range []string{"input[-1]", "input[2]", "input[9007199254740993]"} {
		source := []byte(countedTextPrelude + "activity Run(Text) -> Result computes `return Result{offset: int64(" + expression + ")}`\n")
		generated, err := Generate("bounds.gooo", source, "Run")
		if err != nil {
			continue // A statically invalid constant index is rejected before evaluation.
		}
		_, err = evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Run", generated.Report.RecordTypes,
			[]assemblyspec.ValueCase{{Inputs: `["a"]`, Expected: `{"offset":0}`}})
		if err == nil || !strings.Contains(err.Error(), "Text byte offset") {
			t.Fatal("invalid byte index accepted", expression, err)
		}
	}
}

func TestCountedTextSearchKeepsInputLimit(t *testing.T) {
	generated, err := Generate("limit.gooo", []byte(countedTextPrelude+countedTextSearch), "Run")
	if err != nil {
		t.Fatal(err)
	}
	inputs, _ := json.Marshal([]any{strings.Repeat("a", 1025), 97})
	_, err = evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Run", generated.Report.RecordTypes,
		[]assemblyspec.ValueCase{{Inputs: string(inputs), Expected: `{"offset":0}`}})
	if err == nil {
		t.Fatal("Text input beyond the 1024-byte contract was accepted")
	}
}
