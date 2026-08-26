package semantic

import (
	"errors"
	"fmt"
	"strings"
)

func (e InferenceEdge) Normalized() (InferenceEdge, error) { return e.normalized() }
func (e InferenceEdge) Validate() error                    { _, err := e.normalized(); return err }
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
func (c SemanticChangeClaim) Validate() error                          { _, err := c.normalized(); return err }

type inferenceContract struct {
	phase                    InferencePhase
	layer                    AuthorityLayer
	effect                   AuthorityEffect
	catalog, policy, profile bool
}
