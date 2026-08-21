package query

import (
	"encoding/json"
	"fmt"
)

func (request ResolutionRequest) Normalize() (ResolutionRequest, FactSelection, error) {
	if request.Schema != ResolutionSchema {
		return ResolutionRequest{}, 0, envelopeError(
			ErrInvalidResolution, "invalid_resolution_schema", "schema must be gooo-query/resolution/v1",
		)
	}
	business, err := ParseID(request.Business.String())
	if err != nil {
		return ResolutionRequest{}, 0, envelopeError(ErrInvalidResolution, "invalid_business", err.Error())
	}
	if request.MaxDepth != ResolutionMaxDepth {
		return ResolutionRequest{}, 0, envelopeError(
			ErrInvalidResolution, "invalid_resolution_depth", fmt.Sprintf("must be exactly %d", ResolutionMaxDepth),
		)
	}
	if request.Limit < 1 || request.Limit > MaxEnvelopeLimit {
		return ResolutionRequest{}, 0, envelopeError(
			ErrInvalidResolution, "invalid_resolution_limit", fmt.Sprintf("must be 1..%d", MaxEnvelopeLimit),
		)
	}
	layer, selection, err := normalizeLayer(request.Layer)
	if err != nil {
		return ResolutionRequest{}, 0, err
	}
	request.Business = ID(business)
	request.Layer = layer
	return request, selection, nil
}
func (request ResolutionRequest) CanonicalJSON() ([]byte, error) {
	normalized, _, err := request.Normalize()
	if err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}
func (request ResolutionRequest) CanonicalDigest() (string, error) {
	canonical, err := request.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}
