package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"unicode"
	"unicode/utf8"
)

func (server *Server) hover(params TextDocumentPositionParams) (*Hover, bool) {
	server.mu.RLock()
	document, ok := server.documents[params.TextDocument.URI]
	if ok {
		copyValue := documentCopy(document)
		document = &copyValue
	}
	server.mu.RUnlock()
	if !ok {
		return nil, false
	}
	symbol, ok := symbolAtPosition(*document, params.Position)
	if !ok {
		return nil, false
	}
	rangeValue := symbol.SelectionRange
	if name, start, end, found := wordAt(document.text, params.Position); found {
		if candidate, candidateOK := symbolNamed(allSymbols(document.result), name); candidateOK && candidate.Name == symbol.Name {
			if value, err := byteRange(document.text, start, end); err == nil {
				rangeValue = value
			}
		}
	}
	return &Hover{Contents: MarkupContent{Kind: "plaintext", Value: symbol.Detail}, Range: &rangeValue}, true
}

func (server *Server) completion(uri string) *CompletionList {
	items := []CompletionItem{
		{Label: "activity", Kind: int(SymbolKeyword), Detail: "gooo keyword"},
		{Label: "entity", Kind: int(SymbolKeyword), Detail: "gooo keyword"},
		{Label: "namespace", Kind: int(SymbolKeyword), Detail: "gooo keyword"},
		{Label: "package", Kind: int(SymbolKeyword), Detail: "gooo keyword"},
	}
	server.mu.RLock()
	document, ok := server.documents[uri]
	if ok {
		copyValue := documentCopy(document)
		document = &copyValue
	}
	server.mu.RUnlock()
	if ok {
		for _, symbol := range allSymbols(document.result) {
			item := CompletionItem{Label: symbol.Name, Kind: int(symbol.Kind), Detail: symbol.Detail}
			if symbol.ID != "" {
				item.Documentation = "semantic ID: " + symbol.ID
			}
			items = append(items, item)
		}
	}
	sort.SliceStable(items, func(left, right int) bool { return items[left].Label < items[right].Label })
	return &CompletionList{Items: uniqueCompletionItems(items)}
}

func (server *Server) definition(params TextDocumentPositionParams) []Location {
	server.mu.RLock()
	document, ok := server.documents[params.TextDocument.URI]
	if ok {
		copyValue := documentCopy(document)
		document = &copyValue
	}
	server.mu.RUnlock()
	if !ok {
		return nil
	}
	if document.result.semanticChecked && !document.result.semanticValid {
		return nil
	}
	target, ok := definitionTargetForDocument(*document, params.Position)
	if !ok {
		return nil
	}
	symbol, found := resolveDefinitionSymbol(allSymbols(document.result), target)
	if !found {
		return nil
	}
	return []Location{{URI: params.TextDocument.URI, Range: symbol.SelectionRange}}
}

func symbolAtPosition(document document, position Position) (Symbol, bool) {
	for _, symbol := range allSymbols(document.result) {
		if positionInRange(position, symbol.SelectionRange) || (symbol.hasIdentity && positionInRange(position, symbol.identityRange)) {
			return symbol, true
		}
	}
	for _, reference := range document.result.References {
		if !positionInRange(position, reference.Range) || reference.ID == "" {
			continue
		}
		if symbol, ok := symbolByID(allSymbols(document.result), reference.ID); ok {
			return symbol, true
		}
	}
	name, _, _, ok := wordAt(document.text, position)
	if !ok {
		return Symbol{}, false
	}
	return symbolNamed(allSymbols(document.result), name)
}

func symbolByID(symbols []Symbol, id string) (Symbol, bool) {
	var match Symbol
	for _, symbol := range symbols {
		if symbol.ID != id {
			continue
		}
		if match.ID != "" {
			return Symbol{}, false
		}
		match = symbol
	}
	return match, match.ID != ""
}

func definitionTargetForDocument(document document, position Position) (string, bool) {
	for _, symbol := range allSymbols(document.result) {
		if symbol.hasIdentity && positionInRange(position, symbol.identityRange) {
			return symbol.ID, true
		}
	}
	for _, reference := range document.result.References {
		if reference.ID != "" && positionInRange(position, reference.Range) {
			return reference.ID, true
		}
	}
	return definitionTarget(document.text, position, allSymbols(document.result))
}

func definitionTarget(source string, position Position, symbols []Symbol) (string, bool) {
	for _, symbol := range symbols {
		if symbol.hasIdentity && positionInRange(position, symbol.identityRange) {
			return symbol.ID, true
		}
	}
	target, _, _, ok := wordAt(source, position)
	return target, ok
}

func positionInRange(position Position, value Range) bool {
	if position.Line < value.Start.Line || position.Line > value.End.Line {
		return false
	}
	if position.Line == value.Start.Line && position.Character < value.Start.Character {
		return false
	}
	if position.Line == value.End.Line && position.Character >= value.End.Character {
		return false
	}
	return true
}

func resolveDefinitionSymbol(symbols []Symbol, target string) (Symbol, bool) {
	var match Symbol
	for _, symbol := range symbols {
		if symbol.Name != target && symbol.ID != target {
			continue
		}
		if match.Name != "" || match.ID != "" {
			return Symbol{}, false
		}
		match = symbol
	}
	if match.Name == "" && match.ID == "" {
		return Symbol{}, false
	}
	return match, true
}

func (server *Server) refresh(ctx context.Context, uri string) error {
	server.mu.RLock()
	document, exists := server.documents[uri]
	if exists {
		version, source := document.version, document.text
		server.mu.RUnlock()
		result, err := server.parse(ctx, uri, source)
		if err != nil {
			return err
		}
		server.mu.Lock()
		defer server.mu.Unlock()
		current, stillOpen := server.documents[uri]
		if !stillOpen || current.version != version || current.text != source {
			return ErrStaleResult
		}
		current.result = result
		return nil
	}
	server.mu.RUnlock()
	return nil
}

func documentCopy(value *document) document {
	return document{version: value.version, text: value.text, result: cloneParseResult(value.result)}
}

func featureErrorResponse(id json.RawMessage, err error, ctx context.Context) (*responseEnvelope, [][]byte, error) {
	if errors.Is(err, ErrStaleResult) {
		return responseOrNil(id, contentModified, "content modified during request"), nil, nil
	}
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}
	return responseOrNil(id, internalError, err.Error()), nil, nil
}

func symbolNamed(symbols []Symbol, name string) (Symbol, bool) {
	for _, symbol := range symbols {
		if symbol.Name == name {
			return symbol, true
		}
	}
	return Symbol{}, false
}

func uniqueCompletionItems(items []CompletionItem) []CompletionItem {
	result := make([]CompletionItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, exists := seen[item.Label]; exists {
			continue
		}
		seen[item.Label] = struct{}{}
		result = append(result, item)
	}
	return result
}

func wordAt(source string, position Position) (string, int, int, bool) {
	offset, err := PositionToOffset(source, position)
	if err != nil {
		return "", 0, 0, false
	}
	start, end := identifierBounds(source, offset)
	if start == end {
		return "", 0, 0, false
	}
	return source[start:end], start, end, true
}

func identifierBounds(source string, offset int) (int, int) {
	if offset > 0 && (offset == len(source) || !identifierAt(source, offset)) {
		_, size := utf8.DecodeLastRuneInString(source[:offset])
		if size > 0 && identifierAt(source, offset-size) {
			offset -= size
		}
	}
	start := offset
	for start > 0 {
		value, size := utf8.DecodeLastRuneInString(source[:start])
		if !isIdentifier(value) {
			break
		}
		start -= size
	}
	end := offset
	for end < len(source) {
		value, size := utf8.DecodeRuneInString(source[end:])
		if !isIdentifier(value) {
			break
		}
		end += size
	}
	return start, end
}

func identifierAt(source string, offset int) bool {
	if offset < 0 || offset >= len(source) {
		return false
	}
	value, _ := utf8.DecodeRuneInString(source[offset:])
	return isIdentifier(value)
}

func isIdentifier(value rune) bool {
	return value == '_' || unicode.IsLetter(value) || unicode.IsDigit(value)
}

func byteRange(source string, start, end int) (Range, error) {
	startPosition, err := OffsetToPosition(source, start)
	if err != nil {
		return Range{}, err
	}
	endPosition, err := OffsetToPosition(source, end)
	if err != nil {
		return Range{}, err
	}
	return Range{Start: startPosition, End: endPosition}, nil
}
