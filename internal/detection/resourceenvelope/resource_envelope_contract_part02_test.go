package resourceenvelope

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestResourceEnvelopeContractInputPermutationInvariance(t *testing.T) {
	corpus := loadContractCorpus(t)
	base := contractByName(t, corpus, "five-sample-boundary")
	want := independentObservation(base)
	for left, right := 0, len(base.Samples)-1; left < right; left, right = left+1, right-1 {
		base.Samples[left], base.Samples[right] = base.Samples[right], base.Samples[left]
	}
	if got := independentObservation(base); got != want {
		t.Fatalf("permuted samples changed observation: got=%#v want=%#v", got, want)
	}
}
func TestResourceEnvelopeContractCanonicalReplayEquality(t *testing.T) {
	corpus := loadContractCorpus(t)
	base := contractByName(t, corpus, "five-sample-boundary")
	first, err := json.Marshal(independentObservation(base))
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(independentObservation(base))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("canonical replay changed: first=%s second=%s", first, second)
	}
}
func TestResourceEnvelopeContractRejectsUnknownJSONField(t *testing.T) {
	raw := readCorpus(t)
	var document map[string]json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	document["unregistered"] = json.RawMessage(`true`)
	mutated, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := decodeStrict(mutated, &contractCorpus{}); err == nil {
		t.Fatal("unknown JSON field was accepted")
	}
}
func TestResourceEnvelopeContractRejectsTrailingJSON(t *testing.T) {
	raw := append(readCorpus(t), []byte(`
{"trailing":true}`)...)
	if err := decodeStrict(raw, &contractCorpus{}); err == nil {
		t.Fatal("trailing JSON was accepted")
	}
}
