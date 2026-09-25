package metarecognition

import (
	"testing"
)

func TestReplayCanonicalizesRootPathAndOrder(t *testing.T) {
	original := Corpus()
	original[0].Baseline.Roots = []string{"root://a", "root://z"}
	original[0].Baseline.Path.IDs = []string{"path://a", "path://z"}
	perturbed := append([]Case(nil), original...)
	perturbed[0].Baseline.Roots = []string{"root://z", "root://a"}
	perturbed[0].Baseline.Path.IDs = []string{"path://z", "path://a"}
	relocated := append([]Case(nil), original...)
	relocated[0].Baseline.WorkspaceRoot = "/physical/root-b"
	relocated[0].Baseline.SourcePath = "/physical/root-b/case-01.go"
	commands := append([]CommandAssertion(nil), perturbed[6].Baseline.Commands...)
	for left, right := 0, len(commands)-1; left < right; left, right = left+1, right-1 {
		commands[left], commands[right] = commands[right], commands[left]
	}
	perturbed[6].Baseline.Commands = commands
	for left, right := 0, len(perturbed)-1; left < right; left, right = left+1, right-1 {
		perturbed[left], perturbed[right] = perturbed[right], perturbed[left]
	}
	first, err := ReplayJSON(original)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ReplayJSON(perturbed)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("replay changed under root/path/order permutation")
	}
	third, err := ReplayJSON(relocated)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(third) {
		t.Fatal("replay changed under physical workspace relocation")
	}
	relocated[0].Baseline.SourcePath = "/physical/root-b/other.go"
	different, err := ReplayJSON(relocated)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) == string(different) {
		t.Fatal("replay discarded a genuinely different relative source path")
	}
	decoded, err := DecodeReplayJSON(second)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Cases) != len(original) {
		t.Fatalf("decoded cases = %d, want %d", len(decoded.Cases), len(original))
	}
	if decoded.Cases[0].ID != "case-01" {
		t.Fatalf("decoded cases are not canonicalized: first=%q", decoded.Cases[0].ID)
	}
}
