package languagesyntax

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func TestSourceFamilyCannotBeRemovedOrSubstitutedByResealingAReport(t *testing.T) {
	const head = "0000000000000000000000000000000000000000"
	repository := os.DirFS("../../../..")
	registry, err := os.ReadFile("../../../../examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	report := Evaluate(repository, head, registry, languageconcept.BuildArtifact(repository))
	if err := Validate(report, head); err != nil || report.Decision != DecisionPass {
		t.Fatal("registered source-family baseline failed", err)
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"group", "member", "order", "bytes", "lines", "kind"} {
		var changed Report
		if err := json.Unmarshal(raw, &changed); err != nil {
			t.Fatal(err)
		}
		for i := range changed.Cases {
			item := &changed.Cases[i]
			if item.Definition.ID != "explicit-consumer-continuation" {
				continue
			}
			switch mode {
			case "group":
				changed.Source.CaseSources = nil
			case "member":
				item.Evidence.SourceMembers = item.Evidence.SourceMembers[:1]
			case "order":
				slices.Reverse(item.Evidence.SourceMembers)
			case "bytes":
				item.Evidence.SourceMembers[1].SourceDigest = item.Evidence.SourceMembers[0].SourceDigest
			case "lines":
				item.Evidence.SourceLines++
			case "kind":
				item.Evidence.SourceDigestKind = ""
			}
		}
		for i := range changed.Cases {
			changed.Cases[i].EvidenceDigest = caseDigest(changed.Cases[i], changed.Source)
		}
		changed = finish(changed)
		if err := Validate(changed, head); err == nil {
			t.Fatal("recomputed report seal hid a changed source family", mode)
		}
	}
}
