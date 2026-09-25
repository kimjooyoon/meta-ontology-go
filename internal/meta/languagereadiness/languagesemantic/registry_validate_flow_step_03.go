package languagesemantic

import (
	"fmt"
)

func registry_validateFlowStep03(flow *registry_validateFlowState) {
	if len(flow.slot00.Cases) != FixedTotal {
		{
			flow.result0 = fmt.Errorf("registry has %d cases, want %d", len(flow.slot00.Cases), FixedTotal)
			flow.done = true
			return
		}
	}
}
