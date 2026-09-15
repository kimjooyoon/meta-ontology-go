package query

import (
	"encoding/json"
	"errors"
)

func (response ResolutionResponse) CanonicalJSON() ([]byte, error) {
	canonical := response
	canonical.Hash = ""
	return json.Marshal(canonical)
}

func (response ResolutionResponse) CanonicalDigest() (string, error) {
	canonical, err := response.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}

func (response *ResolutionResponse) seal() error {
	digest, err := response.CanonicalDigest()
	if err != nil {
		return err
	}
	response.Hash = digest
	return nil
}

func (graph Graph) rejectedResolution(request ResolutionRequest, err error) (ResolutionResponse, error) {
	response := ResolutionResponse{
		Schema: ResolutionSchema, Status: ResponseError, Request: request,
		Metadata: resolutionMetadata(graph.Metadata(), ResponseError),
		Error:    &EnvelopeError{Code: resolutionErrorCode(err), Message: err.Error()},
	}
	if normalized, _, normalizeErr := request.Normalize(); normalizeErr == nil {
		response.Request = normalized
		response.RequestHash, _ = normalized.CanonicalDigest()
	}
	if sealErr := response.seal(); sealErr != nil {
		return ResolutionResponse{}, sealErr
	}
	return response, err
}

func resolutionErrorCode(err error) string {
	if envelope, ok := errors.AsType[*EnvelopeError](err); ok {
		return envelope.Code
	}
	if errors.Is(err, ErrUnknownEndpoint) {
		return "unknown_endpoint"
	}
	if errors.Is(err, ErrInvalidResolution) {
		return "invalid_resolution"
	}
	return "query_rejected"
}
