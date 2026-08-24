package verticalsliceclosureshadow

import "encoding/json"

const (
	boundaryReasonSatisfied  = "BOUNDARY_EXACT"
	boundaryReasonMissing    = "BOUNDARY_EVIDENCE_UNAVAILABLE"
	boundaryReasonUnknown    = "BOUNDARY_DECISION_UNKNOWN"
	boundaryReasonDecode     = "BOUNDARY_DECODE_FAILED"
	boundaryReasonSchema     = "BOUNDARY_SCHEMA_MISMATCH"
	boundaryReasonFailure    = "BOUNDARY_KNOWN_FAILURE"
	boundaryReasonResolution = "BOUNDARY_RESOLUTION_NOT_EXACT"
	boundaryReasonSubject    = "BOUNDARY_SUBJECT_MISMATCH"
	boundaryReasonContract   = "BOUNDARY_CONTRACT_MISMATCH"
	boundaryReasonWrite      = "BOUNDARY_OBSERVER_EFFECT"
	boundaryReasonLink       = "BOUNDARY_LINK_MISMATCH"
	boundaryReasonDependency = "BOUNDARY_DEPENDENCY_UNKNOWN"
)

func inspectBoundary(spec boundarySpec, raw []byte, head string) (BoundaryResult, artifactEnvelope) {
	result := BoundaryResult{ID: spec.ID, Schema: spec.Schema, MetaOperation: spec.MetaOperation,
		Target: spec.Target, LinksTotal: spec.LinkTarget}
	if len(raw) == 0 {
		setUnknown(&result, boundaryReasonMissing, false)
		return result, artifactEnvelope{}
	}
	result.EvidenceAvailable = true
	result.EvidenceDigest = digestBytes(raw)
	var artifact artifactEnvelope
	if json.Unmarshal(raw, &artifact) != nil {
		setBlocked(&result, boundaryReasonDecode)
		return result, artifact
	}
	result.HeadSHA = artifact.Source.ExpectedHeadSHA
	if spec.ID == "release" {
		result.HeadSHA = artifact.HeadSHA
	}
	result.ReportDigest = artifact.ReportDigest
	result.RepositoryWrites = maxInt(artifact.RepositoryWrites, artifact.Summary.RepositoryWrites)
	if artifact.Decision != "PASS" && artifact.Decision != DecisionFailClosed {
		setUnknown(&result, boundaryReasonUnknown, true)
		return result, artifact
	}
	if artifact.Schema != spec.Schema {
		setBlocked(&result, boundaryReasonSchema)
		return result, artifact
	}
	if artifact.Decision == DecisionFailClosed {
		result.KnownFailure = true
		setBlocked(&result, boundaryReasonFailure)
		return result, artifact
	}
	if artifact.Resolution != ResolutionExact {
		setBlocked(&result, boundaryReasonResolution)
		return result, artifact
	}
	if !validArtifactSubject(spec.ID, artifact, head) || !validDigest(artifact.ReportDigest) {
		setBlocked(&result, boundaryReasonSubject)
		return result, artifact
	}
	if observedMetaOperation(spec.ID, artifact) != spec.MetaOperation {
		setBlocked(&result, boundaryReasonContract)
		return result, artifact
	}
	if result.RepositoryWrites != 0 || mutationAuthorityPresent(spec.ID, artifact) {
		setBlocked(&result, boundaryReasonWrite)
		return result, artifact
	}
	result.Value, result.Status = observeBoundary(spec.ID, artifact)
	if result.Status != StatusSatisfied || result.Value != spec.Target {
		setBlocked(&result, boundaryReasonContract)
		return result, artifact
	}
	result.Resolution, result.Reason = ResolutionExact, boundaryReasonSatisfied
	return result, artifact
}

func setUnknown(result *BoundaryResult, reason string, top bool) {
	result.Status, result.Resolution, result.Reason = StatusUnknown, ResolutionLower, reason
	result.UnknownTopDecision = top
}

func setBlocked(result *BoundaryResult, reason string) {
	result.Status, result.Resolution, result.Reason = StatusBlocked, ResolutionInvariant, reason
}
