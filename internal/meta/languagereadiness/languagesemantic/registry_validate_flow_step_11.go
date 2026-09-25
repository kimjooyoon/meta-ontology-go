package languagesemantic

import (
	"fmt"
)

func registry_validateFlowStep11(flow *registry_validateFlowState) {
	for _, name := range flow.slot05 {
		count := 0
		for _, definition := range flow.slot00.Cases {
			if definition.Law == name {
				count++
			}
		}
		if count != 1 {
			{
				flow.result0 = fmt.Errorf("registry law %q occurs %d times, want 1", name, count)
				flow.done = true
				return
			}
		}
	}
}
