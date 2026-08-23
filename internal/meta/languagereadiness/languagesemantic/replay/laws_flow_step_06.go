package replay

import (
	"fmt"
)

func lawsFlowStep06(flow *lawsFlowState) {
	if len(flow.slot05) == 0 {
		{
			flow.result0, flow.result1 = LawObservation{}, fmt.Errorf("semantic law anchor %s has no deterministic facts", flow.slot00)
			flow.done = true
			return
		}
	}
}
