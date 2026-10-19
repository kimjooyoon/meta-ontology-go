package syntaxregistration

import (
	"bytes"
	"context"
	"testing"
)

func TestGoooRegistrationAppendBindsTheActualDenominator(t *testing.T) {
	data, request := fixture(t)
	previous := data[denominatorPath(request.BaseVersion)].Data
	_, _, capability, err := corpusTotals(data[corpusPath].Data)
	if err != nil {
		t.Fatal(err)
	}
	record, value, err := composeRegistrationAppend(context.Background(), request,
		previous, capability, contractSource)
	if err != nil || !record.Accepted || record.NextVersion != request.BaseVersion+1 ||
		record.NextTarget != capability+1 || record.BaselineDigest != digest(previous) ||
		record.SourceDigest != request.SourceDigest || value.SourceSHA256 != digest(contractSource) ||
		value.ActivityID != "gooo://syntax-registration/compose-append" || value.ProjectionSHA256 == "" {
		t.Fatalf("Gooo append lost the observed source boundary: %+v %+v %v", record, value, err)
	}
	next, err := generateDenominator(previous, record)
	if err != nil || len(next) == 0 {
		t.Fatalf("Gooo append could not construct the denominator: %v", err)
	}
	record.NextTarget++
	if _, err := generateDenominator(previous, record); err == nil {
		t.Fatal("changed Gooo result entered the denominator writer")
	}
	record.NextTarget--
	record.BaselineDigest = digest([]byte("different denominator"))
	if _, err := generateDenominator(previous, record); err == nil {
		t.Fatal("changed historical denominator entered the writer")
	}
}

func TestGoooRegistrationAppendSourceAndCountCounterexamples(t *testing.T) {
	data, request := fixture(t)
	previous := data[denominatorPath(request.BaseVersion)].Data
	_, _, capability, err := corpusTotals(data[corpusPath].Data)
	if err != nil {
		t.Fatal(err)
	}
	changed := bytes.Replace(contractSource, []byte("next_target: input.base_target + 1"),
		[]byte("next_target: input.base_target + 2"), 1)
	if bytes.Equal(changed, contractSource) {
		t.Fatal("append source mutation anchor was not found")
	}
	_, changedOrigin, err := composeRegistrationAppend(context.Background(), request,
		previous, capability, changed)
	if failureReason(err) != "REGISTRATION_APPEND_CONTRACT_MISMATCH" ||
		changedOrigin.SourceSHA256 == digest(contractSource) {
		t.Fatalf("changed Gooo append result was not rejected with its source: %+v %v", changedOrigin, err)
	}
	if _, _, err := composeRegistrationAppend(context.Background(), request,
		previous, capability-1, contractSource); failureReason(err) != "REGISTRATION_APPEND_REFUTED" {
		t.Fatalf("contradictory corpus count was accepted: %v", err)
	}
}

func TestGoooRegistrationAppendProjectionNeedsItsSourceEdge(t *testing.T) {
	changed := bytes.Replace(contractSource,
		[]byte("GenerateSyntaxDenominator(PinnedRegistrationInput, AppendRecord)"),
		[]byte("GenerateSyntaxDenominator(PinnedRegistrationInput)"), 1)
	if bytes.Equal(changed, contractSource) {
		t.Fatal("append dependency mutation anchor was not found")
	}
	if _, err := bindContractSource(changed); err == nil {
		t.Fatal("denominator projection lost its Gooo append source edge")
	}
}
