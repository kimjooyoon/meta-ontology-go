package languagesemantic

import (
	"fmt"
)

func registry_validateFlowStep07(flow *registry_validateFlowState) {
	if flow.slot02 != expectedSources || flow.slot03 != expectedLaws || flow.slot04 != expectedRejections {
		{
			flow.result0 = fmt.Errorf("registry topology is sources=%d laws=%d rejections=%d, want %d/%d/%d", flow.slot02, flow.slot03, flow.slot04, expectedSources, expectedLaws, expectedRejections)
			flow.done = true
			return
		}
	}
}
