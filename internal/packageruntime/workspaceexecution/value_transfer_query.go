package workspaceexecution

import (
	"context"
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type ProgramValueOrigins struct {
	Schema               string               `json:"schema"`
	SelectedSourceSHA256 string               `json:"selected_source_sha256"`
	RootActivityID       string               `json:"root_activity_id"`
	Analysis             string               `json:"analysis"`
	Truncated            bool                 `json:"truncated"`
	Rows                 []ProgramValueOrigin `json:"rows"`
	Issues               []OriginQueryIssue   `json:"issues"`
	Scope                string               `json:"scope"`
}

type ProgramValueOrigin struct {
	RootID      string              `json:"root_id"`
	RootKind    string              `json:"root_kind"`
	Kind        string              `json:"kind"`
	ActivityID  string              `json:"activity_id"`
	CandidateID string              `json:"candidate_id,omitempty"`
	HoleID      string              `json:"hole_id,omitempty"`
	Path        []ProgramOriginStep `json:"path"`
	Locations   []OriginUseLocation `json:"locations,omitempty"`
}

type ProgramOriginStep struct {
	ActivityID        string `json:"activity_id"`
	BodySHA256        string `json:"body_sha256"`
	NodeID            string `json:"node_id"`
	Kind              string `json:"kind"`
	Via               string `json:"via"`
	Expression        string `json:"expression"`
	Start             int    `json:"start_byte"`
	End               int    `json:"end_byte"`
	SnippetsTruncated bool   `json:"snippets_truncated,omitempty"`
}

type valueTransferActivity struct {
	id, name, body string
	arity          int
	graph          bodycodegen.BodyValueOrigins
	nodes          map[string]bodycodegen.BodyValueOrigin
	calls          map[string]bodycodegen.BodyValueCall
	origins        []*bodycodegen.SourceFillOrigin
}

type valueTransferCatalog struct {
	byID   map[string]*valueTransferActivity
	byName map[string]*valueTransferActivity
	binds  map[string]string
}

// InspectSelectedValueOrigins derives a bounded static view of the selected
// source in a saved receipt. It does not reinterpret observed runtime inputs.
func InspectSelectedValueOrigins(ctx context.Context, result Result, query OriginQuery) (ProgramValueOrigins, error) {
	out := ProgramValueOrigins{Schema: "gooo/selected-program-value-origins/v1", Analysis: "MAPPED",
		SelectedSourceSHA256: result.Composition.SelectedSourceSHA256,
		Scope:                "current static analysis of retained selected source; use-to-origin paths across pure calls and binds; one shortest path per root and origin; branch alternatives are joined; zero program or model execution"}
	if ctx == nil || query.Limit < 1 || query.Limit > 128 {
		return out, fmt.Errorf("value-origin query requires context and limit 1..128")
	}
	catalog, err := prepareValueTransferCatalog(ctx, result)
	if err != nil {
		return out, err
	}
	root := catalog.byID[query.ActivityID]
	if query.ActivityID == "" {
		root = catalog.byName[result.Program.Entry.LoweredName]
	}
	if root == nil {
		return out, fmt.Errorf("value-origin root activity is absent from the selected plan")
	}
	out.RootActivityID = root.id
	walker := valueTransferWalker{ctx: ctx, result: result, query: query, catalog: catalog, out: &out, issues: map[string]bool{}, emitted: map[string]bool{}}
	if len(root.graph.Roots) == 0 {
		walker.issue(root.id, "ROOT_VALUE_ORIGINS_UNAVAILABLE")
	}
	for _, id := range root.graph.Roots {
		if err := walker.walk(root, id); err != nil {
			return out, err
		}
		if out.Truncated {
			break
		}
	}
	if len(out.Issues) > 0 || out.Truncated {
		out.Analysis = "PARTIAL"
	}
	sort.Slice(out.Rows, func(i, j int) bool {
		a, b := out.Rows[i], out.Rows[j]
		if a.RootID != b.RootID {
			return a.RootID < b.RootID
		}
		if a.ActivityID != b.ActivityID {
			return a.ActivityID < b.ActivityID
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.HoleID != b.HoleID {
			return a.HoleID < b.HoleID
		}
		return a.Path[len(a.Path)-1].NodeID < b.Path[len(b.Path)-1].NodeID
	})
	return out, nil
}
