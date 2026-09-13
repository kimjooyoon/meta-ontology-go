package main

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func marshalFieldGraphDump(source []byte, ir semantic.IR) ([]byte, error) {
	dump := newGraphDump(source, ir)
	dump.Lowering = graphStatus{Status: "available", Reason: "explicit entity-fields V1 lowering completed with a bounded context"}
	payload, err := json.Marshal(dump)
	if err != nil {
		return nil, err
	}
	if len(payload)+1 > maxGraphDumpBytes {
		return nil, errGraphDumpLimit
	}
	return append(payload, '\n'), nil
}
