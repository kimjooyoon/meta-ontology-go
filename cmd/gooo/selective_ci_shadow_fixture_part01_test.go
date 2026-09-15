package main

import (
	"errors"
	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
)

type shadowMapReader map[string][]byte

func (r shadowMapReader) ReadFile(name string) ([]byte, error) {
	data, ok := r[name]
	if !ok {
		return nil, errors.New("missing fixture file")
	}
	return append([]byte{}, data...), nil
}

type shadowFixture struct {
	files         map[string][]byte
	base          analyzersci.Snapshot
	head          analyzersci.Snapshot
	planInput     plannersci.Input
	proofInput    proofsci.Input
	laneInput     lanesci.Input
	entityID      string
	otherID       string
	commandID     string
	commandCPU    uint64
	commandMemory uint64
	sourceBase    string
}
