package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strings"
)

func baselineReceiptPaths(input Input, registry map[string]CodeBinding, receipts []CouplingReceipt, root semantic.ID, edges map[string]semantic.InferenceEdge, claims map[string]semantic.SemanticChangeClaim) bool {
	for _, receipt := range receipts {
		binding := registry[receipt.SurfaceID]
		final, exists := semantic.InferenceEdge{}, false
		for _, pathID := range receipt.OriginPathIDs {
			candidate, ok := edges[pathID]
			if ok && candidate.Kind == semantic.InferenceIndependentVerification {
				final, exists = candidate, true
				break
			}
		}
		claim, claimExists := claims[receipt.ClaimRecordID]
		if !exists || !claimExists || final.Kind != semantic.InferenceIndependentVerification || final.ObjectID.String() != receipt.ReceiptID || final.SubjectID.String() != binding.CodeSymbolID || claim.ObjectID.String() != receipt.ReceiptID || claim.SubjectID.String() != binding.SemanticOwnerID {
			return false
		}
		if len(receipt.EvidenceRefs) == 0 {
			return false
		}
		current := final
		seen := map[string]bool{}
		for {
			if seen[current.RecordID.String()] {
				return false
			}
			seen[current.RecordID.String()] = true
			if current.SubjectID == root {
				if current.Kind != semantic.InferenceAuthoritativeDeclaration || current.ObjectID.String() != binding.SemanticOwnerID {
					return false
				}
				break
			}
			var previous []semantic.InferenceEdge
			for _, candidate := range input.Path.Edges {
				if candidate.ObjectID == current.SubjectID {
					previous = append(previous, candidate)
				}
			}
			if len(previous) != 1 {
				return false
			}
			current = previous[0]
		}
	}
	return true
}
func baselineSemantic(ir SemanticIR) (baselineSemanticView, bool) {
	seen := map[string]bool{}
	facts := make([]string, 0, len(ir.Nodes)+len(ir.Relations))
	for _, node := range ir.Nodes {
		if !baselineID(node.ID) || node.Kind == "" || node.Namespace == "" || seen[node.ID] {
			return baselineSemanticView{}, false
		}
		seen[node.ID] = true
		facts = append(facts, "node\t"+node.ID+"\t"+node.Kind+"\t"+node.Namespace)
	}
	for _, relation := range ir.Relations {
		if !baselineID(relation.Subject) || !baselineID(relation.Object) || relation.Predicate == "" || !seen[relation.Subject] || !seen[relation.Object] {
			return baselineSemanticView{}, false
		}
		facts = append(facts, "relation\t"+relation.Subject+"\t"+relation.Predicate+"\t"+relation.Object)
	}
	sort.Strings(facts)
	return baselineSemanticView{facts: facts, digest: baselineHash("semantic-ir-v1\n" + strings.Join(facts, "\n") + "\n")}, true
}
