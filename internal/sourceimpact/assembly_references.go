package sourceimpact

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

func activityBodySurfaces(node semantic.Node) []bodySurface {
	surfaces := []bodySurface{{label: "computes", body: node.ValueProgram}}
	spec := node.Assembly
	if spec == nil {
		return surfaces
	}
	surfaces = append(surfaces, bodySurface{label: "baseline", body: spec.Baseline})
	enclosing := []string{node.ValueProgram, spec.Baseline}
	for _, choice := range spec.Choices {
		if choice.Kind == "field_value" || choice.Kind == "field_update" {
			surfaces = append(surfaces, bodySurface{label: "choice:" + choice.ID, body: "return " + choice.Alternative, enclosing: enclosing})
		}
	}
	if plan := spec.FillPlan; plan != nil {
		for _, candidate := range plan.Candidates {
			for _, fill := range candidate.Fills {
				surfaces = append(surfaces, bodySurface{label: "fill:" + candidate.ID + ":" + fill.HoleID, body: "return " + fill.Expression, enclosing: enclosing})
			}
		}
	}
	return surfaces
}

func inspectAssemblyDependencies(s *snapshot, node semantic.Node, names map[string]string, revision string) {
	id, spec := string(node.ID), node.Assembly
	if spec == nil {
		return
	}
	if observation := spec.Observation; observation != nil && observation.OracleActivity != "" {
		if target, ok := names[observation.OracleActivity]; ok {
			s.addEdge(target, id, "observation_oracle", "observe:oracle")
		} else {
			s.addUnknown(revision, id, "observe:oracle", observation.OracleActivity, "undeclared oracle activity")
		}
	}
	if criteria := spec.Criteria; criteria != nil {
		for _, predicate := range criteria.Predicates {
			surface := "criteria:" + predicate.ID
			if target, ok := names[predicate.Activity]; ok {
				s.addEdge(target, id, "assembly_criterion", surface)
			} else {
				s.addUnknown(revision, id, surface, predicate.Activity, "undeclared criterion activity")
			}
		}
	}
	if plan := spec.FillPlan; plan != nil && plan.Generation != nil {
		s.addUnknown(revision, id, "fill_generation", "", "generated candidate call set has not been materialized")
	}
	if spec.Search != nil {
		s.addUnknown(revision, id, "search", "", "generated candidate call set has not been materialized")
	}
}
