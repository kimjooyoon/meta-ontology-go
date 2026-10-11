package rollbackfixedpoint

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func appliedOriginal(t *testing.T) (Report, []byte) {
	t.Helper()
	path := "testdata/applied-original-8f45c78e.json.gz"
	if root := os.Getenv("GOOO_ROLLBACK_TEST_DRAFT"); root != "" {
		path = filepath.Join(root, "internal/meta/languagereadiness/rollbackfixedpoint", path)
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
	var report Report
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	return report, raw
}

func TestAppliedObservationKeepsOriginalRejectedRecoveryBytes(t *testing.T) {
	report, raw := appliedOriginal(t)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, Encode(Build(report.Source))) {
		t.Fatal("original rollback report changed")
	}
	if !IsKnownAppliedTerminal(report) {
		t.Fatal("conformant sandbox observation was not recognized")
	}
	if report.Decision != DecisionFailClosed || report.Summary.Satisfied != 8 ||
		report.Summary.NotSatisfied != 2 || report.Summary.RecoveredFixedPoints != 0 ||
		report.Summary.AuthorizedPromotions != 0 || report.Source.Transformation.Effects != 2 {
		t.Fatalf("observation became recovery or authorization: %+v", report)
	}
}

func TestAppliedObservationRejectsContradictoryEvidence(t *testing.T) {
	original, _ := appliedOriginal(t)
	mutations := map[string]func(*Source){
		"collection":         func(s *Source) { s.CollectionError = "missing" },
		"head":               func(s *Source) { s.ExpectedHeadSHA = "" },
		"guard head":         func(s *Source) { s.Guard.HeadSHA = "" },
		"ledger head":        func(s *Source) { s.Transformation.HeadSHA = "" },
		"guard decision":     func(s *Source) { s.Guard.Decision = "UNKNOWN" },
		"guard counts":       func(s *Source) { s.Guard.Satisfied++ },
		"guard unresolved":   func(s *Source) { s.Guard.Unresolved = 0 },
		"guard authority":    func(s *Source) { s.Guard.RepositoryMutationAuthorized = true },
		"guard writes":       func(s *Source) { s.Guard.RepositoryWrites++ },
		"writes":             func(s *Source) { s.RepositoryWrites++ },
		"workspace":          func(s *Source) { s.Transformation.WorkspaceMode = "LIVE" },
		"boundary":           func(s *Source) { s.Transformation.WriteBoundary = "SOURCE" },
		"source changed":     func(s *Source) { s.Transformation.SourceWorkspaceUnchanged = false },
		"effect count":       func(s *Source) { s.Transformation.Effects++ },
		"applied count":      func(s *Source) { s.Transformation.AppliedEffects-- },
		"receipt count":      func(s *Source) { s.Transformation.ReceiptCount-- },
		"refuted":            func(s *Source) { s.Transformation.RefutedEffects++ },
		"failure":            func(s *Source) { s.Transformation.FailureCount++ },
		"unknown":            func(s *Source) { s.Transformation.UnknownCount++ },
		"direct unknown":     func(s *Source) { s.Transformation.DirectUnknownCount++ },
		"dependency unknown": func(s *Source) { s.Transformation.DependencyBlockedUnknownCount++ },
		"causal digest":      func(s *Source) { s.Transformation.UnknownCausalDigest = "" },
		"ledger decision":    func(s *Source) { s.Transformation.Decision = "UNKNOWN" },
		"outcome":            func(s *Source) { s.Transformation.OperationOutcome = "UNKNOWN" },
		"receipt decision":   func(s *Source) { s.Transformation.ReceiptDecision = "REFUTED" },
		"promotion":          func(s *Source) { s.Transformation.PromotionAuthorized = true },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			source := original.Source
			mutate(&source)
			if IsKnownAppliedTerminal(Build(source)) {
				t.Fatal("contradiction recognized as known terminal")
			}
		})
	}
	tampered := original
	tampered.ReportDigest = ""
	if IsKnownAppliedTerminal(tampered) {
		t.Fatal("unbound report recognized")
	}
}
