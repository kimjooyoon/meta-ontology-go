package languagesemantic

import (
	"fmt"
	"strings"
)

func registry_validateFlowStep02(flow *registry_validateFlowState) {
	if strings.TrimSpace(flow.slot00.Version) == "" {
		{
			flow.result0 = fmt.Errorf("registry version is empty")
			flow.done = true
			return
		}
	}
}
