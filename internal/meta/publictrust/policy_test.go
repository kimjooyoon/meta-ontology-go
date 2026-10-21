package publictrust

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCanonicalPolicyBindsGeneratedRows(t *testing.T) {
	sourcePath := filepath.Join("..", "..", "..", "examples", "public-trust-surface", "main.gooo")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := Load(CanonicalPolicyPath, source)
	if err != nil {
		t.Fatal(err)
	}
	summary := SummaryFor(policy, 0, 0)
	if summary.MetaRowsBound != ExpectedMetaRows || summary.TotalActiveBadges != ExpectedActiveBadges || summary.UniqueClaims != ExpectedMetaRows || summary.UniqueTargets != ExpectedMetaRows {
		t.Fatalf("summary = %#v", summary)
	}
	if summary.UnsupportedBadges != 0 || summary.DuplicateClaims != 0 || summary.DuplicateTargets != 0 {
		t.Fatalf("unsupported or duplicate rows: %#v", summary)
	}
	if summary.StateRows[DecisionClosed] != 11 || summary.StateRows[DecisionUnknown] != 3 || summary.StateRows[DecisionRefuted] != 2 {
		t.Fatalf("state rows = %#v", summary.StateRows)
	}
	block := RenderBadgeBlock(policy)
	if !strings.Contains(block, "#### Security / Supply Chain") || strings.Contains(block, "Scorecard") || strings.Contains(block, "Branch protection") {
		t.Fatalf("rendered block contains an unavailable or refuted badge: %s", block)
	}
}

func TestNewGoooToolchainBadgeCanRegenerateEvaluator(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "..", CanonicalPolicyPath))
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.ReplaceAll(string(source), ExpectedGoVersion, "1.99.1")
	if changed == string(source) {
		t.Fatal("Gooo toolchain badge version was not found")
	}
	policy, generated, err := GenerateNamed(CanonicalPolicyPath, []byte(changed))
	if err != nil || policy.GoVersion != "1.99.1" || !regexp.MustCompile(`(?m)^\s*GoVersion\s*=\s*"1\.99\.1"$`).Match(generated) {
		t.Fatalf("new Gooo badge did not project into Go: %v", err)
	}
	if _, err := Load(CanonicalPolicyPath, []byte(changed)); err == nil {
		t.Fatal("new source bypassed committed evaluator regeneration")
	}
}
