package workspaceexecution

import (
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestAccessFactBudget8941(t *testing.T) {
	out := ProgramValueOrigins{}
	w := valueTransferWalker{out: &out, issues: map[string]bool{}}
	a := &valueTransferActivity{id: "urn:access8941:activity", body: "input.values[211]"}
	node := bodycodegen.BodyValueOrigin{Kind: "index_operand", Decoded: bodycodegen.FillOriginRange{Start: 13, End: 16}}
	for i := range 260 {
		node.ID = strconv.Itoa(i)
		w.recordAccess("root8941", a, node)
		w.recordAccess("root8941", a, node)
	}
	if len(out.Accesses) != 256 || !out.AccessesTruncated || len(w.accesses) != 256 {
		t.Fatal("occurrence cap and duplicate handling", len(out.Accesses), len(w.accesses), out.AccessesTruncated)
	}
	filtered := ProgramValueOrigins{}
	w = valueTransferWalker{out: &filtered, query: OriginQuery{CandidateID: "candidate8941"}, issues: map[string]bool{}}
	w.recordAccess("root8941", a, node)
	if len(filtered.Accesses) != 0 {
		t.Fatal("candidate-filtered query exposed unrelated access inventory")
	}
	w.query.CandidateID = ""
	node.Decoded.Start = -1
	w.recordAccess("root8941", a, node)
	if len(filtered.Accesses) != 0 || len(filtered.Issues) != 1 || filtered.Issues[0].Reason != "ACCESS_SOURCE_RANGE_UNAVAILABLE" {
		t.Fatal("invalid access range was emitted", filtered)
	}
	t.Log("8941: structural budget check caps access records at256, deduplicates visits, omits candidate-filtered inventory and diagnoses invalid range")
}
