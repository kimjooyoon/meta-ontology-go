package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// Compile the emitted Go independently. Expected values are authored here;
// neither candidate observations nor generated assertions supply the oracle.
func assertNominalNativeCases(t *testing.T, result Result, input string, cases [][2]string) {
	t.Helper()
	name := recordTypeByName(result.Report.RecordTypes, input)
	if name == nil {
		t.Fatal("missing input record", input)
	}
	dir := t.TempDir()
	var body strings.Builder
	// The generated package name is independent of source record display names.
	start := strings.Index(result.Source, "package ") + len("package ")
	end := strings.IndexByte(result.Source[start:], '\n') + start
	fmt.Fprintf(&body, "package %s\nimport (\"testing\";\"encoding/json\";\"reflect\";\"strings\")\n", result.Source[start:end])
	body.WriteString("func TestIndependentValues(t *testing.T) {\n")
	for i, c := range cases {
		fmt.Fprintf(&body, `{
 var input %s
 if err := json.Unmarshal([]byte(%s), &input); err != nil { t.Fatal(err) }
 actual, err := json.Marshal(%s(input))
 if err != nil { t.Fatal(err) }
 var got, want any
 decoder := json.NewDecoder(strings.NewReader(string(actual)))
 decoder.UseNumber()
 if err = decoder.Decode(&got); err != nil { t.Fatal(err) }
 decoder = json.NewDecoder(strings.NewReader(%s))
 decoder.UseNumber()
 if err = decoder.Decode(&want); err != nil { t.Fatal(err) }
 if !reflect.DeepEqual(got, want) { t.Fatalf("case %d: %%s", actual) }
}
`, name.GoName, strconv.Quote(c[0]), result.Report.Activity, strconv.Quote(c[1]), i)
	}
	body.WriteString("}\n")
	for name, content := range map[string]string{"go.mod": "module nominal-oracle\n\ngo 1.27.3\n", "generated.go": result.Source, "generated_test.go": body.String()} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-p", "1", "./...")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("independent generated Go: %v\n%s", err, output)
	}
}

func TestNominalRecordAssemblyNativeItemSum(t *testing.T) {
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.GenerateSourceAssembly(context.Background(), "nested.gooo", nominalRecordFixture(t), "Total")
	if err != nil {
		t.Fatal(err)
	}
	assertNominalNativeCases(t, r, "Basket", [][2]string{
		{`{"items":[]}`, `{"total":0}`},
		{`{"items":[{"quantity":2,"price":7}]}`, `{"total":14}`},
		{`{"items":[{"quantity":2,"price":7},{"quantity":2,"price":7}]}`, `{"total":28}`},
		{`{"items":[{"quantity":1,"price":9007199254740993}]}`, `{"total":9007199254740993}`},
	})
}

func nominalCopySource(body string, cases [][2]string) []byte {
	source := `package copyrecords
namespace copyrecords
entity Integer id "urn:gooo:type:integer"
entity Line id "urn:copy:line" fields {
 field amount id "urn:copy:amount" type integer required one
}
entity Envelope id "urn:copy:envelope" fields {
 field child id "urn:copy:child" type Line required one
 field maybe id "urn:copy:maybe" type Line optional one
 field items id "urn:copy:items" type Line required many
 field more id "urn:copy:more" type Line optional many
}
`
	source += "activity Copy(Envelope) -> Envelope computes " + strconv.Quote(body) + " assembling {\n"
	source += ` choice "child" field_value at "0" alternative "Line{amount: -1}" intent "Preserve the original child record. 원래 자식 레코드를 보존한다."` + "\n"
	for _, c := range cases {
		source += " value_case " + strconv.Quote("["+c[0]+"]") + " -> " + strconv.Quote(c[1]) + "\n"
	}
	return []byte(source + " attempts \"2\"\n}\n")
}

func TestNominalRecordAssemblyPreservesAllFourShapes(t *testing.T) {
	cases := [][2]string{
		{`{"child":{"amount":0},"items":[]}`, `{"child":{"amount":0},"items":[]}`},
		{`{"child":{"amount":1},"maybe":{"amount":0},"items":[{"amount":2},{"amount":2}],"more":[]}`, `{"child":{"amount":1},"maybe":{"amount":0},"items":[{"amount":2},{"amount":2}],"more":[]}`},
		{`{"child":{"amount":3},"items":[],"more":[{"amount":9007199254740993}]}`, `{"child":{"amount":3},"items":[],"more":[{"amount":9007199254740993}]}`},
	}
	source := nominalCopySource("return Envelope{child: input.child, maybe: input.maybe, items: input.items, more: input.more}", cases)
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "copy.gooo", source, "Copy")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.Passed != 3 || r.Total != 3 || r.ModelCalls != 0 || r.FieldsPassed != 12 {
		t.Fatalf("nominal transport: %+v", r)
	}
	for _, c := range r.Cases {
		for _, field := range c.Fields {
			if field.TypeID != "urn:copy:line" {
				t.Fatalf("lost nominal type identity: %+v", field)
			}
		}
	}
	assertNominalNativeCases(t, result, "Envelope", cases)
	if _, err := RealizeSourceAssembly(context.Background(), "copy.gooo", source, result); err != nil {
		t.Fatal(err)
	}
	var present map[string]json.RawMessage
	if err := json.Unmarshal(r.Cases[1].Actual, &present); err != nil || string(present["more"]) != "[]" {
		t.Fatal("present empty optional record array disappeared", err)
	}
}

func TestNominalRecordAssemblyOptionalPresenceAndChildReads(t *testing.T) {
	cases := [][2]string{
		{`{"child":{"amount":7},"items":[]}`, `{"child":{"amount":7},"items":[]}`},
		{`{"child":{"amount":7},"maybe":{"amount":0},"items":[]}`, `{"child":{"amount":0},"maybe":{"amount":0},"items":[]}`},
	}
	body := `let amount = 0
 if input.maybe != nil { amount = input.maybe.amount } else { amount = input.child.amount }
 return Envelope{child: Line{amount: amount}, maybe: input.maybe, items: input.items, more: input.more}`
	assertNominalAssemblyProgram(t, nominalCopySource(body, cases), cases)
}

func TestNominalRecordAssemblyChildCopyAndRecordArrayLiteral(t *testing.T) {
	cases := [][2]string{
		{`{"child":{"amount":7},"items":[]}`, `{"child":{"amount":8},"maybe":{"amount":7},"items":[{"amount":7},{"amount":7}]}`},
		{`{"child":{"amount":9007199254740993},"items":[],"more":[]}`, `{"child":{"amount":9007199254740994},"maybe":{"amount":9007199254740993},"items":[{"amount":9007199254740993},{"amount":9007199254740993}],"more":[]}`},
	}
	body := `let original = input.child
 let child = original
 child.amount = child.amount + 1
 return Envelope{child: child, maybe: &original, items: []Line{original, original}, more: input.more}`
	assertNominalAssemblyProgram(t, nominalCopySource(body, cases), cases)
}

func assertNominalAssemblyProgram(t *testing.T, source []byte, cases [][2]string) {
	t.Helper()
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.GenerateSourceAssembly(context.Background(), "nominal.gooo", source, "Copy")
	if err != nil {
		t.Fatal(err)
	}
	if r.Report.RecordAssembly.Passed != len(cases) || r.Report.RecordAssembly.ModelCalls != 0 {
		t.Fatalf("nominal evaluator mismatch: %+v", r.Report.RecordAssembly)
	}
	assertNominalNativeCases(t, r, "Envelope", cases)
	if _, err := RealizeSourceAssembly(context.Background(), "nominal.gooo", source, r); err != nil {
		t.Fatal(err)
	}
}

func TestNominalRecordAssemblyIndexedRecordRead(t *testing.T) {
	cases := [][2]string{
		{`{"child":{"amount":7},"items":[{"amount":0}]}`, `{"child":{"amount":0},"items":[{"amount":0}]}`},
		{`{"child":{"amount":7},"items":[{"amount":9007199254740993}]}`, `{"child":{"amount":9007199254740993},"items":[{"amount":9007199254740993}]}`},
	}
	body := `return Envelope{child: input.items[0], maybe: input.maybe, items: input.items, more: input.more}`
	assertNominalAssemblyProgram(t, nominalCopySource(body, cases), cases)
}

func TestNominalRecordAssemblyEqualityUsesRequiredChildValues(t *testing.T) {
	source := `package nestedcompare
 namespace nestedcompare
 entity Boolean id "urn:gooo:type:boolean"
 entity Line id "urn:eq:line" fields { field amount id "urn:eq:amount" type integer required one }
 entity Box id "urn:eq:box" fields { field line id "urn:eq:field-line" type Line required one }
 entity Result id "urn:eq:result" fields { field same id "urn:eq:same" type boolean required one }
 `
	body := `let same = input == Box{line: Line{amount: input.line.amount}}
 return Result{same: same}`
	cases := [][2]string{
		{`{"line":{"amount":0}}`, `{"same":true}`},
		{`{"line":{"amount":9007199254740993}}`, `{"same":true}`},
	}
	source += "activity Same(Box) -> Result computes " + strconv.Quote(body) + " assembling {\n"
	source += `choice "same" field_value at "2" alternative "!same" intent "Compare the required record values. 필수 레코드의 값을 비교한다."` + "\n"
	for _, c := range cases {
		source += "value_case " + strconv.Quote("["+c[0]+"]") + " -> " + strconv.Quote(c[1]) + "\n"
	}
	source += "attempts \"2\"\n}\n"
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "eq.gooo", []byte(source), "Same")
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.RecordAssembly.Passed != len(cases) {
		t.Fatal("required record equality used pointer identity")
	}
	assertNominalNativeCases(t, result, "Box", cases)
}
