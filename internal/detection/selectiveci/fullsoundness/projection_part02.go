package fullsoundness

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
