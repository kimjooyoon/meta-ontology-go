package verify

import (
	"encoding/json"
	"fmt"
	"sort"
)

// VerifyCoupling evaluates one immutable snapshot without reading or writing
// repository state. Unknown inputs remain UNKNOWN; no N/A path exists.
func VerifyCoupling(input CouplingInput) CouplingEvidence {
	evidence := CouplingEvidence{
		Schema: CouplingEvidenceSchemaVersion, Observer: CouplingEvidenceSchemaVersion,
		ObserverAvailable: CouplingObserverAvailable, RawDecision: CouplingDecisionUnknown,
		Enforcement:    CouplingEnforcementBlock,
		EnvelopeDigest: input.Envelope.TupleDigest(), RegistryDigest: input.Envelope.RegistryDigest,
	}
	if input.Schema != CouplingEvidenceSchemaVersion {
		return withFailures(evidence, couplingFailure("missing-input", CouplingOwnerUnavailable, "observer input schema is missing or stale"))
	}
	if err := validateEnvelope(input.Envelope); err != nil {
		return withFailures(evidence, couplingFailure("wrong-tuple", CouplingOwnerUnavailable, err.Error()))
	}
	registry, err := input.Registry.Normalized()
	if err != nil {
		return withFailures(evidence, couplingFailure("registry-invalid", CouplingOwnerUnavailable, err.Error()))
	}
	if registry.Digest() != input.Envelope.RegistryDigest {
		return withFailures(evidence, couplingFailure("registry-mismatch", CouplingOwnerUnavailable, "envelope registry digest does not match canonical registry"))
	}
	surfaces, err := registry.resolve(input.ChangedSites)
	if err != nil {
		return withFailures(evidence, couplingFailure(reasonFromError(err), CouplingOwnerUnavailable, err.Error()))
	}
	evidence.RegistryDigest = registry.Digest()
	evidence.SurfaceResults = make([]CouplingSurfaceResult, 0, len(surfaces))
	bySurface := make(map[string]CouplingSurface, len(surfaces))
	for _, surface := range surfaces {
		bySurface[surface.SurfaceID] = surface
		evidence.SurfaceResults = append(evidence.SurfaceResults, CouplingSurfaceResult{SurfaceID: surface.SurfaceID, Decision: CouplingDecisionPass})
	}
	receipts := make(map[string][]CouplingReceipt)
	for _, receipt := range input.Receipts {
		receipts[receipt.SurfaceID] = append(receipts[receipt.SurfaceID], receipt)
		if _, ok := bySurface[receipt.SurfaceID]; !ok {
			evidence.Failures = append(evidence.Failures, couplingFailure("orphan-receipt", CouplingOwnerUnavailable, receipt.SurfaceID))
		}
	}
	for _, surface := range surfaces {
		matches := receipts[surface.SurfaceID]
		if len(matches) == 0 {
			evidence.Failures = append(evidence.Failures, couplingFailure("missing-receipt", surface.SemanticOwnerID, surface.SurfaceID))
			setSurfaceDecision(&evidence, surface.SurfaceID, CouplingDecisionFailClosed, "missing-receipt")
			continue
		}
		if len(matches) != 1 {
			evidence.Failures = append(evidence.Failures, couplingFailure("duplicate-receipt", surface.SemanticOwnerID, surface.SurfaceID))
			setSurfaceDecision(&evidence, surface.SurfaceID, CouplingDecisionFailClosed, "duplicate-receipt")
			continue
		}
		receipt := matches[0]
		failures := validateReceipt(receipt, surface, input.Envelope)
		if len(failures) > 0 {
			evidence.Failures = append(evidence.Failures, failures...)
			for _, failure := range failures {
				setSurfaceDecision(&evidence, surface.SurfaceID, CouplingDecisionFailClosed, failure.Code)
			}
		}
		evidence.Receipts = append(evidence.Receipts, receipt.Evidence())
	}
	evidence.Failures = uniqueFailures(evidence.Failures)
	if len(evidence.Failures) == 0 {
		evidence.RawDecision = CouplingDecisionPass
		evidence.Enforcement = CouplingEnforcementAllow
	} else {
		evidence.RawDecision = CouplingDecisionFailClosed
		evidence.Enforcement = CouplingEnforcementBlock
	}
	sort.Slice(evidence.SurfaceResults, func(i, j int) bool {
		return evidence.SurfaceResults[i].SurfaceID < evidence.SurfaceResults[j].SurfaceID
	})
	sort.Slice(evidence.Receipts, func(i, j int) bool { return evidence.Receipts[i].SurfaceID < evidence.Receipts[j].SurfaceID })
	return evidence
}

func withFailures(evidence CouplingEvidence, failures ...CouplingFailure) CouplingEvidence {
	evidence.Failures = uniqueFailures(failures)
	return evidence
}

func setSurfaceDecision(evidence *CouplingEvidence, surfaceID, decision, reason string) {
	for index := range evidence.SurfaceResults {
		if evidence.SurfaceResults[index].SurfaceID != surfaceID {
			continue
		}
		evidence.SurfaceResults[index].Decision = decision
		evidence.SurfaceResults[index].ReasonCodes = append(evidence.SurfaceResults[index].ReasonCodes, reason)
		return
	}
}

func reasonFromError(err error) string {
	message := err.Error()
	for _, reason := range []string{"surface-unregistered", "ambiguous-origin", "surface-not-applicable", "no-changed-sites", "registry-invalid", "missing-input"} {
		if len(message) >= len(reason) && message[:len(reason)] == reason {
			return reason
		}
	}
	return "missing-input"
}

func (e CouplingEvidence) ValidateCanonical() error {
	if e.Schema != CouplingEvidenceSchemaVersion || e.ObserverAvailable != CouplingObserverAvailable {
		return fmt.Errorf("invalid coupling evidence envelope")
	}
	if e.RawDecision != CouplingDecisionPass && e.RawDecision != CouplingDecisionFailClosed && e.RawDecision != CouplingDecisionUnknown {
		return fmt.Errorf("invalid raw coupling decision")
	}
	if e.Enforcement != CouplingEnforcementAllow && e.Enforcement != CouplingEnforcementBlock {
		return fmt.Errorf("invalid coupling enforcement")
	}
	for _, receipt := range e.Receipts {
		payload, err := jsonPayload(receipt)
		if err != nil || payload != receipt.CanonicalPayload {
			return fmt.Errorf("noncanonical receipt evidence %q", receipt.ReceiptID)
		}
	}
	return nil
}

func jsonPayload(receipt CouplingReceiptEvidence) (string, error) {
	data, err := json.Marshal(receipt.couplingReceiptPayload)
	return string(data), err
}
