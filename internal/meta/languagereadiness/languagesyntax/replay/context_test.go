package replay

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestCompanionSourceKeepsStandaloneSyntaxAndPackageSemantics(t *testing.T) {
	files := fstest.MapFS{
		"base.gooo":      &fstest.MapFile{Data: []byte("package sample\nnamespace sample\nentity Input id \"urn:sample:input\" fields {\n field value id \"urn:sample:value\" type integer required one\n}\n")},
		"extension.gooo": &fstest.MapFile{Data: []byte("package sample\nnamespace sample\nactivity Use(Input) -> Input id \"urn:sample:use\" computes \"return input\"\n")},
	}
	if single := ExecuteWithOrderedCollections(files, "extension.gooo", "VALID", ""); single.ObservedDecision == DecisionPass {
		t.Fatal("a missing declaration passed file-only semantic replay")
	}
	got := ExecuteWithOrderedCollectionsAndContext(files, "extension.gooo", []string{"base.gooo"}, "VALID", "")
	if got.ObservedDecision != DecisionPass || !got.ASTReplayed || !got.ByteReplayed || !got.SemanticReplayed || !got.GetPut || !got.PutGet {
		t.Fatalf("package-aware replay failed: %+v", got)
	}
	files["base.gooo"] = &fstest.MapFile{Data: []byte("package other\nnamespace sample\nentity Input id \"urn:sample:input\"\n")}
	invalid := ExecuteWithOrderedCollectionsAndContext(files, "extension.gooo", []string{"base.gooo"}, "VALID", "")
	if invalid.ObservedDecision != DecisionClosed || len(invalid.Diagnostics) == 0 || !strings.Contains(invalid.Diagnostics[0], "different package headers") {
		t.Fatalf("mismatched context was accepted: %+v", invalid)
	}
}
