package provenance

import (
	"errors"
	"strings"
	"testing"
)

func TestCurrentContractIsExplicitAndNonAuthoritative(t *testing.T) {
	contract := CurrentContract()
	if contract.Version != ContractVersion || contract.Format == "" || len(contract.Required) < 10 || len(contract.NegativeCases) < 10 {
		t.Fatalf("incomplete provenance contract: %#v", contract)
	}
	for _, deferred := range contract.Deferred {
		if strings.TrimSpace(deferred) == "" {
			t.Fatal("empty deferred contract boundary")
		}
	}
	for _, required := range contract.Required {
		if strings.Contains(strings.ToLower(required), "github") || strings.Contains(strings.ToLower(required), "credential") {
			t.Fatalf("authority was smuggled into storage fields: %q", required)
		}
	}
}

func TestMissingAuthorityCriticalFieldsFailClosed(t *testing.T) {
	store := New(t.TempDir() + "/missing.jsonl")
	missingSpan := testRecord("event/missing-span", "semantic/missing-span", StatusVerified)
	missingSpan.SourceSpan = SourceSpan{}
	if err := store.Append(missingSpan); err == nil {
		t.Fatal("record without source span was accepted")
	}
	missingDigest := testRecord("event/missing-digest", "semantic/missing-digest", StatusVerified)
	missingDigest.GraphDigest = ""
	if err := store.Append(missingDigest); err == nil {
		t.Fatal("record without graph digest was accepted")
	}
}

func TestExactDuplicateWithDifferentBytesIsConflict(t *testing.T) {
	store := New(t.TempDir() + "/duplicate.jsonl")
	record := testRecord("event/duplicate", "semantic/duplicate", StatusVerified)
	if err := store.Append(record); err != nil {
		t.Fatal(err)
	}
	record.Attributes["new"] = "value"
	if err := store.Append(record); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate with different bytes was accepted: %v", err)
	}
}
