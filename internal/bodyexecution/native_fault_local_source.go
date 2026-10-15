package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

// RecordedFaultLocalRHS joins a compound operation to its retained local RHS.
// The result locates the selected Gooo carrier, not an original workspace file.
func RecordedFaultLocalRHS(runtime CompositionRuntime, fault CompositionFaultSite) *bodycodegen.BodyExpressionSource {
	if fault.Variable == "" || fault.Assignment != "/=" || fault.ProjectionSHA256 == "" {
		return nil
	}
	var selected *bodycodegen.BodyExpressionSource
	matches := 0
	for _, local := range runtime.LocalSites {
		if local.RootActivityID != fault.RootActivityID || local.ActivityID != fault.ActivityID || local.Name != fault.Variable ||
			local.Operation != fault.Assignment || local.ProjectionSHA256 != fault.ProjectionSHA256 ||
			local.Start <= fault.Start || local.End != fault.End {
			continue
		}
		matches++
		source := local.Source.Value
		if source != nil && source.Status == "MAPPED" && source.View == "selected_gooo_source" &&
			source.SourceSHA256 != "" && source.SourceSHA256 == runtime.SelectedSourceSHA256 && source.Start >= 0 && source.End > source.Start {
			selected = source
		}
	}
	if matches != 1 {
		return nil
	}
	return selected
}
