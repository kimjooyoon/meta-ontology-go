package replay

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func ObserveLaws(anchorPath string, input semantic.IR) (LawObservation, error) {
	flow := &lawsFlowState{slot00: anchorPath, slot01: input}
	for _, step := range lawsFlowSteps {
		step(flow)
		if flow.done {
			break
		}
	}
	return flow.result0, flow.result1
}
