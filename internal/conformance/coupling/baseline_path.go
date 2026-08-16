package coupling

import (
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func baselineClaims(receipts []CouplingReceipt, before, after, delta string) bool {
	for _, receipt := range receipts {
		if receipt.ChangeClaim == ClaimNoDelta {
			if before != after || receipt.ReceiptKind != ReceiptNoSemanticDelta || receipt.SemanticDelta != "" || receipt.SemanticDeltaDigest != "" || receipt.AuthoritativeSourceRef != "" {
				return false
			}
		} else if receipt.ChangeClaim == ClaimDelta {
			if before == after || receipt.ReceiptKind != ReceiptSemanticDelta || receipt.SemanticDelta != delta || receipt.SemanticDeltaDigest != baselineHash(delta) || receipt.AuthoritativeSourceRef == "" {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func baselineManifest(input Input, before, after, registry string) bool {
	manifest := input.Manifest
	if !manifest.Complete || manifest.BeforeSnapshotDigest != baselineStateSnapshot(input.AuthoritySourceBefore, before, registry, input.Config) || manifest.AfterSnapshotDigest != baselineStateSnapshot(input.AuthoritySourceAfter, after, registry, input.Config) || manifest.ToolchainDigest != input.Config.ToolchainDigest || manifest.ProfileDigest != input.Config.Profile.Digest || manifest.RegistryDigest != registry {
		return false
	}
	if manifest.ZeroChange {
		return len(input.Changes) == 0 && before == after && len(input.Receipts) == 0 && len(input.Path.Edges) == 0 && len(input.Path.Claims) == 0 && len(input.Path.Evidence) == 0 && len(input.Roots) == 0
	}
	return true
}

func baselinePath(input Input, registry map[string]CodeBinding, receipts []CouplingReceipt, before, after, delta string) bool {
	if !baselinePathHeader(input, receipts) {
		return false
	}
	root, edges, claims, ok := baselinePathParts(input, before, after, delta)
	if !ok {
		return false
	}
	return baselineReceiptPaths(input, registry, receipts, root, edges, claims)
}

func baselinePathHeader(input Input, receipts []CouplingReceipt) bool {
	return input.Path.Version == semantic.InferencePathSchemaVersion && len(input.Roots) == 1 && len(input.Path.Edges) > 0 && len(input.Path.Claims) == len(receipts) && len(input.Path.Evidence) > 0
}

func baselinePathParts(input Input, before, after, delta string) (semantic.ID, map[string]semantic.InferenceEdge, map[string]semantic.SemanticChangeClaim, bool) {
	root, err := semantic.ParseIdentity(input.Roots[0])
	if err != nil {
		return "", nil, nil, false
	}
	edges := map[string]semantic.InferenceEdge{}
	for _, edge := range input.Path.Edges {
		if !edge.Kind.Valid() || edge.RecordID == "" || !baselineID(edge.RecordID.String()) || edges[edge.RecordID.String()].RecordID != "" {
			return "", nil, nil, false
		}
		edges[edge.RecordID.String()] = edge
	}
	claims := map[string]semantic.SemanticChangeClaim{}
	for _, claim := range input.Path.Claims {
		if !baselineID(claim.RecordID.String()) || claims[claim.RecordID.String()].RecordID != "" || claim.Before.Semantic != before || claim.After.Semantic != after || !claim.Kind.Valid() {
			return "", nil, nil, false
		}
		if claim.Kind == semantic.SemanticDelta && (claim.CanonicalDelta != delta || claim.DeltaDigest != baselineHash(delta)) || claim.Kind == semantic.NoSemanticDelta && (before != after || claim.CanonicalDelta != "" || claim.DeltaDigest != "") {
			return "", nil, nil, false
		}
		claims[claim.RecordID.String()] = claim
	}
	evidenceIDs := map[string]bool{}
	for _, evidence := range input.Path.Evidence {
		if !baselineID(evidence.ID.String()) || evidence.Before.Semantic != before || evidence.After.Semantic != after || evidenceIDs[evidence.ID.String()] {
			return "", nil, nil, false
		}
		evidenceIDs[evidence.ID.String()] = true
	}
	return root, edges, claims, true
}

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

func baselineDelta(before, after []string) string {
	left, right := map[string]bool{}, map[string]bool{}
	for _, fact := range before {
		left[fact] = true
	}
	for _, fact := range after {
		right[fact] = true
	}
	var removed, added []string
	for fact := range left {
		if !right[fact] {
			removed = append(removed, fact)
		}
	}
	for fact := range right {
		if !left[fact] {
			added = append(added, fact)
		}
	}
	sort.Strings(removed)
	sort.Strings(added)
	if len(removed) == 0 && len(added) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("semantic-delta-v1\n")
	for _, fact := range removed {
		builder.WriteString("removed\t" + fact + "\n")
	}
	for _, fact := range added {
		builder.WriteString("added\t" + fact + "\n")
	}
	return builder.String()
}
