package coupling

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// EvaluateBaseline is a deliberately separate full-suite implementation. It
// re-derives registry closure, semantic equality, receipt predicates, and
// resource bindings from the typed input without calling Evaluate or any of
// the oracle's private helpers.
func EvaluateBaseline(input Input) BaselineResult {
	result := BaselineResult{FullSuite: true, ObservationCounts: baselineCounts(input)}
	if !baselineResourceBindingsEqual(input.Config.ResourceBinding, input.ResourceRegistry) {
		return baselineUnknown(result, input, ReasonResourceUnbound)
	}
	resources, ok, reason := baselineResources(input.ResourceReceipts, input.Config.ResourceBinding)
	result.Resources = resources
	result.WorkUnits = resources.WorkUnits
	if !ok {
		return baselineUnknown(result, input, reason)
	}
	if input.Schema != SchemaV1 || input.Config.ToolchainDigest == "" || input.Config.Profile.ID == "" || input.Config.Profile.Version == "" || input.Config.Profile.Digest == "" || input.AuthoritySourceBefore == "" || input.AuthoritySourceAfter == "" || len(input.Registry) == 0 || !input.Manifest.Complete || (len(input.Changes) == 0 && !input.Manifest.ZeroChange) {
		return baselineUnknown(result, input, ReasonRequiredInputMissing)
	}
	if !baselineDigest(input.Config.ToolchainDigest) || !baselineDigest(input.Config.Profile.Digest) {
		return baselineUnknown(result, input, ReasonRequiredInputMissing)
	}
	before, ok := baselineSemantic(input.SemanticBefore)
	if !ok {
		return baselineFail(result, input, ReasonPathMalformed)
	}
	after, ok := baselineSemantic(input.SemanticAfter)
	if !ok {
		return baselineFail(result, input, ReasonPathMalformed)
	}
	registry, ok := baselineRegistry(input.Registry)
	if !ok || registry.digest != input.RegistryDigest {
		return baselineFail(result, input, ReasonDigestMismatch)
	}
	if !baselineManifest(input, before.digest, after.digest, registry.digest) {
		if !input.Manifest.Complete {
			return baselineUnknown(result, input, ReasonRequiredInputMissing)
		}
		return baselineFail(result, input, ReasonDigestMismatch)
	}
	changed, ok := baselineChanged(input.Changes, registry.bySymbol)
	if !ok {
		return baselineFail(result, input, ReasonChangedSurface)
	}
	result.LocalizedSurfaces = append([]string(nil), changed...)
	if receiptsOK, receiptReason := baselineReceipts(input, registry.bySurface, changed, before.digest, after.digest); !receiptsOK {
		if receiptReason == ReasonMissingReceipt || receiptReason == ReasonStaleReceipt {
			return baselineUnknown(result, input, receiptReason)
		}
		return baselineFail(result, input, receiptReason)
	}
	delta := baselineDelta(before.facts, after.facts)
	if !baselineClaims(input.Receipts, before.digest, after.digest, delta) {
		return baselineFail(result, input, ReasonInvalidDelta)
	}
	if len(changed) > 0 && !baselinePath(input, registry.bySurface, input.Receipts, before.digest, after.digest, delta) {
		return baselineFail(result, input, ReasonPathClosure)
	}
	result.Decision, result.Reason, result.LocalizedSurfaces = DecisionPass, ReasonNone, nil
	return result
}

func baselineResourceBindingsEqual(left, right ResourceBindingConfig) bool {
	return left.ProviderID != "" && left.ProviderID == right.ProviderID && left.ObserverID != "" && left.ObserverID == right.ObserverID && left.ProviderDigest != "" && left.ProviderDigest == right.ProviderDigest && left.ObserverDigest != "" && left.ObserverDigest == right.ObserverDigest && left.SnapshotDigest != "" && left.SnapshotDigest == right.SnapshotDigest && left.SourceDigest != "" && left.SourceDigest == right.SourceDigest
}

func Compare(input Input) Comparison {
	oracle := Evaluate(input)
	baseline := EvaluateBaseline(input)
	comparison := Comparison{Oracle: oracle, Baseline: baseline, OutcomeMatch: oracle.Decision == baseline.Decision, ReasonMatch: oracle.Reason == baseline.Reason, LocalizationMatch: sameSurfaceSet(oracle.ChangedSurfaces, baseline.LocalizedSurfaces)}
	if oracle.Decision == DecisionPass {
		comparison.LocalizationMatch = len(baseline.LocalizedSurfaces) == 0
	}
	if comparison.OutcomeMatch && comparison.ReasonMatch && comparison.LocalizationMatch {
		comparison.Finding = "NO_UNIQUE_BENEFIT"
	} else {
		comparison.Finding = "UNIQUE_BENEFIT_NOT_ESTABLISHED"
	}
	return comparison
}

type baselineSemanticView struct {
	digest string
	facts  []string
}

type baselineRegistryView struct {
	bySurface map[string]CodeBinding
	bySymbol  map[string]CodeBinding
	digest    string
}

func baselineCounts(input Input) ObservationCounts {
	counts := ObservationCounts{RegistryBindings: uint64(len(input.Registry)), ReceiptRecords: uint64(len(input.Receipts)), PathEdges: uint64(len(input.Path.Edges)), PathClaims: uint64(len(input.Path.Claims)), PathEvidence: uint64(len(input.Path.Evidence)), ResourceReceipts: uint64(len(input.ResourceReceipts))}
	for _, change := range input.Changes {
		if change.BeforeDigest != change.AfterDigest {
			counts.ChangedCodeSurfaces++
		}
	}
	for _, edge := range input.Path.Edges {
		if edge.Kind == semantic.InferenceObservationCandidate {
			counts.CandidateObservations++
		}
		if edge.Kind == semantic.InferenceAcceptedLift {
			counts.AcceptedLifts++
		}
	}
	return counts
}

func baselineResources(receipts []ExternalResourceReceipt, expected ResourceBindingConfig) (ResourceObservation, bool, Reason) {
	seen := map[string]bool{}
	var result ResourceObservation
	for _, receipt := range receipts {
		if !baselineID(receipt.ReceiptID) || !baselineID(expected.ProviderID) || !baselineID(expected.ObserverID) || !baselineDigest(expected.ProviderDigest) || !baselineDigest(expected.ObserverDigest) || !baselineDigest(expected.SnapshotDigest) || !baselineDigest(expected.SourceDigest) || !baselineDigest(receipt.ProviderDigest) || !baselineDigest(receipt.ObserverDigest) || !baselineDigest(receipt.SnapshotDigest) || !baselineDigest(receipt.SourceDigest) || !baselineDigest(receipt.BindingDigest) || expected.ProviderDigest != baselineProviderDigest(expected.ProviderID) || expected.ObserverDigest != baselineObserverDigest(expected.ObserverID) || expected.SourceDigest != baselineSourceDigest(expected.ProviderID, expected.ObserverID, expected.SnapshotDigest) || receipt.ProviderDigest != expected.ProviderDigest || receipt.ObserverDigest != expected.ObserverDigest || receipt.SnapshotDigest != expected.SnapshotDigest || receipt.SourceDigest != expected.SourceDigest || !receipt.Present || !receipt.Independent || receipt.State != "CURRENT" || receipt.BindingDigest != baselineResourceDigest(receipt) {
			return ResourceObservation{}, false, ReasonResourceUnbound
		}
		if seen[receipt.Metric] {
			return ResourceObservation{}, false, ReasonResourceUnbound
		}
		seen[receipt.Metric] = true
		switch receipt.Metric {
		case "cpu-core-ns":
			if receipt.Unit != "ns" {
				return ResourceObservation{}, false, ReasonResourceUnbound
			}
			result.CPUCoreNS = receipt.Value
		case "peak-memory-bytes":
			if receipt.Unit != "bytes" {
				return ResourceObservation{}, false, ReasonResourceUnbound
			}
			result.PeakMemoryBytes = receipt.Value
		case "work-units":
			if receipt.Unit != "units" {
				return ResourceObservation{}, false, ReasonResourceUnbound
			}
			result.WorkUnits = receipt.Value
		default:
			return ResourceObservation{}, false, ReasonResourceUnbound
		}
	}
	if len(seen) != 3 {
		return ResourceObservation{}, false, ReasonResourceUnbound
	}
	return result, true, ReasonNone
}

func baselineRegistry(bindings []CodeBinding) (baselineRegistryView, bool) {
	result := baselineRegistryView{bySurface: map[string]CodeBinding{}, bySymbol: map[string]CodeBinding{}}
	parts := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		if !baselineID(binding.RegisteredSurfaceID) || !baselineID(binding.CodeSymbolID) || !baselineID(binding.SemanticOwnerID) || binding.SourceMapID == "" || !baselineDigest(binding.BindingDigest) || binding.BindingDigest != baselineBindingDigest(binding) {
			return baselineRegistryView{}, false
		}
		if _, exists := result.bySurface[binding.RegisteredSurfaceID]; exists {
			return baselineRegistryView{}, false
		}
		if _, exists := result.bySymbol[binding.CodeSymbolID]; exists {
			return baselineRegistryView{}, false
		}
		result.bySurface[binding.RegisteredSurfaceID] = binding
		result.bySymbol[binding.CodeSymbolID] = binding
		parts = append(parts, baselineBindingCanonical(binding))
	}
	sort.Strings(parts)
	result.digest = baselineHash(strings.Join(parts, "\n") + "\n")
	return result, true
}

func baselineChanged(changes []CodeChange, symbols map[string]CodeBinding) ([]string, bool) {
	seen := map[string]bool{}
	result := make([]string, 0)
	for _, change := range changes {
		if !baselineID(change.CodeSymbolID) || !baselineDigest(change.BeforeDigest) || !baselineDigest(change.AfterDigest) || seen[change.CodeSymbolID] {
			return nil, false
		}
		seen[change.CodeSymbolID] = true
		if change.BeforeDigest == change.AfterDigest {
			continue
		}
		binding, exists := symbols[change.CodeSymbolID]
		if !exists {
			return nil, false
		}
		result = append(result, binding.RegisteredSurfaceID)
	}
	sort.Strings(result)
	return result, true
}

func baselineReceipts(input Input, registry map[string]CodeBinding, changed []string, before, after string) (bool, Reason) {
	wanted := map[string]bool{}
	for _, surface := range changed {
		wanted[surface] = true
	}
	seen := map[string]bool{}
	for _, receipt := range input.Receipts {
		if !baselineID(receipt.ReceiptID) {
			return false, ReasonStaleReceipt
		}
		if receipt.State == "STALE" {
			return false, ReasonStaleReceipt
		}
		if receipt.State != "CURRENT" || !wanted[receipt.SurfaceID] || seen[receipt.SurfaceID] {
			if seen[receipt.SurfaceID] {
				return false, ReasonDuplicateReceipt
			}
			if !wanted[receipt.SurfaceID] {
				return false, ReasonOrphanReceipt
			}
			return false, ReasonStaleReceipt
		}
		seen[receipt.SurfaceID] = true
		binding := registry[receipt.SurfaceID]
		if receipt.SemanticOwnerID != binding.SemanticOwnerID || receipt.CodeSymbolID != binding.CodeSymbolID || receipt.SourceMapBindingDigest != binding.BindingDigest {
			return false, ReasonRegistryBinding
		}
		if receipt.RegistryDigest != input.RegistryDigest || receipt.ToolchainDigest != input.Config.ToolchainDigest || receipt.ProfileDigest != input.Config.Profile.Digest || receipt.BeforeIRDigest != before || receipt.AfterIRDigest != after || receipt.SnapshotDigest != baselineSnapshot(input, before, after) || receipt.AuthoritySourceBeforeDigest != baselineHash(input.AuthoritySourceBefore) || receipt.AuthoritySourceAfterDigest != baselineHash(input.AuthoritySourceAfter) {
			return false, ReasonStaleReceipt
		}
	}
	if len(seen) != len(wanted) {
		return false, ReasonMissingReceipt
	}
	return len(input.Receipts) == len(wanted), ReasonOrphanReceipt
}

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
	if input.Path.Version != semantic.InferencePathSchemaVersion || len(input.Roots) != 1 || len(input.Path.Edges) == 0 || len(input.Path.Claims) != len(receipts) || len(input.Path.Evidence) == 0 {
		return false
	}
	root, err := semantic.ParseIdentity(input.Roots[0])
	if err != nil {
		return false
	}
	edges := map[string]semantic.InferenceEdge{}
	for _, edge := range input.Path.Edges {
		if !edge.Kind.Valid() || edge.RecordID == "" || !baselineID(edge.RecordID.String()) || edges[edge.RecordID.String()].RecordID != "" {
			return false
		}
		edges[edge.RecordID.String()] = edge
	}
	claims := map[string]semantic.SemanticChangeClaim{}
	for _, claim := range input.Path.Claims {
		if !baselineID(claim.RecordID.String()) || claims[claim.RecordID.String()].RecordID != "" || claim.Before.Semantic != before || claim.After.Semantic != after || !claim.Kind.Valid() {
			return false
		}
		if claim.Kind == semantic.SemanticDelta && (claim.CanonicalDelta != delta || claim.DeltaDigest != baselineHash(delta)) || claim.Kind == semantic.NoSemanticDelta && (before != after || claim.CanonicalDelta != "" || claim.DeltaDigest != "") {
			return false
		}
		claims[claim.RecordID.String()] = claim
	}
	evidenceIDs := map[string]bool{}
	for _, evidence := range input.Path.Evidence {
		if !baselineID(evidence.ID.String()) || evidence.Before.Semantic != before || evidence.After.Semantic != after || evidenceIDs[evidence.ID.String()] {
			return false
		}
		evidenceIDs[evidence.ID.String()] = true
	}
	for _, receipt := range receipts {
		binding := registry[receipt.SurfaceID]
		final, exists := edges[receipt.OriginPathID]
		claim, claimExists := claims[receipt.ClaimRecordID]
		if !exists || !claimExists || final.Kind != semantic.InferenceIndependentVerification || final.ObjectID.String() != receipt.ReceiptID || final.SubjectID.String() != binding.CodeSymbolID || claim.ObjectID.String() != receipt.ReceiptID || claim.SubjectID.String() != binding.SemanticOwnerID {
			return false
		}
		if len(receipt.EvidenceRefs) == 0 || len(claim.Evidence) != len(receipt.EvidenceRefs) {
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

func baselineBindingDigest(binding CodeBinding) string {
	return baselineHash("binding-v1\n" + binding.RegisteredSurfaceID + "\n" + binding.CodeSymbolID + "\n" + binding.SemanticOwnerID + "\n" + binding.SourceMapID + "\n")
}

func baselineBindingCanonical(binding CodeBinding) string {
	return binding.RegisteredSurfaceID + "\t" + binding.CodeSymbolID + "\t" + binding.SemanticOwnerID + "\t" + binding.SourceMapID + "\t" + binding.BindingDigest
}

func baselineSnapshot(input Input, before, after string) string {
	return baselineHash("snapshot-v1\n" + baselineHash(input.AuthoritySourceBefore) + "\n" + baselineHash(input.AuthoritySourceAfter) + "\n" + before + "\n" + after + "\n" + input.RegistryDigest + "\n" + input.Config.ToolchainDigest + "\n" + input.Config.Profile.ID + "\n" + input.Config.Profile.Version + "\n" + input.Config.Profile.Digest + "\n")
}

func baselineStateSnapshot(source, semanticDigest, registry string, config EvaluationConfig) string {
	return baselineHash("state-snapshot-v1\n" + baselineHash(source) + "\n" + semanticDigest + "\n" + registry + "\n" + config.ToolchainDigest + "\n" + config.Profile.ID + "\n" + config.Profile.Version + "\n" + config.Profile.Digest + "\n")
}

func baselineResourceDigest(receipt ExternalResourceReceipt) string {
	return baselineHash("resource-binding-v1\n" + receipt.ReceiptID + "\n" + receipt.Metric + "\n" + fmt.Sprint(receipt.Value) + "\n" + receipt.Unit + "\n" + receipt.ProviderDigest + "\n" + receipt.ObserverDigest + "\n" + receipt.SnapshotDigest + "\n" + receipt.SourceDigest + "\n")
}

func baselineProviderDigest(id string) string {
	return baselineHash("resource-provider-v1\n" + id + "\n")
}

func baselineObserverDigest(id string) string {
	return baselineHash("resource-observer-v1\n" + id + "\n")
}

func baselineSourceDigest(providerID, observerID, snapshot string) string {
	return baselineHash("resource-source-v1\n" + providerID + "\n" + observerID + "\n" + snapshot + "\n")
}

func baselineHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func baselineDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}

func baselineID(value string) bool {
	_, err := semantic.ParseIdentity(value)
	return err == nil
}

func baselineReceiptReason(input Input, changed []string) Reason {
	if len(input.Receipts) < len(changed) {
		return ReasonMissingReceipt
	}
	return ReasonOrphanReceipt
}

func baselineUnknown(result BaselineResult, input Input, reason Reason) BaselineResult {
	result.Decision, result.Reason = DecisionUnknown, reason
	result.LocalizedSurfaces = baselineChangedLabels(input)
	return result
}

func baselineFail(result BaselineResult, input Input, reason Reason) BaselineResult {
	result.Decision, result.Reason = DecisionFailClosed, reason
	result.LocalizedSurfaces = baselineChangedLabels(input)
	return result
}

func baselineChangedLabels(input Input) []string {
	result := make([]string, 0)
	for _, change := range input.Changes {
		if change.BeforeDigest != change.AfterDigest {
			result = append(result, change.CodeSymbolID)
		}
	}
	sort.Strings(result)
	return result
}

func sameSurfaceSet(left, right []string) bool {
	left, right = append([]string(nil), left...), append([]string(nil), right...)
	sort.Strings(left)
	sort.Strings(right)
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
