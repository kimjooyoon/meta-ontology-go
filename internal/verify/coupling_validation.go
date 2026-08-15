package verify

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateEnvelope(e CouplingEnvelope) error {
	if e.Schema != CouplingEnvelopeSchemaVersion || e.ContractDigest != CouplingContractDigest || e.SchemaDigest != semantic.StableHashString(CouplingEvidenceSchemaVersion) || e.SemanticPathSchema != semantic.InferencePathSchemaVersion {
		return fmt.Errorf("invalid coupling envelope schema digest")
	}
	if e.Repository == "" || e.Event == "" || e.Ref == "" || e.EventRef == "" || e.BaseRef == "" || e.CheckoutRef != e.HeadSHA || e.BaseSHA == e.HeadSHA || e.PRNumber <= 0 || e.RunID <= 0 || e.RunAttempt <= 0 {
		return fmt.Errorf("incomplete or mismatched CI tuple")
	}
	if e.ContractDigest == "" || !validCommit(e.BaseSHA) || !validCommit(e.HeadSHA) || !validCommit(e.WorkflowSHA) || !validCommit(e.SemanticSourceHead) || e.SemanticSourceHead != CouplingSemanticSourceHead {
		return fmt.Errorf("invalid revision or semantic foundation binding")
	}
	for _, digest := range []string{e.CatalogDigest, e.PolicyDigest, e.RegistryDigest, e.ProfileDigest, e.ToolchainDigest, e.SnapshotDigest} {
		if !validHexDigest(digest) {
			return fmt.Errorf("missing or malformed envelope digest")
		}
	}
	return nil
}

func validateReceipt(r CouplingReceipt, surface CouplingSurface, envelope CouplingEnvelope) []CouplingFailure {
	failures := make([]CouplingFailure, 0)
	owner := surface.SemanticOwnerID
	need := func(condition bool, reason, detail string) {
		if !condition {
			failures = append(failures, couplingFailure(reason, owner, detail))
		}
	}
	need(r.ReceiptID != "", "missing-receipt-id", "receipt ID is empty")
	need(r.SurfaceID == surface.SurfaceID, "surface-mismatch", "receipt surface does not match registry")
	need(r.SemanticOwnerID == surface.SemanticOwnerID, "surface-owner-mismatch", "receipt owner does not match registry")
	need(r.CodeSymbolID == surface.CodeSymbolID, "surface-symbol-mismatch", "receipt symbol does not match registry")
	need(r.EnvelopeDigest == envelope.TupleDigest(), "wrong-tuple", "receipt is not bound to the current envelope")
	need(r.SnapshotDigest == envelope.SnapshotDigest, "stale-snapshot", "receipt snapshot is stale")
	need(r.RegistryDigest == envelope.RegistryDigest, "registry-mismatch", "receipt registry digest differs")
	need(r.ProfileDigest == surface.ProfileDigest && r.ProfileDigest == envelope.ProfileDigest, "profile-mismatch", "receipt profile differs")
	need(r.ToolchainDigest == surface.ToolchainDigest && r.ToolchainDigest == envelope.ToolchainDigest, "toolchain-mismatch", "receipt toolchain differs")
	need(r.SourceMapBindingDigest == surface.SourceMapBindingDigest, "source-map-mismatch", "receipt source map binding differs")
	need(validHexDigest(r.BeforeIRDigest) && validHexDigest(r.AfterIRDigest), "digest-mismatch", "IR digests are malformed")
	need(validHexDigest(r.RuleDigest) && containsCouplingString(surface.RuleDigests, r.RuleDigest), "rule-mismatch", "receipt rule is not registered")
	need(r.State == CouplingCurrent, "stale-snapshot", "receipt is not current")
	need(r.ChangeClaim == "DELTA" || r.ChangeClaim == "NO_DELTA", "invalid-claim", "unknown change claim")
	need(r.ReceiptKind == expectedReceiptKind(r.ChangeClaim), "claim-kind-mismatch", "claim and receipt kind disagree")
	need(validHexDigest(r.SourceMapBindingDigest), "source-map-mismatch", "source map binding digest is malformed")
	need(r.PathDigest != "" && r.PathDigest == r.Path.StableHash(), "path-digest-mismatch", "typed path digest is not canonical")
	need(r.CanonicalPayload != "", "noncanonical-evidence", "canonical payload is missing")
	if payload, err := r.ExpectedCanonicalPayload(); err != nil || payload != r.CanonicalPayload {
		failures = append(failures, couplingFailure("noncanonical-evidence", owner, "receipt payload is not canonical"))
	}
	if r.ChangeClaim == "DELTA" {
		need(r.BeforeIRDigest != r.AfterIRDigest, "delta-without-change", "DELTA requires unequal IR digests")
		need(r.CanonicalDelta != "" && validHexDigest(r.DeltaDigest) && semantic.StableHashString(r.CanonicalDelta) == r.DeltaDigest, "delta-without-source", "DELTA requires a canonical delta digest")
		need(validHexDigest(r.AuthoritySourceBeforeDigest) && validHexDigest(r.AuthoritySourceAfterDigest) && r.AuthoritySourceBeforeDigest != r.AuthoritySourceAfterDigest, "delta-without-source", "DELTA requires an updated authoritative source")
	} else {
		need(r.BeforeIRDigest == r.AfterIRDigest, "no-delta-without-equality", "NO_DELTA requires equal IR digests")
		need(r.CanonicalDelta == "" && r.DeltaDigest == "", "no-delta-without-equality", "NO_DELTA cannot carry a delta")
		need(validHexDigest(r.AuthoritySourceBeforeDigest) && validHexDigest(r.AuthoritySourceAfterDigest) && r.AuthoritySourceBeforeDigest == r.AuthoritySourceAfterDigest, "no-delta-without-equality", "NO_DELTA requires equal authoritative source snapshots")
	}
	failures = append(failures, validatePath(r, surface, envelope)...)
	return uniqueFailures(failures)
}

func validatePath(r CouplingReceipt, surface CouplingSurface, envelope CouplingEnvelope) []CouplingFailure {
	failures := make([]CouplingFailure, 0)
	owner := surface.SemanticOwnerID
	if len(r.Path.Edges) == 0 || len(r.Path.Claims) != 1 {
		return []CouplingFailure{couplingFailure("path-incomplete", owner, "one finite path and one change claim are required")}
	}
	if err := r.Path.Validate(); err != nil {
		if r.ChangeClaim == "DELTA" && hasCandidatePath(r.Path.Edges) && !hasAuthorityPath(r.Path.Edges) {
			return []CouplingFailure{couplingFailure("observation-promotion", owner, "candidate observation was presented as authority")}
		}
		return []CouplingFailure{couplingFailure("path-incomplete", owner, err.Error())}
	}
	chain, err := semantic.NewInferencePathChain(r.Path.Edges...)
	if err != nil || len(chain.Edges) == 0 || chain.Edges[len(chain.Edges)-1].Kind != semantic.InferenceIndependentVerification || chain.Edges[len(chain.Edges)-1].ObjectID.String() != surface.SemanticOwnerID {
		return []CouplingFailure{couplingFailure("path-incomplete", owner, "typed path does not close at the semantic owner")}
	}
	claim := r.Path.Claims[0]
	wantKind := semantic.NoSemanticDelta
	if r.ChangeClaim == "DELTA" {
		wantKind = semantic.SemanticDelta
	}
	if claim.Kind != wantKind || claim.Before.Semantic != r.BeforeIRDigest || claim.After.Semantic != r.AfterIRDigest || claim.CanonicalDelta != r.CanonicalDelta || claim.DeltaDigest != r.DeltaDigest || claim.ObjectID.String() != surface.SemanticOwnerID {
		failures = append(failures, couplingFailure("claim-kind-mismatch", owner, "semantic claim does not match receipt"))
	}
	if r.ChangeClaim == "DELTA" && !hasAuthorityPath(r.Path.Edges) {
		failures = append(failures, couplingFailure("observation-promotion", owner, "candidate observation did not reach accepted authority"))
	}
	if !sameCouplingStrings(r.OriginPathIDs, edgeIDs(r.Path.Edges)) {
		failures = append(failures, couplingFailure("path-incomplete", owner, "origin path IDs do not match typed edges"))
	}
	if !sameCouplingStrings(r.EvidenceRefs, evidenceIDs(r.Path.Evidence)) {
		failures = append(failures, couplingFailure("missing-evidence", owner, "receipt evidence references do not match the path"))
	}
	for _, edge := range r.Path.Edges {
		if edge.Rule.Digest != r.RuleDigest || edge.Before.Semantic != r.BeforeIRDigest || edge.After.Semantic != r.AfterIRDigest {
			failures = append(failures, couplingFailure("path-binding-mismatch", owner, "typed edge is not bound to receipt digests"))
		}
		if edge.Before.Source != r.AuthoritySourceBeforeDigest || edge.After.Source != r.AuthoritySourceAfterDigest {
			failures = append(failures, couplingFailure("source-binding-mismatch", owner, "typed edge source snapshots are not bound to receipt digests"))
		}
		if !couplingControlsMatch(edge.Controls, surface, envelope) {
			failures = append(failures, couplingFailure("controls-mismatch", owner, "typed edge controls are not bound to the current envelope or profile"))
		}
	}
	for _, claim := range r.Path.Claims {
		if claim.Rule.Digest != r.RuleDigest {
			failures = append(failures, couplingFailure("path-binding-mismatch", owner, "semantic claim is not bound to registry controls"))
		}
		if claim.Before.Source != r.AuthoritySourceBeforeDigest || claim.After.Source != r.AuthoritySourceAfterDigest {
			failures = append(failures, couplingFailure("source-binding-mismatch", owner, "semantic claim source snapshots are not bound to receipt digests"))
		}
		if !couplingControlsMatch(claim.Controls, surface, envelope) {
			failures = append(failures, couplingFailure("controls-mismatch", owner, "semantic claim controls are not bound to the current envelope or profile"))
		}
	}
	return uniqueFailures(failures)
}

func couplingControlsMatch(controls semantic.InferenceControls, surface CouplingSurface, envelope CouplingEnvelope) bool {
	if controls.CatalogDigest != "" && controls.CatalogDigest != envelope.CatalogDigest {
		return false
	}
	if controls.PolicyDigest != "" && controls.PolicyDigest != envelope.PolicyDigest {
		return false
	}
	if controls.Profile.Digest == "" {
		return controls.Profile.ID == "" && controls.Profile.Version == ""
	}
	return controls.Profile.ID == surface.ProfileID && controls.Profile.Version == surface.ProfileVersion &&
		controls.Profile.Digest == surface.ProfileDigest && controls.Profile.Digest == envelope.ProfileDigest
}

func hasAuthorityPath(edges []semantic.InferenceEdge) bool {
	for _, edge := range edges {
		if edge.Kind == semantic.InferenceAuthoritativeDeclaration || edge.Kind == semantic.InferenceAcceptedLift {
			return true
		}
	}
	return false
}

func hasCandidatePath(edges []semantic.InferenceEdge) bool {
	for _, edge := range edges {
		if edge.Kind == semantic.InferenceObservationCandidate {
			return true
		}
	}
	return false
}

func edgeIDs(edges []semantic.InferenceEdge) []string {
	result := make([]string, 0, len(edges))
	for _, edge := range edges {
		result = append(result, edge.RecordID.String())
	}
	return sortedStrings(result)
}

func evidenceIDs(evidence []semantic.InferenceEvidence) []string {
	result := make([]string, 0, len(evidence))
	for _, record := range evidence {
		result = append(result, record.ID.String())
	}
	return sortedStrings(result)
}

func expectedReceiptKind(claim string) string {
	if claim == "DELTA" {
		return semantic.SemanticDelta.String()
	}
	return semantic.NoSemanticDelta.String()
}

func containsCouplingString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sameCouplingStrings(left, right []string) bool {
	left = sortedStrings(left)
	right = sortedStrings(right)
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

func uniqueFailures(values []CouplingFailure) []CouplingFailure {
	seen := make(map[string]struct{}, len(values))
	result := make([]CouplingFailure, 0, len(values))
	for _, value := range values {
		key := value.Code + "\x00" + value.Owner + "\x00" + value.Detail
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Code != result[j].Code {
			return result[i].Code < result[j].Code
		}
		return result[i].Detail < result[j].Detail
	})
	return result
}

func couplingFailure(reason, owner, detail string) CouplingFailure {
	domain, retry := couplingFailureRoute(reason)
	if strings.TrimSpace(owner) == "" {
		owner = CouplingOwnerUnavailable
	}
	return CouplingFailure{Code: CouplingFailureCodePrefix + reason, Domain: domain, Owner: owner, Retry: retry, Detail: detail}
}

func couplingFailureRoute(reason string) (string, bool) {
	switch reason {
	case "registry-mismatch", "registry-invalid", "noncanonical-evidence", "surface-owner-mismatch", "observation-promotion", "source-binding-mismatch", "controls-mismatch":
		return CouplingDomainIntegrity, false
	case "missing-receipt", "orphan-receipt", "duplicate-receipt", "surface-mismatch", "surface-symbol-mismatch", "delta-without-change", "delta-without-source", "no-delta-without-equality":
		return CouplingDomainFeature, false
	default:
		return CouplingDomainDependency, true
	}
}
