package replay

import (
	"fmt"
)

func lawsFlowStep05(flow *lawsFlowState) {
	if len(flow.slot04) == 0 {
		{
			flow.result0, flow.result1 = LawObservation{}, fmt.Errorf("semantic law anchor %s has no nodes", flow.slot00)
			flow.done = true
			return
		}
	}
}
