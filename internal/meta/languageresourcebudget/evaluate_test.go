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
	if normal.Decision != "PASS" || normal.Resolution != "EXACT" || normal.ResourceResolution != "RUNNER_SCOPED" || normal.Semantic.Decision != "PASS" || normal.Summary.Coordinates != (Counter{Satisfied: 19, Total: 19, BasisPoints: 10000}) || normal.Transitions[1].To != "DISCHARGED" {
		t.Fatalf("normal=%#v", normal)
	}
	overbudget := input
	overbudget.Observations = append([]Observation(nil), input.Observations...)
	overbudget.Observations[0].WallTimeNS = 2000000001
	limited := Evaluate(overbudget, "over-budget")
	if limited.Decision != "FAIL_CLOSED" || limited.Resolution != "EXACT" || limited.Reason != "RESOURCE_BUDGET_EXCEEDED" || limited.Semantic.Decision != "PASS" || limited.Semantic.Resolution != "EXACT" || limited.Interpretation != "SEMANTIC_EXACT_RESOURCE_CLAIM_REFUTED" || limited.Transitions[1].To != "REFUTED" || limited.Transitions[2].To != "DISCHARGED" {
		t.Fatalf("overbudget=%#v", limited)
	}
}

func TestEffectsBoundaryRequiresPerOperationSnapshots(t *testing.T) {
	input := fixtureInput()
	input.Producer.WriteSets[0].RepositoryWrites = 1
	violated := Evaluate(input, "write-set-violation")
	if violated.Decision != "FAIL_CLOSED" || violated.Resolution != "EXACT" || violated.ReadOnlyResolution != "EXACT" || violated.Transitions[2].To != "REFUTED" || violated.Transitions[2].Reason != "EFFECT_BOUNDARY_VIOLATED" {
		t.Fatalf("violated=%#v", violated)
	}
	input = fixtureInput()
	input.Producer.WriteSets[0].MutationAuthority = true
	authorized := Evaluate(input, "authority-violation")
	if authorized.Transitions[2].To != "REFUTED" || authorized.ReadOnlyResolution != "EXACT" {
		t.Fatalf("authorized=%#v", authorized)
	}

	input = fixtureInput()
	input.Producer.WriteSets = nil
	unknown := Evaluate(input, "write-set-missing")
	if unknown.Decision != "FAIL_CLOSED" || unknown.Resolution != "LOWER_RESOLUTION" || unknown.ReadOnlyResolution != "LOWER_RESOLUTION" || unknown.Transitions[2].To != "OPEN" || unknown.Transitions[2].Reason != "EFFECT_OBSERVATION_MISSING" {
		t.Fatalf("unknown=%#v", unknown)
	}
}

func TestEvaluateLocalizesMissingResourceCoordinate(t *testing.T) {
	input := fixtureInput()
	input.Observations = append([]Observation(nil), input.Observations[:6]...)
	report := Evaluate(input, "missing-sample")
	if report.Transitions[1].To != "OPEN" || report.Transitions[1].Reason != "RESOURCE_SAMPLE_MISSING" || report.Summary.Resources[0].MissingSamples != 0 || report.Summary.Resources[1].MissingSamples != 0 || report.Summary.Resources[2].MissingSamples != 1 {
		t.Fatalf("missing=%#v", report)
	}
	for _, indicator := range report.Indicators {
		if strings.HasPrefix(indicator.ID, "resource.source-check") || strings.HasPrefix(indicator.ID, "resource.project-manifest") {
			if indicator.Status != "SATISFIED" && indicator.Status != "NOT_APPLICABLE" {
				t.Fatalf("unrelated indicator=%#v", indicator)
			}
		}
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
	activitySource := []byte("package resourcebudget\nnamespace resourcebudget\n\nactivity PayOrder(Order) -> Receipt\n")
	entitySource := []byte("package resourcebudget\nnamespace resourcebudget\n\nentity Order id \"gooo://resource-budget/order\"\nentity Receipt id \"gooo://resource-budget/receipt\"\n")
	sources := []RawSource{{Filename: contract.SourcePaths[0], ContentBase64: base64.StdEncoding.EncodeToString(activitySource)}, {Filename: contract.SourcePaths[1], ContentBase64: base64.StdEncoding.EncodeToString(entitySource)}}
	producer := ProducerEvidence{SourceFiles: sources, SourceFileCount: 2}
	meaning, err := reconstructSourceMeaning(Input{Contract: contract, Producer: producer})
	if err != nil {
		panic(err)
	}
	source := sourceReceiptFixture(contract)
	artifact := artifactFixture()
	observations := make([]Observation, 0, 9)
	for _, spec := range contract.Operations {
		payload := source
		kind := "RECEIPT"
		if spec.ID != "source-check" {
			payload, kind = artifact, "GENERATED"
		}
		for sequence := 1; sequence <= contract.SamplesPerOp; sequence++ {
			receiptBytes, generatedBytes := int64(300), int64(0)
			if kind == "GENERATED" {
				receiptBytes, generatedBytes = 0, 600
			}
			observations = append(observations, Observation{Schema: ObservationSchema, SubjectSHA: strings.Repeat("a", 40), Producer: Producer, Consumer: Consumer, Operation: spec.ID, Stage: spec.Stage, Step: spec.Step, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice, Reason: "RUNNER_RESOURCE_OBSERVED", Sequence: sequence, ExitCode: 0, WallTimeNS: 10000000 + int64(sequence), PeakRSSKiB: 1000, ReceiptBytes: receiptBytes, GeneratedBytes: generatedBytes, OutputDigest: digestBytes(payload), SourceRawDigest: meaning.SourceDigest, SourceSemanticDigest: meaning.SemanticDigest, EntryDigest: meaning.TargetDigest, TargetDigest: meaning.TargetDigest})
		}
		producer.RawOutputs = append(producer.RawOutputs, RawOutput{Operation: spec.ID, Sequence: 1, Kind: kind, PayloadBase64: base64.StdEncoding.EncodeToString(payload)})
	}
	producer.SourceReceiptBase64 = base64.StdEncoding.EncodeToString(source)
	producer.ArtifactBase64 = base64.StdEncoding.EncodeToString(artifact)
	producer.ReplayBase64 = base64.StdEncoding.EncodeToString(artifact)
	producer.SourceDigest = meaning.SourceDigest
	producer.Runner = Runner{OS: "Linux", Architecture: "x86_64", Image: "ubuntu-latest", ImageVersion: "20260827.1", GoVersion: "go1.27.0"}
	producer.WriteSets = make([]WriteSetObservation, 0, len(contract.Operations))
	for _, spec := range contract.Operations {
		producer.WriteSets = append(producer.WriteSets, WriteSetObservation{Operation: spec.ID, Schema: "gooo/meta-resource-budget-write-set/v1", Producer: Producer, Consumer: Consumer, BeforeTreeDigest: strings.Repeat("a", 40), AfterTreeDigest: strings.Repeat("a", 40), BeforeStatusObserved: true, AfterStatusObserved: true, WriteSetDigest: statusDigest(nil, nil), ChangedPaths: []string{}, Reason: "NET_REPOSITORY_STATE_UNCHANGED_ACROSS_OPERATION_WINDOW", AuthorityObserved: true, SampleStart: 1, SampleEnd: 3})
	}
	return Input{Schema: InputSchema, ExpectedHead: strings.Repeat("a", 40), EvidenceClass: "CURRENT_EVIDENCE", ContractDigest: digestValue(contract), Contract: contract, Producer: producer, Observations: observations}
}

func sourceReceiptFixture(contract Contract) []byte {
	value, _ := json.Marshal(map[string]any{"schema_version": "gooo/diagnostics/v1", "command": "check", "status": "ok", "file": contract.SourcePaths[0], "diagnostics": []any{}})
	return value
}

func artifactFixture() []byte {
	value, _ := json.Marshal(map[string]any{"schema": "gooo/operation-manifest/v1", "decision": "PASS", "resolution": "EXACT", "reason": "OPERATION_MANIFEST_EMITTED", "kind": "operation-manifest", "subject_digest": "sha256:" + strings.Repeat("e", 64), "package": map[string]string{"name": "resourcebudget", "namespace": "resourcebudget"}, "operation": map[string]any{"activity": "PayOrder", "inputs": []map[string]string{{"name": "Order", "id": "gooo://resource-budget/order"}}, "output": map[string]string{"name": "Receipt", "id": "gooo://resource-budget/receipt"}}, "effects": Effects{}, "digest": "sha256:" + strings.Repeat("f", 64)})
	return value
}
