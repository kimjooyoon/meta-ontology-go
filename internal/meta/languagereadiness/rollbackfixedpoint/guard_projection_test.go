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

func guardProjectionFixture(t *testing.T, name string) []byte {
	t.Helper()
	packed, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(packed))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func collectedGuardProjection(t *testing.T) (Source, Report, []byte) {
	t.Helper()
	root := t.TempDir()
	guard := filepath.Join(root, "guard.json")
	ledger := filepath.Join(root, "ledger.json")
	guardRaw := guardProjectionFixture(t, "guard-original-23b83809.json.gz")
	if err := os.WriteFile(guard, guardRaw, 0600); err != nil {
		t.Fatal(err)
	}
	ledgerRaw := guardProjectionFixture(t, "transformation-original-23b83809.json.gz")
	if err := os.WriteFile(ledger, ledgerRaw, 0600); err != nil {
		t.Fatal(err)
	}
	raw := guardProjectionFixture(t, "rollback-lossy-original-23b83809.json.gz")
	var legacy Report
	if err := json.Unmarshal(raw, &legacy); err != nil {
		t.Fatal(err)
	}
	source := Collect(guard, ledger, "23b83809300c153247e96bbf357fa8fc4635bc8a")
	if source.CollectionError != "" {
		t.Fatal(source.CollectionError)
	}
	return source, legacy, raw
}

func TestCollectedGuardRetainsKnownRefusalWithoutPromoting(t *testing.T) {
	source, legacy, raw := collectedGuardProjection(t)
	if err := Validate(legacy); err != nil || !bytes.Equal(Encode(Build(legacy.Source)), raw) {
		t.Fatal("historical lossy observation bytes changed", err)
	}
	if IsKnownAppliedTerminal(legacy) {
		t.Fatal("missing known-refusal count was inferred into the old report")
	}
	if source.Guard.Satisfied != 9 || source.Guard.NotSatisfied != 1 ||
		source.Guard.Unresolved != 2 || source.Guard.Total != 12 {
		t.Fatal("canonical guard populations were lost", source.Guard)
	}
	prior := source
	prior.Guard.NotSatisfied = 0
	if prior != legacy.Source {
		t.Fatal("collect changed original source evidence beyond the retained refusal count")
	}
	report := Build(source)
	if err := Validate(report); err != nil || !IsKnownAppliedTerminal(report) {
		t.Fatal("complete non-promoting sandbox observation was not recognized", err)
	}
	if report.Decision != DecisionFailClosed || report.Reason != ReasonRejected ||
		report.Summary.Satisfied != 8 || report.Summary.NotSatisfied != 2 ||
		report.Summary.RecoveredFixedPoints != 0 || report.Summary.AuthorizedPromotions != 0 ||
		report.RepositoryMutationAuthorized || report.RepositoryWrites != 0 {
		t.Fatal("retained refusal became recovery or promotion", report)
	}
	if !bytes.Contains(Encode(report), []byte(`"not_satisfied": 1`)) {
		t.Fatal("known refusal was omitted from the new projection")
	}
}

func TestCollectedGuardRejectsContradictoryPopulationsAndAuthority(t *testing.T) {
	original, _, _ := collectedGuardProjection(t)
	mutations := map[string]func(*Source){
		"missing refusal":     func(s *Source) { s.Guard.NotSatisfied = 0 },
		"negative refusal":    func(s *Source) { s.Guard.NotSatisfied = -1 },
		"excess refusal":      func(s *Source) { s.Guard.NotSatisfied = 13 },
		"overcount":           func(s *Source) { s.Guard.NotSatisfied = 2 },
		"negative satisfied":  func(s *Source) { s.Guard.Satisfied = -1 },
		"excess satisfied":    func(s *Source) { s.Guard.Satisfied = 13 },
		"negative unresolved": func(s *Source) { s.Guard.Unresolved = -1 },
		"excess unresolved":   func(s *Source) { s.Guard.Unresolved = 13 },
		"different total":     func(s *Source) { s.Guard.Total = 13 },
		"guard authority":     func(s *Source) { s.Guard.RepositoryMutationAuthorized = true },
		"ledger authority":    func(s *Source) { s.Transformation.PromotionAuthorized = true },
		"source writes":       func(s *Source) { s.RepositoryWrites = 1 },
		"source changed":      func(s *Source) { s.Transformation.SourceWorkspaceUnchanged = false },
		"unknown effect":      func(s *Source) { s.Transformation.UnknownCount = 1 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			source := original
			mutate(&source)
			if IsKnownAppliedTerminal(Build(source)) {
				t.Fatal("contradiction or mutation authority became a completed observation")
			}
		})
	}
}
