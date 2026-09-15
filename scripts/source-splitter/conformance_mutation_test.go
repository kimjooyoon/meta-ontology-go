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
		for index := range value.Candidates {
			if !bytes.Contains(value.Candidates[index].Data, []byte("fmt.Sprint")) {
				continue
			}
			value.Candidates[index].Data = bytes.Replace(value.Candidates[index].Data, []byte(`"fmt"`), []byte(`"strings"`), 1)
			break
		}
	case "ORDER":
		indexes := initializationCandidateIndexes(*value)
		if len(indexes) >= 2 {
			first, second := indexes[0], indexes[1]
			value.Candidates[first].Data, value.Candidates[second].Data = value.Candidates[second].Data, value.Candidates[first].Data
			value.Candidates[first].DeclarationOrder, value.Candidates[second].DeclarationOrder =
				value.Candidates[second].DeclarationOrder, value.Candidates[first].DeclarationOrder
		}
	case "DUPLICATE_BODY":
		indexes := initializationCandidateIndexes(*value)
		if len(indexes) >= 2 {
			index := indexes[1]
			value.Candidates[index].Data = bytes.Replace(value.Candidates[index].Data,
				[]byte("println(1)"), []byte("println(2)"), 1)
		}
	case "DUPLICATE_ORDINAL":
		indexes := initializationCandidateIndexes(*value)
		if len(indexes) >= 2 {
			value.Candidates[indexes[1]].DeclarationOrder[0].Ordinal = value.Candidates[indexes[0]].DeclarationOrder[0].Ordinal
		}
	case "ORDINAL_EXCHANGE":
		indexes := initializationCandidateIndexes(*value)
		if len(indexes) >= 2 {
			first, second := indexes[0], indexes[1]
			value.Candidates[first].DeclarationOrder[0].Ordinal, value.Candidates[second].DeclarationOrder[0].Ordinal =
				value.Candidates[second].DeclarationOrder[0].Ordinal, value.Candidates[first].DeclarationOrder[0].Ordinal
		}
	case "PACKAGE":
		value.Candidates[0].Data = bytes.Replace(value.Candidates[0].Data, []byte("package fixture"), []byte("package other"), 1)
	case "EVIDENCE_MISSING":
		value.EvidenceComplete, value.Write.Complete = false, false
	}
}

func initializationCandidateIndexes(value conformance.SplitGoEvidence) []int {
	indexes := make([]int, 0)
	for index, candidate := range value.Candidates {
		if bytes.Contains(candidate.Data, []byte("func init()")) {
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func indicatorDecision(report conformance.Report, id string) conformance.Decision {
	for _, indicator := range report.Indicators {
		if indicator.ID == id {
			return indicator.Decision
		}
	}
	return ""
}
