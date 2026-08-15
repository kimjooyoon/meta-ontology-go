package semantic

import (
	"errors"
	"fmt"
	"sort"
	"strings"
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

func (e InferenceEdge) Normalized() (InferenceEdge, error) { return e.normalized() }

func (e InferenceEdge) Validate() error { _, err := e.normalized(); return err }

func (c SemanticChangeClaim) normalized() (SemanticChangeClaim, error) {
	base, err := c.InferenceRecord.normalized()
	if err != nil {
		return SemanticChangeClaim{}, err
	}
	if !c.Kind.Valid() {
		return SemanticChangeClaim{}, fmt.Errorf("unknown semantic change kind %q", c.Kind)
	}
	if base.Authority.Layer != AuthoritySemantic {
		return SemanticChangeClaim{}, errors.New("semantic change claim must bind the semantic authority layer")
	}
	delta := strings.TrimSpace(c.CanonicalDelta)
	digest := strings.TrimSpace(c.DeltaDigest)
	if base.Before.Semantic == "" || base.After.Semantic == "" {
		return SemanticChangeClaim{}, errors.New("semantic change claim requires before and after semantic digests")
	}
	switch c.Kind {
	case SemanticDelta:
		if base.Before.Semantic == base.After.Semantic {
			return SemanticChangeClaim{}, errors.New("semantic delta requires unequal semantic snapshot digests")
		}
		if delta == "" || digest == "" {
			return SemanticChangeClaim{}, errors.New("semantic delta requires a canonical delta and digest")
		}
		if StableHashString(delta) != digest {
			return SemanticChangeClaim{}, errors.New("semantic delta digest does not match its canonical delta")
		}
	case NoSemanticDelta:
		if base.Before.Semantic != base.After.Semantic {
			return SemanticChangeClaim{}, errors.New("no semantic delta requires equal semantic snapshot digests")
		}
		if delta != "" || digest != "" {
			return SemanticChangeClaim{}, errors.New("no semantic delta cannot carry a canonical delta")
		}
		if base.Authority.Effect != AuthorityNoChange {
			return SemanticChangeClaim{}, errors.New("no semantic delta must use the no-change authority effect")
		}
	}
	if c.Kind == SemanticDelta && base.Authority.Effect != AuthorityDelta {
		return SemanticChangeClaim{}, errors.New("semantic delta must use the semantic-delta authority effect")
	}
	c.InferenceRecord = base
	c.CanonicalDelta = delta
	c.DeltaDigest = digest
	return c, nil
}

func (c SemanticChangeClaim) Normalized() (SemanticChangeClaim, error) { return c.normalized() }

func (c SemanticChangeClaim) Validate() error { _, err := c.normalized(); return err }

type inferenceContract struct {
	phase                    InferencePhase
	layer                    AuthorityLayer
	effect                   AuthorityEffect
	catalog, policy, profile bool
}

func inferenceContractFor(kind InferenceKind) inferenceContract {
	switch kind {
	case InferenceAuthoritativeDeclaration:
		return inferenceContract{PhaseDeclaration, AuthoritySource, AuthorityDeclare, false, false, false}
	case InferenceDeterministicDerivation:
		return inferenceContract{PhaseDerivation, AuthoritySemantic, AuthorityDerive, false, false, false}
	case InferenceDerivedProjection:
		return inferenceContract{PhaseProjection, AuthorityDerived, AuthorityProject, false, false, true}
	case InferenceObservationCandidate:
		return inferenceContract{PhaseObservation, AuthorityCandidate, AuthorityObserve, true, false, false}
	case InferenceAcceptedLift:
		return inferenceContract{PhaseLift, AuthoritySemantic, AuthorityLift, false, true, false}
	default:
		return inferenceContract{PhaseVerification, AuthorityVerification, AuthorityVerify, false, true, false}
	}
}
