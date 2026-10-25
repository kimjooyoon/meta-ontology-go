package languagesyntax

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func expectedCaseSources() []CaseSourceSet {
	return []CaseSourceSet{{CaseID: "explicit-consumer-continuation", Members: []string{
		"internal/meta/executioncontinuation/policy/source.gooo",
		"internal/meta/executioncontinuation/policy/source_recovery.gooo",
	}}}
}

func caseSourcesFor(registry Registry, id string) []string {
	for _, group := range registry.CaseSources {
		if group.CaseID == id {
			return slices.Clone(group.Members)
		}
	}
	return nil
}

func cloneCaseSources(groups []CaseSourceSet) []CaseSourceSet {
	cloned := slices.Clone(groups)
	for i := range cloned {
		cloned[i].Members = slices.Clone(cloned[i].Members)
	}
	return cloned
}

func validateCaseSources(registry Registry) error {
	cases := make(map[string]CaseDefinition, len(registry.Cases))
	for _, definition := range registry.Cases {
		cases[definition.ID] = definition
	}
	seen := map[string]bool{}
	for _, group := range registry.CaseSources {
		definition, found := cases[group.CaseID]
		if !found || seen[group.CaseID] {
			return fmt.Errorf("source set needs one known, unique case")
		}
		seen[group.CaseID] = true
		if err := validCaseMembers(definition, group.Members); err != nil {
			return err
		}
	}
	return nil
}

func validCaseMembers(definition CaseDefinition, names []string) error {
	if definition.Kind != KindValid || len(names) < 2 || len(names) > sourcebundle.MaxSources || names[0] != definition.Path {
		return fmt.Errorf("source set requires a valid case and its primary path followed by explicit members")
	}
	seen := map[string]bool{}
	for _, name := range names {
		key := strings.ToLower(name)
		if !fs.ValidPath(name) || !strings.HasSuffix(name, ".gooo") || strings.ContainsAny(name, "\\\x00:") || seen[key] {
			return fmt.Errorf("source set member is invalid or duplicated")
		}
		seen[key] = true
	}
	return nil
}

// ReplayCaseSources uses the case's exact profile and an explicit source set.
// Empty membership retains the original single-file evidence representation.
func ReplayCaseSources(repository fs.FS, definition CaseDefinition, names []string) replay.Result {
	if len(names) == 0 {
		return ReplayCase(repository, definition)
	}
	if !ValidProfileSelection(definition) || validCaseMembers(definition, names) != nil {
		return replay.Result{ObservedDecision: replay.DecisionUnknown, Diagnostics: []string{"registry.case-sources-mismatch"}}
	}
	support := syntax.CurrentEntityFieldsSupport()
	switch {
	case definition.EntityFieldsProfile != "":
		support, _ = replay.EntityFieldsProfileSupport(definition.EntityFieldsProfile)
	case definition.OrderedCollections:
		support = syntax.EntityFieldsV5Support()
	case definition.EntityFields && !definition.ImplicitActivityPorts:
		support = syntax.EntityFieldsV1Support()
	}
	result := replay.ExecuteSourceSet(repository, names, support, definition.ImplicitActivityPorts && !definition.OrderedCollections)
	if definition.EntityFieldsProfile != "" {
		result.EntityFieldsProfile, result.ProfileDigest = support.Profile.ID, "sha256:"+support.Profile.Digest
	}
	return result
}
