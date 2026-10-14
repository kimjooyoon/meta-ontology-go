package workspaceexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"

func bindWorkspaceFieldFileSources(program Program, runtime *bodyexecution.CompositionRuntime, composition bodyexecution.Composition, fills []BodyFillStep) {
	if runtime.FieldSources == nil {
		return
	}
	selections := workspaceLocalFillSelections(composition, fills)
	bindings := make([]bodyexecution.CompositionFieldFileBinding, 0, len(runtime.FieldSources.Bindings))
	for _, site := range runtime.FieldSources.Bindings {
		locations := runtime.FieldSites
		if site.Kind == "use" {
			locations = runtime.FieldUseSites
		}
		identity := runtime.FieldBindings[locations[site.SiteIndex].Binding]
		input, found := program.fillInputs[identity.ActivityID]
		selection := localSelectionForSource(selections, identity.ActivityID, site.Source.Value)
		binding := bodyexecution.CompositionFieldFileBinding{Kind: site.Kind, SiteIndex: site.SiteIndex,
			Target: localFileLocation(input, found, site.Source.Target, selection), Value: localFileLocation(input, found, site.Source.Value, selection),
			Construction: fileExpressionConstruction(input, found, site.Source.Value, selection, "field_expression")}
		if found && site.Kind == "write" {
			field := runtime.FieldSites[site.SiteIndex].Field
			if origin := workspaceRecordChoiceOrigin(composition, identity, field, site.Source); origin != nil {
				binding.Choice = workspaceRecordChoiceSource(input, *origin)
				if origin.TargetDecoded != nil {
					binding.Target = recordChoiceTemplateLocation(input, origin.TemplateBodySHA, *origin.TargetDecoded)
				}
			}
		}
		bindings = append(bindings, binding)
	}
	bodyexecution.RecordFieldFileSources(runtime, bindings)
}
