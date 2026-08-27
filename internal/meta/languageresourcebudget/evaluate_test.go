package languageresourcebudget

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestEvaluateSeparatesSemanticMeaningFromRunnerBudget(t *testing.T) {
	input := fixtureInput()
	normal := Evaluate(input, "normal")
	if normal.Decision != "PASS" || normal.Resolution != "EXACT" || normal.ResourceResolution != "RUNNER_SCOPED" ||
		normal.Semantic.Decision != "PASS" || normal.Summary.Coordinates != (Counter{22, 22, 10000}) ||
		normal.Transitions[1].To != "DISCHARGED" {
		t.Fatalf("normal=%#v", normal)
	}
	overbudget := input
	overbudget.Observations = append([]Observation(nil), input.Observations...)
	overbudget.Observations[0].WallTimeNS = 2000000001
	limited := Evaluate(overbudget, "over-budget")
	if limited.Decision != "FAIL_CLOSED" || limited.Resolution != "EXACT" || limited.Reason != "RESOURCE_BUDGET_EXCEEDED" ||
		limited.Semantic.Decision != "PASS" || limited.Semantic.Resolution != "EXACT" || limited.Interpretation != "SEMANTIC_EXACT_RESOURCE_CLAIM_REFUTED" ||
		limited.Transitions[1].To != "REFUTED" {
		t.Fatalf("overbudget=%#v", limited)
	}
}

func TestEvaluateLowersOnlyResourceResolutionForMissingSample(t *testing.T) {
	input := fixtureInput()
	input.Observations = input.Observations[:len(input.Observations)-1]
	report := Evaluate(input, "missing-sample")
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "LOWER_RESOLUTION" || report.Reason != "RESOURCE_SAMPLE_MISSING" ||
		report.Semantic.Decision != "PASS" || report.Semantic.Resolution != "EXACT" || report.ResourceResolution != "LOWER_RESOLUTION" ||
		report.Transitions[1].To != "OPEN" || report.Transitions[1].Reason != "RESOURCE_SAMPLE_MISSING" {
		t.Fatalf("missing=%#v", report)
	}
}

func TestEvaluateAndValidateReportAreDeterministic(t *testing.T) {
	input := fixtureInput()
	first, second := Evaluate(input, "normal"), Evaluate(input, "normal")
	if !reflect.DeepEqual(first, second) {
		t.Fatal("reducer changed its report between replays")
	}
	if err := ValidateReport(input, first); err != nil {
		t.Fatal(err)
	}
}

func fixtureInput() Input {
	contract := CanonicalContract()
	source := sourceReceiptFixture(contract)
	artifact := artifactFixture()
	observations := make([]Observation, 0, 9)
	for _, spec := range contract.Operations {
		for sequence := 1; sequence <= contract.SamplesPerOp; sequence++ {
			receiptBytes, generatedBytes := int64(300), int64(0)
			if spec.Output == "GENERATED" {
				receiptBytes, generatedBytes = 0, 600
			}
			outputDigest := "sha256:" + strings.Repeat("c", 64)
			if sequence == 1 {
				if spec.ID == "source-check" {
					outputDigest = digestBytes(sourceReceiptFixture(contract))
				}
				if spec.ID == "project-manifest" || spec.ID == "replay-manifest" {
					outputDigest = digestBytes(artifactFixture())
				}
			}
			observations = append(observations, Observation{Schema: ObservationSchema, SubjectSHA: strings.Repeat("a", 40), Producer: Producer, Consumer: Consumer,
				Operation: spec.ID, Stage: spec.Stage, Step: spec.Step, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
				Reason: "RUNNER_RESOURCE_OBSERVED", Sequence: sequence, ExitCode: 0, WallTimeNS: 10000000 + int64(sequence), PeakRSSKiB: 1000,
				ReceiptBytes: receiptBytes, GeneratedBytes: generatedBytes, OutputDigest: outputDigest})
		}
	}
	return Input{Schema: InputSchema, ExpectedHead: strings.Repeat("a", 40), Contract: contract, Producer: ProducerEvidence{SourceReceiptBase64: base64.StdEncoding.EncodeToString(source), ArtifactBase64: base64.StdEncoding.EncodeToString(artifact), ReplayBase64: base64.StdEncoding.EncodeToString(artifact), SourceDigest: "sha256:" + strings.Repeat("d", 64), SourceFiles: 2, GoFiles: 0, Runner: Runner{OS: "Linux", Architecture: "x86_64", Image: "ubuntu-latest", ImageVersion: "20260827.1", GoVersion: "go1.27.0"}, Effects: Effects{}}, Observations: observations}
}

func sourceReceiptFixture(contract Contract) []byte {
	value, _ := json.Marshal(map[string]any{"schema_version": "gooo/diagnostics/v1", "command": "check", "status": "ok", "file": contract.SourcePaths[0], "diagnostics": []any{}})
	return value
}

func artifactFixture() []byte {
	value, _ := json.Marshal(map[string]any{"schema": "gooo/operation-manifest/v1", "decision": "PASS", "resolution": "EXACT", "reason": "OPERATION_MANIFEST_EMITTED", "kind": "operation-manifest", "subject_digest": "sha256:" + strings.Repeat("e", 64), "operation": map[string]string{"activity": "PayOrder"}, "effects": Effects{}, "digest": "sha256:" + strings.Repeat("f", 64)})
	return value
}
