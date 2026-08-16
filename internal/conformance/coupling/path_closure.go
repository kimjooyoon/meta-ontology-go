package coupling

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

func receiptsClosePath(roots []semantic.ID, receipts map[string]CouplingReceipt, registry registryView, edges map[semantic.ID]semantic.InferenceEdge, claims map[semantic.ID]semantic.SemanticChangeClaim, evidence map[semantic.ID]semantic.InferenceEvidence) bool {
	rootSet := make(map[semantic.ID]struct{}, len(roots))
	for _, root := range roots {
		rootSet[root] = struct{}{}
	}
	for _, edge := range edges {
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
		chain, ok := selectedOracleChain(receipt.OriginPathIDs, edges)
		if !ok || len(chain) == 0 {
			return false
		}
		pathEdge := chain[len(chain)-1]
		first := chain[0]
		if pathEdge.Kind != semantic.InferenceIndependentVerification || pathEdge.ObjectID != receiptID || pathEdge.SubjectID.String() != binding.CodeSymbolID || first.Kind != semantic.InferenceAuthoritativeDeclaration || len(first.SourceRoots) != 1 || first.SourceRoots[0] != first.SubjectID || first.ObjectID.String() != binding.SemanticOwnerID {
			return false
		}
		if _, isRoot := rootSet[first.SourceRoots[0]]; !isRoot {
			return false
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
		if !sameReceiptEvidence(receipt.EvidenceRefs, chain, claim, evidence) || !hasIndependentEvidence(pathEdge.Evidence, evidence) {
			return false
		}
	}
	if len(usedClaims) != len(claims) || len(usedEdges) != len(edges) || len(usedEvidence) != len(evidence) {
		return false
	}
	return rootsUsed(roots, edges, usedEdges)
}

func sameReceiptEvidence(ids []string, chain []semantic.InferenceEdge, claim semantic.SemanticChangeClaim, evidence map[semantic.ID]semantic.InferenceEvidence) bool {
	if len(ids) != len(sortedUnique(ids)) {
		return false
	}
	want := make(map[string]struct{})
	for _, ref := range claim.Evidence {
		want[ref.ID.String()] = struct{}{}
	}
	for _, edge := range chain {
		for _, ref := range edge.Evidence {
			want[ref.ID.String()] = struct{}{}
		}
	}
	actual := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		parsed, err := semantic.ParseIdentity(id)
		if err != nil {
			return false
		}
		if _, ok := evidence[parsed]; !ok {
			return false
		}
		actual[id] = struct{}{}
	}
	if len(actual) != len(want) {
		return false
	}
	for id := range want {
		if _, ok := actual[id]; !ok {
			return false
		}
	}
	return true
}

// selectedOracleChain independently orders the exact edge IDs named by a
// receipt. It intentionally does not call semantic path normalization: the
// oracle owns the closure, start/end, fork, cycle, and disconnected checks.
func selectedOracleChain(ids []string, edges map[semantic.ID]semantic.InferenceEdge) ([]semantic.InferenceEdge, bool) {
	if len(ids) == 0 {
		return nil, false
	}
	selected := make(map[semantic.ID]semantic.InferenceEdge, len(ids))
	for _, rawID := range ids {
		id, err := semantic.ParseIdentity(rawID)
		if err != nil {
			return nil, false
		}
		if _, duplicate := selected[id]; duplicate {
			return nil, false
		}
		edge, ok := edges[id]
		if !ok {
			return nil, false
		}
		selected[id] = edge
	}
	bySubject := make(map[semantic.ID][]semantic.InferenceEdge, len(selected))
	objects := make(map[semantic.ID]struct{}, len(selected))
	for _, edge := range selected {
		bySubject[edge.SubjectID] = append(bySubject[edge.SubjectID], edge)
		objects[edge.ObjectID] = struct{}{}
	}
	var start semantic.ID
	starts := 0
	for subject := range bySubject {
		if _, hasIncoming := objects[subject]; !hasIncoming {
			start = subject
			starts++
		}
	}
	if starts != 1 {
		return nil, false
	}
	ordered := make([]semantic.InferenceEdge, 0, len(selected))
	visited := make(map[semantic.ID]struct{}, len(selected))
	for {
		outgoing := bySubject[start]
		if len(outgoing) == 0 {
			break
		}
		if len(outgoing) != 1 {
			return nil, false
		}
		edge := outgoing[0]
		if _, duplicate := visited[edge.RecordID]; duplicate {
			return nil, false
		}
		visited[edge.RecordID] = struct{}{}
		ordered = append(ordered, edge)
		start = edge.ObjectID
	}
	if len(ordered) != len(selected) {
		return nil, false
	}
	return ordered, true
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
