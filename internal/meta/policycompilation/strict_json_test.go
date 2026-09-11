package policycompilation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func declaredCaseSourceFixture(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../../examples/meta-policy-compilation/policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func declaredCaseDocumentFixture(t *testing.T, source []byte) []byte {
	t.Helper()
	policy, err := Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(Case{
		ID: "declared-case", ValidatorExpectation: "PASS",
		EvidenceClass: "SYNTHETIC_FIXTURE", Provenance: "CALLER_SUPPLIED",
		ProducerAvailable: true, ConsumerAvailable: true,
		ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest,
		ObservedGeneratedJudgeDigest: policy.SemanticDigest, ObservedIndependentDigest: policy.SemanticDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func declaredCaseStrictIntervention(source []byte) ([]byte, error) {
	// This is a bounded fixture transformation, not a general source rewriter.
	// The native intervention contract permits one transition and its one case
	// decision to change. Missing or ambiguous bindings are errors, not guesses.
	lines := bytes.SplitAfter(source, []byte("\n"))
	transitionLine, decisionLine := -1, -1
	transitionCount, caseCount, decisionCount := 0, 0, 0
	inTargetCase := false
	for index, line := range lines {
		fields := strings.Fields(string(line))
		if len(fields) == 4 && fields[0] == "transition" && fields[1] == `"SEMANTIC_EQUIVALENCE"` && fields[2] == "->" && fields[3] == `"PASS"` {
			transitionLine = index
			transitionCount++
		}
		if len(fields) >= 2 && fields[0] == "case" {
			inTargetCase = fields[1] == `"SEMANTIC_EQUIVALENCE"`
			if inTargetCase {
				caseCount++
			}
		}
		if inTargetCase && len(fields) == 2 && fields[0] == "decision" && fields[1] == `"PASS"` {
			decisionLine = index
			decisionCount++
		}
	}
	if transitionCount != 1 || caseCount != 1 || decisionCount != 1 {
		return nil, fmt.Errorf("strict fixture intervention requires one transition, case, and decision; found %d/%d/%d", transitionCount, caseCount, decisionCount)
	}
	lines[transitionLine] = bytes.Replace(lines[transitionLine], []byte(`"PASS"`), []byte(`"FAIL_CLOSED"`), 1)
	lines[decisionLine] = bytes.Replace(lines[decisionLine], []byte(`"PASS"`), []byte(`"FAIL_CLOSED"`), 1)
	return bytes.Join(lines, nil), nil
}

func TestDeclaredCaseInterventionRejectsUnboundOrAmbiguousTargets(t *testing.T) {
	source := declaredCaseSourceFixture(t)
	transition := []byte(`transition "SEMANTIC_EQUIVALENCE" -> "PASS"`)
	caseHeader := []byte(`case "SEMANTIC_EQUIVALENCE"`)
	for _, test := range []struct {
		name   string
		source []byte
	}{
		{"missing transition", bytes.Replace(source, transition, []byte(`transition "OTHER" -> "PASS"`), 1)},
		{"duplicate transition", append(append([]byte(nil), source...), append([]byte("\n"), transition...)...)},
		{"missing case", bytes.Replace(source, caseHeader, []byte(`case "OTHER"`), 1)},
		{"duplicate case", append(append([]byte(nil), source...), []byte("\n    case \"SEMANTIC_EQUIVALENCE\" {\n        decision \"PASS\"\n    }\n")...)},
		{"missing decision", bytes.ReplaceAll(source, []byte(`decision "PASS"`), []byte(`decision "FAIL_CLOSED"`))},
		{"unrelated case decision", []byte("transition \"SEMANTIC_EQUIVALENCE\" -> \"PASS\"\ncase \"SEMANTIC_EQUIVALENCE\" {\n}\ncase \"OTHER\" {\n decision \"PASS\"\n}\n")},
	} {
		t.Run(test.name, func(t *testing.T) {
			if result, err := declaredCaseStrictIntervention(test.source); err == nil || result != nil {
				t.Fatal("unbound or ambiguous fixture intervention produced a candidate")
			}
		})
	}
}

func TestEvaluateDeclaredCasePreservesSourcePolicyAuthority(t *testing.T) {
	source := declaredCaseSourceFixture(t)
	strict, err := declaredCaseStrictIntervention(source)
	if err != nil {
		t.Fatal(err)
	}
	beforeLines, afterLines := bytes.Split(source, []byte("\n")), bytes.Split(strict, []byte("\n"))
	if len(beforeLines) != len(afterLines) {
		t.Fatal("fixture intervention changed the source line population")
	}
	changed := 0
	for index, before := range beforeLines {
		if bytes.Equal(before, afterLines[index]) {
			continue
		}
		changed++
		if !bytes.Equal(afterLines[index], bytes.Replace(before, []byte(`"PASS"`), []byte(`"FAIL_CLOSED"`), 1)) {
			t.Fatal("fixture intervention changed bytes outside the two decision tokens")
		}
	}
	if changed != 2 {
		t.Fatalf("fixture intervention changed %d lines, want exactly two", changed)
	}
	var baseline DeclaredCaseEvaluation
	for index, policySource := range [][]byte{source, strict} {
		document := declaredCaseDocumentFixture(t, policySource)
		report, err := EvaluateDeclaredCase("policy.gooo", policySource, document, "metapolicycompilation", "metapolicycompilation")
		if err != nil {
			t.Fatal(err)
		}
		if index == 0 {
			baseline = report
		} else if report.SourceDigest == baseline.SourceDigest || report.SemanticDigest == baseline.SemanticDigest {
			t.Fatal("bound policy intervention did not change both source and semantic identities")
		}
		want := []string{"PASS", "FAIL_CLOSED"}[index]
		if report.SourceDecision.Decision != want || report.SourceDecision.MatchedCondition != "SEMANTIC_EQUIVALENCE" {
			t.Fatalf("Gooo source decision = %+v, want %s at SEMANTIC_EQUIVALENCE", report.SourceDecision, want)
		}
		var input Case
		if err := json.Unmarshal(document, &input); err != nil {
			t.Fatal(err)
		}
		policy, err := Compile(policySource)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(report.SourceDecision, EvaluateSourcePolicy(policy, input)) {
			t.Fatal("explanation changed the existing source decision or UNKNOWN context")
		}
		if report.Schema != DeclaredCaseEvaluationSchema || report.SourceDigest != DigestBytes(policySource) || report.SemanticDigest != policy.SemanticDigest || report.InputDigest != DigestBytes(document) || report.EffectiveInputDigest != DigestBytes(document) || len(report.Fields) != 11 {
			t.Fatalf("source/input/schema binding = %+v", report)
		}
		if report.EvaluationScope != "CONDITIONAL_SOURCE_POLICY_EVALUATION" || report.ExternalEvidenceState != "UNKNOWN_NOT_VERIFIED" || report.FullConformanceState != "UNKNOWN_NOT_EXECUTED" || report.GeneratedExecutionObserved || report.MutationAuthority != 0 || report.PromotionAuthority != 0 {
			t.Fatalf("caller declarations were promoted to execution/evidence/authority: %+v", report)
		}
	}
}

func TestEvaluateDeclaredCaseSeparatesMissingNullAndFalse(t *testing.T) {
	source := declaredCaseSourceFixture(t)
	var first DeclaredCaseEvaluation
	digests := map[string]bool{}
	for index, test := range []struct {
		document string
		binding  string
	}{
		{document: "{}", binding: "MISSING_DEFAULTED"},
		{document: "{\"producer_available\":false}", binding: "DECLARED"},
		{document: "{\"producer_available\":null}", binding: "NULL_DEFAULTED"},
		{document: "{\n  \"producer_available\": false\n}", binding: "DECLARED"},
	} {
		report, err := EvaluateDeclaredCase("policy.gooo", source, []byte(test.document), "metapolicycompilation", "metapolicycompilation")
		if err != nil {
			t.Fatal(err)
		}
		if index == 0 {
			first = report
		}
		if report.EffectiveInputDigest != first.EffectiveInputDigest || !reflect.DeepEqual(report.SourceDecision, first.SourceDecision) || digests[report.InputDigest] {
			t.Fatal("raw documents and equal effective inputs were conflated, or source defaults changed")
		}
		digests[report.InputDigest] = true
		found := 0
		for _, field := range report.Fields {
			if field.JSONField != "producer_available" {
				continue
			}
			found++
			if field.Binding != test.binding || field.GoType != "bool" || string(field.Effective) != "false" {
				t.Fatalf("presence or effective value = %+v", field)
			}
			wantSupplied := map[string]string{"MISSING_DEFAULTED": "", "NULL_DEFAULTED": "null", "DECLARED": "false"}[test.binding]
			if string(field.Supplied) != wantSupplied {
				t.Fatalf("supplied value = %q, want %q", field.Supplied, wantSupplied)
			}
		}
		if found != 1 || len(report.Fields) != 11 {
			t.Fatal("Case field binding is not exactly once per existing schema field")
		}
	}
}

func TestEvaluateDeclaredCaseRejectsUnbindableDocuments(t *testing.T) {
	source := declaredCaseSourceFixture(t)
	for _, document := range []string{
		"null",
		"[]",
		"{}{}",
		"{\"unexpected\":true}",
		"{\"producer_available\":\"false\"}",
		"{\"producer_available\":false,\"producer_available\":true}",
		"{\"id\":\"one\",\"ID\":\"two\"}",
		"{\"Producer_Available\":true}",
	} {
		if _, err := EvaluateDeclaredCase("policy.gooo", source, []byte(document), "metapolicycompilation", "metapolicycompilation"); err == nil {
			t.Fatalf("unbindable document was accepted: %s", document)
		}
	}
	if _, err := EvaluateDeclaredCase("policy.gooo", source, []byte("{}"), "different", "metapolicycompilation"); err == nil {
		t.Fatal("declared package identity mismatch was accepted")
	}
}

func TestDecodeStrictJSONRejectsDuplicateObjectKeys(t *testing.T) {
	tests := []struct {
		name  string
		input string
		path  string
	}{
		{
			name:  "same value",
			input: `{"value":"same","value":"same"}`,
			path:  `$["value"]`,
		},
		{
			name:  "conflicting values",
			input: `{"value":"first","value":"second"}`,
			path:  `$["value"]`,
		},
		{
			name:  "escaped key alias",
			input: `{"value":"first","va\u006cue":"second"}`,
			path:  `$["value"]`,
		},
		{
			name:  "nested object",
			input: `{"outer":{"value":"first","value":"second"}}`,
			path:  `$["outer"]["value"]`,
		},
		{
			name:  "nested array object",
			input: `{"items":[{"value":"first","value":"second"}]}`,
			path:  `$["items"][0]["value"]`,
		},
	}
	type nestedValue struct {
		Value string `json:"value"`
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var value struct {
				Value string        `json:"value"`
				Outer nestedValue   `json:"outer"`
				Items []nestedValue `json:"items"`
			}
			err := decodeStrictJSON([]byte(test.input), &value)
			if err == nil {
				t.Fatalf("duplicate object key input %q was accepted", test.input)
			}
			if !strings.Contains(err.Error(), `duplicate JSON object key "value"`) {
				t.Fatalf("error = %q, want duplicate-key diagnostic", err)
			}
			if !strings.Contains(err.Error(), test.path) {
				t.Fatalf("error = %q, want location %q", err, test.path)
			}
		})
	}
}

func TestDecodeStrictJSONAllowsSeparateObjectsAndExplicitZeroValues(t *testing.T) {
	var values []struct {
		Value string `json:"value"`
	}
	if err := decodeStrictJSON([]byte(`[{"value":"first"},{"value":"second"}]`), &values); err != nil {
		t.Fatalf("same key in separate objects = %v, want valid document", err)
	}
	if len(values) != 2 || values[0].Value != "first" || values[1].Value != "second" {
		t.Fatalf("separate object values = %#v, want both values preserved", values)
	}

	var explicit struct {
		Flag  bool   `json:"flag"`
		Value string `json:"value"`
	}
	if err := decodeStrictJSON([]byte(`{"flag":false,"value":""}`), &explicit); err != nil {
		t.Fatalf("explicit false/empty values = %v, want valid document", err)
	}
	if explicit.Flag || explicit.Value != "" {
		t.Fatalf("explicit zero values = %#v, want false and empty", explicit)
	}
	var missing struct {
		Flag  bool   `json:"flag"`
		Value string `json:"value"`
	}
	if err := decodeStrictJSON([]byte(`{}`), &missing); err != nil {
		t.Fatalf("missing values = %v, want valid document", err)
	}
	if missing.Flag || missing.Value != "" {
		t.Fatalf("missing values = %#v, want zero values", missing)
	}
}

func TestDecodeStrictJSONRetainsUnknownFieldRejection(t *testing.T) {
	var value struct {
		Value string `json:"value"`
	}
	if err := decodeStrictJSON([]byte(`{"value":"ok","unknown":true}`), &value); err == nil {
		t.Fatal("unknown field was accepted")
	}
}

func TestDecodeStrictJSONAcceptsWhitespaceAndRejectsTrailingDocuments(t *testing.T) {
	var value struct {
		Value string `json:"value"`
	}
	if err := decodeStrictJSON([]byte("{\"value\":\"ok\"}\n \t"), &value); err != nil || value.Value != "ok" {
		t.Fatalf("whitespace suffix = %#v, want valid document", err)
	}
	for _, input := range []string{"{\"value\":\"ok\"}{}", "{\"value\":\"ok\"} trailing"} {
		var got struct {
			Value string `json:"value"`
		}
		if err := decodeStrictJSON([]byte(input), &got); err == nil {
			t.Fatalf("trailing input %q was accepted", input)
		}
	}
}

func TestGeneratedJudgeRejectsTrailingDocuments(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-policy-compilation/policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	judgePath := filepath.Join(t.TempDir(), "judge.go")
	if err := os.WriteFile(judgePath, GenerateJudge(policy), 0o600); err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(filepath.Dir(judgePath), "judge")
	build := exec.Command("go", "build", "-o", binaryPath, judgePath)
	build.Env = append(os.Environ(), "GO111MODULE=off", "GOTOOLCHAIN=go1.27.0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build generated judge: %v: %s", err, output)
	}
	input, err := json.Marshal(map[string]string{"id": "strict-eof"})
	if err != nil {
		t.Fatal(err)
	}
	run := func(suffix string) error {
		command := exec.Command(binaryPath)
		command.Stdin = bytes.NewBuffer(append(append(append([]byte{}, input...), '\n'), []byte(suffix)...))
		return command.Run()
	}
	if err := run(" \t\n"); err != nil {
		t.Fatalf("whitespace suffix rejected: %v", err)
	}
	if err := run("{}\n"); err == nil {
		t.Fatal("second JSON document was accepted")
	}
	if err := run("trailing"); err == nil {
		t.Fatal("malformed trailing bytes were accepted")
	}
}

func TestGeneratedJudgeRejectsDuplicateObjectKeys(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-policy-compilation/policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	tempDir := t.TempDir()
	judgePath := filepath.Join(tempDir, "judge.go")
	if err := os.WriteFile(judgePath, GenerateJudge(policy), 0o600); err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(tempDir, "judge")
	build := exec.Command("go", "build", "-o", binaryPath, judgePath)
	build.Env = append(os.Environ(), "GO111MODULE=off", "GOTOOLCHAIN=go1.27.0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build generated judge: %v: %s", err, output)
	}
	run := func(input string) (string, error) {
		command := exec.Command(binaryPath)
		command.Stdin = strings.NewReader(input)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		err := command.Run()
		return stderr.String(), err
	}
	for _, test := range []struct {
		name  string
		input string
		key   string
		path  string
	}{
		{name: "same value", input: `{"id":"same","id":"same"}`, key: "id", path: `$["id"]`},
		{name: "escaped key alias", input: `{"id":"first","i\u0064":"second"}`, key: "id", path: `$["id"]`},
		{name: "repeated boolean", input: `{"producer_available":false,"producer_available":true}`, key: "producer_available", path: `$["producer_available"]`},
	} {
		stderr, err := run(test.input)
		t.Run(test.name, func(t *testing.T) {
			if err == nil {
				t.Fatalf("duplicate object key input %q was accepted", test.input)
			}
			var exitError *exec.ExitError
			if !errors.As(err, &exitError) || exitError.ExitCode() != 2 {
				t.Fatalf("error = %v, want process exit status 2", err)
			}
			if !strings.Contains(stderr, `duplicate JSON object key "`+test.key+`"`) {
				t.Fatalf("stderr = %q, want duplicate-key diagnostic", stderr)
			}
			if !strings.Contains(stderr, test.path) || !strings.Contains(stderr, "byte offset") {
				t.Fatalf("stderr = %q, want path %q and byte offsets", stderr, test.path)
			}
		})
	}
	if stderr, err := run(`{"id":"","producer_available":false,"consumer_available":false}`); err != nil {
		t.Fatalf("explicit false/empty generated input = %v, want accepted", err)
	} else if stderr != "" {
		t.Fatalf("explicit false/empty generated input stderr = %q, want empty", stderr)
	}
}
