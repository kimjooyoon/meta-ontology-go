package cache

// FutureStageEvidence records the planned host without claiming it works.
func FutureStageEvidence() StageEvidence {
	return StageEvidence{Stage: GoooHostedStage, Status: EvidenceDeferred,
		Authority: "gooo-hosted implementation and parity checks are not available"}
}

// Valid reports whether s is one of the two declared host stages.
func (s HostStage) Valid() bool {
	return s == GoHostedStage || s == GoooHostedStage
}

// String returns the stable serialized stage name.
func (s HostStage) String() string { return string(s) }
