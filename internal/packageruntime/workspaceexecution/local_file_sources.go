package workspaceexecution

import (
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

type localFillKey struct{ activity, body string }
type localFillSelection struct {
	origin    *bodycodegen.SourceFillOrigin
	ambiguous bool
}

func workspaceLocalFillSelections(composition bodyexecution.Composition, fills []BodyFillStep) map[localFillKey]localFillSelection {
	result := map[localFillKey]localFillSelection{}
	add := func(generation bodycodegen.Result) {
		fill := generation.Report.BodyFill
		if fill == nil || fill.SourceOrigin == nil {
			return
		}
		o := fill.SourceOrigin
		key := localFillKey{o.ActivityID, o.SelectedBodySHA256}
		prior := result[key]
		if prior.origin != nil && !reflect.DeepEqual(prior.origin, o) {
			prior.ambiguous = true
		}
		if prior.origin == nil {
			prior.origin = o
		}
		result[key] = prior
	}
	for _, fill := range fills {
		add(fill.Generation)
	}
	for _, step := range composition.Steps {
		add(step.Generation)
	}
	for _, step := range composition.Preparations {
		add(step.Generation)
	}
	return result
}

func bindWorkspaceLocalFileSources(program Program, runtime *bodyexecution.CompositionRuntime, composition bodyexecution.Composition, fills ...[]BodyFillStep) {
	var steps []BodyFillStep
	for _, group := range fills {
		steps = append(steps, group...)
	}
	selections := workspaceLocalFillSelections(composition, steps)
	bindings := make([]bodyexecution.CompositionLocalFileBinding, 0, len(runtime.LocalSites)+len(runtime.LocalUseSites))
	for index, site := range runtime.LocalSites {
		input, found := program.fillInputs[site.ActivityID]
		selection := localSelectionForSource(selections, site.ActivityID, site.Source.Value)
		bindings = append(bindings, bodyexecution.CompositionLocalFileBinding{Kind: "write", SiteIndex: index,
			Declaration:  localFileLocation(input, found, site.Source.Declaration, selection),
			Value:        localFileLocation(input, found, site.Source.Value, selection),
			Construction: localFileConstruction(input, found, site.Source.Value, selection)})
	}
	for index, site := range runtime.LocalUseSites {
		input, found := program.fillInputs[site.ActivityID]
		selection := localSelectionForSource(selections, site.ActivityID, site.Source)
		bindings = append(bindings, bodyexecution.CompositionLocalFileBinding{Kind: "use", SiteIndex: index,
			Declaration:  localFileLocation(input, found, site.Declaration, selection),
			Value:        localFileLocation(input, found, site.Source, selection),
			Construction: localFileConstruction(input, found, site.Source, selection)})
	}
	bodyexecution.RecordLocalFileSources(runtime, bindings)
	bindWorkspaceFieldFileSources(program, runtime, composition, steps)
}

func localSelectionForSource(selections map[localFillKey]localFillSelection, activity string, source *bodycodegen.BodyExpressionSource) localFillSelection {
	if source == nil {
		return localFillSelection{}
	}
	return selections[localFillKey{activity, source.BodySHA256}]
}
