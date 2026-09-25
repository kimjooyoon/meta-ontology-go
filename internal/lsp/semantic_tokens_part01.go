package lsp

import (
	"context"
)

const (
	semanticTokenEntity    = "entity"
	semanticTokenActivity  = "activity"
	semanticTokenReference = "reference"
	semanticTokenSymbol    = "symbol"
)

var canonicalSemanticTokenTypes = []string{
	semanticTokenEntity, semanticTokenActivity, semanticTokenReference, semanticTokenSymbol,
}

type semanticTokenSpan struct {
	rangeValue Range
	tokenType  int
	semanticID string
}

func (server *Server) semanticTokensRequest(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params SemanticTokensParams
	if decodeParams(request.Params, &params) != nil || params.TextDocument == nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid semantic tokens parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	document, exists := server.referenceDocument(params.TextDocument.URI)
	if !exists {
		return resultResponse(request.ID, nil), nil, nil
	}
	return resultResponse(request.ID, semanticTokensForDocument(document)), nil, nil
}

// semanticTokensForDocument projects only the canonical AST-facing symbols
// and references. Semantic IR, provenance, and filesystem state are absent
// from the wire response; semanticID only supports in-process
// cross-feature correspondence tests.
func semanticTokensForDocument(document document) SemanticTokens {
	spans := semanticTokenSpansForDocument(document)
	return SemanticTokens{Data: semanticTokenData(spans)}
}
func semanticTokenSpansForDocument(document document) []semanticTokenSpan {
	if document.result.semanticChecked && !document.result.semanticValid {
		return nil
	}
	spans := make([]semanticTokenSpan, 0, len(document.result.Symbols)+len(document.result.References))
	for _, symbol := range document.result.Symbols {
		if document.result.semanticChecked && symbol.ID == "" {
			continue
		}
		if !validSemanticTokenRange(symbol.SelectionRange) {
			continue
		}
		spans = append(spans, semanticTokenSpan{
			rangeValue: symbol.SelectionRange, tokenType: semanticTokenType(symbol.Kind), semanticID: symbol.ID,
		})
	}
	for _, reference := range document.result.References {
		if document.result.semanticChecked && reference.ID == "" {
			continue
		}
		if validSemanticTokenRange(reference.Range) {
			spans = append(spans, semanticTokenSpan{rangeValue: reference.Range, tokenType: 2, semanticID: reference.ID})
		}
	}
	spans = canonicalSemanticTokenSpans(spans)
	return spans
}
