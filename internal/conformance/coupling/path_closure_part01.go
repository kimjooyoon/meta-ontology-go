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
