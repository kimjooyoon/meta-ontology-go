package coupling

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

func validatePath(input Input, registry registryView, receipts map[string]CouplingReceipt, beforeDigest, afterDigest, deltaText string) pathView {
	view := pathView{counts: ObservationCounts{PathEdges: uint64(len(input.Path.Edges)), PathClaims: uint64(len(input.Path.Claims)), PathEvidence: uint64(len(input.Path.Evidence))}}
	if len(receipts) == 0 {
		if len(input.Path.Edges) == 0 && len(input.Path.Claims) == 0 && len(input.Path.Evidence) == 0 && len(input.Roots) == 0 {
			view.digest = pathDigest(input.Path)
			return view
		}
		view.decision, view.reason = DecisionFailClosed, ReasonPathClosure
		return view
	}
	if input.Path.Version != semantic.InferencePathSchemaVersion {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMalformed
		return view
	}
	roots, issue := parseUniqueIDs(input.Roots)
	if issue != "" {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMalformed
		return view
	}
	if len(roots) == 0 {
		view.decision, view.reason = DecisionUnknown, ReasonPathMissing
		return view
	}
	rootSet := make(map[semantic.ID]struct{}, len(roots))
	for _, root := range roots {
		rootSet[root] = struct{}{}
	}
	evidence, issue := collectEvidence(input.Path.Evidence, beforeDigest, afterDigest, input.Config)
	if issue != "" {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMalformed
		return view
	}
	edges, issue := collectEdges(input.Path.Edges, evidence, beforeDigest, afterDigest, input.Config)
	if issue != "" {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMalformed
		return view
	}
	claims, issue := collectClaims(input.Path.Claims, evidence, beforeDigest, afterDigest, deltaText)
	if issue != "" {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMalformed
		return view
	}
	if !declarationRootsMatch(edges, rootSet) {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMissing
		return view
	}
	if !receiptsClosePath(roots, receipts, registry, edges, claims, evidence) {
		view.decision, view.reason = DecisionFailClosed, ReasonPathClosure
		return view
	}
	view.digest = pathDigest(input.Path)
	for _, edge := range input.Path.Edges {
		switch edge.Kind {
		case semantic.InferenceObservationCandidate:
			view.counts.CandidateObservations++
		case semantic.InferenceAcceptedLift:
			view.counts.AcceptedLifts++
		}
	}
	return view
}

func collectEvidence(records []semantic.InferenceEvidence, beforeDigest, afterDigest string, config EvaluationConfig) (map[semantic.ID]semantic.InferenceEvidence, string) {
	result := make(map[semantic.ID]semantic.InferenceEvidence, len(records))
	for _, record := range records {
		if !validID(record.ID.String()) || !validDigest(record.Digest) || record.Before.Semantic != beforeDigest || record.After.Semantic != afterDigest || !validSnapshot(record.Before) || !validSnapshot(record.After) || !validControls(record.Controls, config) {
			return nil, "evidence"
		}
		if _, duplicate := result[record.ID]; duplicate {
			return nil, "duplicate-evidence"
		}
		result[record.ID] = record
	}
	return result, ""
}

func collectEdges(records []semantic.InferenceEdge, evidence map[semantic.ID]semantic.InferenceEvidence, beforeDigest, afterDigest string, config EvaluationConfig) (map[semantic.ID]semantic.InferenceEdge, string) {
	result := make(map[semantic.ID]semantic.InferenceEdge, len(records))
	for _, edge := range records {
		if !validRecord(edge.InferenceRecord, evidence, beforeDigest, afterDigest, config) || !edge.Kind.Valid() {
			return nil, "edge"
		}
		if _, duplicate := result[edge.RecordID]; duplicate {
			return nil, "duplicate-edge"
		}
		if !validKindBinding(edge) {
			return nil, "kind-binding"
		}
		if edge.Kind == semantic.InferenceObservationCandidate && edge.Before.Semantic != edge.After.Semantic {
			return nil, "candidate-authority"
		}
		if edge.Kind == semantic.InferenceAuthoritativeDeclaration && len(edge.SourceRoots) == 0 {
			return nil, "declaration-root"
		}
		if edge.Kind == semantic.InferenceAcceptedLift {
			if edge.AcceptanceReceipt == "" || !sourceBackedReceipt(edge.AcceptanceReceipt, edge.Evidence, evidence) {
				return nil, "accepted-lift"
			}
		} else if edge.AcceptanceReceipt != "" {
			return nil, "unexpected-receipt"
		}
		result[edge.RecordID] = edge
	}
	return result, ""
}

func collectClaims(records []semantic.SemanticChangeClaim, evidence map[semantic.ID]semantic.InferenceEvidence, beforeDigest, afterDigest, deltaText string) (map[semantic.ID]semantic.SemanticChangeClaim, string) {
	result := make(map[semantic.ID]semantic.SemanticChangeClaim, len(records))
	for _, claim := range records {
		if !validRecord(claim.InferenceRecord, evidence, beforeDigest, afterDigest, EvaluationConfig{}) {
			return nil, "claim"
		}
		if claim.Authority.Layer != semantic.AuthoritySemantic || !claim.Kind.Valid() {
			return nil, "claim-kind"
		}
		switch claim.Kind {
		case semantic.SemanticDelta:
			if claim.Before.Semantic == claim.After.Semantic || claim.Authority.Effect != semantic.AuthorityDelta || claim.CanonicalDelta != deltaText || claim.DeltaDigest != digestBytes([]byte(deltaText)) {
				return nil, "delta-claim"
			}
		case semantic.NoSemanticDelta:
			if claim.Before.Semantic != claim.After.Semantic || claim.Authority.Effect != semantic.AuthorityNoChange || claim.CanonicalDelta != "" || claim.DeltaDigest != "" {
				return nil, "no-delta-claim"
			}
		}
		if _, duplicate := result[claim.RecordID]; duplicate {
			return nil, "duplicate-claim"
		}
		result[claim.RecordID] = claim
	}
	return result, ""
}

func validRecord(record semantic.InferenceRecord, evidence map[semantic.ID]semantic.InferenceEvidence, beforeDigest, afterDigest string, config EvaluationConfig) bool {
	if !validID(record.RecordID.String()) || !validID(record.SubjectID.String()) || !validID(record.ObjectID.String()) ||
		!validID(record.Rule.ID.String()) || !validToken(record.Rule.Version) || !validDigest(record.Rule.Digest) ||
		!record.Phase.Phase.Valid() || record.Phase.Ordinal == 0 || record.Before.Semantic != beforeDigest || record.After.Semantic != afterDigest ||
		!validSnapshot(record.Before) || !validSnapshot(record.After) || !record.Authority.Layer.Valid() || !record.Authority.Effect.Valid() ||
		len(record.Evidence) == 0 {
		return false
	}
	if !validControls(record.Controls, config) {
		return false
	}
	seen := make(map[semantic.ID]struct{}, len(record.Evidence))
	for _, ref := range record.Evidence {
		if !validID(ref.ID.String()) || !validDigest(ref.Digest) {
			return false
		}
		if _, duplicate := seen[ref.ID]; duplicate {
			return false
		}
		seen[ref.ID] = struct{}{}
		ev, ok := evidence[ref.ID]
		if !ok || ev.Digest != ref.Digest || ev.Before != record.Before || ev.After != record.After || !controlsEqual(ev.Controls, record.Controls) {
			return false
		}
	}
	return true
}

func validKindBinding(edge semantic.InferenceEdge) bool {
	if edge.Kind == semantic.InferenceAuthoritativeDeclaration {
		return edge.Phase.Phase == semantic.PhaseDeclaration && edge.Authority.Layer == semantic.AuthoritySource && edge.Authority.Effect == semantic.AuthorityDeclare
	}
	if edge.Kind == semantic.InferenceDeterministicDerivation {
		return edge.Phase.Phase == semantic.PhaseDerivation && edge.Authority.Layer == semantic.AuthoritySemantic && edge.Authority.Effect == semantic.AuthorityDerive
	}
	if edge.Kind == semantic.InferenceDerivedProjection {
		return edge.Phase.Phase == semantic.PhaseProjection && edge.Authority.Layer == semantic.AuthorityDerived && edge.Authority.Effect == semantic.AuthorityProject && edge.Controls.Profile.ID != ""
	}
	if edge.Kind == semantic.InferenceObservationCandidate {
		return edge.Phase.Phase == semantic.PhaseObservation && edge.Authority.Layer == semantic.AuthorityCandidate && edge.Authority.Effect == semantic.AuthorityObserve && edge.Controls.CatalogDigest != ""
	}
	if edge.Kind == semantic.InferenceAcceptedLift {
		return edge.Phase.Phase == semantic.PhaseLift && edge.Authority.Layer == semantic.AuthoritySemantic && edge.Authority.Effect == semantic.AuthorityLift && edge.Controls.PolicyDigest != ""
	}
	return edge.Phase.Phase == semantic.PhaseVerification && edge.Authority.Layer == semantic.AuthorityVerification && edge.Authority.Effect == semantic.AuthorityVerify && edge.Controls.PolicyDigest != ""
}

func validControls(controls semantic.InferenceControls, config EvaluationConfig) bool {
	if controls.CatalogDigest != "" && !validDigest(controls.CatalogDigest) || controls.PolicyDigest != "" && !validDigest(controls.PolicyDigest) {
		return false
	}
	profile := controls.Profile
	if profile.ID == "" && profile.Version == "" && profile.Digest == "" {
		return true
	}
	if profile.ID == "" || profile.Version == "" || !validDigest(profile.Digest) {
		return false
	}
	if config.Profile.ID != "" && (profile.ID != config.Profile.ID || profile.Version != config.Profile.Version || profile.Digest != config.Profile.Digest) {
		return false
	}
	return true
}

func validSnapshot(snapshot semantic.SnapshotDigests) bool {
	return (snapshot.Source != "" && validDigest(snapshot.Source)) || (snapshot.Semantic != "" && validDigest(snapshot.Semantic))
}

func controlsEqual(left, right semantic.InferenceControls) bool {
	return left.CatalogDigest == right.CatalogDigest && left.PolicyDigest == right.PolicyDigest && left.Profile == right.Profile
}

func sourceBackedReceipt(id semantic.ID, refs []semantic.EvidenceReference, evidence map[semantic.ID]semantic.InferenceEvidence) bool {
	for _, ref := range refs {
		if ref.ID == id {
			record, ok := evidence[id]
			return ok && record.SourceBacked && record.Before.Source != "" && record.After.Source != ""
		}
	}
	return false
}

func declarationRootsMatch(edges map[semantic.ID]semantic.InferenceEdge, roots map[semantic.ID]struct{}) bool {
	seen := make(map[semantic.ID]struct{})
	for _, edge := range edges {
		if edge.Kind != semantic.InferenceAuthoritativeDeclaration {
			continue
		}
		for _, root := range edge.SourceRoots {
			if _, ok := roots[root]; !ok {
				return false
			}
			seen[root] = struct{}{}
		}
	}
	return len(seen) == len(roots)
}
