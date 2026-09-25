package languagesemantic

import (
	replay "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"
)

type evaluateFlowState struct {
	done    bool
	slot00  Input
	slot01  Registry
	slot02  []byte
	slot03  error
	slot04  []byte
	slot05  syntaxReceipt
	slot06  string
	slot07  []CaseResult
	slot08  []string
	slot09  []string
	slot10  []string
	slot11  []string
	slot12  map[string]replay.Observation
	slot13  *replay.Observation
	slot14  replay.LawObservation
	slot15  error
	slot16  map[string]syntaxCase
	slot17  Report
	result0 Report
	result1 error
}
