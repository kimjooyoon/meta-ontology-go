package resourcevector

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestR4F01Vectors(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if corpus.Schema != CorpusSchemaV1 || len(corpus.Cases) != 1 {
		t.Fatalf("corpus = %q/%d", corpus.Schema, len(corpus.Cases))
	}
	if corpus.CanonicalDigest != CorpusDigest(corpus) {
		t.Fatalf("corpus digest = %q, want %q", corpus.CanonicalDigest, CorpusDigest(corpus))
	}
	row := corpus.Cases[0]
	got := Evaluate(row.Input)
	if got.Decision != DecisionPass || got.Reason != ReasonNone || !got.ProofValid {
		t.Fatalf("r4-f-01 output = %#v", got)
	}
	wantSelected := Vector{CPUCoreNS: 33, MemoryBytes: 224, PeakMemoryBytes: 128, WorkUnits: 18, AffectedStableIDs: 2, ApplicablePressures: 2, IndependentGroups: 2, UniquePROVRecords: 12, FinitePROVPaths: 3, ClosureNumerator: 3, ClosureDenominator: 3}
	wantFull := Vector{CPUCoreNS: 36, MemoryBytes: 272, PeakMemoryBytes: 128, WorkUnits: 19, AffectedStableIDs: 3, ApplicablePressures: 3, IndependentGroups: 3, UniquePROVRecords: 15, FinitePROVPaths: 4, ClosureNumerator: 4, ClosureDenominator: 4}
	if got.Selected == nil || got.Full == nil || *got.Selected != wantSelected || *got.Full != wantFull {
		t.Fatalf("vectors selected=%#v full=%#v", got.Selected, got.Full)
	}
	if row.Expected.Selected == nil || row.Expected.Full == nil || *row.Expected.Selected != wantSelected || *row.Expected.Full != wantFull {
		t.Fatalf("fixture expected vectors selected=%#v full=%#v", row.Expected.Selected, row.Expected.Full)
	}
	if got.InputDigest != row.Expected.InputDigest || got.CanonicalOutputDigest != row.Expected.CanonicalOutputDigest || got.ReplayDigest != row.Expected.ReplayDigest {
		t.Fatalf("fixture digests got=%#v expected=%#v", got, row.Expected)
	}
	t.Logf("input=%s output=%s replay=%s corpus=%s", got.InputDigest, got.CanonicalOutputDigest, got.ReplayDigest, CorpusDigest(corpus))
}

func TestStrictJSONAndExpectedIsolation(t *testing.T) {
	input := R4F01()
	data, err := EncodeInputJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	roundTrip, err := DecodeInput(data)
	if err != nil || CanonicalInputDigest(roundTrip) != CanonicalInputDigest(input) {
		t.Fatalf("strict input round trip changed digest: err=%v", err)
	}
	if _, err := DecodeInput(append(data, []byte(` {"schema":"extra"}`)...)); err == nil {
		t.Fatal("trailing JSON value accepted")
	}
	duplicate := strings.Replace(string(data), `"schema": "gooo/selective-ci-resource-vector/v1"`, `"schema": "gooo/selective-ci-resource-vector/v1","schema": "gooo/selective-ci-resource-vector/v1"`, 1)
	if _, err := DecodeInput([]byte(duplicate)); err == nil {
		t.Fatal("duplicate JSON key accepted")
	}
	unknown := strings.TrimSuffix(strings.TrimSpace(string(data)), "}") + `,"unknown":true}`
	if _, err := DecodeInput([]byte(unknown)); err == nil {
		t.Fatal("unknown JSON field accepted")
	}
	left := Evaluate(input)
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	corpus.Cases[0].Expected.Decision = DecisionFailClosed
	right := Evaluate(corpus.Cases[0].Input)
	if left.InputDigest != right.InputDigest || !reflect.DeepEqual(left.Selected, right.Selected) || left.Decision != right.Decision {
		t.Fatal("expected-only mutation changed replay output")
	}
	outputJSON, err := EncodeOutputJSON(left)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(outputJSON), "promotion_authorized") || left.PromotionAuthorized() {
		t.Fatal("promotion authorization escaped the oracle")
	}
}

func TestPermutationAndRootRelocation(t *testing.T) {
	input := R4F01()
	permuted := mustClone(t, input)
	for left, right := 0, len(permuted.Commands)-1; left < right; left, right = left+1, right-1 {
		permuted.Commands[left], permuted.Commands[right] = permuted.Commands[right], permuted.Commands[left]
	}
	for left, right := 0, len(permuted.Paths)-1; left < right; left, right = left+1, right-1 {
		permuted.Paths[left], permuted.Paths[right] = permuted.Paths[right], permuted.Paths[left]
	}
	reverse(permuted.SelectedCommandIDs)
	reverse(permuted.FullCommandIDs)
	for index := range permuted.Commands {
		reversePressures(permuted.Commands[index].Pressures)
	}
	for index := range permuted.Paths {
		reverse(permuted.Paths[index].RecordIDs)
	}
	permuted.Root = "/relocated/r4-f-01"
	for index := range permuted.Commands {
		permuted.Commands[index].Path = relocate(permuted.Commands[index].Path, "/workspace/r4-f-01", permuted.Root)
	}
	for index := range permuted.Paths {
		permuted.Paths[index].Path = relocate(permuted.Paths[index].Path, "/workspace/r4-f-01", permuted.Root)
	}
	left, right := Evaluate(input), Evaluate(permuted)
	if left.InputDigest != right.InputDigest || left.ReplayDigest != right.ReplayDigest || !reflect.DeepEqual(left.Selected, right.Selected) || !reflect.DeepEqual(left.Full, right.Full) {
		t.Fatalf("permutation/root relocation changed replay left=%#v right=%#v", left, right)
	}
}

func TestComponentwiseCeilingsDoNotCompensate(t *testing.T) {
	input := R4F01()
	input.Ceilings.Selected.MemoryBytes = U64(223)
	input.Ceilings.Selected.CPUCoreNS = U64(1_000_000)
	got := Evaluate(input)
	if got.Decision != DecisionUnknown || got.Reason != ReasonCeilingExceeded || !contains(got.LimitFailures, "selected:memory_bytes") {
		t.Fatalf("memory ceiling result = %#v", got)
	}
	if got.Selected == nil || got.Selected.MemoryBytes != 224 {
		t.Fatalf("selected vector lost computed value = %#v", got.Selected)
	}
}

func TestMissingFieldsAndDuplicateRecordsAreNotZeroed(t *testing.T) {
	missing := R4F01()
	missing.Commands[0].CPUCoreNS = nil
	missing.Ceilings.Selected.CPUCoreNS = nil
	got := Evaluate(missing)
	if got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput || got.Selected != nil || got.Full != nil {
		t.Fatalf("missing resource/ceiling result = %#v", got)
	}
	missingPROV := R4F01()
	missingPROV.Paths[0].RecordIDs = nil
	prov := Evaluate(missingPROV)
	if prov.Decision != DecisionUnknown || prov.Reason != ReasonMissingPROV || prov.Selected != nil || prov.Full != nil {
		t.Fatalf("missing PROV result = %#v", prov)
	}
	baseline := EvaluatePlainDAGRetry(missingPROV)
	if baseline.Decision != DecisionUnknown || baseline.Reason != ReasonMissingPROV || baseline.Vector == nil || baseline.Vector.UniquePROVRecords != nil || baseline.Vector.FinitePROVPaths != nil {
		t.Fatalf("baseline missing PROV result = %#v", baseline)
	}
	duplicateID := R4F01()
	duplicateID.Commands = append(duplicateID.Commands, duplicateID.Commands[0])
	if got := Evaluate(duplicateID); got.Decision != DecisionFailClosed || got.Reason != ReasonDuplicateID {
		t.Fatalf("duplicate command ID = %#v", got)
	}
	duplicateRecord := R4F01()
	duplicateRecord.Paths[1].RecordIDs[0] = duplicateRecord.Paths[0].RecordIDs[0]
	if got := Evaluate(duplicateRecord); got.Decision != DecisionFailClosed || got.Reason != ReasonDuplicateRecord {
		t.Fatalf("duplicate PROV record = %#v", got)
	}
	duplicatePressure := R4F01()
	duplicatePressure.Commands[1].Pressures = append(duplicatePressure.Commands[1].Pressures, duplicatePressure.Commands[0].Pressures[0])
	if got := Evaluate(duplicatePressure); got.Decision != DecisionFailClosed || got.Reason != ReasonDuplicateID {
		t.Fatalf("duplicate pressure ID = %#v", got)
	}
	duplicatePath := R4F01()
	duplicatePath.Paths[1].ID = duplicatePath.Paths[0].ID
	if got := Evaluate(duplicatePath); got.Decision != DecisionFailClosed || got.Reason != ReasonDuplicatePath {
		t.Fatalf("duplicate path = %#v", got)
	}
}

func TestFairBaselinesAndFiniteProof(t *testing.T) {
	input := R4F01()
	comparison := Compare(input)
	if comparison.Finding != NoUniqueBenefit || comparison.TypedConfigFullSuite.Decision != DecisionPass || comparison.PlainDAGRetry.Decision != DecisionPass {
		t.Fatalf("comparison = %#v", comparison)
	}
	nonFinite := R4F01()
	nonFinite.Paths[0].Finite = Bool(false)
	got := Evaluate(nonFinite)
	if got.Decision != DecisionPass || got.ProofValid || got.Selected == nil || got.Selected.FinitePROVPaths != 2 {
		t.Fatalf("non-finite proof result = %#v", got)
	}
}

func mustClone(t *testing.T, input Input) Input {
	t.Helper()
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	copy, err := DecodeInput(data)
	if err != nil {
		t.Fatal(err)
	}
	return copy
}

func reverse(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func reversePressures(values []PressureRecord) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func relocate(value, oldRoot, newRoot string) string {
	return strings.Replace(value, oldRoot, newRoot, 1)
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
