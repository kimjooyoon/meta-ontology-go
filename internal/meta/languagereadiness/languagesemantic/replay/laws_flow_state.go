package replay

import (
	semantic "github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type lawsFlowState struct {
	done    bool
	slot00  string
	slot01  semantic.IR
	slot02  semantic.IR
	slot03  error
	slot04  []semantic.Node
	slot05  []semantic.Fact
	slot06  semantic.IR
	slot07  semantic.Node
	slot08  semantic.IR
	slot09  semantic.IR
	slot10  semantic.Fact
	slot11  semantic.IR
	result0 LawObservation
	result1 error
}
