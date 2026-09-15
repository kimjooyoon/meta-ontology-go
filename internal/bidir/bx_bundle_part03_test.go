package bidir

import (
	_ "embed"
	"reflect"
	"testing"
)

func assertBundleBase(t *testing.T, got map[string]bxBundleArtifact, want BXBaseEvidence) {
	t.Helper()
	wantArtifacts := map[string]BXArtifactEvidence{"dsl": want.DSL, "ir": want.IR, "go": want.Go, "source_map": want.SourceMap, "evidence": want.Evidence, "provenance": want.Provenance}
	for name, artifact := range wantArtifacts {
		if got[name].Hash != artifact.Hash || got[name].Count != artifact.Count {
			t.Fatalf("bundle base %s mismatch: got=%#v want=%#v", name, got[name], artifact)
		}
	}
}
func assertBundleDelta(t *testing.T, got bxBundleDelta, want BXDeltaEvidence) {
	t.Helper()
	if got.SequenceHash != want.SequenceHash || got.OrderHash != want.OrderHash || got.PortOrderHash != want.PortOrderHash || got.RelationOrderHash != want.RelationOrderHash || got.PartialObservation != want.PartialObservation {
		t.Fatalf("bundle delta hashes mismatch: got=%#v want=%#v", got, want)
	}
	if !reflect.DeepEqual(got.PortSequence, want.PortSequence) || !reflect.DeepEqual(got.RelationSequence, want.RelationSequence) || !reflect.DeepEqual(got.Candidates, want.Candidates) {
		t.Fatal("bundle delta sequence or candidate evidence mismatch")
	}
	if !reflect.DeepEqual(got.Closure.Touched, idsAsStrings(want.Locality.Touched)) || !reflect.DeepEqual(got.Closure.Affected, idsAsStrings(want.Locality.Affected)) || !reflect.DeepEqual(got.Closure.Members, idsAsStrings(want.ClosureMembers)) || got.Closure.Hash != want.LocalityClosureHash {
		t.Fatal("bundle locality closure evidence mismatch")
	}
	if !reflect.DeepEqual(got.Evidence.IDs, want.EvidenceSpans.IDs) || !reflect.DeepEqual(got.Evidence.FactKeys, want.EvidenceSpans.FactKeys) || !reflect.DeepEqual(got.Evidence.Records, bundleRecords(want.EvidenceSpans.Records)) || got.Evidence.IDCount != want.EvidenceSpans.IDCount || got.Evidence.SpanCount != want.EvidenceSpans.SpanCount || got.Evidence.Hash != want.EvidenceHash || got.Evidence.EvidenceIDAuthority != want.EvidenceSpans.EvidenceIDAuthority {
		t.Fatalf("bundle evidence ID/span set mismatch: got=%#v want=%#v", got.Evidence, want.EvidenceSpans)
	}
}
func bundleRecords(records []BXEvidenceRecord) []bxBundleEvidenceRecord {
	values := make([]bxBundleEvidenceRecord, len(records))
	for index, record := range records {
		values[index] = bxBundleEvidenceRecord{EvidenceID: record.EvidenceID, FactKey: record.FactKey, Span: spanText(record.Span), HasSpan: record.HasSpan}
	}
	return values
}
func assertBundleTransaction(t *testing.T, got bxBundleTransaction, want BXTransactionEvidence) {
	t.Helper()
	if got.Observer != want.ObserverKind || got.Observed != want.Observed || got.Atomic != want.Atomic || got.NoWrite != want.NoWrite || got.Deferred != want.Deferred || got.Before != stateBundle(want.Before) || got.After != stateBundle(want.After) {
		t.Fatalf("bundle transaction mismatch: got=%#v want=%#v", got, want)
	}
}
func stateBundle(state BXStateEvidence) bxBundleState {
	return bxBundleState{Semantic: state.Semantic, Source: state.Source, Region: state.Region, Slot: state.Slot, Bytes: state.Bytes, LStat: state.LStat}
}
func idsAsStrings(ids []ID) []string {
	values := make([]string, len(ids))
	for index, id := range ids {
		values[index] = string(id)
	}
	return values
}
