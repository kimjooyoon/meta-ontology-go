package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func TestBoundedTextGeneratedInterpretationFresh1905(t *testing.T) {
	source, err := os.ReadFile("../../examples/bounded-text/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate("bounded.gooo", source, "Prefix")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(generated.Source, `gooo_max_utf8_bytes:"1030"`) {
		t.Fatal("body projection lost its source field description", generated.Source)
	}
	var cases []assemblyspec.ValueCase
	for _, value := range []string{"", strings.Repeat("x", 1024), strings.Repeat("한", 341) + "x", strings.Repeat("\x00", 1024)} {
		input, _ := json.Marshal([]string{value})
		expected, _ := json.Marshal(map[string]string{"value": "prefix" + value})
		cases = append(cases, assemblyspec.ValueCase{Inputs: string(input), Expected: string(expected)})
	}
	observed, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Prefix", generated.Report.RecordTypes, cases)
	if err != nil || len(observed) != len(cases) {
		t.Fatal("interpreter did not complete the declared cases", err, observed)
	}
	for i, result := range observed {
		if !result.Passed || !bytes.Equal(result.Actual, []byte(cases[i].Expected)) {
			t.Fatal("interpreted value differs from independent expectation", i, result)
		}
	}
}

func TestBoundedTextInterpreterRefusalsFresh1905(t *testing.T) {
	source, err := os.ReadFile("../../examples/bounded-text/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, prefix, input, reason, expected string
	}{
		{"ingress", "prefix", strings.Repeat("x", 1025), "1024", `{"value":"sentinel"}`},
		{"output", "prefix!", strings.Repeat("x", 1024), "gooo://examples/bounded-text/expanded/value", `{"value":"sentinel"}`},
		{"undeclared", "prefix", "", "declared", `{"value":"sentinel","extra":""}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := bytes.Replace(source, []byte(`"prefix"`), []byte(`"`+tc.prefix+`"`), 1)
			generated, err := Generate("bounded.gooo", changed, "Prefix")
			if err != nil {
				t.Fatal(err)
			}
			input, _ := json.Marshal([]string{tc.input})
			_, err = evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Prefix", generated.Report.RecordTypes,
				[]assemblyspec.ValueCase{{Inputs: string(input), Expected: tc.expected}})
			if err == nil || !strings.Contains(err.Error(), tc.reason) {
				t.Fatal("interpreter admitted a value outside the declared boundary", err)
			}
		})
	}
}

func TestTextBoundGenerationIdentityFresh1905(t *testing.T) {
	source, err := os.ReadFile("../../examples/bounded-text/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	before, err := Generate("bounded.gooo", source, "Prefix")
	if err != nil {
		t.Fatal(err)
	}
	after, err := Generate("bounded.gooo", bytes.Replace(source, []byte(`"1030"`), []byte(`"1031"`), 1), "Prefix")
	if err != nil {
		t.Fatal(err)
	}
	if before.Report.SourceDigest == after.Report.SourceDigest || before.Report.PlanSHA256 == after.Report.PlanSHA256 {
		t.Fatal("a changed source domain reused the old generation identity")
	}
	if before.Report.ProgramDigest != after.Report.ProgramDigest || before.Report.RecordTypes[0].Fields[0].TextBound == after.Report.RecordTypes[0].Fields[0].TextBound {
		t.Fatal("value-program identity and declared field domain were conflated")
	}
}

func TestBoundedInterpreterRefusesMalformedUTF8ResultFresh1905(t *testing.T) {
	source, err := os.ReadFile("../../examples/bounded-text/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	source = bytes.Replace(source, []byte(`"prefix" + input`), []byte(`input[1:]`), 1)
	generated, err := Generate("bounded.gooo", source, "Prefix")
	if err != nil {
		t.Fatal(err)
	}
	_, err = evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Prefix", generated.Report.RecordTypes,
		[]assemblyspec.ValueCase{{Inputs: `["한"]`, Expected: `{"value":""}`}})
	if err == nil || !strings.Contains(err.Error(), "valid UTF-8") {
		t.Fatal("interpreter accepted a malformed string result", err)
	}
}
