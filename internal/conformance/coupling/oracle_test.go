package coupling

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestCorpusExpectations(t *testing.T) {
	for _, row := range testCorpus() {
		got := Evaluate(row.Input)
		if got.Decision != row.Expected.Decision || got.Reason != row.Expected.Reason ||
			!reflect.DeepEqual(got.AcceptedSurfaces, row.Expected.AcceptedSurfaces) ||
			!reflect.DeepEqual(got.ChangedSurfaces, row.Expected.ChangedSurfaces) ||
			!reflect.DeepEqual(got.ReceiptSurfaces, row.Expected.ReceiptSurfaces) ||
			got.ObservationCounts != row.Expected.ObservationCounts || got.Resources != row.Expected.Resources {
			t.Errorf("%s got decision=%s/%s accepted=%v/%v changed=%v/%v receipts=%v/%v counts=%+v/%+v resources=%+v/%+v", row.Name, got.Decision, got.Reason, got.AcceptedSurfaces, row.Expected.AcceptedSurfaces, got.ChangedSurfaces, row.Expected.ChangedSurfaces, got.ReceiptSurfaces, row.Expected.ReceiptSurfaces, got.ObservationCounts, row.Expected.ObservationCounts, got.Resources, row.Expected.Resources)
		}
	}
}

func TestBaselineIsFullSuiteAndFair(t *testing.T) {
	for _, row := range testCorpus()[:4] {
		comparison := Compare(row.Input)
		if !comparison.OutcomeMatch || !comparison.ReasonMatch || !comparison.LocalizationMatch || comparison.Finding != "NO_UNIQUE_BENEFIT" {
			t.Fatalf("%s comparison=%+v", row.Name, comparison)
		}
		if !comparison.Baseline.FullSuite || comparison.Baseline.Resources != comparison.Oracle.Resources || comparison.Baseline.WorkUnits != comparison.Oracle.Resources.WorkUnits {
			t.Fatalf("%s baseline resource binding=%+v oracle=%+v", row.Name, comparison.Baseline, comparison.Oracle.Resources)
		}
	}
}

func TestStrictJSONAndPermutationInvariance(t *testing.T) {
	input := testCorpus()[0].Input
	data, err := EncodeInputJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeInput(data)
	if err != nil {
		t.Fatal(err)
	}
	if CanonicalInputDigest(input) != CanonicalInputDigest(decoded) || Evaluate(input).CanonicalOutputDigest != Evaluate(decoded).CanonicalOutputDigest {
		t.Fatal("JSON round trip changed canonical result")
	}
	if _, err := DecodeInput(append(data, []byte("\n{}")...)); err == nil {
		t.Fatal("trailing JSON value accepted")
	}
	duplicate := strings.Replace(string(data), `"schema": "gooo/coupling/v1"`, `"schema": "gooo/coupling/v1","schema": "gooo/coupling/v1"`, 1)
	if _, err := DecodeInput([]byte(duplicate)); err == nil {
		t.Fatal("duplicate JSON key accepted")
	}
	unknown := strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "}") + `,"unexpected":true}`
	if _, err := DecodeInput([]byte(unknown)); err == nil {
		t.Fatal("unknown JSON field accepted")
	}

	permuted := cloneInput(input)
	reverseBindings(permuted.Registry)
	reverseChanges(permuted.Changes)
	reverseReceipts(permuted.Receipts)
	reverseResources(permuted.ResourceReceipts)
	reverseStrings(permuted.Roots)
	reverseEdges(permuted.Path.Edges)
	reverseClaims(permuted.Path.Claims)
	reverseEvidence(permuted.Path.Evidence)
	if CanonicalInputDigest(input) != CanonicalInputDigest(permuted) || !reflect.DeepEqual(Evaluate(input), Evaluate(permuted)) {
		t.Fatal("presentation ordering changed canonical result")
	}
}

func TestPresentationAndExpectedMutationIsolation(t *testing.T) {
	row := testCorpus()[0]
	base := Evaluate(row.Input)
	mutations := []func(*Input){
		func(input *Input) { input.FixtureID = "fixture-label/changed" },
		func(input *Input) {
			input.SemanticBefore.Nodes[0].Name = "Renamed"
			input.SemanticBefore.Nodes[0].Aliases = []string{"AliasChanged"}
		},
		func(input *Input) {
			input.SemanticAfter.Nodes[0].Name = "Renamed"
			input.SemanticAfter.Nodes[0].Aliases = []string{"AliasChanged"}
		},
		func(input *Input) {
			input.Registry[0].PackageLabel = "other"
			input.Registry[0].FileLabel = "other.go"
			input.Registry[0].SourceSpan = "99:1-99:2"
		},
	}
	for index, mutate := range mutations {
		input := cloneInput(row.Input)
		mutate(&input)
		got := Evaluate(input)
		if got.InputDigest != base.InputDigest || got.CanonicalOutputDigest != base.CanonicalOutputDigest || got.ReplayDigest != base.ReplayDigest || got.Decision != base.Decision || got.Reason != base.Reason {
			t.Fatalf("presentation mutation %d affected authority result: base=%+v got=%+v", index, base, got)
		}
	}
	mutatedCase := row
	mutatedCase.Name = "case-label-changed"
	mutatedCase.Expected = FixtureExpectation{
		Decision: DecisionFailClosed, Reason: ReasonDigestMismatch,
		AcceptedSurfaces: []string{"expected-only"}, ChangedSurfaces: []string{"expected-only"}, ReceiptSurfaces: []string{"expected-only"},
		ObservationCounts: ObservationCounts{RegistryBindings: 99, ResourceReceipts: 99}, Resources: ResourceObservation{CPUCoreNS: 99, PeakMemoryBytes: 98, WorkUnits: 97},
	}
	if got := Evaluate(mutatedCase.Input); !reflect.DeepEqual(got, base) {
		t.Fatalf("case name or expected-only mutation affected actual result: base=%+v got=%+v", base, got)
	}
	if got := CanonicalInputDigest(row.Input); got == CanonicalInputDigest(func() Input {
		input := cloneInput(row.Input)
		input.SemanticBefore.Nodes[0].ID = "urn:gooo:entity:renamed"
		return input
	}()) {
		t.Fatal("stable identity mutation did not affect input digest")
	}
	authority := cloneInput(row.Input)
	authority.AuthoritySourceAfter += "\n"
	if CanonicalInputDigest(authority) == base.InputDigest {
		t.Fatal("authoritative source mutation did not affect input digest")
	}
}

func TestAdversarialPathClosureAndEvidencePartitions(t *testing.T) {
	base := testCorpus()[0].Input
	cases := []struct {
		name   string
		mutate func(*Input)
	}{
		{name: "disconnected", mutate: func(input *Input) { input.Path.Edges[1].ObjectID = semantic.MustIdentity("urn:gooo:term:disconnected") }},
		{name: "fork", mutate: func(input *Input) {
			edge := input.Path.Edges[1]
			edge.RecordID = semantic.MustIdentity("urn:gooo:path:fork")
			input.Path.Edges = append(input.Path.Edges, edge)
		}},
		{name: "cycle", mutate: func(input *Input) { input.Path.Edges[1].ObjectID = semantic.MustIdentity("urn:gooo:source:billing") }},
		{name: "wrong-endpoint", mutate: func(input *Input) {
			input.Path.Edges[len(input.Path.Edges)-1].ObjectID = semantic.MustIdentity("urn:gooo:receipt:wrong")
		}},
		{name: "receipt-evidence-omission", mutate: func(input *Input) { input.Receipts[0].EvidenceRefs = nil }},
		{name: "extra-unrelated-evidence", mutate: func(input *Input) {
			extra := input.Path.Evidence[0]
			extra.ID = semantic.MustIdentity("urn:gooo:evidence:unrelated")
			extra.Digest = digestText("evidence:unrelated")
			input.Path.Evidence = append(input.Path.Evidence, extra)
		}},
		{name: "duplicate-evidence", mutate: func(input *Input) { input.Path.Evidence = append(input.Path.Evidence, input.Path.Evidence[0]) }},
		{name: "candidate-becomes-authority", mutate: func(input *Input) {
			candidate := makeCouplingInput(false, true)
			*input = candidate
			input.Path.Edges[2].After.Semantic = digestText("different")
		}},
	}
	for _, test := range cases {
		input := cloneInput(base)
		test.mutate(&input)
		if got := Evaluate(input); got.Decision == DecisionPass {
			t.Errorf("%s mutation falsely passed: %+v", test.name, got)
		}
	}
	permuted := cloneInput(base)
	reverseEdges(permuted.Path.Edges)
	reverseClaims(permuted.Path.Claims)
	reverseEvidence(permuted.Path.Evidence)
	if got := Evaluate(permuted); got.Decision != DecisionPass || got.CanonicalOutputDigest != Evaluate(base).CanonicalOutputDigest {
		t.Fatal("reordered path IDs changed the result")
	}
}

func TestResourceBindingsAndMeasuredZero(t *testing.T) {
	base := testCorpus()[0].Input
	arbitrary := cloneInput(base)
	arbitrary.ResourceReceipts[0].ProviderDigest = digestText("arbitrary-provider")
	if got := Evaluate(arbitrary); got.Decision != DecisionUnknown || got.Reason != ReasonResourceUnbound {
		t.Fatalf("arbitrary provider was accepted: %+v", got)
	}
	missing := cloneInput(base)
	missing.ResourceReceipts[0].Present = false
	if got := Evaluate(missing); got.Decision != DecisionUnknown || got.Reason != ReasonResourceUnbound {
		t.Fatalf("missing resource presence was accepted: %+v", got)
	}
	zero := cloneInput(base)
	zero.ResourceReceipts[0].Value = 0
	zero.ResourceReceipts[0].BindingDigest = resourceBindingDigest(zero.ResourceReceipts[0])
	if got := Evaluate(zero); got.Decision != DecisionPass || got.Resources.CPUCoreNS != 0 {
		t.Fatalf("explicit measured zero rejected: %+v", got)
	}
}

func TestNoWriteDoesNotMutateInput(t *testing.T) {
	input := testCorpus()[3].Input
	before, err := CanonicalInputBytes(input)
	if err != nil {
		t.Fatal(err)
	}
	_ = Evaluate(input)
	after, err := CanonicalInputBytes(input)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("oracle wrote to its input")
	}
}

func reverseBindings(values []CodeBinding) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseChanges(values []CodeChange) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseReceipts(values []CouplingReceipt) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseResources(values []ExternalResourceReceipt) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseStrings(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseEdges(values []semantic.InferenceEdge) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseClaims(values []semantic.SemanticChangeClaim) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseEvidence(values []semantic.InferenceEvidence) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func TestCorpusJSONDoesNotContainExpectedLabels(t *testing.T) {
	row := testCorpus()[0]
	data, err := EncodeInputJSON(row.Input)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), row.Name) || strings.Contains(string(data), string(row.Expected.Reason)) {
		// Fixture metadata is not encoded in Input JSON.
		t.Fatal("fixture expectation escaped into input JSON")
	}
}
