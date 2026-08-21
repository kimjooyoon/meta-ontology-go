package query

func (graph Graph) rejectedResponse(request Request, err error) (Response, error) {
	response := Response{
		Schema:   QueryEnvelopeSchema,
		Status:   ResponseError,
		Request:  request,
		Metadata: envelopeMetadata(graph.Metadata()),
		Error:    &EnvelopeError{Code: errorCode(err), Message: err.Error()},
	}
	if normalized, normalizeErr := request.Normalize(); normalizeErr == nil {
		response.Request = normalized
		response.RequestHash, _ = normalized.CanonicalDigest()
	}
	if sealErr := response.seal(); sealErr != nil {
		return Response{}, sealErr
	}
	return response, err
}
