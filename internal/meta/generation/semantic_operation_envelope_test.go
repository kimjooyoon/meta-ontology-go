package generation

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestSemanticOperationEnvelopeFixedDenominator(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", "self-improvement-minimal-loop", "operation-envelope.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	wantActivities := SemanticOperationActivityNames()
	wantScenarios := SemanticOperationScenarioIDs()
	wantDecisions := map[string]string{
		"C1": "CLOSED", "C2": "CLOSED",
		"U1": "UNKNOWN", "U2": "UNKNOWN",
		"R1": "REFUTED", "R2": "REFUTED",
	}
	wantReasons := map[string]string{
		"C1": "EXACT_RESULT", "C2": "EXACT_RESULT",
		"U1": "DIRECT_MISSING", "U2": "STALE",
		"R1": "EFFECT_ESCALATION", "R2": "REPLAY_COLLISION",
	}
	wantMetrics := map[string]EnvelopeMetrics{
		"C1": {OperationRequests: 0, OperationResults: 1, EffectsGranted: 0, EffectsUsed: 0},
		"C2": {OperationRequests: 1, OperationResults: 1, EffectsGranted: 1, EffectsUsed: 1, ReplayComparisons: 1},
		"U1": {OperationRequests: 1, OperationResults: 0, EffectsGranted: 1, EffectsUsed: 0},
		"U2": {OperationRequests: 1, OperationResults: 1, EffectsGranted: 1, EffectsUsed: 0, StaleRejections: 1},
		"R1": {OperationRequests: 1, OperationResults: 1, EffectsGranted: 1, EffectsUsed: 1, EffectEscalationsRefuted: 1},
		"R2": {OperationRequests: 1, OperationResults: 1, EffectsGranted: 1, EffectsUsed: 1, ReplayComparisons: 1, ReplayMismatches: 1},
	}
	counts := map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0}
	for _, scenarioID := range wantScenarios {
		firstDir := t.TempDir()
		run, err := GenerateSemanticOperationEnvelope(source, scenarioID, firstDir)
		if err != nil {
			t.Fatalf("%s generate: %v", scenarioID, err)
		}
		if run.IR.Decision.Decision != wantDecisions[scenarioID] || run.IR.Decision.Reason != wantReasons[scenarioID] {
			t.Fatalf("%s generated decision: got %s/%s", scenarioID, run.IR.Decision.Decision, run.IR.Decision.Reason)
		}
		if !reflect.DeepEqual(run.IR.Activities, wantActivities) {
			t.Fatalf("%s activity graph mismatch: got %v", scenarioID, run.IR.Activities)
		}
		if len(run.Artifacts) != 6 {
			t.Fatalf("%s generated %d artifacts, want 6", scenarioID, len(run.Artifacts))
		}
		assertEnvelopeMetrics(t, scenarioID, run.Receipt.Metrics, wantMetrics[scenarioID])
		if scenarioID == "U1" || scenarioID == "U2" {
			assertEnvelopeUnknownFields(t, run.Receipt.Decision.Unknown)
		}
		verified, err := VerifySemanticOperationEnvelope(firstDir)
		if err != nil {
			t.Fatalf("%s independent verification: %v", scenarioID, err)
		}
		if verified.Decision != wantDecisions[scenarioID] || verified.Reason != wantReasons[scenarioID] {
			t.Fatalf("%s verified decision: got %s/%s", scenarioID, verified.Decision, verified.Reason)
		}
		if verified.Metrics != run.Receipt.Metrics {
			t.Fatalf("%s verifier metrics differ", scenarioID)
		}
		counts[verified.Decision]++

		secondDir := t.TempDir()
		if _, err := GenerateSemanticOperationEnvelope(source, scenarioID, secondDir); err != nil {
			t.Fatalf("%s replay generate: %v", scenarioID, err)
		}
		for _, artifact := range run.Artifacts {
			first, err := os.ReadFile(filepath.Join(firstDir, artifact.Name))
			if err != nil {
				t.Fatal(err)
			}
			second, err := os.ReadFile(filepath.Join(secondDir, artifact.Name))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(first, second) {
				t.Fatalf("%s artifact %s is not deterministic", scenarioID, artifact.Name)
			}
		}
	}
	wantCounts := map[string]int{"CLOSED": 2, "UNKNOWN": 2, "REFUTED": 2}
	if !reflect.DeepEqual(counts, wantCounts) {
		t.Fatalf("decision counts: got %v, want %v", counts, wantCounts)
	}
}

func assertEnvelopeMetrics(t *testing.T, scenarioID string, got, want EnvelopeMetrics) {
	t.Helper()
	if got.OperationRequests != want.OperationRequests || got.OperationResults != want.OperationResults ||
		got.EffectsGranted != want.EffectsGranted || got.EffectsUsed != want.EffectsUsed ||
		got.ReplayComparisons != want.ReplayComparisons || got.ReplayMismatches != want.ReplayMismatches ||
		got.StaleRejections != want.StaleRejections || got.EffectEscalationsRefuted != want.EffectEscalationsRefuted {
		t.Fatalf("%s metric core: got %+v, want core %+v", scenarioID, got, want)
	}
	if got.InputDescendantDirs != 0 || got.InputRegularFiles != 1 || got.InputGoPhysicalLines != 0 ||
		got.InputGoooPhysicalLines <= 0 || got.OutputArtifactFiles != 6 || got.PeakRSSKib != 0 || got.WallMS != 0 ||
		got.RepositoryWrites != 0 || got.LocalTestExecutions != 0 {
		t.Fatalf("%s metric contract: got %+v", scenarioID, got)
	}
}

func assertEnvelopeUnknownFields(t *testing.T, unknown *EnvelopeUnknownState) {
	t.Helper()
	if unknown == nil {
		t.Fatal("unknown decision has no unknown evidence")
	}
	data, err := json.Marshal(unknown)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(fields))
	for field := range fields {
		got = append(got, field)
	}
	sort.Strings(got)
	want := []string{"blocked_by", "next_operation", "reason", "stage", "step", "unknown_class"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unknown fields: got %v, want %v", got, want)
	}
}

func TestSemanticOperationEnvelopeRejectsNonEmptyOutput(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", "self-improvement-minimal-loop", "operation-envelope.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	outputDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outputDir, "unrelated.txt"), []byte("caller-owned"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateSemanticOperationEnvelope(source, "C1", outputDir); err == nil {
		t.Fatal("expected non-empty caller-owned output to be rejected")
	}
}
