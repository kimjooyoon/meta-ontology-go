package fullsoundness

import (
	"encoding/json"
	"errors"
)

var errProjectionOverflow = errors.New("decision projection overflow")

type decisionProjection struct {
	Status              Decision         `json:"status"`
	Reason              Reason           `json:"reason"`
	Semantic            decisionSemantic `json:"semantic"`
	Resource            decisionResource `json:"resource"`
	ExecutionAuthorized bool             `json:"execution_authorized"`
	CIAuthorized        bool             `json:"ci_authorized"`
	CanonicalDigest     string           `json:"canonical_digest"`
}

type decisionSemantic struct {
	FullCount          uint64   `json:"full_count"`
	SelectedCount      uint64   `json:"selected_count"`
	FullPassCount      uint64   `json:"full_pass_count"`
	FullFailCount      uint64   `json:"full_fail_count"`
	SelectedPassCount  uint64   `json:"selected_pass_count"`
	SelectedFailCount  uint64   `json:"selected_fail_count"`
	FullFailureIDs     []string `json:"full_failure_ids"`
	SelectedFailureIDs []string `json:"selected_failure_ids"`
	OmittedIDs         []string `json:"omitted_ids"`
}

type decisionResource struct {
	CPUFullNS              int64         `json:"cpu_full_ns"`
	CPUSelectedNS          int64         `json:"cpu_selected_ns"`
	CPUSavedNS             int64         `json:"cpu_saved_ns"`
	FullMaxRSSBytes        int64         `json:"full_max_rss_bytes"`
	SelectedMaxRSSBytes    int64         `json:"selected_max_rss_bytes"`
	RSSSavedBytes          int64         `json:"rss_saved_bytes"`
	FullReadBytes          int64         `json:"full_read_bytes"`
	SelectedReadBytes      int64         `json:"selected_read_bytes"`
	ReadSavedBytes         int64         `json:"read_saved_bytes"`
	FullWriteBytes         int64         `json:"full_write_bytes"`
	SelectedWriteBytes     int64         `json:"selected_write_bytes"`
	WriteSavedBytes        int64         `json:"write_saved_bytes"`
	FullCPUUtilization     Utilization   `json:"full_cpu_utilization"`
	SelectedCPUUtilization Utilization   `json:"selected_cpu_utilization"`
	ResourceClass          ResourceClass `json:"resource_class"`
}

func (output Output) DecisionCanonicalJSON() ([]byte, error) {
	projection, err := output.decisionProjection()
	if err != nil {
		return nil, err
	}
	return json.Marshal(projection)
}

func (output Output) DecisionStableDigest() (string, error) {
	data, err := output.DecisionCanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}

func (output Output) decisionProjection() (decisionProjection, error) {
	output = normalizeOutput(output)
	semantic, err := semanticProjection(output)
	if err != nil {
		return decisionProjection{}, err
	}
	resource, err := resourceProjection(output.ResourceVector)
	if err != nil {
		return decisionProjection{}, err
	}
	return decisionProjection{Status: output.Decision, Reason: output.Reason, Semantic: semantic, Resource: resource}, nil
}

func semanticProjection(output Output) (decisionSemantic, error) {
	if !output.SemanticEvaluated {
		return emptyDecisionSemantic(), nil
	}
	fullFail := uint64(len(output.FullFailureCommandIDs))
	selectedFail := uint64(len(output.SelectedFailureCommandIDs))
	fullPass, ok := subtractUint64(output.CommandCount, fullFail)
	if !ok {
		return decisionSemantic{}, errProjectionOverflow
	}
	selectedPass, ok := subtractUint64(output.SelectedCommandCount, selectedFail)
	if !ok {
		return decisionSemantic{}, errProjectionOverflow
	}
	return decisionSemantic{FullCount: output.CommandCount, SelectedCount: output.SelectedCommandCount, FullPassCount: fullPass, FullFailCount: fullFail, SelectedPassCount: selectedPass, SelectedFailCount: selectedFail, FullFailureIDs: output.FullFailureCommandIDs, SelectedFailureIDs: output.SelectedFailureCommandIDs, OmittedIDs: output.OmittedCommandIDs}, nil
}

func emptyDecisionSemantic() decisionSemantic {
	return decisionSemantic{FullFailureIDs: []string{}, SelectedFailureIDs: []string{}, OmittedIDs: []string{}}
}

func resourceProjection(vector *ResourceVector) (decisionResource, error) {
	if vector == nil {
		return decisionResource{ResourceClass: ResourceNotComputed}, nil
	}
	full := vector.Full
	selected := vector.Selected
	cpuSaved, ok := subtractInt64(full.CPUCoreNS, selected.CPUCoreNS)
	if !ok {
		return decisionResource{}, errProjectionOverflow
	}
	rssSaved, ok := subtractInt64(full.PeakRSSBytes, selected.PeakRSSBytes)
	if !ok {
		return decisionResource{}, errProjectionOverflow
	}
	readSaved, ok := subtractInt64(full.ReadBytes, selected.ReadBytes)
	if !ok {
		return decisionResource{}, errProjectionOverflow
	}
	writeSaved, ok := subtractInt64(full.WriteBytes, selected.WriteBytes)
	if !ok {
		return decisionResource{}, errProjectionOverflow
	}
	return decisionResource{CPUFullNS: full.CPUCoreNS, CPUSelectedNS: selected.CPUCoreNS, CPUSavedNS: cpuSaved, FullMaxRSSBytes: full.PeakRSSBytes, SelectedMaxRSSBytes: selected.PeakRSSBytes, RSSSavedBytes: rssSaved, FullReadBytes: full.ReadBytes, SelectedReadBytes: selected.ReadBytes, ReadSavedBytes: readSaved, FullWriteBytes: full.WriteBytes, SelectedWriteBytes: selected.WriteBytes, WriteSavedBytes: writeSaved, FullCPUUtilization: full.Utilization, SelectedCPUUtilization: selected.Utilization, ResourceClass: vector.Class}, nil
}

func subtractUint64(left, right uint64) (uint64, bool) {
	if left < right {
		return 0, false
	}
	return left - right, true
}
