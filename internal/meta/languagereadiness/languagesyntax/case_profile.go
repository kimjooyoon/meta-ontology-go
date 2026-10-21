package languagesyntax

import (
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

// ValidProfileSelection keeps explicit profile identities separate from legacy
// flag routes. A request cannot silently combine two parser configurations.
func ValidProfileSelection(definition CaseDefinition) bool {
	if len(definition.ContextSources) > 0 && (!definition.OrderedCollections || definition.EntityFields || definition.ImplicitActivityPorts || definition.EntityFieldsProfile != "") {
		return false
	}
	if definition.EntityFieldsProfile == "" {
		return true
	}
	_, known := replay.EntityFieldsProfileSupport(definition.EntityFieldsProfile)
	return known && !definition.EntityFields && !definition.ImplicitActivityPorts && !definition.OrderedCollections
}

func ReplayCase(repository fs.FS, definition CaseDefinition) replay.Result {
	if !ValidProfileSelection(definition) {
		return replay.Result{ObservedDecision: replay.DecisionClosed, Diagnostics: []string{"registry.profile-mismatch"}}
	}
	if definition.EntityFieldsProfile != "" {
		return replay.ExecuteWithEntityFieldsProfile(repository, definition.Path, definition.Kind,
			definition.ExpectedDiagnostic, definition.EntityFieldsProfile)
	}
	if definition.OrderedCollections {
		if len(definition.ContextSources) > 0 {
			return replay.ExecuteWithOrderedCollectionsAndContext(repository, definition.Path, definition.ContextSources, definition.Kind, definition.ExpectedDiagnostic)
		}
		return replay.ExecuteWithOrderedCollections(repository, definition.Path, definition.Kind, definition.ExpectedDiagnostic)
	}
	if definition.ImplicitActivityPorts {
		return replay.ExecuteWithImplicitActivityPorts(repository, definition.Path, definition.Kind, definition.ExpectedDiagnostic)
	}
	if definition.EntityFields {
		return replay.ExecuteWithEntityFieldsSupport(repository, definition.Path, definition.Kind, definition.ExpectedDiagnostic)
	}
	return replay.Execute(repository, definition.Path, definition.Kind, definition.ExpectedDiagnostic)
}
