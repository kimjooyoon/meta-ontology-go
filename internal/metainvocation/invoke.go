package metainvocation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func Invoke(program Program, entry string, rawInput []byte) (Report, error) {
	if err := validateProgram(program); err != nil {
		return Report{}, err
	}
	bound, ok := program.Operations[entry]
	if !ok || bound.Program != operationPlan {
		return Report{}, fmt.Errorf("entry %q is not bound to %q", entry, operationPlan)
	}
	if program.Operations["VerifyCIPlan"].Program != operationVerify {
		return Report{}, fmt.Errorf("verification operation is not bound")
	}
	changeSet, decodeReason := decodeChangeSet(rawInput)
	inputDigest := bytesDigest(rawInput)
	if decodeReason != "" {
		return buildReport(program, entry, changeSet.CaseID, inputDigest, DecisionClosed, ResolutionExact, nil, nil, decodeReason), nil
	}
	if reason, file := validateChangeSet(changeSet); reason != "" {
		return buildReport(program, entry, changeSet.CaseID, inputDigest, DecisionClosed, ResolutionExact, nil, nil, reasonWithFile(reason, file)), nil
	}
	checks, unknowns := selectChecks(program, changeSet)
	if len(unknowns) != 0 {
		return buildReport(program, entry, changeSet.CaseID, inputDigest, DecisionUnknown, ResolutionLower, nil, unknowns, ""), nil
	}
	return buildReport(program, entry, changeSet.CaseID, inputDigest, DecisionPass, ResolutionExact, checks, nil, ""), nil
}

func decodeChangeSet(raw []byte) (ChangeSet, string) {
	changeSet := ChangeSet{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&changeSet); err != nil {
		return changeSet, "INPUT_DECODE:decode-change-set:" + err.Error()
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return changeSet, "INPUT_DECODE:reject-trailing-content"
	}
	return changeSet, ""
}
