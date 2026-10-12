package sourceimpact

import (
	"context"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func traceImpacts(ctx context.Context, report *Report, before, after snapshot) error {
	rows := map[string]*Affected{}
	for _, side := range []struct {
		name string
		s    snapshot
	}{{"before", before}, {"after", after}} {
		adjacent := map[string][]Edge{}
		for _, edge := range side.s.edges {
			adjacent[edge.Dependency] = append(adjacent[edge.Dependency], edge)
		}
		causes := map[string]Cause{}
		queue := []string{}
		for _, change := range report.Changes {
			if _, exists := side.s.nodes[change.ID]; exists {
				causes[change.ID] = Cause{Revision: side.name, ChangedID: change.ID}
				queue = append(queue, change.ID)
			}
		}
		for cursor := 0; cursor < len(queue); cursor++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			id := queue[cursor]
			cause := causes[id]
			row := rows[id]
			if row == nil {
				value := affectedRow(id, before, after)
				row = &value
				rows[id] = row
			}
			row.Causes = append(row.Causes, cause)
			for _, edge := range adjacent[id] {
				if _, seen := causes[edge.Dependent]; seen {
					continue
				}
				via := edge
				causes[edge.Dependent] = Cause{Revision: side.name, ChangedID: cause.ChangedID, Depth: cause.Depth + 1, Via: &via}
				queue = append(queue, edge.Dependent)
			}
		}
	}
	ids := make([]string, 0, len(rows))
	for id := range rows {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		report.Affected = append(report.Affected, *rows[id])
	}
	return nil
}

func affectedRow(id string, before, after snapshot) Affected {
	left := before.nodes[id]
	right, present := after.nodes[id]
	name := right.Name
	if !present {
		name = left.Name
	}
	return Affected{ID: id, Name: name, PresentAfter: present, Causes: []Cause{}, BeforeObligations: obligations(left), AfterObligations: obligations(right), State: "RECHECK_REQUIRED"}
}

func obligations(node semantic.Node) *Obligations {
	s := node.Assembly
	if s == nil {
		return nil
	}
	r := &Obligations{Cases: len(s.Cases) + len(s.ValueCases), Conditions: len(s.ConditionCases), Holdouts: len(s.HoldoutCases) + len(s.ValueHoldoutCases), MaxAttempts: s.MaxAttempts}
	if s.Observation != nil {
		r.ObservationInputs = len(s.Observation.Inputs) + len(s.Observation.ValueInputs)
		r.Oracle = s.Observation.OracleActivity
	}
	if s.Search != nil {
		r.GeneratedCandidateLimit = s.Search.MaxCandidates
	}
	if s.FillPlan != nil {
		r.FillCandidates = len(s.FillPlan.Candidates)
		if s.FillPlan.Generation != nil {
			r.GeneratedCandidateLimit = s.FillPlan.Generation.MaxCandidates
		}
	}
	return r
}
