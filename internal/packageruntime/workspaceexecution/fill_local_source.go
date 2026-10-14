package workspaceexecution

import (
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// Workspace fills precede the final composition. Keep their assembly input
// attached to local writes without changing an existing lexical site identity.
func bindWorkspaceFillLocalSources(runtime *bodyexecution.CompositionRuntime, fills []BodyFillStep) {
	for index := range runtime.LocalSites {
		site := &runtime.LocalSites[index]
		value := site.Source.Value
		if value == nil || value.Status != "MAPPED" || site.Source.ConstructionReason != "" {
			continue
		}
		local := bodycodegen.BodyLocalOrigin{ActivityID: site.ActivityID,
			BodySHA256: value.BodySHA256, Start: value.DecodedStart, End: value.DecodedEnd}
		for _, step := range fills {
			fill := step.Generation.Report.BodyFill
			if fill == nil || fill.SourceOrigin == nil {
				continue
			}
			link := fill.SourceOrigin.LocalOrigin(local)
			if link == nil {
				continue
			}
			if site.Source.Construction != nil && !reflect.DeepEqual(site.Source.Construction, link) {
				site.Source.Construction, site.Source.ConstructionReason = nil, "AMBIGUOUS_ASSEMBLY_INPUT"
				break
			}
			site.Source.Construction = link
		}
	}
}
