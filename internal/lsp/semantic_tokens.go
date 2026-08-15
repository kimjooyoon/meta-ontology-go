package lsp

import (
	"context"
	"sort"
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

func canonicalSemanticTokenSpans(spans []semanticTokenSpan) []semanticTokenSpan {
	result := append([]semanticTokenSpan(nil), spans...)
	sort.Slice(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.rangeValue.Start != second.rangeValue.Start {
			return positionLess(first.rangeValue.Start, second.rangeValue.Start)
		}
		if first.rangeValue.End != second.rangeValue.End {
			return positionLess(first.rangeValue.End, second.rangeValue.End)
		}
		return first.tokenType < second.tokenType
	})
	unique := result[:0]
	for _, span := range result {
		if len(unique) > 0 && unique[len(unique)-1] == span {
			continue
		}
		unique = append(unique, span)
	}
	return unique
}

func semanticTokenData(spans []semanticTokenSpan) []uint32 {
	data := make([]uint32, 0, len(spans)*5)
	previousLine, previousCharacter := 0, 0
	for _, span := range spans {
		lineDelta := span.rangeValue.Start.Line - previousLine
		characterDelta := span.rangeValue.Start.Character
		if lineDelta == 0 {
			characterDelta -= previousCharacter
		}
		if lineDelta < 0 || characterDelta < 0 {
			continue
		}
		length := span.rangeValue.End.Character - span.rangeValue.Start.Character
		if length <= 0 {
			continue
		}
		data = append(data, uint32(lineDelta), uint32(characterDelta), uint32(length), uint32(span.tokenType), 0)
		previousLine, previousCharacter = span.rangeValue.Start.Line, span.rangeValue.Start.Character
	}
	return data
}

func validSemanticTokenRange(value Range) bool {
	return value.Start.Line >= 0 && value.Start.Character >= 0 &&
		value.End.Line == value.Start.Line && value.End.Character > value.Start.Character
}

func semanticTokenType(kind SymbolKind) int {
	switch kind {
	case SymbolClass:
		return 0
	case SymbolFunction:
		return 1
	default:
		return 3
	}
}
