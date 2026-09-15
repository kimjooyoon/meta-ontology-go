package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func analyzerKind(kind semantic.Kind) (SymbolKind, bool) {
	switch kind {
	case semantic.Entity:
		return KindEntity, true
	case semantic.Activity:
		return KindActivity, true
	default:
		return "", false
	}
}
