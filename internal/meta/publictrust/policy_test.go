package publictrust

import (
	"os"
	"path/filepath"
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
