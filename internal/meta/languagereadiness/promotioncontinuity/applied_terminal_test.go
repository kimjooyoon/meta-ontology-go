package promotioncontinuity

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/rollbackfixedpoint"
)

func originalAppliedRecovery(t *testing.T) (rollbackfixedpoint.Report, []byte) {
	t.Helper()
	path := "../rollbackfixedpoint/testdata/applied-original-8f45c78e.json.gz"
	if root := os.Getenv("GOOO_ROLLBACK_TEST_DRAFT"); root != "" {
		path = filepath.Join(root, "internal/meta/languagereadiness/rollbackfixedpoint/testdata/applied-original-8f45c78e.json.gz")
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var report rollbackfixedpoint.Report
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	return report, raw
}

func appliedEvidence(t *testing.T) (string, GuardEvidence, RecoveryEvidence) {
	t.Helper()
	original, raw := originalAppliedRecovery(t)
	file := filepath.Join(t.TempDir(), "recovery.json")
	if err := os.WriteFile(file, raw, 0600); err != nil {
		t.Fatal(err)
	}
	recovery, err := readRecovery(file)
	if err != nil {
		t.Fatal(err)
	}
	g := original.Source.Guard
	guard := GuardEvidence{Schema: guardSchema, FileSHA256: g.FileSHA256,
		ReportDigest: g.ReportDigest, HeadSHA: g.HeadSHA, Decision: g.Decision,
		Reason: g.Reason, Resolution: g.Resolution, Satisfied: g.Satisfied,
		Total: g.Total, Unresolved: g.Unresolved}
	return original.Source.ExpectedHeadSHA, guard, recovery
}

func TestAppliedRecoveryPreservesNonPromotingContinuity(t *testing.T) {
	head, guard, recovery := appliedEvidence(t)
	report := Evaluate(head, guard, recovery)
	if !IsKnownNonPromotingTerminal(report) || report.Decision != DecisionFailClosed || report.Resolution != "EXACT" {
		t.Fatalf("conformant sandbox observation was lost: %+v", report)
	}
	if report.Summary.AuthorizedGuardReceipts != 0 || report.Summary.AuthorizedRecoveryRoutes != 0 ||
		report.RepositoryMutationAuthorized || report.RepositoryWrites != 0 ||
		report.Source.Recovery.RecoveredFixedPoints != 0 || report.Source.Recovery.TransformationEffects != 2 {
		t.Fatal("sandbox effects were erased or converted to authorization")
	}
	for _, coordinate := range report.Coordinates {
		if coordinate.Status == "SATISFIED" && strings.Contains(coordinate.ID, "authoriz") {
			t.Fatal(coordinate)
		}
	}
	for _, proof := range report.Proofs {
		if proof.Passed && strings.Contains(proof.MetaOperation, "authoriz") {
			t.Fatal(proof)
		}
	}
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.ReportDigest != Evaluate(head, guard, recovery).ReportDigest {
		t.Fatal("unstable classification")
	}
}

func TestAppliedRecoveryRejectsUnknownOrContradictoryConsumerFields(t *testing.T) {
	head, guard, original := appliedEvidence(t)
	mutations := map[string]func(*RecoveryEvidence){
		"decision":          func(v *RecoveryEvidence) { v.Decision = "UNKNOWN" },
		"reason":            func(v *RecoveryEvidence) { v.Reason = "ROLLBACK_FIXED_POINT_RECOVERED" },
		"mode":              func(v *RecoveryEvidence) { v.Mode = "RECOVERED_FIXED_POINT" },
		"head":              func(v *RecoveryEvidence) { v.TransformationHeadSHA = strings.Repeat("a", 40) },
		"counts":            func(v *RecoveryEvidence) { v.Satisfied = 10 },
		"guard counts":      func(v *RecoveryEvidence) { v.GuardSatisfied++ },
		"guard unknown":     func(v *RecoveryEvidence) { v.GuardDecision = "UNKNOWN" },
		"recovered":         func(v *RecoveryEvidence) { v.RecoveredFixedPoints++ },
		"authorized":        func(v *RecoveryEvidence) { v.AuthorizedPromotions++ },
		"effects":           func(v *RecoveryEvidence) { v.TransformationEffects++ },
		"applied":           func(v *RecoveryEvidence) { v.TransformationAppliedEffects-- },
		"receipt":           func(v *RecoveryEvidence) { v.TransformationReceiptCount-- },
		"refuted":           func(v *RecoveryEvidence) { v.TransformationRefutedEffects++ },
		"failed":            func(v *RecoveryEvidence) { v.TransformationFailureCount++ },
		"unknown":           func(v *RecoveryEvidence) { v.TransformationUnknownCount++ },
		"direct unknown":    func(v *RecoveryEvidence) { v.TransformationDirectUnknownCount++ },
		"dependent unknown": func(v *RecoveryEvidence) { v.TransformationDependencyBlockedUnknownCount++ },
		"workspace":         func(v *RecoveryEvidence) { v.TransformationWorkspaceMode = "LIVE" },
		"boundary":          func(v *RecoveryEvidence) { v.WriteBoundary = "SOURCE" },
		"source changed":    func(v *RecoveryEvidence) { v.SourceWorkspaceUnchanged = false },
		"writes":            func(v *RecoveryEvidence) { v.SourceRepositoryWrites++ },
		"guard writes":      func(v *RecoveryEvidence) { v.GuardRepositoryWrites++ },
		"authority":         func(v *RecoveryEvidence) { v.MutationAuthorized = true },
		"guard authority":   func(v *RecoveryEvidence) { v.GuardMutationAuthorized = true },
		"promotion":         func(v *RecoveryEvidence) { v.TransformationAuthorization = true },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			value := original
			mutate(&value)
			value.TransformationCausalBindingDigest = causalBindingDigest(value)
			report := Evaluate(head, guard, value)
			if IsKnownNonPromotingTerminal(report) || report.Decision == DecisionPass {
				t.Fatal("contradiction accepted")
			}
		})
	}
	guard.PromotionAuthorized = true
	if IsKnownNonPromotingTerminal(Evaluate(head, guard, original)) {
		t.Fatal("guard authorization mismatch accepted")
	}
}

func TestReadAppliedRecoveryRejectsMalformedCanonicalRecords(t *testing.T) {
	_, raw := originalAppliedRecovery(t)
	unknown := bytes.Replace(raw, []byte("{\n"), []byte("{\n  \"future_field\":true,\n"), 1)
	tampered := bytes.Replace(raw, []byte("\"effects\": 2"), []byte("\"effects\": 3"), 1)
	for name, data := range map[string][]byte{"trailing": append(append([]byte{}, raw...), []byte("{}")...), "unknown": unknown, "tampered": tampered} {
		t.Run(name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "recovery.json")
			if err := os.WriteFile(file, data, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := readRecovery(file); err == nil {
				t.Fatal("malformed original receipt accepted")
			}
		})
	}
}
