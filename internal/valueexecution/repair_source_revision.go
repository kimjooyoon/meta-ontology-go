package valueexecution

// ProposeSourceRevisionFromHandoff derives a source-revision candidate from a
// validated repair handoff. The handoff supplies the trigger reason; callers
// still provide the exact source and replacement and must evaluate the result
// independently before any later explicit execution.
func ProposeSourceRevisionFromHandoff(filename string, source []byte, handoff RepairHandoff, request SourceRevisionRequest) ([]byte, SourceRevision, error) {
	if err := ValidateRepairHandoff(handoff); err != nil {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "validate-repair-handoff", err.Error())
	}
	if request.TriggerReason != "" && request.TriggerReason != handoff.TriggerReason {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "reject-trigger-reason-mismatch", request.TriggerReason)
	}
	request.TriggerReason = handoff.TriggerReason
	candidate, revision, err := ProposeSourceRevision(filename, source, request)
	if err != nil {
		return nil, SourceRevision{}, err
	}
	revision.RepairHandoffDigest = DigestRepairHandoff(handoff)
	revision.RepairCandidateID = handoff.CandidateID
	revision.CandidateID = ""
	revision.CandidateID = "gooo://source-revision/" + digestValue(revision)[len("sha256:"):len("sha256:")+16]
	return candidate, revision, nil
}

func DigestRepairHandoff(handoff RepairHandoff) string {
	return digestValue(handoff)
}
