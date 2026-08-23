package languagesemantic

import (
	fmt "fmt"
	filepath "path/filepath"
	strings "strings"
)

func registryValidateDefinitionStep(flow *registry_validateFlowState, definition Definition) bool {
	if strings.TrimSpace(definition.ID) == "" {
		{
			flow.result0 = fmt.Errorf("registry contains an empty case id")
			flow.done = true
			return true
		}
	}
	if _, exists := flow.slot01[definition.ID]; exists {
		{
			flow.result0 = fmt.Errorf("registry contains duplicate case %q", definition.ID)
			flow.done = true
			return true
		}
	}
	flow.slot01[definition.ID] = struct{}{}
	if definition.ProofChoice != "FOUNDATION" && definition.ProofChoice != "COHERENCE" && definition.ProofChoice != "REGRESSION" {
		{
			flow.result0 = fmt.Errorf("case %q has invalid proof choice %q", definition.ID, definition.ProofChoice)
			flow.done = true
			return true
		}
	}
	if strings.TrimSpace(definition.MetaOperation) == "" {
		{
			flow.result0 = fmt.Errorf("case %q has no meta operation", definition.ID)
			flow.done = true
			return true
		}
	}
	switch definition.Kind {
	case CaseSource:
		flow.slot02++
		clean := filepath.ToSlash(filepath.Clean(definition.Path))
		if clean == "." || strings.HasPrefix(clean, "../") || filepath.Ext(clean) != ".gooo" {
			{
				flow.result0 = fmt.Errorf("source case %q has invalid path %q", definition.ID, definition.Path)
				flow.done = true
				return true
			}
		}
	case CaseLaw:
		flow.slot03++
		if _, ok := knownLaws[definition.Law]; !ok {
			{
				flow.result0 = fmt.Errorf("law case %q has unknown law %q", definition.ID, definition.Law)
				flow.done = true
				return true
			}
		}
	case CaseUpstreamRejection:
		flow.slot04++
		if definition.UpstreamCase == "" {
			{
				flow.result0 = fmt.Errorf("upstream rejection %q has no upstream case", definition.ID)
				flow.done = true
				return true
			}
		}
	default:
		{
			flow.result0 = fmt.Errorf("case %q has unknown kind %q", definition.ID, definition.Kind)
			flow.done = true
			return true
		}
	}
	return false
}
