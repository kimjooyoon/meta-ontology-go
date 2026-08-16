package coupling

import (
	"encoding/json"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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

func receiptsClosePath(roots []semantic.ID, receipts map[string]CouplingReceipt, registry registryView, edges map[semantic.ID]semantic.InferenceEdge, claims map[semantic.ID]semantic.SemanticChangeClaim, evidence map[semantic.ID]semantic.InferenceEvidence) bool {
	rootSet := make(map[semantic.ID]struct{}, len(roots))
	for _, root := range roots {
		rootSet[root] = struct{}{}
	}
	byObject := make(map[semantic.ID][]semantic.InferenceEdge)
	for _, edge := range edges {
		byObject[edge.ObjectID] = append(byObject[edge.ObjectID], edge)
		if edge.Kind != semantic.InferenceAuthoritativeDeclaration && len(edge.SourceRoots) != 0 {
			return false
		}
	}
	usedEdges := make(map[semantic.ID]struct{}, len(edges))
	usedClaims := make(map[semantic.ID]struct{}, len(claims))
	usedEvidence := make(map[semantic.ID]struct{}, len(evidence))
	for surface, receipt := range receipts {
		binding, ok := registry.bySurface[surface]
		if !ok {
			return false
		}
		receiptID, err := semantic.ParseIdentity(receipt.ReceiptID)
		if err != nil {
			return false
		}
		pathID, err := semantic.ParseIdentity(receipt.OriginPathID)
		if err != nil {
			return false
		}
		pathEdge, ok := edges[pathID]
		if !ok || pathEdge.Kind != semantic.InferenceIndependentVerification || pathEdge.ObjectID != receiptID || pathEdge.SubjectID.String() != binding.CodeSymbolID {
			return false
		}
		chain := make([]semantic.InferenceEdge, 0, len(edges))
		seenChain := make(map[semantic.ID]struct{})
		current := pathEdge
		for {
			if _, duplicate := seenChain[current.RecordID]; duplicate {
				return false
			}
			seenChain[current.RecordID] = struct{}{}
			chain = append(chain, current)
			if _, isRoot := rootSet[current.SubjectID]; isRoot {
				if current.Kind != semantic.InferenceAuthoritativeDeclaration || len(current.SourceRoots) != 1 || current.SourceRoots[0] != current.SubjectID || current.ObjectID.String() != binding.SemanticOwnerID {
					return false
				}
				break
			}
			predecessors := byObject[current.SubjectID]
			if len(predecessors) != 1 {
				return false
			}
			current = predecessors[0]
		}
		if !chainHasProjection(chain, binding) {
			return false
		}
		for _, edge := range chain {
			if _, duplicate := usedEdges[edge.RecordID]; duplicate {
				continue
			}
			usedEdges[edge.RecordID] = struct{}{}
			for _, ref := range edge.Evidence {
				usedEvidence[ref.ID] = struct{}{}
			}
		}
		claimID, err := semantic.ParseIdentity(receipt.ClaimRecordID)
		if err != nil {
			return false
		}
		claim, ok := claims[claimID]
		if !ok || claim.SubjectID.String() != binding.SemanticOwnerID || claim.ObjectID != receiptID || !claimMatchesReceipt(claim.Kind, receipt.ChangeClaim) {
			return false
		}
		if _, duplicate := usedClaims[claimID]; duplicate {
			return false
		}
		usedClaims[claimID] = struct{}{}
		for _, ref := range claim.Evidence {
			usedEvidence[ref.ID] = struct{}{}
		}
		if !sameEvidenceIDs(receipt.EvidenceRefs, claim.Evidence) || !hasIndependentEvidence(pathEdge.Evidence, evidence) {
			return false
		}
	}
	if len(usedClaims) != len(claims) || len(usedEdges) != len(edges) || len(usedEvidence) != len(evidence) {
		return false
	}
	return rootsUsed(roots, edges, usedEdges)
}

func chainHasProjection(chain []semantic.InferenceEdge, binding CodeBinding) bool {
	hasOwner := false
	for _, edge := range chain {
		if edge.SubjectID.String() == binding.SemanticOwnerID || edge.ObjectID.String() == binding.SemanticOwnerID {
			hasOwner = true
		}
	}
	if !hasOwner {
		return false
	}
	for _, edge := range chain {
		if edge.Kind == semantic.InferenceDerivedProjection && edge.ObjectID.String() == binding.CodeSymbolID {
			return true
		}
	}
	return false
}

func claimMatchesReceipt(kind semantic.SemanticChangeKind, claim ChangeClaim) bool {
	switch claim {
	case ClaimDelta:
		return kind == semantic.SemanticDelta
	case ClaimNoDelta:
		return kind == semantic.NoSemanticDelta
	default:
		return false
	}
}

func rootsUsed(roots []semantic.ID, edges map[semantic.ID]semantic.InferenceEdge, used map[semantic.ID]struct{}) bool {
	seen := make(map[semantic.ID]struct{})
	for id := range used {
		edge := edges[id]
		if edge.Kind == semantic.InferenceAuthoritativeDeclaration {
			seen[edge.SubjectID] = struct{}{}
		}
	}
	if len(seen) != len(roots) {
		return false
	}
	for _, root := range roots {
		if _, ok := seen[root]; !ok {
			return false
		}
	}
	return true
}

func sameEvidenceIDs(ids []string, refs []semantic.EvidenceReference) bool {
	if len(ids) != len(sortedUnique(ids)) {
		return false
	}
	left := sortedUnique(ids)
	right := make([]string, 0, len(refs))
	for _, ref := range refs {
		right = append(right, ref.ID.String())
	}
	right = sortedUnique(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func hasIndependentEvidence(refs []semantic.EvidenceReference, evidence map[semantic.ID]semantic.InferenceEvidence) bool {
	for _, ref := range refs {
		if record, ok := evidence[ref.ID]; ok && record.Independent {
			return true
		}
	}
	return false
}

func countIndependentEdges(edges map[semantic.ID]semantic.InferenceEdge) int {
	count := 0
	for _, edge := range edges {
		if edge.Kind == semantic.InferenceIndependentVerification {
			count++
		}
	}
	return count
}

func parseUniqueIDs(values []string) ([]semantic.ID, string) {
	result := make([]semantic.ID, 0, len(values))
	seen := make(map[semantic.ID]struct{}, len(values))
	for _, value := range values {
		id, err := semantic.ParseIdentity(value)
		if err != nil {
			return nil, "invalid-id"
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, "duplicate-id"
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, ""
}

func pathDigest(path semantic.InferencePathV1) string {
	raw := pathToWire(path)
	normalizePathWire(&raw)
	data, _ := json.Marshal(raw)
	return digestBytes(data)
}

func normalizePathWire(path *wirePath) {
	sort.Slice(path.Edges, func(i, j int) bool { return path.Edges[i].RecordID < path.Edges[j].RecordID })
	sort.Slice(path.Claims, func(i, j int) bool { return path.Claims[i].RecordID < path.Claims[j].RecordID })
	sort.Slice(path.Evidence, func(i, j int) bool { return path.Evidence[i].ID < path.Evidence[j].ID })
	for i := range path.Edges {
		sort.Strings(path.Edges[i].SourceRoots)
		sort.Slice(path.Edges[i].Evidence, func(a, b int) bool { return path.Edges[i].Evidence[a].ID < path.Edges[i].Evidence[b].ID })
	}
	for i := range path.Claims {
		sort.Slice(path.Claims[i].Evidence, func(a, b int) bool { return path.Claims[i].Evidence[a].ID < path.Claims[i].Evidence[b].ID })
	}
}
