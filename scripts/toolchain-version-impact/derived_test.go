package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoooDerivedOutputsUseExactCurrentSource(t *testing.T) {
	root, projected, rules := derivedFixture(t)
	if len(rules) != 3 {
		t.Fatalf("expected three declared Gooo outputs, got %d", len(rules))
	}
	for _, rule := range rules {
		candidate, err := deriveCandidate(root, rule, projected)
		if err != nil {
			t.Fatalf("derive %s: %v", rule.TargetPath, err)
		}
		if !strings.Contains(string(candidate.After), "1.99.1") ||
			(rule.TargetPath != readmeTargetPath && strings.Contains(string(candidate.After), "1.27.3")) {
			t.Fatalf("derived target kept the old version: %s", rule.TargetPath)
		}
		original, err := os.ReadFile(filepath.Join(root, rule.TargetPath))
		if err != nil || digest(original) != digest(candidate.Before) {
			t.Fatalf("derivation changed source target %s: %v", rule.TargetPath, err)
		}
	}
}

func TestGoooDerivedOutputsRefuseStaleTargets(t *testing.T) {
	root, projected, rules := derivedFixture(t)
	for _, rule := range rules {
		path := filepath.Join(root, rule.TargetPath)
		original, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		stale := append([]byte(nil), original...)
		if rule.TargetPath == readmeTargetPath {
			stale = []byte(strings.Replace(string(original), "Go 1.27.3 toolchain", "Go 1.27.2 toolchain", 1))
			if string(stale) == string(original) {
				t.Fatal("README badge fixture was not found")
			}
		} else {
			stale = append(stale, '\n')
		}
		if err := os.WriteFile(path, stale, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := deriveCandidate(root, rule, projected); err == nil {
			t.Fatalf("stale target accepted: %s", rule.TargetPath)
		}
		if err := os.WriteFile(path, original, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	forged := rules[0]
	forged.TargetPath = "../../outside"
	if _, err := deriveCandidate(root, forged, projected); err == nil {
		t.Fatal("undeclared derived target accepted")
	}
}

func TestDerivedProvenanceNamesTheProjectedGoooSource(t *testing.T) {
	for _, target := range []string{publicTrustTargetPath, readmeTargetPath} {
		file := projectedFile{Path: target, InputSHA256: "sha256:old", OutputSHA256: "sha256:new",
			DerivationRule: "urn:gooo:toolchain-projection:test", SourcePath: publicTrustSourcePath,
			SourceOutputSHA256: "sha256:projected-source"}
		nodes := derivedGraphNodes("sha256:plan", "urn:plan", "urn:decision", file)
		activity := nodes[1].(map[string]any)
		output := nodes[2].(map[string]any)
		sourceID := projectionOutputID(file.SourcePath, file.SourceOutputSHA256)
		used := activity["prov:used"].([]any)
		derivedFrom := output["prov:wasDerivedFrom"].([]any)
		if used[0].(map[string]string)["@id"] != sourceID ||
			derivedFrom[0].(map[string]string)["@id"] != sourceID {
			t.Fatalf("derived %s has no projected Gooo origin", target)
		}
		wantAncestors := 1
		if target == readmeTargetPath {
			wantAncestors = 2
		}
		if len(derivedFrom) != wantAncestors {
			t.Fatalf("derived %s has wrong ancestor count", target)
		}
	}
}

func TestMissingDerivedTargetHashKeepsPlanOpen(t *testing.T) {
	observed := report{Version: "1.27.3", ModuleSHA256: "sha256:module", ActiveOccurrences: 1,
		ClassifiedOccurrences: 1, Occurrences: []occurrence{{Path: publicTrustSourcePath,
			SHA256: "sha256:source", Classified: true, StartByte: 0, EndByte: len("1.27.3")}}}
	plan, err := buildProjectionPlan(observed, "1.27.4")
	if err != nil || len(plan.DerivedOutputs) != 2 || plan.UnresolvedDerivedOutputs != 2 {
		t.Fatalf("missing derived target preimages were not exposed: %+v, %v", plan, err)
	}
}

func derivedFixture(t *testing.T) (string, map[string][]byte, []derivedOutputRule) {
	t.Helper()
	root := t.TempDir()
	paths := []string{compatibilitySourcePath, compatibilityTargetPath, publicTrustSourcePath,
		publicTrustTargetPath, readmeTargetPath}
	for _, path := range paths {
		data, err := os.ReadFile(filepath.Join("..", "..", path))
		if err != nil {
			t.Fatal(err)
		}
		output := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(output, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	projected := make(map[string][]byte, 2)
	var sites []occurrence
	targetHashes := map[string]string{}
	for _, path := range []string{compatibilityTargetPath, publicTrustTargetPath, readmeTargetPath} {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		targetHashes[path] = digest(data)
	}
	for _, path := range []string{compatibilitySourcePath, publicTrustSourcePath} {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		projected[path] = []byte(strings.ReplaceAll(string(data), "1.27.3", "1.99.1"))
		sites = append(sites, occurrence{Path: path, SHA256: digest(data), Classified: true})
	}
	return root, projected, derivedOutputRules(sites, targetHashes)
}
