package workspaceexecution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// Package execute can materialize fills before constructing the final graph.
// Those origins belong to BodyFills rather than the final composition's steps.
func bindWorkspaceFillConditionSources(runtime *bodyexecution.CompositionRuntime, fills []BodyFillStep) {
	for i := range runtime.ConditionSites {
		site := &runtime.ConditionSites[i]
		if site.Source == nil || site.Source.Status != "MAPPED" {
			continue
		}
		condition := bodycodegen.BodyConditionOrigin{ActivityID: site.ActivityID,
			BodySHA256: site.Source.BodySHA256, Start: site.Source.DecodedStart, End: site.Source.DecodedEnd}
		for _, step := range fills {
			fill := step.Generation.Report.BodyFill
			if fill == nil || fill.SourceOrigin == nil {
				continue
			}
			if link := fill.SourceOrigin.ConditionOrigin(condition); link != nil {
				site.Source.Construction = link
				break
			}
		}
	}
}
