package coupling

import (
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
