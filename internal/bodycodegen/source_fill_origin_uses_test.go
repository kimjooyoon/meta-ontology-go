package bodycodegen

import (
	"strings"
	"testing"
)

func TestRecordedOriginUsePaths5527(t *testing.T) {
	const body = "let ready = input > 89; let output = input; if ready { output = input + 173 }; return output"
	flow := InspectBodyValueOrigins(body, 1)
	start := strings.Index(body, "89")
	origin := SourceFillOrigin{SelectedBodySHA256: flow.BodySHA256, DataFlow: &flow,
		Fragments: []SourceFillFragment{{Kind: "declared_candidate", HoleID: "threshold5527", SelectedDecoded: FillOriginRange{start, start + 2}}}}
	uses, truncated, err := SourceFillUses(origin, 8)
	if err != nil || truncated || len(uses) != 2 {
		t.Fatal(err, truncated, uses)
	}
	control := false
	for _, use := range uses {
		if use.Relation != "static_dependency" || len(use.Path) < 2 {
			t.Fatal("indirect path lost", use)
		}
		for _, step := range use.Path {
			if step.Via == "control" {
				control = true
			}
		}
	}
	if !control {
		t.Fatal("return path omitted controlling candidate")
	}
	limited, cut, err := SourceFillUses(origin, 1)
	if err != nil || !cut || len(limited) != 1 {
		t.Fatal("row limit", err, limited, cut)
	}
	for i := range flow.Nodes {
		if flow.Nodes[i].Kind == "declaration" {
			flow.Nodes[i].Inputs = append(flow.Nodes[i].Inputs, flow.Nodes[i].ID)
		}
	}
	if _, _, err := SourceFillUses(origin, 8); err != nil {
		t.Fatal("bounded cycle rejected", err)
	}
	flow.Nodes[0].Inputs = append(flow.Nodes[0].Inputs, "missing5527")
	if _, _, err := SourceFillUses(origin, 8); err == nil {
		t.Fatal("dangling recorded reference hidden")
	}
	t.Log("5527: two root/hole uses; return path contains a control edge; one-row truncation explicit; cycles terminate; missing node reported")
}
