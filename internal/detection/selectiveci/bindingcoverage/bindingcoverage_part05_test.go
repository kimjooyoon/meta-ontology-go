package bindingcoverage

import (
	"math"
	"strings"
	"testing"
)

func TestStrictJSON(t *testing.T) {
	encoded, err := EncodeInputJSON(fixtureInput())
	if err != nil {
		t.Fatal(err)
	}
	if got := ClassifyJSON(encoded); got.Decision != DecisionExact {
		t.Fatalf("encoded fixture got %s/%s", got.Decision, got.Reason)
	}
	canonical := strings.TrimSpace(string(encoded))
	duplicate := strings.Replace(canonical, `"schema_version":"`+SchemaVersion+`"`, `"schema_version":"`+SchemaVersion+`","schema_version":"`+SchemaVersion+`"`, 1)
	unknown := strings.TrimSuffix(canonical, "}") + `,"extra":true}`
	trailing := canonical + " {}"
	for name, data := range map[string]string{"duplicate": duplicate, "unknown": unknown, "trailing": trailing} {
		t.Run(name, func(t *testing.T) {
			if got := ClassifyJSON([]byte(data)); got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput {
				t.Fatalf("got %s/%s", got.Decision, got.Reason)
			}
		})
	}
	missingLists := fixtureInput()
	missingLists.RequiredBindings = nil
	if got := Observe(missingLists); got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput {
		t.Fatalf("nil list got %s/%s", got.Decision, got.Reason)
	}
	missingLists = fixtureInput()
	missingLists.Partitions = nil
	if got := Observe(missingLists); got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput {
		t.Fatalf("nil partitions got %s/%s", got.Decision, got.Reason)
	}
	missingLists = fixtureInput()
	missingLists.PrecedenceRegistry = nil
	if got := Observe(missingLists); got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput {
		t.Fatalf("nil precedence got %s/%s", got.Decision, got.Reason)
	}
	if _, err := DecodeJSON([]byte(unknown)); err == nil {
		t.Fatal("strict decoder accepted unknown field")
	}
	if got := ClassifyJSON([]byte(`{"schema_version":"gooo/other/v1"}`)); got.Decision != DecisionUnknown || got.Reason != ReasonUnknownSchema {
		t.Fatalf("unsupported JSON schema got %s/%s", got.Decision, got.Reason)
	}
}
func TestWorkAccountingOverflow(t *testing.T) {
	if _, ok := addUint64(math.MaxUint64, 1); ok {
		t.Fatal("addition overflow was accepted")
	}
	if _, ok := workUnits(math.MaxUint64, 0, 1); ok {
		t.Fatal("work overflow was accepted")
	}
	if got, ok := workUnits(9, 18, 18); !ok || got != 45 {
		t.Fatalf("work units = %d/%v, want 45/true", got, ok)
	}
}
func TestSharedEndpointReferences(t *testing.T) {
	got := Observe(sharedEndpointInput())
	if got.Decision != DecisionExact || got.Reason != ReasonComplete {
		t.Fatalf("got %s/%s, want EXACT/COMPLETE", got.Decision, got.Reason)
	}
	if got.RequiredBindingCount != 2 || got.MatchCoveredCount != 2 || got.MismatchCoveredCount != 2 || got.PartitionCount != 4 {
		t.Fatalf("coverage counts = %d/%d/%d/%d, want 2/2/2/4", got.RequiredBindingCount, got.MatchCoveredCount, got.MismatchCoveredCount, got.PartitionCount)
	}
	if got.EndpointReferenceCount != 4 || got.DeterministicWorkUnits != 10 {
		t.Fatalf("endpoint/work counts = %d/%d, want 4/10", got.EndpointReferenceCount, got.DeterministicWorkUnits)
	}
	t.Logf("shared endpoint fixture digest=%s", got.CanonicalDigest)
}
