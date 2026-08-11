package cache

import (
	"errors"
	"testing"
)

func TestHostStageEvidenceSeparatesVerifiedAndDeferred(t *testing.T) {
	current := CurrentStageEvidence()
	if err := current.Validate(); err != nil {
		t.Fatalf("current evidence invalid: %v", err)
	}
	if current.Stage != GoHostedStage || current.Status != EvidenceVerified {
		t.Fatalf("unexpected current evidence: %+v", current)
	}
	future := FutureStageEvidence()
	if err := future.Validate(); err != nil {
		t.Fatalf("future evidence invalid: %v", err)
	}
	if future.Stage != GoooHostedStage || future.Status != EvidenceDeferred {
		t.Fatalf("unexpected future evidence: %+v", future)
	}
	future.Status = EvidenceVerified
	if err := future.Validate(); !errors.Is(err, ErrUnimplementedStage) {
		t.Fatalf("future verification error = %v, want ErrUnimplementedStage", err)
	}
}

func TestStageEvidenceRejectsMalformedClaims(t *testing.T) {
	cases := []StageEvidence{
		{Stage: HostStage("unknown"), Status: EvidenceDeferred, Authority: "test"},
		{Stage: GoHostedStage, Status: EvidenceStatus("unknown"), Authority: "test"},
		{Stage: GoHostedStage, Status: EvidenceVerified},
	}
	for _, evidence := range cases {
		if err := evidence.Validate(); !errors.Is(err, ErrInvalidStageEvidence) &&
			!errors.Is(err, ErrInvalidHostStage) {
			t.Errorf("evidence %+v error = %v", evidence, err)
		}
	}
}

func TestHostStageIsPartOfCacheKeyAndMetadata(t *testing.T) {
	goKey := makeTestKey(t, "v1", "billing")
	futureKey, err := NewKey(KeySpec{
		Version: "v1", Namespace: "billing", ToolVersion: "compiler-1",
		HostStage: GoooHostedStage, Inputs: map[string]any{"source": "main.gooo"},
		Options:   map[string]any{"mode": "fast"},
		Freshness: testFreshnessSpec(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if goKey.HostStage != GoHostedStage || futureKey.HostStage != GoooHostedStage {
		t.Fatalf("unexpected key stages: %q %q", goKey.HostStage, futureKey.HostStage)
	}
	if goKey.String() == futureKey.String() {
		t.Fatal("host stage did not change the cache key")
	}
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(futureKey, []byte("future projection")); err != nil {
		t.Fatal(err)
	}
	metadata, err := cache.GetMetadata(futureKey)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.HostStage != GoooHostedStage {
		t.Fatalf("metadata host stage = %q, want %q", metadata.HostStage, GoooHostedStage)
	}
}

func TestNewKeyRejectsUnknownHostStage(t *testing.T) {
	_, err := NewKey(KeySpec{Namespace: "billing", HostStage: HostStage("future")})
	if !errors.Is(err, ErrInvalidHostStage) {
		t.Fatalf("unknown stage error = %v, want ErrInvalidHostStage", err)
	}
}
