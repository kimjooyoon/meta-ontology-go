package provenance

// OriginChainUpdate is an evidence-linked description of an explicit origin
// replacement. It does not decide whether a changed chain is better.
type OriginChainUpdate struct {
	BeforeDigest  string                `json:"before_digest,omitempty"`
	AfterDigest   string                `json:"after_digest,omitempty"`
	Transition    OriginChainTransition `json:"transition"`
	ChangedStages []OriginChainStage    `json:"changed_stages,omitempty"`
	Reason        string                `json:"reason"`
}

// Verified recomputes the observation identity from its chain and rejects
// caller-supplied status, missing-boundary, reason, or digest drift.
func (observation OriginChainObservation) Verified() bool {
	canonical := ObserveOriginChain(observation.Chain)
	if observation.Status != canonical.Status || observation.Digest != canonical.Digest || observation.Reason != canonical.Reason {
		return false
	}
	if len(observation.Missing) != len(canonical.Missing) {
		return false
	}
	for index := range canonical.Missing {
		if observation.Missing[index] != canonical.Missing[index] {
			return false
		}
	}
	return true
}

// UpdateOriginChain compares two self-consistent complete observations. Any
// incomplete or tampered input remains UNKNOWN rather than being treated as a
// no-op or as an improvement.
func UpdateOriginChain(before, after OriginChainObservation) OriginChainUpdate {
	update := OriginChainUpdate{BeforeDigest: before.Digest, AfterDigest: after.Digest, Transition: OriginChainTransitionUnknown}
	if !before.Verified() || !after.Verified() {
		update.Reason = "origin observation is not self-consistent"
		return update
	}
	if !before.Comparable() || !after.Comparable() {
		update.Reason = "origin comparison requires complete observations"
		return update
	}
	update.Transition = CompareOriginChains(before, after)
	update.ChangedStages = originChainChangedStages(before.Chain, after.Chain)
	if update.Transition == OriginChainTransitionUnchanged {
		update.Reason = "origin chain unchanged"
	} else {
		update.Reason = "origin chain changed"
	}
	return update
}

func originChainChangedStages(before, after OriginChain) []OriginChainStage {
	changed := make([]OriginChainStage, 0, originChainFieldCount)
	if before.DeclarationURI != after.DeclarationURI || before.DeclarationSymbol != after.DeclarationSymbol {
		changed = append(changed, OriginChainStageDeclaration)
	}
	if before.IRNode != after.IRNode {
		changed = append(changed, OriginChainStageIR)
	}
	if before.GeneratedURI != after.GeneratedURI || before.GeneratedSymbol != after.GeneratedSymbol {
		changed = append(changed, OriginChainStageGeneration)
	}
	if before.ReverseObservationURI != after.ReverseObservationURI || before.ReverseObservation != after.ReverseObservation {
		changed = append(changed, OriginChainStageReverseObservation)
	}
	if before.MetricName != after.MetricName || before.MetricValue != after.MetricValue || before.EvidenceDigest != after.EvidenceDigest {
		changed = append(changed, OriginChainStageMetric)
	}
	return changed
}
