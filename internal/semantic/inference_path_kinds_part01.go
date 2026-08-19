package semantic

import (
	"errors"
	"fmt"
	"sort"
)

func (e InferenceEdge) normalized() (InferenceEdge, error) {
	base, err := e.InferenceRecord.normalized()
	if err != nil {
		return InferenceEdge{}, err
	}
	if !e.Kind.Valid() {
		return InferenceEdge{}, fmt.Errorf("unknown inference kind %q", e.Kind)
	}
	contract := inferenceContractFor(e.Kind)
	if base.Phase.Phase != contract.phase || base.Authority.Layer != contract.layer ||
		base.Authority.Effect != contract.effect {
		return InferenceEdge{}, errors.New("phase or authority binding does not match inference kind")
	}
	if contract.catalog && base.Controls.CatalogDigest == "" {
		return InferenceEdge{}, errors.New("catalog digest is required for this inference kind")
	}
	if contract.policy && base.Controls.PolicyDigest == "" {
		return InferenceEdge{}, errors.New("policy digest is required for this inference kind")
	}
	if contract.profile && base.Controls.Profile.ID == "" {
		return InferenceEdge{}, errors.New("profile binding is required for this inference kind")
	}
	if e.Kind == InferenceObservationCandidate {
		if base.Before.Semantic == "" || base.After.Semantic == "" {
			return InferenceEdge{}, errors.New("candidate observation requires semantic authority snapshots")
		}
		if base.Before.Semantic != base.After.Semantic {
			return InferenceEdge{}, errors.New("candidate observation cannot change the semantic authority snapshot")
		}
	}
	roots := append([]ID(nil), e.SourceRoots...)
	if e.Kind == InferenceAuthoritativeDeclaration && len(roots) == 0 {
		return InferenceEdge{}, errors.New("authoritative declaration requires a source root")
	}
	seenRoots := make(map[ID]struct{}, len(roots))
	for i, root := range roots {
		parsed, parseErr := ParseIdentity(root.String())
		if parseErr != nil {
			return InferenceEdge{}, fmt.Errorf("source root: %w", parseErr)
		}
		if _, exists := seenRoots[parsed]; exists {
			return InferenceEdge{}, fmt.Errorf("duplicate source root %s", parsed)
		}
		seenRoots[parsed] = struct{}{}
		roots[i] = parsed
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i] < roots[j] })
	receipt := e.AcceptanceReceipt
	if receipt != "" {
		parsedReceipt, parseErr := ParseIdentity(receipt.String())
		if parseErr != nil {
			return InferenceEdge{}, fmt.Errorf("acceptance receipt: %w", parseErr)
		}
		receipt = parsedReceipt
	}
	if e.Kind == InferenceAcceptedLift && receipt == "" {
		return InferenceEdge{}, errors.New("accepted lift requires an acceptance receipt")
	}
	if e.Kind != InferenceAcceptedLift && receipt != "" {
		return InferenceEdge{}, errors.New("acceptance receipt is only valid for an accepted lift")
	}
	e.InferenceRecord = base
	e.SourceRoots = roots
	e.AcceptanceReceipt = receipt
	return e, nil
}
