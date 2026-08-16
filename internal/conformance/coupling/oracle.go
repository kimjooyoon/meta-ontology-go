package coupling

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type oracleValidation struct {
	decision Decision
	reason   Reason
}

type normalizedSemantic struct {
	digest string
	facts  []string
}

type registryView struct {
	bySurface map[string]CodeBinding
	bySymbol  map[string]CodeBinding
	digest    string
}

type receiptView struct {
	bySurface map[string]CouplingReceipt
	valid     []string
}

type pathView struct {
	digest   string
	counts   ObservationCounts
	decision Decision
	reason   Reason
}

func Evaluate(input Input) Output {
	output := Output{Schema: SchemaV1, FixtureID: input.FixtureID, InputDigest: CanonicalInputDigest(input)}
	output.ObservationCounts = inputObservationCounts(input)
	if !resourceBindingsEqual(input.Config.ResourceBinding, input.ResourceRegistry) {
		return finish(output, DecisionUnknown, ReasonResourceUnbound)
	}

	if issue := validateRequiredInput(input); issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	resources, issue := normalizeResources(input.ResourceReceipts, input.Config.ResourceBinding)
	output.Resources = resources
	output.ObservationCounts.ResourceReceipts = uint64(len(input.ResourceReceipts))
	if issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	before, err := normalizeSemantic(input.SemanticBefore)
	if err != nil {
		return finish(output, DecisionFailClosed, ReasonPathMalformed)
	}
	after, err := normalizeSemantic(input.SemanticAfter)
	if err != nil {
		return finish(output, DecisionFailClosed, ReasonPathMalformed)
	}
	output.SemanticBeforeDigest, output.SemanticAfterDigest = before.digest, after.digest
	deltaText, added, removed := semanticDelta(before.facts, after.facts)
	output.ObservationCounts.AddedSemanticFacts = uint64(len(added))
	output.ObservationCounts.RemovedSemanticFacts = uint64(len(removed))
	if input.RegistryDigest == "" || !validDigest(input.RegistryDigest) {
		return finish(output, DecisionUnknown, ReasonRequiredInputMissing)
	}
	registry, issue := normalizeRegistry(input.Registry)
	if issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	if registry.digest != input.RegistryDigest {
		return finish(output, DecisionFailClosed, ReasonDigestMismatch)
	}
	if issue := validateManifest(input, before.digest, after.digest, registry.digest); issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	changed, issue := resolveChangedSurfaces(input.Changes, registry)
	if issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	output.ChangedSurfaces = changed
	output.ObservationCounts.ChangedRegistered = uint64(len(changed))
	if issue := validateSourceBindings(input, before.digest, after.digest); issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	receipts, issue := validateReceipts(input, registry, changed, before, after, deltaText)
	output.ReceiptSurfaces = receipts.valid
	output.ObservationCounts.ValidReceipts = uint64(len(receipts.valid))
	if issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	path := validatePath(input, registry, receipts.bySurface, before.digest, after.digest, deltaText)
	output.PathClosureDigest = path.digest
	output.ObservationCounts.PathEdges = path.counts.PathEdges
	output.ObservationCounts.PathClaims = path.counts.PathClaims
	output.ObservationCounts.PathEvidence = path.counts.PathEvidence
	output.ObservationCounts.CandidateObservations = path.counts.CandidateObservations
	output.ObservationCounts.AcceptedLifts = path.counts.AcceptedLifts
	if path.decision != "" {
		return finish(output, path.decision, path.reason)
	}
	if len(changed) > 0 {
		output.SemanticDeltaDigest = ""
		if deltaText != "" {
			output.SemanticDeltaDigest = digestBytes([]byte(deltaText))
		}
	}
	return finish(output, DecisionPass, ReasonNone)
}

func finish(output Output, decision Decision, reason Reason) Output {
	output.Decision, output.Reason = decision, reason
	output.ChangedSurfaces = sortedUnique(output.ChangedSurfaces)
	output.ReceiptSurfaces = sortedUnique(output.ReceiptSurfaces)
	output.CanonicalOutputDigest = CanonicalOutputDigest(output)
	output.ReplayDigest = ReplayDigest(output.InputDigest, output.CanonicalOutputDigest)
	return output
}

func validateRequiredInput(input Input) oracleValidation {
	if input.Schema != SchemaV1 || input.Config.ToolchainDigest == "" || input.Config.Profile.ID == "" ||
		input.Config.Profile.Version == "" || input.Config.Profile.Digest == "" ||
		input.AuthoritySourceBefore == "" || input.AuthoritySourceAfter == "" ||
		len(input.Registry) == 0 || len(input.ResourceReceipts) == 0 || !input.Manifest.Complete ||
		(len(input.Changes) == 0 && !input.Manifest.ZeroChange) {
		return oracleValidation{DecisionUnknown, ReasonRequiredInputMissing}
	}
	if !validDigest(input.Config.ToolchainDigest) || !validDigest(input.Config.Profile.Digest) {
		return oracleValidation{DecisionUnknown, ReasonRequiredInputMissing}
	}
	return oracleValidation{}
}

func normalizeResources(receipts []ExternalResourceReceipt, expected ResourceBindingConfig) (ResourceObservation, oracleValidation) {
	seen := make(map[string]struct{}, len(receipts))
	var out ResourceObservation
	for _, receipt := range receipts {
		if !validID(receipt.ReceiptID) || !validID(expected.ProviderID) || !validID(expected.ObserverID) || !validDigest(expected.ProviderDigest) || !validDigest(expected.ObserverDigest) || !validDigest(expected.SnapshotDigest) || !validDigest(expected.SourceDigest) || !validDigest(receipt.ProviderDigest) || !validDigest(receipt.ObserverDigest) || !validDigest(receipt.SnapshotDigest) || !validDigest(receipt.SourceDigest) || !validDigest(receipt.BindingDigest) ||
			expected.ProviderDigest != resourceProviderDigest(expected.ProviderID) || expected.ObserverDigest != resourceObserverDigest(expected.ObserverID) || expected.SourceDigest != resourceSourceDigest(expected.ProviderID, expected.ObserverID, expected.SnapshotDigest) ||
			receipt.ProviderDigest != expected.ProviderDigest || receipt.ObserverDigest != expected.ObserverDigest || receipt.SnapshotDigest != expected.SnapshotDigest || receipt.SourceDigest != expected.SourceDigest ||
			!receipt.Present || !receipt.Independent || receipt.State != "CURRENT" ||
			receipt.BindingDigest != resourceBindingDigest(receipt) {
			return ResourceObservation{}, oracleValidation{DecisionUnknown, ReasonResourceUnbound}
		}
		if _, duplicate := seen[receipt.Metric]; duplicate {
			return ResourceObservation{}, oracleValidation{DecisionFailClosed, ReasonResourceUnbound}
		}
		seen[receipt.Metric] = struct{}{}
		switch receipt.Metric {
		case "cpu-core-ns":
			if receipt.Unit != "ns" {
				return ResourceObservation{}, oracleValidation{DecisionFailClosed, ReasonResourceUnbound}
			}
			out.CPUCoreNS = receipt.Value
		case "peak-memory-bytes":
			if receipt.Unit != "bytes" {
				return ResourceObservation{}, oracleValidation{DecisionFailClosed, ReasonResourceUnbound}
			}
			out.PeakMemoryBytes = receipt.Value
		case "work-units":
			if receipt.Unit != "units" {
				return ResourceObservation{}, oracleValidation{DecisionFailClosed, ReasonResourceUnbound}
			}
			out.WorkUnits = receipt.Value
		default:
			return ResourceObservation{}, oracleValidation{DecisionFailClosed, ReasonResourceUnbound}
		}
	}
	if len(seen) != 3 {
		return ResourceObservation{}, oracleValidation{DecisionUnknown, ReasonResourceUnbound}
	}
	return out, oracleValidation{}
}

func resourceBindingsEqual(left, right ResourceBindingConfig) bool {
	return left.ProviderID != "" && left.ProviderID == right.ProviderID && left.ObserverID != "" && left.ObserverID == right.ObserverID && left.ProviderDigest != "" && left.ProviderDigest == right.ProviderDigest && left.ObserverDigest != "" && left.ObserverDigest == right.ObserverDigest && left.SnapshotDigest != "" && left.SnapshotDigest == right.SnapshotDigest && left.SourceDigest != "" && left.SourceDigest == right.SourceDigest
}

func validateManifest(input Input, beforeDigest, afterDigest, registryDigest string) oracleValidation {
	manifest := input.Manifest
	if !manifest.Complete {
		return oracleValidation{DecisionUnknown, ReasonRequiredInputMissing}
	}
	expectedBefore := stateSnapshotDigest(input.AuthoritySourceBefore, beforeDigest, registryDigest, input.Config)
	expectedAfter := stateSnapshotDigest(input.AuthoritySourceAfter, afterDigest, registryDigest, input.Config)
	if manifest.BeforeSnapshotDigest != expectedBefore || manifest.AfterSnapshotDigest != expectedAfter || manifest.ToolchainDigest != input.Config.ToolchainDigest || manifest.ProfileDigest != input.Config.Profile.Digest || manifest.RegistryDigest != registryDigest {
		return oracleValidation{DecisionFailClosed, ReasonDigestMismatch}
	}
	if manifest.ZeroChange {
		if len(input.Changes) != 0 || beforeDigest != afterDigest || manifest.BeforeSnapshotDigest != manifest.AfterSnapshotDigest || len(input.Receipts) != 0 || len(input.Path.Edges) != 0 || len(input.Path.Claims) != 0 || len(input.Path.Evidence) != 0 || len(input.Roots) != 0 {
			return oracleValidation{DecisionFailClosed, ReasonChangedSurface}
		}
	}
	return oracleValidation{}
}

func inputObservationCounts(input Input) ObservationCounts {
	counts := ObservationCounts{
		RegistryBindings: uint64(len(input.Registry)), ReceiptRecords: uint64(len(input.Receipts)),
		PathEdges: uint64(len(input.Path.Edges)), PathClaims: uint64(len(input.Path.Claims)), PathEvidence: uint64(len(input.Path.Evidence)),
	}
	for _, change := range input.Changes {
		if change.BeforeDigest != change.AfterDigest {
			counts.ChangedCodeSurfaces++
		}
	}
	for _, edge := range input.Path.Edges {
		switch edge.Kind {
		case semantic.InferenceObservationCandidate:
			counts.CandidateObservations++
		case semantic.InferenceAcceptedLift:
			counts.AcceptedLifts++
		}
	}
	return counts
}

func normalizeRegistry(bindings []CodeBinding) (registryView, oracleValidation) {
	view := registryView{bySurface: make(map[string]CodeBinding), bySymbol: make(map[string]CodeBinding)}
	canonical := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		if !validID(binding.RegisteredSurfaceID) || !validID(binding.CodeSymbolID) || !validID(binding.SemanticOwnerID) ||
			!validToken(binding.SourceMapID) || !validDigest(binding.BindingDigest) {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		if _, exists := view.bySurface[binding.RegisteredSurfaceID]; exists {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		if _, exists := view.bySymbol[binding.CodeSymbolID]; exists {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		if expected := bindingDigest(binding); expected != binding.BindingDigest {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		view.bySurface[binding.RegisteredSurfaceID] = binding
		view.bySymbol[binding.CodeSymbolID] = binding
		canonical = append(canonical, bindingCanonical(binding))
	}
	sort.Strings(canonical)
	view.digest = digestBytes([]byte(strings.Join(canonical, "\n") + "\n"))
	return view, oracleValidation{}
}

func resolveChangedSurfaces(changes []CodeChange, registry registryView) ([]string, oracleValidation) {
	seen := make(map[string]struct{}, len(changes))
	changed := make([]string, 0, len(changes))
	for _, change := range changes {
		if !validID(change.CodeSymbolID) || !validDigest(change.BeforeDigest) || !validDigest(change.AfterDigest) {
			return nil, oracleValidation{DecisionFailClosed, ReasonChangedSurface}
		}
		if _, duplicate := seen[change.CodeSymbolID]; duplicate {
			return nil, oracleValidation{DecisionFailClosed, ReasonChangedSurface}
		}
		seen[change.CodeSymbolID] = struct{}{}
		if change.BeforeDigest == change.AfterDigest {
			continue
		}
		binding, exists := registry.bySymbol[change.CodeSymbolID]
		if !exists {
			return nil, oracleValidation{DecisionFailClosed, ReasonSurfaceUnregistered}
		}
		changed = append(changed, binding.RegisteredSurfaceID)
	}
	sort.Strings(changed)
	return changed, oracleValidation{}
}

func validateSourceBindings(input Input, beforeDigest, afterDigest string) oracleValidation {
	if sourceDigest(input.AuthoritySourceBefore) != beforeDigest && input.AuthoritySourceBefore == "" {
		return oracleValidation{DecisionUnknown, ReasonSourceUnbound}
	}
	if !validDigest(sourceDigest(input.AuthoritySourceBefore)) || !validDigest(sourceDigest(input.AuthoritySourceAfter)) {
		return oracleValidation{DecisionUnknown, ReasonSourceUnbound}
	}
	return oracleValidation{}
}

func validateReceipts(input Input, registry registryView, changed []string, before, after normalizedSemantic, deltaText string) (receiptView, oracleValidation) {
	view := receiptView{bySurface: make(map[string]CouplingReceipt)}
	changedSet := make(map[string]struct{}, len(changed))
	for _, surface := range changed {
		changedSet[surface] = struct{}{}
	}
	seenIDs := make(map[string]struct{}, len(input.Receipts))
	for _, receipt := range input.Receipts {
		if !validID(receipt.ReceiptID) {
			return view, oracleValidation{DecisionFailClosed, ReasonStaleReceipt}
		}
		if receipt.State == "STALE" {
			return view, oracleValidation{DecisionUnknown, ReasonStaleReceipt}
		}
		if receipt.State != "CURRENT" {
			return view, oracleValidation{DecisionFailClosed, ReasonStaleReceipt}
		}
		if _, duplicate := seenIDs[receipt.ReceiptID]; duplicate {
			return view, oracleValidation{DecisionFailClosed, ReasonDuplicateReceipt}
		}
		seenIDs[receipt.ReceiptID] = struct{}{}
		if _, exists := changedSet[receipt.SurfaceID]; !exists {
			return view, oracleValidation{DecisionFailClosed, ReasonOrphanReceipt}
		}
		if _, duplicate := view.bySurface[receipt.SurfaceID]; duplicate {
			return view, oracleValidation{DecisionFailClosed, ReasonDuplicateReceipt}
		}
		binding := registry.bySurface[receipt.SurfaceID]
		if receipt.SemanticOwnerID != binding.SemanticOwnerID || receipt.CodeSymbolID != binding.CodeSymbolID ||
			receipt.SourceMapBindingDigest != binding.BindingDigest {
			return view, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		expectedSnapshot := snapshotDigest(input, before.digest, after.digest, registry.digest)
		if receipt.SnapshotDigest != expectedSnapshot || receipt.RegistryDigest != registry.digest {
			return view, oracleValidation{DecisionUnknown, ReasonStaleReceipt}
		}
		if receipt.ToolchainDigest != input.Config.ToolchainDigest || receipt.ProfileDigest != input.Config.Profile.Digest {
			return view, oracleValidation{DecisionUnknown, ReasonStaleReceipt}
		}
		if receipt.BeforeIRDigest != before.digest || receipt.AfterIRDigest != after.digest {
			return view, oracleValidation{DecisionUnknown, ReasonStaleReceipt}
		}
		if receipt.AuthoritySourceBeforeDigest != sourceDigest(input.AuthoritySourceBefore) || receipt.AuthoritySourceAfterDigest != sourceDigest(input.AuthoritySourceAfter) {
			return view, oracleValidation{DecisionUnknown, ReasonStaleReceipt}
		}
		if issue := validateReceiptClaim(receipt, before.digest, after.digest, deltaText); issue.decision != "" {
			return view, issue
		}
		view.bySurface[receipt.SurfaceID] = receipt
		view.valid = append(view.valid, receipt.SurfaceID)
	}
	if len(view.valid) != len(changed) {
		for _, surface := range changed {
			if _, exists := view.bySurface[surface]; !exists {
				return view, oracleValidation{DecisionUnknown, ReasonMissingReceipt}
			}
		}
		return view, oracleValidation{DecisionUnknown, ReasonMissingReceipt}
	}
	sort.Strings(view.valid)
	return view, oracleValidation{}
}

func validateReceiptClaim(receipt CouplingReceipt, beforeDigest, afterDigest, deltaText string) oracleValidation {
	switch receipt.ChangeClaim {
	case ClaimDelta:
		if receipt.ReceiptKind != ReceiptSemanticDelta || beforeDigest == afterDigest {
			return oracleValidation{DecisionFailClosed, ReasonInvalidDelta}
		}
		if receipt.SemanticDelta != deltaText || receipt.SemanticDelta == "" || receipt.SemanticDeltaDigest != digestBytes([]byte(deltaText)) {
			return oracleValidation{DecisionFailClosed, ReasonInvalidDelta}
		}
		if receipt.AuthoritativeSourceRef == "" || receipt.AuthoritySourceAfterDigest == "" {
			return oracleValidation{DecisionFailClosed, ReasonDeltaWithoutSource}
		}
	case ClaimNoDelta:
		if receipt.ReceiptKind != ReceiptNoSemanticDelta || beforeDigest != afterDigest {
			return oracleValidation{DecisionFailClosed, ReasonNoDeltaWithoutEquality}
		}
		if receipt.SemanticDelta != "" || receipt.SemanticDeltaDigest != "" || receipt.AuthoritativeSourceRef != "" {
			return oracleValidation{DecisionFailClosed, ReasonNoDeltaWithoutEquality}
		}
	default:
		return oracleValidation{DecisionFailClosed, ReasonInvalidDelta}
	}
	if len(receipt.EvidenceRefs) == 0 || receipt.OriginPathID == "" || receipt.ClaimRecordID == "" {
		return oracleValidation{DecisionFailClosed, ReasonPathMalformed}
	}
	return oracleValidation{}
}

func normalizeSemantic(ir SemanticIR) (normalizedSemantic, error) {
	facts := make([]string, 0, len(ir.Nodes)+len(ir.Relations))
	nodes := make(map[string]struct{}, len(ir.Nodes))
	for _, node := range ir.Nodes {
		if !validID(node.ID) || !validToken(node.Kind) || !validToken(node.Namespace) {
			return normalizedSemantic{}, fmt.Errorf("invalid semantic node")
		}
		if node.Kind != semantic.Entity.String() && node.Kind != semantic.Activity.String() && node.Kind != semantic.Agent.String() {
			return normalizedSemantic{}, fmt.Errorf("invalid semantic kind")
		}
		if _, duplicate := nodes[node.ID]; duplicate {
			return normalizedSemantic{}, fmt.Errorf("duplicate semantic node")
		}
		nodes[node.ID] = struct{}{}
		facts = append(facts, "node\t"+node.ID+"\t"+node.Kind+"\t"+node.Namespace)
	}
	for _, relation := range ir.Relations {
		if !validID(relation.Subject) || !validID(relation.Object) || !validToken(relation.Predicate) {
			return normalizedSemantic{}, fmt.Errorf("invalid semantic relation")
		}
		if _, ok := nodes[relation.Subject]; !ok {
			return normalizedSemantic{}, fmt.Errorf("relation subject is not registered")
		}
		if _, ok := nodes[relation.Object]; !ok {
			return normalizedSemantic{}, fmt.Errorf("relation object is not registered")
		}
		facts = append(facts, "relation\t"+relation.Subject+"\t"+relation.Predicate+"\t"+relation.Object)
	}
	sort.Strings(facts)
	seen := make(map[string]struct{}, len(facts))
	for _, fact := range facts {
		if _, duplicate := seen[fact]; duplicate {
			return normalizedSemantic{}, fmt.Errorf("duplicate semantic fact")
		}
		seen[fact] = struct{}{}
	}
	canonical := strings.Join(append([]string{"semantic-ir-v1"}, facts...), "\n") + "\n"
	return normalizedSemantic{digest: digestBytes([]byte(canonical)), facts: facts}, nil
}

func semanticDelta(before, after []string) (string, []string, []string) {
	left, right := make(map[string]struct{}, len(before)), make(map[string]struct{}, len(after))
	for _, fact := range before {
		left[fact] = struct{}{}
	}
	for _, fact := range after {
		right[fact] = struct{}{}
	}
	added, removed := make([]string, 0), make([]string, 0)
	for fact := range right {
		if _, ok := left[fact]; !ok {
			added = append(added, fact)
		}
	}
	for fact := range left {
		if _, ok := right[fact]; !ok {
			removed = append(removed, fact)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	if len(added) == 0 && len(removed) == 0 {
		return "", added, removed
	}
	var b strings.Builder
	b.WriteString("semantic-delta-v1\n")
	for _, fact := range removed {
		b.WriteString("removed\t")
		b.WriteString(fact)
		b.WriteByte('\n')
	}
	for _, fact := range added {
		b.WriteString("added\t")
		b.WriteString(fact)
		b.WriteByte('\n')
	}
	return b.String(), added, removed
}

func bindingDigest(binding CodeBinding) string {
	return digestBytes([]byte("binding-v1\n" + binding.RegisteredSurfaceID + "\n" + binding.CodeSymbolID + "\n" + binding.SemanticOwnerID + "\n" + binding.SourceMapID + "\n"))
}

func bindingCanonical(binding CodeBinding) string {
	return binding.RegisteredSurfaceID + "\t" + binding.CodeSymbolID + "\t" + binding.SemanticOwnerID + "\t" + binding.SourceMapID + "\t" + binding.BindingDigest
}

func snapshotDigest(input Input, beforeDigest, afterDigest, registryDigest string) string {
	return digestBytes([]byte("snapshot-v1\n" + sourceDigest(input.AuthoritySourceBefore) + "\n" + sourceDigest(input.AuthoritySourceAfter) + "\n" + beforeDigest + "\n" + afterDigest + "\n" + registryDigest + "\n" + input.Config.ToolchainDigest + "\n" + input.Config.Profile.ID + "\n" + input.Config.Profile.Version + "\n" + input.Config.Profile.Digest + "\n"))
}

func stateSnapshotDigest(source string, semanticDigest, registryDigest string, config EvaluationConfig) string {
	return digestBytes([]byte("state-snapshot-v1\n" + sourceDigest(source) + "\n" + semanticDigest + "\n" + registryDigest + "\n" + config.ToolchainDigest + "\n" + config.Profile.ID + "\n" + config.Profile.Version + "\n" + config.Profile.Digest + "\n"))
}

func sourceDigest(source string) string { return digestBytes([]byte(source)) }

func resourceBindingDigest(receipt ExternalResourceReceipt) string {
	return digestBytes([]byte("resource-binding-v1\n" + receipt.ReceiptID + "\n" + receipt.Metric + "\n" + fmt.Sprint(receipt.Value) + "\n" + receipt.Unit + "\n" + receipt.ProviderDigest + "\n" + receipt.ObserverDigest + "\n" + receipt.SnapshotDigest + "\n" + receipt.SourceDigest + "\n"))
}

func resourceProviderDigest(id string) string {
	return digestBytes([]byte("resource-provider-v1\n" + id + "\n"))
}

func resourceObserverDigest(id string) string {
	return digestBytes([]byte("resource-observer-v1\n" + id + "\n"))
}

func resourceSourceDigest(providerID, observerID, snapshot string) string {
	return digestBytes([]byte("resource-source-v1\n" + providerID + "\n" + observerID + "\n" + snapshot + "\n"))
}

func validID(value string) bool {
	_, err := semantic.ParseIdentity(value)
	return err == nil
}

func validToken(value string) bool {
	if value == "" || value != strings.TrimSpace(value) {
		return false
	}
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}
