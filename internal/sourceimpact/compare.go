package sourceimpact

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func Compare(ctx context.Context, beforeName string, before []byte, afterName string, after []byte) (Report, error) {
	if ctx == nil {
		return Report{}, fmt.Errorf("source impact requires a context")
	}
	left, err := inspect(ctx, beforeName, before, "before")
	if err != nil {
		return Report{}, err
	}
	right, err := inspect(ctx, afterName, after, "after")
	if err != nil {
		return Report{}, err
	}
	r := Report{Schema: Schema, Status: "COMPLETE_STATIC", BeforeSourceSHA256: semantic.StableHash(before),
		AfterSourceSHA256: semantic.StableHash(after), Changes: []Change{}, Affected: []Affected{}, ScopeChanges: compareScope(left.scope, right.scope),
		Unknowns: append(left.unknowns, right.unknowns...), BeforeDeclarations: len(left.nodes), AfterDeclarations: len(right.nodes),
		Scope: "single-file source comparison; potential dependencies in each revision; one nearest-change explanation per affected declaration and revision; body reachability, correctness and saved-result validity require separate execution or replay"}
	ids := unionIDs(left, right)
	for _, id := range ids {
		change := compareNode(id, left, right)
		if len(change.Facets) > 0 {
			r.Changes = append(r.Changes, change)
		}
	}
	if len(r.Unknowns) > 0 {
		r.Status = "PARTIAL_STATIC"
	}
	if err := traceImpacts(ctx, &r, left, right); err != nil {
		return Report{}, err
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return Report{}, err
	}
	r.ReportSHA256 = semantic.StableHash(raw)
	return r, ctx.Err()
}

func unionIDs(left, right snapshot) []string {
	ids := make([]string, 0, len(left.nodes)+len(right.nodes))
	for id := range left.nodes {
		ids = append(ids, id)
	}
	for id := range right.nodes {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return slices.Compact(ids)
}

func compareNode(id string, left, right snapshot) Change {
	before, beforeExists := left.nodes[id]
	after, afterExists := right.nodes[id]
	r := Change{ID: id, BeforeName: before.Name, AfterName: after.Name, Facets: []string{}}
	if !beforeExists {
		r.Facets = append(r.Facets, "added")
		return r
	}
	if !afterExists {
		r.Facets = append(r.Facets, "removed")
		return r
	}
	for _, facet := range []string{"source_scope", "identity_role", "presentation", "fields", "body", "input_sequence", "input_names", "assembly", "dependencies"} {
		if left.facets[id][facet] != right.facets[id][facet] {
			r.Facets = append(r.Facets, facet)
		}
	}
	return r
}

func encoded(value any) string { raw, _ := json.Marshal(value); return string(raw) }

func compareScope(left, right map[string]string) []ScopeChange {
	keys := []string{}
	for key := range left {
		keys = append(keys, key)
	}
	for key := range right {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	keys = slices.Compact(keys)
	changes := []ScopeChange{}
	for _, key := range keys {
		if left[key] != right[key] {
			changes = append(changes, ScopeChange{Field: key, BeforeSHA256: semantic.StableHash([]byte(left[key])), AfterSHA256: semantic.StableHash([]byte(right[key]))})
		}
	}
	return changes
}
