package languagesemantic

import (
	fmt "fmt"
	strings "strings"
)

func registryValidateDefinitionGuard(flow *registry_validateFlowState, definition Definition) bool {
	if strings.TrimSpace(definition.ID) == "" {
		{
			flow.result0 = fmt.Errorf("registry contains an empty case id")
			flow.done = true
			return true
		}
	}
	return false
}
