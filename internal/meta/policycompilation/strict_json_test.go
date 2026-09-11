package policycompilation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGeneratedJudgeInputFieldsFollowTypedSchema(t *testing.T) {
	policy, err := Compile(declaredCaseSourceFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	names := []string{"runtime-input", "renamed-fields", "reordered-and-retyped-fields"}
	for index, shape := range []reflect.Type{
		reflect.TypeFor[generatedJudgeInput](),
		reflect.StructOf([]reflect.StructField{
			{Name: "Enabled", Type: reflect.TypeFor[bool](), Tag: "json:\"enabled\""},
			{Name: "Label", Type: reflect.TypeFor[string](), Tag: "json:\"label\""},
		}),
		reflect.StructOf([]reflect.StructField{
			{Name: "Label", Type: reflect.TypeFor[bool](), Tag: "json:\"renamed_label,omitempty\""},
			{Name: "Enabled", Type: reflect.TypeFor[string](), Tag: "json:\"renamed_enabled\""},
		}),
	} {
		t.Run(names[index], func(t *testing.T) {
			source := "package probe\n\ntype input struct {\n" + generatedJudgeInputFields(shape) + "}\n"
			if index == 0 {
				source = string(GenerateJudge(policy))
			}
			file, err := parser.ParseFile(token.NewFileSet(), "judge.go", source, parser.SkipObjectResolution)
			if err != nil {
				t.Fatal(err)
			}
			var fields *ast.FieldList
			for _, declaration := range file.Decls {
				group, ok := declaration.(*ast.GenDecl)
				if !ok || group.Tok != token.TYPE {
					continue
				}
				for _, specification := range group.Specs {
					named, ok := specification.(*ast.TypeSpec)
					if !ok || named.Name.Name != "input" {
						continue
					}
					structure, ok := named.Type.(*ast.StructType)
					if !ok || fields != nil {
						t.Fatal("generated input declaration is not one struct")
					}
					fields = structure.Fields
				}
			}
			if fields == nil || len(fields.List) != shape.NumField() {
				t.Fatal("generated input field count differs from its typed schema")
			}
			for fieldIndex, observed := range fields.List {
				expected := shape.Field(fieldIndex)
				if len(observed.Names) != 1 || observed.Names[0].Name != expected.Name {
					t.Fatalf("generated input field order/name differs at %d", fieldIndex)
				}
				kind, ok := observed.Type.(*ast.Ident)
				if !ok || kind.Name != expected.Type.String() || observed.Tag == nil {
					t.Fatalf("generated input field type/tag differs at %d", fieldIndex)
				}
				var tag string
				count, err := fmt.Sscanf(observed.Tag.Value, "%q", &tag)
				if err != nil || count != 1 || tag != string(expected.Tag) {
					t.Fatalf("generated JSON tag differs at %d: %q (%v)", fieldIndex, tag, err)
				}
			}
		})
	}
}

func TestGeneratedJudgeDeclaredInputPreservesBindingsAndAuthority(t *testing.T) {
	source := declaredCaseSourceFixture(t)
	policy, err := Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	judgePath := filepath.Join(work, "judge.go")
	binary := filepath.Join(work, "judge-test")
	if err := os.WriteFile(judgePath, GenerateJudge(policy), 0o600); err != nil {
		t.Fatal(err)
	}
	// Compile once for this corpus; each input runs the same generated binary.
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, judgePath)
	build.Dir = work
	build.Env = append(os.Environ(), "GO111MODULE=off", "GOTOOLCHAIN=go1.27.0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build generated judge: %v: %s", err, output)
	}
	run := func(document []byte, arguments ...string) ([]byte, error) {
		command := exec.CommandContext(t.Context(), binary, arguments...)
		command.Dir = work
		command.Stdin = bytes.NewReader(document)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		output, err := command.Output()
		if err != nil {
			return output, fmt.Errorf("generated judge: %w: %s", err, stderr.String())
		}
		return output, nil
	}
	encode := func(value generatedJudgeInput) []byte {
		t.Helper()
		document, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return document
	}
	t.Run("input-schema", func(t *testing.T) {
		output, err := run(nil, "--input-schema")
		if err != nil {
			t.Fatal(err)
		}
		var envelope map[string]json.RawMessage
		if err := decodeStrictJSON(output, &envelope); err != nil || len(envelope) != 9 {
			t.Fatalf("input schema: %v: %s", err, output)
		}
		for key, want := range map[string]string{
			"schema":          "gooo/generated-policy-input-schema/v1",
			"source_digest":   policy.SourceDigest,
			"semantic_digest": policy.SemanticDigest,
			"evaluation_mode": "--declared-input",
		} {
			var got string
			if err := json.Unmarshal(envelope[key], &got); err != nil || got != want {
				t.Fatalf("%s = %s (%v), want %q", key, envelope[key], err, want)
			}
		}
		if string(envelope["policy_evaluation_observed"]) != "false" || string(envelope["mutation_authority"]) != "0" || string(envelope["promotion_authority"]) != "0" {
			t.Fatal("input discovery was promoted into policy evaluation or authority")
		}
		if !bytes.Equal(envelope["default_input"], encode(generatedJudgeInput{})) {
			t.Fatal("input schema defaults differ from the typed runtime ABI")
		}
		var fields []map[string]json.RawMessage
		if err := decodeStrictJSON(envelope["fields"], &fields); err != nil {
			t.Fatal(err)
		}
		shape := reflect.TypeFor[generatedJudgeInput]()
		if len(fields) != shape.NumField() {
			t.Fatal("input schema field count differs from the typed runtime ABI")
		}
		for index, field := range fields {
			expected := shape.Field(index)
			if len(field) != 5 || string(field["required"]) != "false" || string(field["nullable"]) != "true" {
				t.Fatalf("input schema invented required fields or erased null defaulting: %s", envelope["fields"])
			}
			for key, want := range map[string]string{
				"json_field": strings.Split(expected.Tag.Get("json"), ",")[0],
				"go_type":    expected.Type.String(),
			} {
				var got string
				if err := json.Unmarshal(field[key], &got); err != nil || got != want {
					t.Fatalf("schema field %d %s = %s (%v), want %q", index, key, field[key], err, want)
				}
			}
			zero, err := json.Marshal(reflect.Zero(expected.Type).Interface())
			if err != nil || !bytes.Equal(field["default"], zero) {
				t.Fatalf("schema field %d changed its zero value: %v", index, err)
			}
		}
		replay, err := run([]byte("not JSON; input schema does not decode stdin"), "--input-schema")
		if err != nil || !bytes.Equal(output, replay) {
			t.Fatalf("input discovery depends on a supplied case: %v", err)
		}
		defaultOutput, err := run(envelope["default_input"], "--declared-input")
		if err != nil {
			t.Fatal(err)
		}
		var defaultEnvelope map[string]json.RawMessage
		if err := decodeStrictJSON(defaultOutput, &defaultEnvelope); err != nil {
			t.Fatal(err)
		}
		var defaultDecision DecisionResult
		if err := decodeStrictJSON(defaultEnvelope["generated_decision"], &defaultDecision); err != nil {
			t.Fatal(err)
		}
		if defaultDecision.Decision != "UNKNOWN" || !reflect.DeepEqual(defaultDecision, EvaluateSourcePolicy(policy, Case{})) {
			t.Fatalf("schema-driven defaults manufactured policy evidence: %+v", defaultDecision)
		}
	})
	bound := generatedJudgeInput{
		ID: "conditional-pass", ProducerAvailable: true, ConsumerAvailable: true,
		ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest,
		ObservedGeneratedJudgeDigest: policy.SemanticDigest, ObservedIndependentDigest: policy.SemanticDigest,
		UpperDecision: "PASS",
	}
	passDocument := encode(bound)
	upper := bound
	upper.UpperDecision = "FIXED_POINT"
	contradiction := bound
	contradiction.ObservedSourceDigest = "sha256:" + strings.Repeat("0", 64)
	if contradiction.ObservedSourceDigest == policy.SourceDigest {
		contradiction.ObservedSourceDigest = "sha256:" + strings.Repeat("1", 64)
	}
	var firstEffective string
	rawDigests := map[string]bool{}
	for index, test := range []struct {
		name, document, binding, decision, condition string
	}{
		{"missing", "{}", "MISSING_DEFAULTED", "UNKNOWN", "EVIDENCE_UNAVAILABLE"},
		{"false", "{\"producer_available\":false}", "DECLARED", "UNKNOWN", "EVIDENCE_UNAVAILABLE"},
		{"null", "{\"producer_available\":null}", "NULL_DEFAULTED", "UNKNOWN", "EVIDENCE_UNAVAILABLE"},
		{"formatted-false", "{\n  \"producer_available\": false\n}", "DECLARED", "UNKNOWN", "EVIDENCE_UNAVAILABLE"},
		{"conditional-pass", string(passDocument), "DECLARED", "PASS", "SEMANTIC_EQUIVALENCE"},
		{"unknown-upper-decision", string(encode(upper)), "DECLARED", "FAIL_CLOSED", "UNRECOGNIZED_TOP_LEVEL_DECISION"},
		{"source-contradiction", string(encode(contradiction)), "DECLARED", "FAIL_CLOSED", "SOURCE_DIGEST_MISMATCH"},
	} {
		t.Run(test.name, func(t *testing.T) {
			document := []byte(test.document)
			output, err := run(document, "--declared-input")
			if err != nil {
				t.Fatal(err)
			}
			var envelope map[string]json.RawMessage
			if err := decodeStrictJSON(output, &envelope); err != nil || len(envelope) != 14 {
				t.Fatalf("declared envelope: %v: %s", err, output)
			}
			var effective generatedJudgeInput
			var reference Case
			var supplied map[string]json.RawMessage
			if err := json.Unmarshal(document, &effective); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(document, &reference); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(document, &supplied); err != nil {
				t.Fatal(err)
			}
			effectiveBytes := encode(effective)
			wantText := map[string]string{
				"schema":                   "gooo/generated-policy-declared-input/v1",
				"source_digest":            policy.SourceDigest,
				"semantic_digest":          policy.SemanticDigest,
				"input_digest":             DigestBytes(document),
				"effective_input_digest":   DigestBytes(effectiveBytes),
				"evaluation_scope":         "CONDITIONAL_GENERATED_POLICY_EVALUATION",
				"external_evidence_state":  "UNKNOWN_NOT_VERIFIED",
				"full_conformance_state":   "UNKNOWN_NOT_EXECUTED",
				"execution_evidence_state": "SELF_REPORTED_NOT_ATTESTED",
			}
			for key, want := range wantText {
				var got string
				if err := json.Unmarshal(envelope[key], &got); err != nil || got != want {
					t.Fatalf("%s = %s (%v), want %q", key, envelope[key], err, want)
				}
			}
			if string(envelope["generated_execution_observed"]) != "true" || string(envelope["mutation_authority"]) != "0" || string(envelope["promotion_authority"]) != "0" {
				t.Fatal("generated execution, attestation, or authority boundaries were conflated")
			}
			var decision DecisionResult
			if err := decodeStrictJSON(envelope["generated_decision"], &decision); err != nil {
				t.Fatal(err)
			}
			if decision.Decision != test.decision || decision.MatchedCondition != test.condition || !reflect.DeepEqual(decision, EvaluateSourcePolicy(policy, reference)) {
				t.Fatalf("generated decision changed the Gooo reduction or UNKNOWN context: %+v", decision)
			}
			var fields []DeclaredCaseField
			if err := decodeStrictJSON(envelope["fields"], &fields); err != nil {
				t.Fatal(err)
			}
			expectedFields, err := bindDeclaredCaseInputFields(reference, supplied)
			if err != nil {
				t.Fatal(err)
			}
			expectedByName := map[string]DeclaredCaseField{}
			for _, field := range expectedFields {
				expectedByName[field.JSONField] = field
			}
			var effectiveFields map[string]json.RawMessage
			if err := json.Unmarshal(effectiveBytes, &effectiveFields); err != nil {
				t.Fatal(err)
			}
			if len(fields) != 8 || len(fields) != len(effectiveFields) {
				t.Fatal("the generated eight-field input was confused with the eleven-field Case")
			}
			seen := map[string]bool{}
			for _, field := range fields {
				if seen[field.JSONField] || effectiveFields[field.JSONField] == nil || !reflect.DeepEqual(field, expectedByName[field.JSONField]) {
					t.Fatalf("generated field was missing, duplicated, or relabeled: %+v", field)
				}
				seen[field.JSONField] = true
				if field.JSONField == "producer_available" && field.Binding != test.binding {
					t.Fatalf("producer binding = %s, want %s", field.Binding, test.binding)
				}
			}
			if index < 4 {
				if index == 0 {
					firstEffective = wantText["effective_input_digest"]
				}
				if wantText["effective_input_digest"] != firstEffective || rawDigests[wantText["input_digest"]] {
					t.Fatal("equal effective inputs erased distinct raw declarations")
				}
				rawDigests[wantText["input_digest"]] = true
			}
			if test.name == "conditional-pass" {
				replay, err := run(document, "--declared-input")
				if err != nil || !bytes.Equal(output, replay) {
					t.Fatalf("declared report replay changed: %v", err)
				}
			}
		})
	}
	for _, document := range []string{
		"null", "[]", "{}{}", "{\"unexpected\":true}", "{\"producer_available\":\"false\"}",
		"{\"producer_available\":false,\"producer_available\":true}",
		"{\"Producer_Available\":true}", "{\"id\":\"one\",\"ID\":\"two\"}", "{\"id\":}",
	} {
		if output, err := run([]byte(document), "--declared-input"); err == nil || len(output) != 0 {
			t.Fatalf("unbindable declared input emitted a report: %s: %v: %s", document, err, output)
		}
	}
	for _, arguments := range [][]string{
		{"--declared-input", "--unknown"},
		{"--input-schema", "--declared-input"},
	} {
		if output, err := run(passDocument, arguments...); err == nil || len(output) != 0 {
			t.Fatalf("unknown or conflicting arguments silently downgraded the mode: %v: %s", err, output)
		}
	}
	output, err := run(passDocument)
	if err != nil {
		t.Fatal(err)
	}
	var legacy DecisionResult
	var reference Case
	if err := decodeStrictJSON(output, &legacy); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(passDocument, &reference); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(legacy, EvaluateSourcePolicy(policy, reference)) {
		t.Fatal("default generated-judge output changed its legacy decision schema")
	}
}

func TestPolicyDecisionProposalBindsCoordinatesAndPreservesSource(t *testing.T) {
	source := append([]byte("\n"), declaredCaseSourceFixture(t)...)
	untouched := append([]byte(nil), source...)
	revision := PolicyDecisionRevision{
		ExpectedSourceDigest: DigestBytes(source),
		Condition:            ConditionSemanticEquivalence,
		FromDecision:         DecisionPass,
		ToDecision:           DecisionFailClosed,
	}
	proposal, err := ProposePolicyDecisionRevision("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", revision)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, untouched) || proposal.Original.SourceDigest != revision.ExpectedSourceDigest || proposal.Candidate.SourceDigest != DigestBytes([]byte(proposal.CandidateSource)) {
		t.Fatal("policy proposal mutated or detached its source identity")
	}
	if !reflect.DeepEqual(proposal.ChangedCoordinates, []string{"transition.to", "case.resolution.decision"}) || proposal.Original.SemanticDigest == proposal.Candidate.SemanticDigest {
		t.Fatal("policy proposal did not bind two decision coordinates and a semantic change")
	}
	expected := proposal.Original
	expected.SourceDigest = proposal.Candidate.SourceDigest
	expected.SemanticDigest = proposal.Candidate.SemanticDigest
	expected.Reduction.Rules = append([]DecisionRule(nil), proposal.Original.Reduction.Rules...)
	matches := 0
	for index, rule := range expected.Reduction.Rules {
		if rule.Condition == revision.Condition {
			matches++
			expected.Reduction.Rules[index].Decision = revision.ToDecision
		}
	}
	if matches != 1 || !reflect.DeepEqual(expected, proposal.Candidate) {
		t.Fatal("policy proposal changed semantic coordinates beyond the requested decision")
	}
	replay, err := ProposePolicyDecisionRevision("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", revision)
	if err != nil || !reflect.DeepEqual(proposal, replay) {
		t.Fatalf("policy proposal replay changed: %v", err)
	}
	reverse, err := ProposePolicyDecisionRevision("policy.gooo", []byte(proposal.CandidateSource), "metapolicycompilation", "metapolicycompilation", PolicyDecisionRevision{
		ExpectedSourceDigest: proposal.Candidate.SourceDigest,
		Condition:            revision.Condition,
		FromDecision:         revision.ToDecision,
		ToDecision:           revision.FromDecision,
	})
	if err != nil || reverse.Candidate.SemanticDigest != proposal.Original.SemanticDigest {
		t.Fatalf("explicit reverse proposal did not recover the original semantic contract: %v", err)
	}
}

func TestPolicyDecisionProposalRejectsUnboundChanges(t *testing.T) {
	source := declaredCaseSourceFixture(t)
	revision := PolicyDecisionRevision{
		ExpectedSourceDigest: DigestBytes(source),
		Condition:            ConditionSemanticEquivalence,
		FromDecision:         DecisionPass,
		ToDecision:           DecisionFailClosed,
	}
	for _, test := range []struct {
		name   string
		change func(*PolicyDecisionRevision)
	}{
		{"missing-source-digest", func(value *PolicyDecisionRevision) { value.ExpectedSourceDigest = "" }},
		{"stale-source-digest", func(value *PolicyDecisionRevision) { value.ExpectedSourceDigest = DigestBytes([]byte("different source")) }},
		{"missing-condition", func(value *PolicyDecisionRevision) { value.Condition = "" }},
		{"unknown-condition", func(value *PolicyDecisionRevision) { value.Condition = "UNDECLARED_CONDITION" }},
		{"stale-from-decision", func(value *PolicyDecisionRevision) { value.FromDecision = DecisionUnknown }},
		{"unknown-from-decision", func(value *PolicyDecisionRevision) { value.FromDecision = "FIXED_POINT" }},
		{"missing-to-decision", func(value *PolicyDecisionRevision) { value.ToDecision = "" }},
		{"unknown-to-decision", func(value *PolicyDecisionRevision) { value.ToDecision = "FIXED_POINT" }},
		{"unchanged-decision", func(value *PolicyDecisionRevision) { value.ToDecision = value.FromDecision }},
		{"missing-unknown-context", func(value *PolicyDecisionRevision) { value.ToDecision = DecisionUnknown }},
		{"would-discard-unknown-context", func(value *PolicyDecisionRevision) {
			value.Condition, value.FromDecision = ConditionEvidenceUnavailable, DecisionUnknown
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			altered := revision
			test.change(&altered)
			proposal, err := ProposePolicyDecisionRevision("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", altered)
			if err == nil || !reflect.DeepEqual(proposal, PolicyDecisionProposal{}) {
				t.Fatalf("unbound revision produced a candidate: %+v (%v)", proposal, err)
			}
		})
	}
	for _, test := range []struct {
		name, namespace string
		source          []byte
	}{
		{"invalid-source", "metapolicycompilation", []byte("not a Gooo declaration")},
		{"wrong-namespace", "different", source},
	} {
		t.Run(test.name, func(t *testing.T) {
			altered := revision
			altered.ExpectedSourceDigest = DigestBytes(test.source)
			proposal, err := ProposePolicyDecisionRevision("policy.gooo", test.source, "metapolicycompilation", test.namespace, altered)
			if err == nil || !reflect.DeepEqual(proposal, PolicyDecisionProposal{}) {
				t.Fatal("invalid source or identity produced a candidate")
			}
		})
	}
}

func TestPolicyDecisionProposalChangesGeneratedBehavior(t *testing.T) {
	source := declaredCaseSourceFixture(t)
	proposal, err := ProposePolicyDecisionRevision("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", PolicyDecisionRevision{
		ExpectedSourceDigest: DigestBytes(source),
		Condition:            ConditionSemanticEquivalence,
		FromDecision:         DecisionPass,
		ToDecision:           DecisionFailClosed,
	})
	if err != nil {
		t.Fatal(err)
	}
	generated := GenerateJudge(proposal.Candidate)
	work := t.TempDir()
	judgePath, binary := filepath.Join(work, "judge.go"), filepath.Join(work, "proposal-judge")
	if err := os.WriteFile(judgePath, generated, 0o600); err != nil {
		t.Fatal(err)
	}
	// Build one candidate binary; the proposal API itself does not execute it.
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, judgePath)
	build.Dir = work
	build.Env = append(os.Environ(), "GO111MODULE=off", "GOTOOLCHAIN=go1.27.0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build proposed generated judge: %v: %s", err, output)
	}
	input := generatedJudgeInput{
		ID: "source-revision", ProducerAvailable: true, ConsumerAvailable: true,
		ObservedSourceDigest: proposal.Candidate.SourceDigest, ObservedArtifactSourceDigest: proposal.Candidate.SourceDigest,
		ObservedGeneratedJudgeDigest: DigestBytes(generated), ObservedIndependentDigest: DigestBytes(generated),
		UpperDecision: "PASS",
	}
	document, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(t.Context(), binary, "--declared-input")
	command.Dir = work
	command.Stdin = bytes.NewReader(document)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute proposed generated judge: %v: %s", err, output)
	}
	var envelope map[string]json.RawMessage
	if err := decodeStrictJSON(output, &envelope); err != nil {
		t.Fatal(err)
	}
	var decision DecisionResult
	if err := decodeStrictJSON(envelope["generated_decision"], &decision); err != nil {
		t.Fatal(err)
	}
	var reference Case
	if err := json.Unmarshal(document, &reference); err != nil {
		t.Fatal(err)
	}
	if decision.Decision != DecisionFailClosed || decision.MatchedCondition != ConditionSemanticEquivalence || !reflect.DeepEqual(decision, EvaluateSourcePolicy(proposal.Candidate, reference)) {
		t.Fatalf("generated behavior did not follow the proposed Gooo decision: %+v", decision)
	}
	reference.ObservedSourceDigest, reference.ObservedArtifactSourceDigest = proposal.Original.SourceDigest, proposal.Original.SourceDigest
	originalJudgeDigest := DigestBytes(GenerateJudge(proposal.Original))
	reference.ObservedGeneratedJudgeDigest, reference.ObservedIndependentDigest = originalJudgeDigest, originalJudgeDigest
	if before := EvaluateSourcePolicy(proposal.Original, reference); before.Decision != DecisionPass || before.MatchedCondition != ConditionSemanticEquivalence {
		t.Fatalf("the original source policy did not retain its independent baseline decision: %+v", before)
	}
	if string(envelope["mutation_authority"]) != "0" || string(envelope["promotion_authority"]) != "0" || string(envelope["external_evidence_state"]) != "\"UNKNOWN_NOT_VERIFIED\"" {
		t.Fatal("candidate execution was promoted into external evidence or mutation authority")
	}
}


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
	proposal, err := ProposePolicyDecisionRevision("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", PolicyDecisionRevision{
		ExpectedSourceDigest: DigestBytes(source),
		Condition:            ConditionSemanticEquivalence,
		FromDecision:         DecisionPass,
		ToDecision:           DecisionFailClosed,
	})
	if err != nil {
		return nil, err
	}
	return []byte(proposal.CandidateSource), nil
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
