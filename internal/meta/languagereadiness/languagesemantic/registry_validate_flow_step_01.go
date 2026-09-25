package languagesemantic

import (
	"fmt"
)

func registry_validateFlowStep01(flow *registry_validateFlowState) {
	if flow.slot00.Schema != RegistrySchema {
		{
			flow.result0 = fmt.Errorf("registry schema %q is not %q", flow.slot00.Schema, RegistrySchema)
			flow.done = true
			return
		}
	}
}
