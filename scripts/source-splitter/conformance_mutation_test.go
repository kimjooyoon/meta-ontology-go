package main

import (
	"bytes"
	"encoding/json"
	"strings"

	conformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/operationconformance"
)

func cloneConformanceEvidence(value conformance.SplitGoEvidence) conformance.SplitGoEvidence {
	raw, _ := json.Marshal(value)
	var result conformance.SplitGoEvidence
	_ = json.Unmarshal(raw, &result)
	return result
}

func mutateConformanceEvidence(value *conformance.SplitGoEvidence, mutation string) {
	switch mutation {
	case "NONE":
	case "DIRECT_WRITE":
		for index := range value.Write.Events {
			if strings.HasPrefix(value.Write.Events[index].Kind, "RENAME_") {
				value.Write.Events[index].Kind = "DIRECT_WRITE"
				break
			}
		}
	case "FILENAME_DOMAIN":
		value.Candidates[1].Path = strings.Replace(value.Candidates[1].Path, "linux_amd64", "windows_amd64", 1)
	case "HEADER":
		value.Candidates[0].Data = bytes.Replace(value.Candidates[0].Data, []byte("linux && amd64"), []byte("windows"), 1)
	case "IMPORT":
		value.Candidates[0].Data = bytes.Replace(value.Candidates[0].Data, []byte(`"fmt"`), []byte(`"strings"`), 1)
	case "ORDER":
		value.Candidates[0].Data, value.Candidates[1].Data = value.Candidates[1].Data, value.Candidates[0].Data
	case "PACKAGE":
		value.Candidates[0].Data = bytes.Replace(value.Candidates[0].Data, []byte("package fixture"), []byte("package other"), 1)
	case "EVIDENCE_MISSING":
		value.EvidenceComplete, value.Write.Complete = false, false
	}
}

func indicatorDecision(report conformance.Report, id string) conformance.Decision {
	for _, indicator := range report.Indicators {
		if indicator.ID == id {
			return indicator.Decision
		}
	}
	return ""
}
